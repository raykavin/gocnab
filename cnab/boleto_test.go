package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func validBoleto(now time.Time) BoletoPayment {
	return BoletoPayment{
		Barcode:  "34191924500025200001570004013540025876327000",
		Assignor: validPayee(),
		Payer:    validPayee(),
		DueDate:  now.AddDate(0, 0, 5),
		Amount:   20000,
		Date:     now.AddDate(0, 0, 1),
	}
}

func TestBoletoPaymentValidate(t *testing.T) {
	now := time.Now()
	if err := validBoleto(now).validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	shortBarcode := validBoleto(now)
	shortBarcode.Barcode = "123"
	if err := shortBarcode.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for a short barcode")
	}

	zeroAmount := validBoleto(now)
	zeroAmount.Amount = 0
	if err := zeroAmount.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for zero amount")
	}
}

func TestBoletoPaymentToSegments(t *testing.T) {
	now := time.Now()
	b := validBoleto(now)

	segments, err := b.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	if segments[0].Key != layout.SegmentJ || segments[1].Key != layout.SegmentJ52 {
		t.Fatalf("unexpected segment keys: %v, %v", segments[0].Key, segments[1].Key)
	}
	if segments[0].Values[layout.KeyBarcode] != "34191924500025200001570004013540025876327000" {
		t.Fatalf("KeyBarcode = %v", segments[0].Values[layout.KeyBarcode])
	}
	if segments[0].Values[layout.KeyDocumentAmount] != int64(20000) {
		t.Fatalf("KeyDocumentAmount = %v, want 20000 (falls back to Amount)", segments[0].Values[layout.KeyDocumentAmount])
	}
}
