package cnab

import (
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// DARF is a normal DARF tax payment, with principal, fine and interest
// tracked separately (FEBRABAN Segmento N).
type DARF struct {
	// TaxCode is the revenue code ("código da receita").
	TaxCode string
	// Taxpayer identifies who the tax is being paid for.
	Taxpayer Payee
	// ReferenceNumber is the DARF reference number.
	ReferenceNumber string
	// Period is the assessment period ("período de apuração").
	Period time.Time
	// DueDate is the DARF due date.
	DueDate time.Time
	// Principal is the principal tax amount.
	Principal Cents
	// Fine is the fine ("multa") portion. Optional.
	Fine Cents
	// Interest is the interest ("juros") portion. Optional.
	Interest Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (d DARF) total() Cents { return d.Principal + d.Fine + d.Interest }

// taxpayerIdType returns the DARF/GPS segment's own taxpayer
// identification type code: this is a different domain than the general
// documentKind convention used elsewhere ("1" for CPF, "2" for CNPJ), so
// it is not reused across the two.
func taxpayerIdType(d Document) string {
	if d != nil && d.Kind() == "CNPJ" {
		return "01"
	}
	return "02"
}

func (d DARF) validate(now time.Time, opts validateOptions) error {
	if d.TaxCode == "" {
		return &ValidationError{Context: "DARF", Reason: "TaxCode is required"}
	}
	if err := d.Taxpayer.validate(); err != nil {
		return err
	}
	if d.total() <= 0 {
		return &ValidationError{Context: "DARF", Reason: "Principal + Fine + Interest must be greater than zero"}
	}
	return validatePaymentDate(d.Date, now, opts)
}

func (d DARF) toSegments(l Layout) ([]DetailSegment, error) {
	n := layout.Values{
		layout.KeyMovementType:     "0",
		layout.KeyTaxCode:          d.TaxCode,
		layout.KeyTaxpayerIdType:   taxpayerIdType(d.Taxpayer.Registration),
		layout.KeyTaxpayerDocument: d.Taxpayer.Registration.Digits(),
		layout.KeyTaxpayerName:     d.Taxpayer.Name,
		layout.KeyReferenceNumber:  d.ReferenceNumber,
		layout.KeyPeriod:           formatDate(d.Period),
		layout.KeyDueDate:          formatDate(d.DueDate),
		layout.KeyPrincipalAmount:  int64(d.Principal),
		layout.KeyFineAmount:       int64(d.Fine),
		layout.KeyInterestAmount:   int64(d.Interest),
		layout.KeyAmount:           int64(d.total()),
		layout.KeyPaymentDate:      formatDate(d.Date),
		layout.KeyYourNumber:       d.YourNumber,
	}
	return []DetailSegment{{Key: layout.SegmentN, Values: n}}, nil
}

// DARFSimple is a simplified DARF tax payment, carried as a single total
// amount instead of separate principal/fine/interest (FEBRABAN Segmento
// N, "DARF Simples").
type DARFSimple struct {
	// TaxCode is the revenue code ("código da receita").
	TaxCode string
	// Taxpayer identifies who the tax is being paid for.
	Taxpayer Payee
	// DueDate is the DARF due date.
	DueDate time.Time
	// Amount is the total amount to pay.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (d DARFSimple) validate(now time.Time, opts validateOptions) error {
	if d.TaxCode == "" {
		return &ValidationError{Context: "DARFSimple", Reason: "TaxCode is required"}
	}
	if err := d.Taxpayer.validate(); err != nil {
		return err
	}
	if d.Amount <= 0 {
		return &ValidationError{Context: "DARFSimple", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(d.Date, now, opts)
}

func (d DARFSimple) toSegments(l Layout) ([]DetailSegment, error) {
	n := layout.Values{
		layout.KeyMovementType:     "0",
		layout.KeyTaxCode:          d.TaxCode,
		layout.KeyTaxpayerIdType:   taxpayerIdType(d.Taxpayer.Registration),
		layout.KeyTaxpayerDocument: d.Taxpayer.Registration.Digits(),
		layout.KeyTaxpayerName:     d.Taxpayer.Name,
		layout.KeyDueDate:          formatDate(d.DueDate),
		layout.KeyPrincipalAmount:  int64(d.Amount),
		layout.KeyAmount:           int64(d.Amount),
		layout.KeyPaymentDate:      formatDate(d.Date),
		layout.KeyYourNumber:       d.YourNumber,
	}
	return []DetailSegment{{Key: layout.SegmentNSimple, Values: n}}, nil
}

// GPS is a Guia da Previdência Social payment (FEBRABAN Segmento N,
// "GPS").
type GPS struct {
	// Taxpayer identifies who the contribution is being paid for.
	Taxpayer Payee
	// Period is the assessment period ("competência").
	Period time.Time
	// DueDate is the GPS due date.
	DueDate time.Time
	// Amount is the total amount to pay.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (g GPS) validate(now time.Time, opts validateOptions) error {
	if err := g.Taxpayer.validate(); err != nil {
		return err
	}
	if g.Amount <= 0 {
		return &ValidationError{Context: "GPS", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(g.Date, now, opts)
}

func (g GPS) toSegments(l Layout) ([]DetailSegment, error) {
	n := layout.Values{
		layout.KeyMovementType:     "0",
		layout.KeyTaxpayerIdType:   taxpayerIdType(g.Taxpayer.Registration),
		layout.KeyTaxpayerDocument: g.Taxpayer.Registration.Digits(),
		layout.KeyTaxpayerName:     g.Taxpayer.Name,
		layout.KeyPeriod:           formatMonthYear(g.Period),
		layout.KeyDueDate:          formatDate(g.DueDate),
		layout.KeyPrincipalAmount:  int64(g.Amount),
		layout.KeyAmount:           int64(g.Amount),
		layout.KeyPaymentDate:      formatDate(g.Date),
		layout.KeyYourNumber:       g.YourNumber,
	}
	return []DetailSegment{{Key: layout.SegmentNSocial, Values: n}}, nil
}
