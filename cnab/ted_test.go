package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestTEDValidate(t *testing.T) {
	now := time.Now()
	valid := TED{
		Payee:    validPayee(),
		BankCode: "001",
		Account:  validAccount(),
		Amount:   1000,
		Date:     now.AddDate(0, 0, 1),
		Purpose:  PurposeSupplierPayment,
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	cases := []struct {
		name    string
		payment TED
	}{
		{"missing bank code", TED{Payee: validPayee(), Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, 1), Purpose: PurposeOther}},
		{"missing purpose", TED{Payee: validPayee(), BankCode: "001", Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, 1)}},
		{"zero amount", TED{Payee: validPayee(), BankCode: "001", Account: validAccount(), Date: now.AddDate(0, 0, 1), Purpose: PurposeOther}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.payment.validate(now, validateOptions{}); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}

func TestTEDToSegments(t *testing.T) {
	now := time.Now()
	p := TED{
		Payee:    validPayee(),
		BankCode: "001",
		Account:  validAccount(),
		Amount:   2000,
		Date:     now.AddDate(0, 0, 1),
		Purpose:  PurposeSupplierPayment,
	}

	segments, err := p.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	if segments[0].Values[layout.KeyClearingCode] != "018" {
		t.Fatalf("KeyClearingCode = %v, want \"018\" (TED)", segments[0].Values[layout.KeyClearingCode])
	}
	if segments[0].Values[layout.KeyBeneficiaryBankCode] != "001" {
		t.Fatalf("KeyBeneficiaryBankCode = %v, want \"001\"", segments[0].Values[layout.KeyBeneficiaryBankCode])
	}
	if segments[0].Values[layout.KeyPurposeCode] != string(PurposeSupplierPayment) {
		t.Fatalf("KeyPurposeCode = %v", segments[0].Values[layout.KeyPurposeCode])
	}
}
