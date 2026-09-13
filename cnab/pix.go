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

// pixKeyColumnValue returns what Segmento B's PIX key column carries for
// key.
//
// A CPF/CNPJ key is already fully expressed by the beneficiary registration
// columns (kind + document) that sit beside it, so the key column itself
// stays empty for it: that column holds a phone, an e-mail or a random key
// and nothing else. Sicredi names the column
// "chavePixTelefoneEmailChaveAleatoria" and rejects a file that repeats the
// document there ("Caso PIX de CPF/CNPJ: Deve deixar em branco"). The key
// value is still required on the Pix value itself — validate rejects an
// empty one — because it is what says who is being paid; this is only about
// which column the file writes it to.
func pixKeyColumnValue(key PixKey) string {
	switch key.pixKeyType() {
	case PixKeyTypeCPF, PixKeyTypeCNPJ:
		return ""
	default:
		return key.pixKeyValue()
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
		// A PIX addressed by key credits no account of its own, so the
		// beneficiary bank/branch/account columns of Segmento A carry
		// zeros. The account's check digit is spelled out here because it
		// is the only one of the four that is alphanumeric in the FEBRABAN
		// picture (a check digit may be a letter), and an unset
		// alphanumeric renders as a blank — which the banks that validate
		// the column reject, Sicredi among them: "alinhada a direita com
		// zeros esquerda, sem espaços e somente números". The other three
		// are numeric and already zero-fill on their own.
		//
		// Only this payment kind needs it: TED, TEV, CreditAccount and
		// PixBankData all address a real account and validate through
		// Account.validate, which requires a CheckDigit.
		layout.KeyBeneficiaryCheckDigit: "0",
		layout.KeyPayeeName:             p.Payee.Name,
		layout.KeyYourNumber:            p.YourNumber,
		layout.KeyAmount:                int64(p.Amount),
		layout.KeyPaymentDate:           formatDate(p.Date),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(p.Payee.Registration),
		layout.KeyPayeeDocument:     p.Payee.Registration.Digits(),
		layout.KeyPixKeyType:        pixKeyFebrabanCode(p.Key.pixKeyType()),
		layout.KeyPixKeyValue:       pixKeyColumnValue(p.Key),
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentBPix, Values: b},
	}, nil
}

// PixBankAccountKind identifies the kind of account a PixBankData payment
// credits, per the FEBRABAN convention some banks (e.g. Sicredi) require
// alongside the beneficiary's bank data.
type PixBankAccountKind string

const (
	// PixBankAccountChecking is a checking account ("conta corrente").
	PixBankAccountChecking PixBankAccountKind = "01"
	// PixBankAccountPayment is a payment account ("conta de pagamento").
	PixBankAccountPayment PixBankAccountKind = "02"
	// PixBankAccountSavings is a savings account ("conta poupança").
	PixBankAccountSavings PixBankAccountKind = "03"
)

// pixKeyTypeCodeBankData is the FEBRABAN Segmento B code identifying a PIX
// payment addressed by bank account data rather than by key ("chave 5" in
// the Sicredi manual). It is not a PixKeyType/PixKey value: bank-data
// addressing is a distinct payment kind (PixBankData), not one more key
// variant of Pix, so it does not go through pixKeyFebrabanCode.
const pixKeyTypeCodeBankData = "05"

// PixBankData is a PIX transfer addressed by the beneficiary's bank
// account data instead of a PIX key (FEBRABAN Segmentos A e B-Pix).
type PixBankData struct {
	// Payee is the beneficiary receiving the transfer.
	Payee Payee
	// BankCode is the beneficiary's bank COMPE code, written to Segmento
	// A.
	BankCode string
	// ISPB is the beneficiary's bank ISPB code (8 digits). Some banks
	// (e.g. Sicredi) require it alongside BankCode, packed together with
	// the beneficiary's document and AccountKind into whichever field the
	// active Layout binds KeySupplementaryInfo to on Segmento A or on
	// Segmento B-Pix; left blank it renders as 8 zeros there, which a
	// Layout that does not need it simply ignores.
	ISPB string
	// AccountKind identifies the kind of account Account refers to.
	// Required by the same banks that require ISPB; left as the zero
	// value it renders as "00" wherever a Layout expects one of the
	// PixBankAccount* codes.
	AccountKind PixBankAccountKind
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
		layout.KeySupplementaryInfo:     pixBankDataSupplementaryInfo(p),
	}
	b := layout.Values{
		layout.KeyPayeeDocumentKind: documentKind(p.Payee.Registration),
		layout.KeyPayeeDocument:     p.Payee.Registration.Digits(),
		layout.KeyPixKeyType:        pixKeyTypeCodeBankData,
		// Offered on both segments because banks disagree about where the
		// blob belongs: some want it in Segmento A's supplementary info,
		// others in a dedicated range of Segmento B. The Layout binds
		// whichever its manual specifies; the other simply renders blank.
		layout.KeySupplementaryInfo: pixBankDataSupplementaryInfo(p),
	}
	return []DetailSegment{
		{Key: layout.SegmentA, Values: a},
		{Key: layout.SegmentBPix, Values: b},
	}, nil
}

// pixBankDataSupplementaryInfo packs the beneficiary's document, bank
// ISPB and account kind into the fixed-width blob some banks (e.g.
// Sicredi, per its manual's G031 note for PIX "chave 5") expect in
// Segmento A's supplementary info field for a bank-data-addressed PIX
// payment: a 14 digit CPF/CNPJ (zero-padded on the left), an 8 digit
// ISPB and a 2 digit account kind code, 24 characters total. A Layout
// that does not bind KeySupplementaryInfo, or a bank that does not need
// this convention, simply never reads it.
func pixBankDataSupplementaryInfo(p PixBankData) string {
	document := zeroPadLeft(p.Payee.Registration.Digits(), 14)
	ispb := zeroPadLeft(onlyDigits(p.ISPB), 8)
	kind := string(p.AccountKind)
	if kind == "" {
		kind = "00"
	}
	return document + ispb + kind
}
