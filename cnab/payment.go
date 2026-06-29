package cnab

import (
	"fmt"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// Payment is implemented by every payment kind this SDK models:
// CreditAccount, TED, Pix, PixBankData, BoletoPayment, BarcodeTax, DARF,
// DARFSimple, GPS and CancelPayment. The interface is sealed (its methods
// are unexported): a caller passes a Payment value to Batch.AddPayment,
// it never implements one from outside this package.
type Payment interface {
	toSegments(l Layout) ([]DetailSegment, error)
	validate(now time.Time, opts validateOptions) error
}

// validateOptions tunes a business rule that must behave differently
// when a payment is being validated as a cancellation of a previously
// sent payment: a cancellation legitimately references an original
// payment date that is now in the past.
type validateOptions struct {
	AllowPastDate bool
}

// DetailSegment is one 240 character detail line a Payment contributes:
// a payment kind that needs two segments (for example SegmentA and
// SegmentB) returns two DetailSegment values from toSegments, in the
// order they must appear in the file.
type DetailSegment struct {
	Key    layout.RecordKey
	Values layout.Values
}

func validatePaymentDate(date time.Time, now time.Time, opts validateOptions) error {
	if date.IsZero() {
		return &ValidationError{Context: "Payment", Reason: "Date is required"}
	}
	if !opts.AllowPastDate && truncateToDate(date).Before(truncateToDate(now)) {
		return &ValidationError{Context: "Payment", Reason: fmt.Sprintf("Date %s is in the past", date.Format("2006-01-02"))}
	}
	return nil
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// formatDate renders t as an 8 digit DDMMAAAA string, the date format
// used throughout the FEBRABAN CNAB 240 standard.
func formatDate(t time.Time) string {
	return t.Format("02012006")
}

// formatMonthYear renders t as a 6 digit MMAAAA string, used by the GPS
// "competência" field.
func formatMonthYear(t time.Time) string {
	return t.Format("012006")
}

func cloneSegmentValues(v layout.Values) layout.Values {
	out := make(layout.Values, len(v)+2)
	for k, val := range v {
		out[k] = val
	}
	return out
}
