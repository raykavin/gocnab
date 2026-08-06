package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestRecordParseRoundTrip(t *testing.T) {
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fullCoverageFields()})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	line, err := rec.render(layout.Values{layout.KeyPayeeName: "ACME"})
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}

	values, err := rec.parse(line)
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if got := values[layout.KeyPayeeName]; got != "ACME" {
		t.Fatalf("parse() KeyPayeeName = %q, want %q", got, "ACME")
	}
	// RecordType is a const field: it must not appear as data.
	if _, ok := values[layout.Key("RecordType")]; ok {
		t.Fatal("parse() exposed a const field as a Values entry")
	}
}

func TestRecordParseNumericRoundTrip(t *testing.T) {
	fields := []layout.FieldSpec{
		{Name: "Amount", Start: 1, End: 15, Kind: layout.KindNumeric, Decimals: 2, Key: layout.KeyAmount},
		{Name: "Filler", Start: 16, End: 240, Kind: layout.KindAlphanumeric},
	}
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fields})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	line, err := rec.render(layout.Values{layout.KeyAmount: int64(25200)})
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}

	values, err := rec.parse(line)
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if got := values[layout.KeyAmount]; got != "000000000025200" {
		t.Fatalf("parse() KeyAmount = %q, want the raw zero-padded digit string", got)
	}
}

func TestRecordParseRejectsWrongLength(t *testing.T) {
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fullCoverageFields()})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	if _, err := rec.parse("too short"); err == nil {
		t.Fatal("parse() error = nil, want a length error")
	} else if !strings.Contains(err.Error(), "characters") {
		t.Fatalf("error %q does not mention the length problem", err)
	}
}

func TestRecordParseDetectsConstMismatch(t *testing.T) {
	rec, err := compileRecord(layout.RecordSpec{Name: "test", Fields: fullCoverageFields()})
	if err != nil {
		t.Fatalf("compileRecord() error = %v", err)
	}

	line, err := rec.render(layout.Values{layout.KeyPayeeName: "ACME"})
	if err != nil {
		t.Fatalf("render() error = %v", err)
	}
	// Corrupt the RecordType const field (column 1, rendered "0").
	corrupted := "1" + line[1:]

	_, err = rec.parse(corrupted)
	if err == nil {
		t.Fatal("parse() error = nil, want a RecordMismatchError")
	}
	var mismatch *RecordMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("error = %v (%T), want *RecordMismatchError", err, err)
	}
}
