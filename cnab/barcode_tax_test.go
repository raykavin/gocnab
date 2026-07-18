package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestBarcodeTaxValidate(t *testing.T) {
	now := time.Now()
	valid := BarcodeTax{
		Barcode: "83600000000285100060000010120234400710517746",
		DueDate: now.AddDate(0, 0, 5),
		Amount:  2851,
		Date:    now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	shortBarcode := valid
	shortBarcode.Barcode = "123"
	if err := shortBarcode.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for a short barcode")
	}
}

func TestBarcodeTaxToSegments(t *testing.T) {
	now := time.Now()
	b := BarcodeTax{
		Barcode: "83600000000285100060000010120234400710517746",
		DueDate: now.AddDate(0, 0, 5),
		Amount:  2851,
		Date:    now.AddDate(0, 0, 1),
	}

	segments, err := b.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("len(segments) = %d, want 1", len(segments))
	}
	if segments[0].Key != layout.SegmentO {
		t.Fatalf("segments[0].Key = %q, want SegmentO", segments[0].Key)
	}
	if segments[0].Values[layout.KeyAmount] != int64(2851) {
		t.Fatalf("KeyAmount = %v, want 2851", segments[0].Values[layout.KeyAmount])
	}
}
