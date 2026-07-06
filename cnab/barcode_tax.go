package cnab

import (
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// BarcodeTax is a utility bill or tax payment identified by a barcode
// (FEBRABAN Segmento O), such as an electricity bill or a tax slip that
// carries a barcode.
type BarcodeTax struct {
	// Barcode is the bill's 44 digit barcode.
	Barcode string
	// DueDate is the bill's due date.
	DueDate time.Time
	// Amount is the amount to pay.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (b BarcodeTax) validate(now time.Time, opts validateOptions) error {
	if len(onlyDigits(b.Barcode)) != 44 {
		return &ValidationError{Context: "BarcodeTax", Reason: "Barcode must have 44 digits"}
	}
	if b.Amount <= 0 {
		return &ValidationError{Context: "BarcodeTax", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(b.Date, now, opts)
}

func (b BarcodeTax) toSegments(l Layout) ([]DetailSegment, error) {
	o := layout.Values{
		layout.KeyMovementType: "0",
		layout.KeyBarcode:      onlyDigits(b.Barcode),
		layout.KeyDueDate:      formatDate(b.DueDate),
		layout.KeyAmount:       int64(b.Amount),
		layout.KeyPaymentDate:  formatDate(b.Date),
		layout.KeyYourNumber:   b.YourNumber,
	}
	return []DetailSegment{{Key: layout.SegmentO, Values: o}}, nil
}
