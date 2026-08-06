package cnab

import (
	"errors"
	"testing"
	"time"

	"github.com/raykavin/gocnab/internal/engine"
)

// No real bank return-file sample is available for a reference layout that
// represents no actual bank, so these tests build one the same way the
// SDK itself would receive it: generate a remittance with NewRemittance
// exactly as a caller would, then patch the three return-only column
// ranges of Segmento A (columns 155-177 and 231-240 see segment_a.go)
// that a bank fills in on the way back. Every other byte, including every
// structural field the engine computes, is real, generated content.

// patchLine overwrites content's line index i (0-based; each line is 240
// data characters followed by "\r\n", so a stride of 242 bytes) at
// [startCol, endCol] (1-based, inclusive) with value, which must be
// exactly endCol-startCol+1 bytes.
func patchLine(t *testing.T, content []byte, i, startCol, endCol int, value string) {
	t.Helper()
	width := endCol - startCol + 1
	if len(value) != width {
		t.Fatalf("patchLine: value %q has %d bytes, want %d", value, len(value), width)
	}
	offset := i*242 + (startCol - 1)
	copy(content[offset:offset+width], value)
}

func segmentALineIndex(t *testing.T, content []byte, occurrence int) int {
	t.Helper()
	lines := splitLines(content)
	found := 0
	for i, line := range lines {
		recordType, segmentCode, err := engine.ClassifyLine(line)
		if err != nil {
			t.Fatalf("classify line %d: %v", i, err)
		}
		if recordType == engine.RecordTypeDetail && segmentCode == "A" {
			if found == occurrence {
				return i
			}
			found++
		}
	}
	t.Fatalf("did not find segment A occurrence %d in %d lines", occurrence, len(lines))
	return -1
}

func TestParseReturn_AcceptedAndRejected(t *testing.T) {
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}

	tomorrow := time.Now().AddDate(0, 0, 1)
	if err := batch.AddPayment(Pix{
		Key:        EmailKey("fornecedor@exemplo.com"),
		Payee:      validPayee(),
		Amount:     Cents(25200),
		Date:       tomorrow,
		YourNumber: "NF-0001",
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}
	if err := batch.AddPayment(Pix{
		Key:        EmailKey("outro@exemplo.com"),
		Payee:      validPayee(),
		Amount:     Cents(9900),
		Date:       tomorrow,
		YourNumber: "NF-0002",
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	content, err := f.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	first := segmentALineIndex(t, content, 0)
	second := segmentALineIndex(t, content, 1)

	// Simulate the bank settling the first payment: real date and amount,
	// no occurrence code.
	patchLine(t, content, first, 155, 162, "02012026")
	patchLine(t, content, first, 163, 177, "000000000025200")

	// Simulate the bank rejecting the second payment: occurrence code "02"
	// in the first two characters of the field, blank elsewhere, no
	// settlement date/amount.
	patchLine(t, content, second, 231, 240, "02        ")

	result, err := ParseReturn("febraban240", content)
	if err != nil {
		t.Fatalf("ParseReturn() error = %v", err)
	}
	if len(result.Movements) != 2 {
		t.Fatalf("len(Movements) = %d, want 2", len(result.Movements))
	}

	accepted := result.Movements[0]
	if accepted.YourNumber != "NF-0001" {
		t.Errorf("Movements[0].YourNumber = %q, want %q", accepted.YourNumber, "NF-0001")
	}
	if accepted.Amount != 25200 {
		t.Errorf("Movements[0].Amount = %d, want 25200", accepted.Amount)
	}
	if !accepted.Accepted() {
		t.Errorf("Movements[0].Accepted() = false, want true (codes: %v)", accepted.OccurrenceCodes)
	}
	if accepted.SettlementAmount != 25200 {
		t.Errorf("Movements[0].SettlementAmount = %d, want 25200", accepted.SettlementAmount)
	}
	wantDate := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
	if !accepted.SettlementDate.Equal(wantDate) {
		t.Errorf("Movements[0].SettlementDate = %v, want %v", accepted.SettlementDate, wantDate)
	}

	rejected := result.Movements[1]
	if rejected.YourNumber != "NF-0002" {
		t.Errorf("Movements[1].YourNumber = %q, want %q", rejected.YourNumber, "NF-0002")
	}
	if rejected.Accepted() {
		t.Error("Movements[1].Accepted() = true, want false")
	}
	if len(rejected.OccurrenceCodes) != 1 || rejected.OccurrenceCodes[0] != "02" {
		t.Errorf("Movements[1].OccurrenceCodes = %v, want [\"02\"]", rejected.OccurrenceCodes)
	}
	if !rejected.SettlementDate.IsZero() {
		t.Errorf("Movements[1].SettlementDate = %v, want zero", rejected.SettlementDate)
	}
}

func TestParseReturn_UnregisteredLayout(t *testing.T) {
	if _, err := ParseReturn("does-not-exist", []byte{}); err == nil {
		t.Fatal("ParseReturn() error = nil, want an error for an unregistered layout")
	}
}

func TestParseReturn_RejectsWrongLineLength(t *testing.T) {
	_, err := ParseReturn("febraban240", []byte("too short\r\n"))
	if err == nil {
		t.Fatal("ParseReturn() error = nil, want an error for a malformed line")
	}
	var parseErr *ReturnParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error = %v (%T), want *ReturnParseError", err, err)
	}
}

// TestParseReturn_ExposesMismatchAsPublicType confirms a caller outside
// this module who cannot import internal/engine, and so cannot name
// *engine.RecordMismatchError still gets a type they can match against
// with errors.As when a line's const field does not match what the layout
// expects: ParseReturn must translate it to the exported *ReturnParseError,
// never let the internal type escape unwrapped.
func TestParseReturn_ExposesMismatchAsPublicType(t *testing.T) {
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := batch.AddPayment(Pix{
		Key:        EmailKey("fornecedor@exemplo.com"),
		Payee:      validPayee(),
		Amount:     Cents(25200),
		Date:       time.Now().AddDate(0, 0, 1),
		YourNumber: "NF-0001",
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}
	content, err := f.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Corrupt the segment_a "BankCode" const field (columns 1-3). Unlike
	// RecordType (column 8) or SegmentCode (column 14), this does not
	// affect how ParseReturn classifies the line, so it reaches
	// Engine.ParseRecord and fails there instead.
	first := segmentALineIndex(t, content, 0)
	patchLine(t, content, first, 1, 3, "999")

	_, err = ParseReturn("febraban240", content)
	if err == nil {
		t.Fatal("ParseReturn() error = nil, want an error for the corrupted bank code")
	}
	var parseErr *ReturnParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error = %v (%T), want *ReturnParseError", err, err)
	}
	if parseErr.Field != "BankCode" {
		t.Errorf("Field = %q, want %q", parseErr.Field, "BankCode")
	}
	if parseErr.Expected != "000" || parseErr.Got != "999" {
		t.Errorf("Expected/Got = %q/%q, want %q/%q", parseErr.Expected, parseErr.Got, "000", "999")
	}
}
