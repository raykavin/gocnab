package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestTaxpayerIdType(t *testing.T) {
	cnpj, _ := NewCNPJ("11222333000181")
	cpf, _ := NewCPF("11144477735")

	if got := taxpayerIdType(cnpj); got != "01" {
		t.Fatalf("taxpayerIdType(CNPJ) = %q, want \"01\"", got)
	}
	if got := taxpayerIdType(cpf); got != "02" {
		t.Fatalf("taxpayerIdType(CPF) = %q, want \"02\"", got)
	}
}

func TestDARFValidateAndSegments(t *testing.T) {
	now := time.Now()
	valid := DARF{
		TaxCode:         "0220",
		Taxpayer:        validPayee(),
		ReferenceNumber: "12345",
		Period:          now,
		DueDate:         now.AddDate(0, 0, 5),
		Principal:       1000,
		Fine:            100,
		Interest:        50,
		Date:            now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	missingCode := valid
	missingCode.TaxCode = ""
	if err := missingCode.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for missing TaxCode")
	}

	zeroTotal := valid
	zeroTotal.Principal, zeroTotal.Fine, zeroTotal.Interest = 0, 0, 0
	if err := zeroTotal.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for a zero total")
	}

	segments, err := valid.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 1 || segments[0].Key != layout.SegmentN {
		t.Fatalf("unexpected segments: %+v", segments)
	}
	if segments[0].Values[layout.KeyAmount] != int64(1150) {
		t.Fatalf("KeyAmount = %v, want 1150", segments[0].Values[layout.KeyAmount])
	}
}

func TestDARFSimpleValidateAndSegments(t *testing.T) {
	now := time.Now()
	valid := DARFSimple{
		TaxCode:  "0220",
		Taxpayer: validPayee(),
		DueDate:  now.AddDate(0, 0, 5),
		Amount:   1500,
		Date:     now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	segments, err := valid.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 1 || segments[0].Key != layout.SegmentNSimple {
		t.Fatalf("unexpected segments: %+v", segments)
	}
}

func TestGPSValidateAndSegments(t *testing.T) {
	now := time.Now()
	valid := GPS{
		Taxpayer: validPayee(),
		Period:   now,
		DueDate:  now.AddDate(0, 0, 5),
		Amount:   800,
		Date:     now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	zeroAmount := valid
	zeroAmount.Amount = 0
	if err := zeroAmount.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for zero amount")
	}

	segments, err := valid.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 1 || segments[0].Key != layout.SegmentNSocial {
		t.Fatalf("unexpected segments: %+v", segments)
	}
	if segments[0].Values[layout.KeyPeriod] != formatMonthYear(now) {
		t.Fatalf("KeyPeriod = %v", segments[0].Values[layout.KeyPeriod])
	}
}
