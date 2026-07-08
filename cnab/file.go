package cnab

import (
	"fmt"
	"strings"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
	"github.com/raykavin/gocnab/internal/engine"
)

// File is a CNAB 240 remittance file being assembled. Create one with
// NewRemittance, add batches with NewBatch, add payments to each batch
// with Batch.AddPayment, then call Generate to produce the file content.
type File struct {
	config Config
	engine *engine.Engine
	layout Layout

	batches []*Batch
}

// NewRemittance validates cfg and starts a new remittance file.
//
// It returns a *ValidationError when Company, Account or NSA are
// missing/invalid, when Layout is empty, or when Layout does not name a
// layout previously registered with RegisterLayout (the bundled
// "febraban240" reference layout is always available).
func NewRemittance(cfg Config) (*File, error) {
	if err := cfg.Company.validate(); err != nil {
		return nil, err
	}
	if err := cfg.Account.validate(); err != nil {
		return nil, err
	}
	if cfg.NSA <= 0 {
		return nil, &ValidationError{Context: "Config", Reason: "NSA must be greater than zero"}
	}
	if strings.TrimSpace(cfg.Layout) == "" {
		return nil, &ValidationError{Context: "Config", Reason: "Layout is required"}
	}

	l, ok := layout.Lookup(cfg.Layout)
	if !ok {
		return nil, &ValidationError{
			Context: "Config",
			Reason:  fmt.Sprintf("layout %q is not registered (available: %v)", cfg.Layout, layout.Names()),
		}
	}

	eng, err := engine.New(l)
	if err != nil {
		return nil, err
	}

	return &File{config: cfg, engine: eng, layout: l}, nil
}

// NewBatch starts a new batch ("lote") for product settled via service,
// and appends it to the file.
//
// It returns a *ValidationError when product or service is the zero
// value, or a *LimitExceededError once the file already holds
// internal/engine.MaxBatchesPerFile batches.
func (f *File) NewBatch(product BatchProduct, service BatchService) (*Batch, error) {
	if len(f.batches) >= engine.MaxBatchesPerFile {
		return nil, &LimitExceededError{
			Limit:     engine.LimitBatchesPerFile,
			Max:       engine.MaxBatchesPerFile,
			Attempted: len(f.batches) + 1,
		}
	}
	if product.code == "" {
		return nil, &ValidationError{Context: "NewBatch", Reason: "product must not be the zero value"}
	}
	if service.code == "" {
		return nil, &ValidationError{Context: "NewBatch", Reason: "service must not be the zero value"}
	}

	b := &Batch{file: f, product: product, service: service}
	f.batches = append(f.batches, b)
	return b, nil
}

// FileName suggests a file name for the remittance, combining the active
// layout name, the configured NSA and the current date. Banks generally
// only require a specific extension (commonly ".REM"); rename the result
// if your bank expects a different convention.
func (f *File) FileName() (string, error) {
	if len(f.batches) == 0 {
		return "", &ValidationError{Context: "FileName", Reason: "file must have at least one batch"}
	}
	return fmt.Sprintf("%s_%04d_%s.REM", strings.ToUpper(f.engine.LayoutName()), f.config.NSA, time.Now().Format("20060102")), nil
}

// Generate renders the complete CNAB 240 file content: the file header,
// every batch (header, payments and trailer) and the file trailer, each
// line terminated with CRLF.
//
// Generate re-validates that no payment date has become past due since
// it was added (time advances between AddPayment and Generate) and, once
// the engine has rendered the file, cross-checks the total record count
// against what the batches themselves report, returning a
// *TrailerMismatchError if they ever disagree; this should not happen
// through normal use and exists as a defensive guard against a defective
// Layout. Every other structural rule (FEBRABAN limits, sequencing,
// per-batch trailer totals) is enforced by construction inside the
// engine.
func (f *File) Generate() ([]byte, error) {
	if len(f.batches) == 0 {
		return nil, &ValidationError{Context: "Generate", Reason: "file must have at least one batch"}
	}

	now := time.Now()
	if err := f.checkPastDates(now); err != nil {
		return nil, err
	}

	in := engine.FileInput{Header: f.headerValues(), Trailer: layout.Values{}}
	expectedRecords := 1 // file header
	for _, b := range f.batches {
		batchIn := engine.BatchInput{Header: f.batchHeaderValues(b), Trailer: layout.Values{}}
		segCount := 0
		for _, movement := range b.movements {
			lines := make([]engine.DetailLine, len(movement))
			for i, seg := range movement {
				lines[i] = engine.DetailLine{Key: seg.Key, Values: seg.Values}
			}
			batchIn.Movements = append(batchIn.Movements, lines)
			segCount += len(movement)
		}
		in.Batches = append(in.Batches, batchIn)
		expectedRecords += 1 + segCount + 1 // header + details + trailer
	}
	expectedRecords++ // file trailer

	out, err := f.engine.Build(in)
	if err != nil {
		return nil, err
	}

	actualRecords := strings.Count(string(out), "\r\n")
	if actualRecords != expectedRecords {
		return nil, &TrailerMismatchError{
			Field:    "record_count",
			Expected: fmt.Sprintf("%d", expectedRecords),
			Got:      fmt.Sprintf("%d", actualRecords),
		}
	}

	return out, nil
}

func (f *File) headerValues() layout.Values {
	now := time.Now()
	return layout.Values{
		layout.KeyFileSequenceNumber:      int64(f.config.NSA),
		layout.KeyFileGenerationDate:      now.Format("02012006"),
		layout.KeyFileGenerationTime:      now.Format("150405"),
		layout.KeyCompanyRegistrationKind: documentKind(f.config.Company.Registration),
		layout.KeyCompanyRegistration:     f.config.Company.Registration.Digits(),
		layout.KeyCompanyName:             f.config.Company.Name,
		layout.KeyAgreement:               f.config.Company.Agreement,
		layout.KeyBranch:                  f.config.Account.Branch,
		layout.KeyAccountNumber:           f.config.Account.Number,
		layout.KeyAccountCheckDigit:       f.config.Account.CheckDigit,
	}
}

func (f *File) batchHeaderValues(b *Batch) layout.Values {
	v := f.headerValues()
	v[layout.KeyBatchProductCode] = b.product.code
	v[layout.KeyBatchServiceCode] = b.service.code
	return v
}

// checkPastDates re-validates every pending payment's date against now.
// Cancellations (movement type "9") are skipped since they legitimately
// carry the original, now past-due, payment date.
func (f *File) checkPastDates(now time.Time) error {
	today := truncateToDate(now)
	for bi, b := range f.batches {
		for _, movement := range b.movements {
			for _, seg := range movement {
				if seg.Values[layout.KeyMovementType] == "9" {
					continue
				}
				raw, ok := seg.Values[layout.KeyPaymentDate]
				if !ok {
					continue
				}
				s, ok := raw.(string)
				if !ok || s == "" {
					continue
				}
				d, err := time.ParseInLocation("02012006", s, now.Location())
				if err != nil {
					continue
				}
				if truncateToDate(d).Before(today) {
					return &ValidationError{
						Context: fmt.Sprintf("batch %d", bi+1),
						Reason:  fmt.Sprintf("payment date %s is in the past", d.Format("2006-01-02")),
					}
				}
			}
		}
	}
	return nil
}
