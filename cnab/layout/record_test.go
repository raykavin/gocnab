package layout

import (
	"strings"
	"testing"
)

func TestRecordSpecValidateValid(t *testing.T) {
	spec := RecordSpec{
		Name: "test",
		Fields: []FieldSpec{
			{Name: "A", Start: 1, End: 10, Kind: KindAlphanumeric},
			{Name: "B", Start: 11, End: 240, Kind: KindAlphanumeric},
		},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestRecordSpecValidateGap(t *testing.T) {
	spec := RecordSpec{Fields: []FieldSpec{
		{Name: "A", Start: 1, End: 5, Kind: KindAlphanumeric},
		{Name: "B", Start: 10, End: 240, Kind: KindAlphanumeric},
	}}
	err := spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "gap") {
		t.Fatalf("Validate() error = %v, want a gap error", err)
	}
}

func TestRecordSpecValidateOverlap(t *testing.T) {
	spec := RecordSpec{Fields: []FieldSpec{
		{Name: "A", Start: 1, End: 10, Kind: KindAlphanumeric},
		{Name: "B", Start: 5, End: 240, Kind: KindAlphanumeric},
	}}
	err := spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("Validate() error = %v, want an overlap error", err)
	}
}

func TestRecordSpecValidateShortRecord(t *testing.T) {
	spec := RecordSpec{Fields: []FieldSpec{
		{Name: "A", Start: 1, End: 100, Kind: KindAlphanumeric},
	}}
	err := spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "1-240") {
		t.Fatalf("Validate() error = %v, want a record-too-short error", err)
	}
}

func TestRecordSpecValidateInvertedField(t *testing.T) {
	spec := RecordSpec{Fields: []FieldSpec{
		{Name: "A", Start: 10, End: 5, Kind: KindAlphanumeric},
	}}
	if err := spec.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an error for an inverted field")
	}
}

func TestIsKnownRecordKey(t *testing.T) {
	if !IsKnownRecordKey(SegmentA) {
		t.Fatal("IsKnownRecordKey(SegmentA) = false, want true")
	}
	if IsKnownRecordKey(RecordKey("not-a-real-key")) {
		t.Fatal("IsKnownRecordKey(bogus) = true, want false")
	}
	if len(AllRecordKeys) != 13 {
		t.Fatalf("len(AllRecordKeys) = %d, want 13", len(AllRecordKeys))
	}
}

func TestIsKnownKey(t *testing.T) {
	if !IsKnownKey(KeyAmount) {
		t.Fatal("IsKnownKey(KeyAmount) = false, want true")
	}
	if IsKnownKey(Key("not-a-real-key")) {
		t.Fatal("IsKnownKey(bogus) = true, want false")
	}
	if len(AllKeys) == 0 {
		t.Fatal("AllKeys is empty")
	}
}
