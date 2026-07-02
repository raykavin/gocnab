package cnab

import (
	"strings"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// PurposeCode is a TED purpose ("finalidade") code, as published by the
// Central Bank. Confirm the exact table with your bank before relying on
// a specific code in production; the constants below cover the most
// common cases.
type PurposeCode string

const (
	// PurposeSupplierPayment identifies a TED made to pay a supplier.
	PurposeSupplierPayment PurposeCode = "00005"
	// PurposePayroll identifies a TED made to pay payroll.
	PurposePayroll PurposeCode = "00006"
	// PurposeOwnTransfer identifies a TED between accounts of the same
	// holder.
	PurposeOwnTransfer PurposeCode = "00001"
	// PurposeOther identifies a TED that does not fit a more specific
	// purpose code.
	PurposeOther PurposeCode = "00009"
)

// TED is a TED (wire transfer) payment (FEBRABAN Segmentos A e B).
type TED struct {
	// Payee is the beneficiary receiving the transfer.
	Payee Payee
	// BankCode is the beneficiary's bank code (COMPE).
	BankCode string
	// Account is the beneficiary's account at BankCode.
	Account Account
	// Amount is the payment amount.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// Purpose is the TED purpose code.
	Purpose PurposeCode
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (t TED) validate(now time.Time, opts validateOptions) error {
	if err := t.Payee.validate(); err != nil {
		return err
	}
	if err := t.Account.validate(); err != nil {
		return err
	}
	if strings.TrimSpace(t.BankCode) == "" {
		return &ValidationError{Context: "TED", Reason: "BankCode is required"}
	}
	if t.Amount <= 0 {
		return &ValidationError{Context: "TED", Reason: "Amount must be greater than zero"}
	}
	if t.Purpose == "" {
		return &ValidationError{Context: "TED", Reason: "Purpose is required"}
	}
	return validatePaymentDate(t.Date, now, opts)
}

func (t TED) toSegments(l Layout) ([]DetailSegment, error) {
	a := layout.Values{
		layout.KeyMovementType:          "0",
		layout.KeyClearingCode:          "018", // TED (STR/CIP)
		layout.KeyBeneficiaryBankCode:   t.BankCode,
		layout.KeyBeneficiaryBranch:     t.Account.Branch,
		layout.KeyBeneficiaryAccount:    t.Account.Number,
		layout.KeyBeneficiaryCheckDigit: t.Account.CheckDigit,
		layout.KeyPayeeName:             t.Payee.Name,
		layout.KeyYourNumber:            t.YourNumber,
		layout.KeyAmount:                int64(t.Amount),
		layout.KeyPaymentDate:           formatDate(t.Date),
		layout.KeyPurposeCode:           string(t.Purpose),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(t.Payee.Registration),
		layout.KeyPayeeDocument:     t.Payee.Registration.Digits(),
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentB, Values: b},
	}, nil
}
