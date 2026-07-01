package cnab

import (
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// BoletoPayment is a boleto payment (FEBRABAN Segmentos J e J-52).
type BoletoPayment struct {
	// Barcode is the boleto's 44 digit barcode ("linha digitável").
	Barcode string
	// Assignor is the boleto's assignor ("cedente"): who issued it.
	Assignor Payee
	// Payer is the boleto's payer ("sacado"): who is being billed. This
	// is normally the same company sending the remittance file.
	Payer Payee
	// DueDate is the boleto's due date.
	DueDate time.Time
	// DocumentAmount is the boleto's face value. If zero, Amount is used.
	DocumentAmount Cents
	// Discount is a discount applied to the payment. Optional.
	Discount Cents
	// Addition is interest/fine added to the payment. Optional.
	Addition Cents
	// Amount is the amount actually paid.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (b BoletoPayment) validate(now time.Time, opts validateOptions) error {
	if len(onlyDigits(b.Barcode)) != 44 {
		return &ValidationError{Context: "BoletoPayment", Reason: "Barcode must have 44 digits"}
	}
	if err := b.Assignor.validate(); err != nil {
		return err
	}
	if err := b.Payer.validate(); err != nil {
		return err
	}
	if b.Amount <= 0 {
		return &ValidationError{Context: "BoletoPayment", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(b.Date, now, opts)
}

func (b BoletoPayment) toSegments(l Layout) ([]DetailSegment, error) {
	documentAmount := b.DocumentAmount
	if documentAmount == 0 {
		documentAmount = b.Amount
	}

	j := layout.Values{
		layout.KeyMovementType:   "0",
		layout.KeyBarcode:        onlyDigits(b.Barcode),
		layout.KeyAssignorName:   b.Assignor.Name,
		layout.KeyDueDate:        formatDate(b.DueDate),
		layout.KeyDocumentAmount: int64(documentAmount),
		layout.KeyDiscountAmount: int64(b.Discount),
		layout.KeyAdditionAmount: int64(b.Addition),
		layout.KeyAmount:         int64(b.Amount),
		layout.KeyPaymentDate:    formatDate(b.Date),
		layout.KeyYourNumber:     b.YourNumber,
	}
	j52 := layout.Values{
		layout.KeyPayerDocumentKind:    documentKind(b.Payer.Registration),
		layout.KeyPayerDocument:        b.Payer.Registration.Digits(),
		layout.KeyPayerName:            b.Payer.Name,
		layout.KeyAssignorDocumentKind: documentKind(b.Assignor.Registration),
		layout.KeyAssignorDocument:     b.Assignor.Registration.Digits(),
	}
	return []DetailSegment{
		{Key: layout.SegmentJ, Values: j},
		{Key: layout.SegmentJ52, Values: j52},
	}, nil
}
