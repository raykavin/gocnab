package engine

import (
	"strings"
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
)

func fullCoverageFields() []layout.FieldSpec {
	return []layout.FieldSpec{
		{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "0"},
		{Name: "Name", Start: 2, End: 11, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName},
		{Name: "Filler", Start: 12, End: 240, Kind: layout.KindAlphanumeric},
	}
}

func TestCompileRecordValid(t *testing.T) {
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fullCoverageFields()})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}
	if len(rec.fields) != 3 {
		t.Fatalf("compiled %d fields, want 3", len(rec.fields))
	}
}

func TestCompileRecordDetectsGap(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "A", Start: 1, End: 5, Kind: layout.KindAlphanumeric},
		{Name: "B", Start: 10, End: 240, Kind: layout.KindAlphanumeric},
	}
	_, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err == nil {
		t.Fatal("compileRecord() error = nil, want a gap error")
	}
	if !strings.Contains(err.Error(), "gap") {
		t.Fatalf("error %q does not mention the gap", err)
	}
}

func TestCompileRecordDetectsOverlap(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "A", Start: 1, End: 10, Kind: layout.KindAlphanumeric},
		{Name: "B", Start: 5, End: 240, Kind: layout.KindAlphanumeric},
	}
	_, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err == nil {
		t.Fatal("compileRecord() error = nil, want an overlap error")
	}
	if !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("error %q does not mention the overlap", err)
	}
}

func TestCompileRecordDetectsShortRecord(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "A", Start: 1, End: 100, Kind: layout.KindAlphanumeric},
	}
	_, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err == nil {
		t.Fatal("compileRecord() error = nil, want a record-too-short error")
	}
	if !strings.Contains(err.Error(), "1-240") {
		t.Fatalf("error %q does not mention the expected range", err)
	}
}

func TestCompileRecordDetectsInvertedField(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "A", Start: 10, End: 5, Kind: layout.KindAlphanumeric},
	}
	_, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err == nil {
		t.Fatal("compileRecord() error = nil, want an inverted field error")
	}
}

func TestRecordRenderByteExact(t *testing.T) {
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fullCoverageFields()})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	line, err := rec.render(layout.Values{layout.KeyPayeeName: "ACME"})
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}
	if len(line) != 240 {
		t.Fatalf("render() length = %d, want 240", len(line))
	}
	want := "0" + "ACME      " + strings.Repeat(" ", 229)
	if line != want {
		t.Fatalf("render() = %q, want %q", line, want)
	}
}

func TestRecordRenderPropagatesFieldError(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "Amount", Start: 1, End: 2, Kind: layout.KindNumeric, Key: layout.KeyAmount},
		{Name: "Filler", Start: 3, End: 240, Kind: layout.KindAlphanumeric},
	}
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	if _, err := rec.render(layout.Values{layout.KeyAmount: 12345}); err == nil {
		t.Fatal("render() error = nil, want an error for an overflowing field")
	}
}
