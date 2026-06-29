package febraban240

import (
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
	"github.com/raykavin/gocnab/internal/engine"
)

func TestSelfRegistered(t *testing.T) {
	l, ok := layout.Lookup("febraban240")
	if !ok {
		t.Fatal("febraban240 did not self-register")
	}
	if l.Name() != "febraban240" {
		t.Fatalf("Name() = %q, want \"febraban240\"", l.Name())
	}
	if l.Version() == "" {
		t.Fatal("Version() is empty")
	}
}

var allKeys = []layout.RecordKey{
	layout.FileHeader, layout.FileTrailer, layout.BatchHeader, layout.BatchTrailer,
	layout.SegmentA, layout.SegmentB, layout.SegmentBPix, layout.SegmentJ, layout.SegmentJ52,
	layout.SegmentO, layout.SegmentN, layout.SegmentNSimple, layout.SegmentNSocial,
}

func TestEveryRecordCovers240Columns(t *testing.T) {
	l := febraban240{}
	for _, key := range allKeys {
		spec, ok := l.Record(key)
		if !ok {
			t.Fatalf("Record(%q) ok = false, want true", key)
		}

		total := 0
		for _, f := range spec.Fields {
			total += f.Size()
		}
		if total != 240 {
			t.Fatalf("record %q: fields sum to %d columns, want 240", key, total)
		}
	}
}

func TestRecordUnknownKey(t *testing.T) {
	l := febraban240{}
	if _, ok := l.Record(layout.RecordKey("does-not-exist")); ok {
		t.Fatal("Record() ok = true for an unknown key, want false")
	}
}

// TestEngineAcceptsLayout confirms the whole layout compiles cleanly
// through the real engine (no gaps/overlaps across any record), which is
// a stronger check than summing field sizes alone.
func TestEngineAcceptsLayout(t *testing.T) {
	if _, err := engine.New(febraban240{}); err != nil {
		t.Fatalf("engine.New(febraban240{}) error = %v", err)
	}
}
