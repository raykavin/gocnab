package layout

import (
	"fmt"
	"slices"
	"sort"
)

// RecordKey identifies the role a 240 character line plays inside a CNAB
// 240 file. The engine asks a Layout for the RecordSpec of each key it
// needs to render; a Layout that does not support a given key returns
// ok=false from Record, which lets the engine (and the public cnab
// package) reject a payment type that is not compatible with the active
// bank layout before any bytes are written.
type RecordKey string

const (
	// FileHeader is the file header record (record type 0).
	FileHeader RecordKey = "file_header"
	// FileTrailer is the file trailer record (record type 9).
	FileTrailer RecordKey = "file_trailer"
	// BatchHeader is the batch (lote) header record (record type 1).
	BatchHeader RecordKey = "batch_header"
	// BatchTrailer is the batch (lote) trailer record (record type 5).
	BatchTrailer RecordKey = "batch_trailer"
	// SegmentA carries the payment instruction core data: beneficiary bank
	// data, amount and payment date. Used by credit-in-account, TED and
	// PIX payments.
	SegmentA RecordKey = "segment_a"
	// SegmentB complements SegmentA with the beneficiary document and
	// address data. Used by credit-in-account, TED and PIX-by-bank-data
	// payments.
	SegmentB RecordKey = "segment_b"
	// SegmentBPix complements SegmentA with the beneficiary document and
	// PIX key data. Used by PIX-by-key payments. The FEBRABAN standard
	// reuses the same segment letter ("B") for this and for SegmentB with
	// different physical content depending on the payment kind; this SDK
	// models that as two distinct RecordKeys instead of one polymorphic
	// shape, so each stays a single, straightforward physical layout.
	SegmentBPix RecordKey = "segment_b_pix"
	// SegmentJ carries boleto payment data (barcode, due date, amount).
	SegmentJ RecordKey = "segment_j"
	// SegmentJ52 complements SegmentJ with payer and assignor document data.
	SegmentJ52 RecordKey = "segment_j52"
	// SegmentO carries utility bill / barcoded tax payment data.
	SegmentO RecordKey = "segment_o"
	// SegmentN carries a normal DARF tax payment (principal, fine and
	// interest tracked separately).
	SegmentN RecordKey = "segment_n"
	// SegmentNSimple carries a DARF Simples tax payment (a single total
	// amount). Like SegmentB/SegmentBPix, FEBRABAN reuses the "N" segment
	// letter for a physically different layout here; this SDK models it
	// as its own RecordKey.
	SegmentNSimple RecordKey = "segment_n_simple"
	// SegmentNSocial carries a GPS (Guia da Previdência Social) payment.
	SegmentNSocial RecordKey = "segment_n_social"
)

// AllRecordKeys lists every RecordKey the engine knows how to render. A
// Layout does not need to implement all of them; Engine.New skips any key
// for which Layout.Record returns ok=false.
var AllRecordKeys = []RecordKey{
	FileHeader, FileTrailer, BatchHeader, BatchTrailer,
	SegmentA, SegmentB, SegmentBPix, SegmentJ, SegmentJ52,
	SegmentO, SegmentN, SegmentNSimple, SegmentNSocial,
}

// IsKnownRecordKey reports whether key is one of the values in
// AllRecordKeys.
func IsKnownRecordKey(key RecordKey) bool {
	return slices.Contains(AllRecordKeys, key)
}

// RecordSpec describes the full set of 240 columns of one record type or
// segment. Fields must cover columns 1 through 240 with no overlap and no
// gap; the engine validates this once when a Layout is registered.
type RecordSpec struct {
	// Name is a descriptive English identifier used in error messages.
	Name string
	// Fields lists every column range of the record. Order does not need
	// to match column order; the engine sorts by Start before validating
	// and rendering.
	Fields []FieldSpec
}

// Validate checks that r's fields cover columns 1-240 with no gap and no
// overlap, returning a descriptive error otherwise. The engine runs this
// same check when a Layout is used to build an Engine; Validate lets a
// layout author (or a config-driven loader, see NewFromJSON) catch the
// same problem earlier, with the same guarantee.
func (r RecordSpec) Validate() error {
	fields := make([]FieldSpec, len(r.Fields))
	copy(fields, r.Fields)
	sort.Slice(fields, func(i, j int) bool { return fields[i].Start < fields[j].Start })

	expected := 1
	for _, f := range fields {
		if f.End < f.Start {
			return fmt.Errorf("field %q has End (%d) before Start (%d)", f.Name, f.End, f.Start)
		}
		switch {
		case f.Start > expected:
			return fmt.Errorf("gap between columns %d and %d", expected, f.Start-1)
		case f.Start < expected:
			return fmt.Errorf("field %q overlaps a previous field at column %d", f.Name, f.Start)
		}
		expected = f.End + 1
	}
	if expected != 241 {
		return fmt.Errorf("record covers columns 1-%d, want 1-240", expected-1)
	}
	return nil
}
