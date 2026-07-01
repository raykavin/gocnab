package cnab

import (
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// CreditAccount is a same-bank credit-in-account payment (FEBRABAN
// Segmentos A e B, "Crédito em Conta Corrente").
type CreditAccount struct {
	// Payee is the beneficiary receiving the credit.
	Payee Payee
	// Account is the beneficiary's account at the same bank the
	// remittance file is sent to.
	Account Account
	// Amount is the payment amount.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment ("seu
	// número"), such as an invoice number. Optional.
	YourNumber string
}

func (c CreditAccount) validate(now time.Time, opts validateOptions) error {
	if err := c.Payee.validate(); err != nil {
		return err
	}
	if err := c.Account.validate(); err != nil {
		return err
	}
	if c.Amount <= 0 {
		return &ValidationError{Context: "CreditAccount", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(c.Date, now, opts)
}

func (c CreditAccount) toSegments(l Layout) ([]DetailSegment, error) {
	a := layout.Values{
		layout.KeyMovementType:          "0",
		layout.KeyClearingCode:          "000", // crédito em conta, mesmo banco
		layout.KeyBeneficiaryBranch:     c.Account.Branch,
		layout.KeyBeneficiaryAccount:    c.Account.Number,
		layout.KeyBeneficiaryCheckDigit: c.Account.CheckDigit,
		layout.KeyPayeeName:             c.Payee.Name,
		layout.KeyYourNumber:            c.YourNumber,
		layout.KeyAmount:                int64(c.Amount),
		layout.KeyPaymentDate:           formatDate(c.Date),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(c.Payee.Registration),
		layout.KeyPayeeDocument:     c.Payee.Registration.Digits(),
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentB, Values: b},
	}, nil
}
