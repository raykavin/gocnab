package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestCreditAccountValidate(t *testing.T) {
	now := time.Now()
	valid := CreditAccount{
		Payee:   validPayee(),
		Account: validAccount(),
		Amount:  1000,
		Date:    now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	cases := []struct {
		name    string
		payment CreditAccount
	}{
		{"missing payee", CreditAccount{Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, 1)}},
		{"missing account", CreditAccount{Payee: validPayee(), Amount: 1000, Date: now.AddDate(0, 0, 1)}},
		{"zero amount", CreditAccount{Payee: validPayee(), Account: validAccount(), Date: now.AddDate(0, 0, 1)}},
		{"missing date", CreditAccount{Payee: validPayee(), Account: validAccount(), Amount: 1000}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.payment.validate(now, validateOptions{}); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}

func TestCreditAccountToSegments(t *testing.T) {
	now := time.Now()
	p := CreditAccount{
		Payee:      validPayee(),
		Account:    validAccount(),
		Amount:     1000,
		Date:       now.AddDate(0, 0, 1),
		YourNumber: "NF-1",
	}

	segments, err := p.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	if segments[0].Key != layout.SegmentA {
		t.Fatalf("segments[0].Key = %q, want SegmentA", segments[0].Key)
	}
	if segments[0].Values[layout.KeyClearingCode] != "000" {
		t.Fatalf("KeyClearingCode = %v, want \"000\" (same-bank credit)", segments[0].Values[layout.KeyClearingCode])
	}
	if segments[1].Key != layout.SegmentB {
		t.Fatalf("segments[1].Key = %q, want SegmentB", segments[1].Key)
	}
	if segments[0].Values[layout.KeyAmount] != int64(1000) {
		t.Fatalf("KeyAmount = %v, want 1000", segments[0].Values[layout.KeyAmount])
	}
	if segments[1].Values[layout.KeyPayeeDocument] != validPayee().Registration.Digits() {
		t.Fatalf("KeyPayeeDocument = %v", segments[1].Values[layout.KeyPayeeDocument])
	}
}
