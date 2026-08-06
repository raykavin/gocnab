package engine

import (
	"strings"
	"testing"
)

func TestClassifyLine(t *testing.T) {
	tests := []struct {
		name            string
		recordTypeChar  string
		segmentCodeChar string
		wantType        string
		wantSegment     string
	}{
		{"file header", "0", " ", RecordTypeFileHeader, ""},
		{"batch header", "1", " ", RecordTypeBatchHeader, ""},
		{"segment A", "3", "A", RecordTypeDetail, "A"},
		{"segment B", "3", "B", RecordTypeDetail, "B"},
		{"batch trailer", "5", " ", RecordTypeBatchTrailer, ""},
		{"file trailer", "9", " ", RecordTypeFileTrailer, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := buildClassifyLine(tt.recordTypeChar, tt.segmentCodeChar)
			gotType, gotSegment, err := ClassifyLine(line)
			if err != nil {
				t.Fatalf("ClassifyLine() error = %v", err)
			}
			if gotType != tt.wantType {
				t.Errorf("recordType = %q, want %q", gotType, tt.wantType)
			}
			if gotSegment != tt.wantSegment {
				t.Errorf("segmentCode = %q, want %q", gotSegment, tt.wantSegment)
			}
		})
	}
}

func TestClassifyLineRejectsShortLine(t *testing.T) {
	if _, _, err := ClassifyLine("too short"); err == nil {
		t.Fatal("ClassifyLine() error = nil, want an error for a short line")
	}
}

// buildClassifyLine places recordTypeChar at column 8 and segmentCodeChar
// at column 14 of an otherwise blank 240 character line, matching every
// febraban240 RecordSpec's fixed positions for these two fields.
func buildClassifyLine(recordTypeChar, segmentCodeChar string) string {
	buf := []byte(strings.Repeat(" ", recordWidth))
	buf[recordTypeColumn-1] = recordTypeChar[0]
	buf[segmentCodeColumn-1] = segmentCodeChar[0]
	return string(buf)
}
