package cnab

import (
	"strings"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// PixKeyType identifies which kind of PIX key a PixKey value carries.
type PixKeyType string

const (
	PixKeyTypePhone  PixKeyType = "phone"
	PixKeyTypeEmail  PixKeyType = "email"
	PixKeyTypeCPF    PixKeyType = "cpf"
	PixKeyTypeCNPJ   PixKeyType = "cnpj"
	PixKeyTypeRandom PixKeyType = "random"
)

// PixKey is a PIX addressing key. It is implemented by PhoneKey,
// EmailKey, CPFKey, CNPJKey and RandomKey; the interface is sealed (its
// methods are unexported).
type PixKey interface {
	pixKeyType() PixKeyType
	pixKeyValue() string
}

// PhoneKey is a PIX key expressed as a phone number, e.g. "+5551998765432".
type PhoneKey string

func (k PhoneKey) pixKeyType() PixKeyType { return PixKeyTypePhone }
func (k PhoneKey) pixKeyValue() string    { return string(k) }

// EmailKey is a PIX key expressed as an e-mail address.
type EmailKey string

func (k EmailKey) pixKeyType() PixKeyType { return PixKeyTypeEmail }
func (k EmailKey) pixKeyValue() string    { return string(k) }

// CPFKey is a PIX key expressed as an individual's CPF.
type CPFKey string

func (k CPFKey) pixKeyType() PixKeyType { return PixKeyTypeCPF }
func (k CPFKey) pixKeyValue() string    { return string(k) }

// CNPJKey is a PIX key expressed as a company's CNPJ.
type CNPJKey string

func (k CNPJKey) pixKeyType() PixKeyType { return PixKeyTypeCNPJ }
func (k CNPJKey) pixKeyValue() string    { return string(k) }

// RandomKey is a PIX key expressed as a random ("chave aleatória") UUID
// issued by the PIX system.
type RandomKey string

func (k RandomKey) pixKeyType() PixKeyType { return PixKeyTypeRandom }
func (k RandomKey) pixKeyValue() string    { return string(k) }

// pixKeyFebrabanCode maps a PixKeyType to the 2 digit code the FEBRABAN
// standard uses to identify it on Segmento B.
func pixKeyFebrabanCode(t PixKeyType) string {
	switch t {
	case PixKeyTypePhone:
		return "01"
	case PixKeyTypeEmail:
		return "02"
	case PixKeyTypeCPF, PixKeyTypeCNPJ:
		return "03"
	case PixKeyTypeRandom:
		return "04"
	default:
		return "00"
	}
}

// Pix is a PIX transfer addressed by key (FEBRABAN Segmentos A e B).
type Pix struct {
	// Key is the beneficiary's PIX key.
	Key PixKey
	// Payee is the beneficiary receiving the transfer.
	Payee Payee
	// Amount is the payment amount.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (p Pix) validate(now time.Time, opts validateOptions) error {
	if p.Key == nil || strings.TrimSpace(p.Key.pixKeyValue()) == "" {
		return &ValidationError{Context: "Pix", Reason: "Key is required"}
	}
	if err := p.Payee.validate(); err != nil {
		return err
	}
	if p.Amount <= 0 {
		return &ValidationError{Context: "Pix", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(p.Date, now, opts)
}

func (p Pix) toSegments(l Layout) ([]DetailSegment, error) {
	a := layout.Values{
		layout.KeyMovementType: "0",
		layout.KeyClearingCode: "009", // PIX (SPI)
		layout.KeyPayeeName:    p.Payee.Name,
		layout.KeyYourNumber:   p.YourNumber,
		layout.KeyAmount:       int64(p.Amount),
		layout.KeyPaymentDate:  formatDate(p.Date),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(p.Payee.Registration),
		layout.KeyPayeeDocument:     p.Payee.Registration.Digits(),
		layout.KeyPixKeyType:        pixKeyFebrabanCode(p.Key.pixKeyType()),
		layout.KeyPixKeyValue:       p.Key.pixKeyValue(),
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentBPix, Values: b},
	}, nil
}

// PixBankData is a PIX transfer addressed by the beneficiary's bank
// account data instead of a PIX key (FEBRABAN Segmentos A e B).
type PixBankData struct {
	// Payee is the beneficiary receiving the transfer.
	Payee Payee
	// BankCode is the beneficiary's bank code (COMPE or ISPB).
	BankCode string
	// Account is the beneficiary's account at BankCode.
	Account Account
	// Amount is the payment amount.
	Amount Cents
	// Date is the date the payment should be settled.
	Date time.Time
	// YourNumber is the payer's own reference for this payment. Optional.
	YourNumber string
}

func (p PixBankData) validate(now time.Time, opts validateOptions) error {
	if err := p.Payee.validate(); err != nil {
		return err
	}
	if err := p.Account.validate(); err != nil {
		return err
	}
	if strings.TrimSpace(p.BankCode) == "" {
		return &ValidationError{Context: "PixBankData", Reason: "BankCode is required"}
	}
	if p.Amount <= 0 {
		return &ValidationError{Context: "PixBankData", Reason: "Amount must be greater than zero"}
	}
	return validatePaymentDate(p.Date, now, opts)
}

func (p PixBankData) toSegments(l Layout) ([]DetailSegment, error) {
	a := layout.Values{
		layout.KeyMovementType:          "0",
		layout.KeyClearingCode:          "009", // PIX (SPI)
		layout.KeyBeneficiaryBankCode:   p.BankCode,
		layout.KeyBeneficiaryBranch:     p.Account.Branch,
		layout.KeyBeneficiaryAccount:    p.Account.Number,
		layout.KeyBeneficiaryCheckDigit: p.Account.CheckDigit,
		layout.KeyPayeeName:             p.Payee.Name,
		layout.KeyYourNumber:            p.YourNumber,
		layout.KeyAmount:                int64(p.Amount),
		layout.KeyPaymentDate:           formatDate(p.Date),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(p.Payee.Registration),
		layout.KeyPayeeDocument:     p.Payee.Registration.Digits(),
		layout.KeyPixKeyType:        "bank_data",
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentB, Values: b},
	}, nil
}
