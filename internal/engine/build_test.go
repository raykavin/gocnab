package engine

import (
	"strings"
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
)

// testLayout is a small, self-contained Layout used only by these tests.
// Every record is 240 columns wide (verified by compileRecord itself), but
// most of that width is an unused filler so test expectations stay easy
// to read and verify byte by byte.
type testLayout struct{}

func (testLayout) Name() string    { return "test240" }
func (testLayout) Version() string { return "001" }

func (testLayout) Record(key layout.RecordKey) (layout.RecordSpec, bool) {
	switch key {
	case layout.FileHeader:
		return layout.RecordSpec{Name: "file_header", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "0"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Const: "0000"},
			{Name: "Filler", Start: 6, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	case layout.FileTrailer:
		return layout.RecordSpec{Name: "file_trailer", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "9"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Const: "9999"},
			{Name: "BatchCount", Start: 6, End: 8, Kind: layout.KindNumeric, Key: layout.KeyBatchCount},
			{Name: "RecordCount", Start: 9, End: 14, Kind: layout.KindNumeric, Key: layout.KeyFileRecordCount},
			{Name: "Filler", Start: 15, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	case layout.BatchHeader:
		return layout.RecordSpec{Name: "batch_header", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "1"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Key: layout.KeyBatchNumber},
			{Name: "Filler", Start: 6, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	case layout.BatchTrailer:
		return layout.RecordSpec{Name: "batch_trailer", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "5"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Key: layout.KeyBatchNumber},
			{Name: "RecordCount", Start: 6, End: 11, Kind: layout.KindNumeric, Key: layout.KeyBatchRecordCount},
			{Name: "Amount", Start: 12, End: 26, Kind: layout.KindNumeric, Key: layout.KeyBatchAmount},
			{Name: "Filler", Start: 27, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	case layout.SegmentA:
		return layout.RecordSpec{Name: "segment_a", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "3"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Key: layout.KeyBatchNumber},
			{Name: "Sequence", Start: 6, End: 10, Kind: layout.KindNumeric, Key: layout.KeySequence},
			{Name: "SegmentCode", Start: 11, End: 11, Kind: layout.KindAlphanumeric, Const: "A"},
			{Name: "Amount", Start: 12, End: 26, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			{Name: "Filler", Start: 27, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	case layout.SegmentB:
		return layout.RecordSpec{Name: "segment_b", Fields: []layout.FieldSpec{
			{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "3"},
			{Name: "BatchNumber", Start: 2, End: 5, Kind: layout.KindNumeric, Key: layout.KeyBatchNumber},
			{Name: "Sequence", Start: 6, End: 10, Kind: layout.KindNumeric, Key: layout.KeySequence},
			{Name: "SegmentCode", Start: 11, End: 11, Kind: layout.KindAlphanumeric, Const: "B"},
			{Name: "Filler", Start: 12, End: 240, Kind: layout.KindAlphanumeric},
		}}, true
	default:
		return layout.RecordSpec{}, false
	}
}

func segA(amount int64) layout.Values {
	return layout.Values{layout.KeyAmount: amount}
}

func segB() layout.Values {
	return layout.Values{}
}

func TestBuildSingleBatchByteExact(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	in := FileInput{
		Batches: []BatchInput{
			{
				Movements: [][]DetailLine{
					{
						{Key: layout.SegmentA, Values: segA(1000)},
						{Key: layout.SegmentB, Values: segB()},
					},
					{
						{Key: layout.SegmentA, Values: segA(2500)},
					},
				},
			},
		},
	}

	out, err := e.Build(in)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	lines := splitLines(t, out, 7)

	assertLine(t, lines[0], "0"+"0000"+strings.Repeat(" ", 235))
	assertLine(t, lines[1], "1"+"0001"+strings.Repeat(" ", 235))
	assertLine(t, lines[2], "3"+"0001"+"00001"+"A"+"000000000001000"+strings.Repeat(" ", 214))
	assertLine(t, lines[3], "3"+"0001"+"00002"+"B"+strings.Repeat(" ", 229))
	assertLine(t, lines[4], "3"+"0001"+"00003"+"A"+"000000000002500"+strings.Repeat(" ", 214))
	assertLine(t, lines[5], "5"+"0001"+"000005"+"000000000003500"+strings.Repeat(" ", 214))
	assertLine(t, lines[6], "9"+"9999"+"001"+"000007"+strings.Repeat(" ", 226))
}

func TestBuildCRLFTermination(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	out, err := e.Build(FileInput{Batches: []BatchInput{{Movements: [][]DetailLine{
		{{Key: layout.SegmentA, Values: segA(100)}},
	}}}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !strings.HasSuffix(string(out), "\r\n") {
		t.Fatal("Build() output does not end with CRLF")
	}
	if strings.Contains(string(out), "\n\n") {
		t.Fatal("Build() output has an unexpected blank line")
	}

	lineCount := strings.Count(string(out), "\r\n")
	// file header + batch header + 1 detail + batch trailer + file trailer
	expectedLines := 5
	if lineCount != expectedLines {
		t.Fatalf("got %d CRLF-terminated lines, want %d", lineCount, expectedLines)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(out), "\r\n"), "\r\n") {
		if len(line) != 240 {
			t.Fatalf("line %q has length %d, want 240", line, len(line))
		}
	}
}

func TestBuildMultipleBatchesSequencing(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	in := FileInput{
		Batches: []BatchInput{
			{Movements: [][]DetailLine{{{Key: layout.SegmentA, Values: segA(100)}}}},
			{Movements: [][]DetailLine{
				{{Key: layout.SegmentA, Values: segA(200)}},
				{{Key: layout.SegmentA, Values: segA(300)}},
			}},
		},
	}

	out, err := e.Build(in)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	lines := splitLines(t, out, 1+3+4+1) // file header + (header+1detail+trailer) + (header+2detail+trailer) + file trailer

	// batch 1: header, detail (batch number 0001, sequence 00001), trailer
	assertField(t, lines[1], 2, 5, "0001")
	assertField(t, lines[2], 2, 5, "0001")
	assertField(t, lines[2], 6, 10, "00001")
	assertField(t, lines[3], 2, 5, "0001")
	assertField(t, lines[3], 6, 11, "000003") // header + 1 detail + trailer

	// batch 2: header, 2 details (batch number 0002, sequence resets to 00001)
	assertField(t, lines[4], 2, 5, "0002")
	assertField(t, lines[5], 2, 5, "0002")
	assertField(t, lines[5], 6, 10, "00001")
	assertField(t, lines[6], 2, 5, "0002")
	assertField(t, lines[6], 6, 10, "00002")
	assertField(t, lines[7], 2, 5, "0002")
	assertField(t, lines[7], 6, 11, "000004") // header + 2 details + trailer

	// file trailer: batch count = 2, record count = 9
	assertField(t, lines[8], 6, 8, "002")
	assertField(t, lines[8], 9, 14, "000009")
}

func TestBuildBatchesPerFileLimit(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	batches := make([]BatchInput, MaxBatchesPerFile+1)
	for i := range batches {
		batches[i] = BatchInput{Movements: [][]DetailLine{{{Key: layout.SegmentA, Values: segA(1)}}}}
	}

	_, err = e.Build(FileInput{Batches: batches})
	if err == nil {
		t.Fatal("Build() error = nil, want a LimitError")
	}
	limitErr, ok := err.(*LimitError)
	if !ok {
		t.Fatalf("Build() error type = %T, want *LimitError", err)
	}
	if limitErr.Limit != LimitBatchesPerFile {
		t.Fatalf("LimitError.Limit = %q, want %q", limitErr.Limit, LimitBatchesPerFile)
	}
}

func TestBuildMovementsPerBatchLimit(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	movements := make([][]DetailLine, MaxMovementsPerBatch+1)
	for i := range movements {
		movements[i] = []DetailLine{{Key: layout.SegmentA, Values: segA(1)}}
	}

	_, err = e.Build(FileInput{Batches: []BatchInput{{Movements: movements}}})
	if err == nil {
		t.Fatal("Build() error = nil, want a LimitError")
	}
	limitErr, ok := err.(*LimitError)
	if !ok {
		t.Fatalf("Build() error type = %T, want *LimitError", err)
	}
	if limitErr.Limit != LimitMovementsPerBatch {
		t.Fatalf("LimitError.Limit = %q, want %q", limitErr.Limit, LimitMovementsPerBatch)
	}
	if limitErr.Batch != 1 {
		t.Fatalf("LimitError.Batch = %d, want 1", limitErr.Batch)
	}
}

func TestBuildUnsupportedSegmentErrors(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = e.Build(FileInput{Batches: []BatchInput{
		{Movements: [][]DetailLine{{{Key: layout.SegmentJ, Values: layout.Values{}}}}},
	}})
	if err == nil {
		t.Fatal("Build() error = nil, want an error for an unsupported segment")
	}
	batchErr, ok := err.(*BatchError)
	if !ok {
		t.Fatalf("Build() error type = %T, want *BatchError", err)
	}
	if batchErr.Batch != 1 {
		t.Fatalf("BatchError.Batch = %d, want 1", batchErr.Batch)
	}
}

func TestBuildFieldErrorIsWrappedWithBatchContext(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// The Amount field is only 15 columns wide; this value does not fit
	// and must surface as an error identifying the batch and record.
	huge := int64(1)
	for i := 0; i < 20; i++ {
		huge *= 10
	}

	_, err = e.Build(FileInput{Batches: []BatchInput{
		{Movements: [][]DetailLine{{{Key: layout.SegmentA, Values: segA(huge)}}}},
	}})
	if err == nil {
		t.Fatal("Build() error = nil, want a field render error")
	}
	if !strings.Contains(err.Error(), "batch 1") {
		t.Fatalf("error %q does not mention the batch", err)
	}
}

func TestNewRejectsNilLayout(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) error = nil, want a SpecError")
	}
}

func TestNewRejectsEmptyVersion(t *testing.T) {
	if _, err := New(emptyVersionLayout{}); err == nil {
		t.Fatal("New() error = nil, want a SpecError for an empty version")
	}
}

type emptyVersionLayout struct{}

func (emptyVersionLayout) Name() string    { return "empty-version" }
func (emptyVersionLayout) Version() string { return "" }
func (emptyVersionLayout) Record(layout.RecordKey) (layout.RecordSpec, bool) {
	return layout.RecordSpec{}, false
}

func TestEngineSupports(t *testing.T) {
	e, err := New(testLayout{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !e.Supports(layout.SegmentA) {
		t.Fatal("Supports(SegmentA) = false, want true")
	}
	if e.Supports(layout.SegmentJ) {
		t.Fatal("Supports(SegmentJ) = true, want false")
	}
	if e.LayoutName() != "test240" {
		t.Fatalf("LayoutName() = %q, want %q", e.LayoutName(), "test240")
	}
	if e.LayoutVersion() != "001" {
		t.Fatalf("LayoutVersion() = %q, want %q", e.LayoutVersion(), "001")
	}
}

// --- test helpers ---

func splitLines(t *testing.T, out []byte, want int) []string {
	t.Helper()
	s := strings.TrimSuffix(string(out), "\r\n")
	lines := strings.Split(s, "\r\n")
	if len(lines) != want {
		t.Fatalf("got %d lines, want %d (output: %q)", len(lines), want, string(out))
	}
	for i, line := range lines {
		if len(line) != 240 {
			t.Fatalf("line %d has length %d, want 240", i, len(line))
		}
	}
	return lines
}

func assertLine(t *testing.T, line, wantPrefix string) {
	t.Helper()
	if len(wantPrefix) != 240 {
		t.Fatalf("test bug: wantPrefix length = %d, want 240", len(wantPrefix))
	}
	if line != wantPrefix {
		t.Fatalf("line = %q\nwant  = %q", line, wantPrefix)
	}
}

func assertField(t *testing.T, line string, start, end int, want string) {
	t.Helper()
	got := line[start-1 : end]
	if got != want {
		t.Fatalf("columns %d-%d = %q, want %q (line: %q)", start, end, got, want, line)
	}
}
