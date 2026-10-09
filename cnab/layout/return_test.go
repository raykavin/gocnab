package layout

import "testing"

// stubReturnLayout is a stubLayout that also declares return records.
type stubReturnLayout struct {
	stubLayout
	returnRecords map[RecordKey]RecordSpec
}

func (s stubReturnLayout) ReturnRecord(key RecordKey) (RecordSpec, bool) {
	spec, ok := s.returnRecords[key]
	return spec, ok
}

func TestForReturnLeavesAPlainLayoutUnchanged(t *testing.T) {
	plain := stubLayout{
		name:    "plain",
		version: "001",
		records: map[RecordKey]RecordSpec{SegmentA: {Name: "remittance A"}},
	}

	got := ForReturn(plain)
	if _, ok := got.(stubLayout); !ok {
		t.Fatalf("ForReturn() = %T, want the stubLayout itself", got)
	}
	if spec, _ := got.Record(SegmentA); spec.Name != "remittance A" {
		t.Errorf("Record(SegmentA).Name = %q, want %q", spec.Name, "remittance A")
	}
}

func TestForReturnPrefersTheReturnRecord(t *testing.T) {
	l := stubReturnLayout{
		stubLayout: stubLayout{
			name:    "split",
			version: "002",
			records: map[RecordKey]RecordSpec{
				FileHeader: {Name: "remittance header"},
				SegmentA:   {Name: "remittance A"},
			},
		},
		returnRecords: map[RecordKey]RecordSpec{
			SegmentA: {Name: "return A"},
			SegmentZ: {Name: "return Z"},
		},
	}

	got := ForReturn(l)
	if got.Name() != "split" || got.Version() != "002" {
		t.Errorf("Name()/Version() = %q/%q, want %q/%q", got.Name(), got.Version(), "split", "002")
	}

	cases := []struct {
		key  RecordKey
		want string
	}{
		{SegmentA, "return A"},            // declared on both sides: the return one wins
		{FileHeader, "remittance header"}, // not declared for the return: falls back
		{SegmentZ, "return Z"},            // return-only record
	}
	for _, c := range cases {
		spec, ok := got.Record(c.key)
		if !ok {
			t.Errorf("Record(%s) ok = false, want true", c.key)
			continue
		}
		if spec.Name != c.want {
			t.Errorf("Record(%s).Name = %q, want %q", c.key, spec.Name, c.want)
		}
	}
	if _, ok := got.Record(SegmentJ); ok {
		t.Error("Record(SegmentJ) ok = true, want false (declared on neither side)")
	}

	// The view answers one way per record: it must not look like a
	// ReturnLayout that could offer a second answer.
	if _, ok := got.(ReturnLayout); ok {
		t.Error("ForReturn() result implements ReturnLayout, want a plain Layout")
	}
	// The layout itself still writes remittances with its own records.
	if spec, _ := l.Record(SegmentA); spec.Name != "remittance A" {
		t.Errorf("l.Record(SegmentA).Name = %q, want %q", spec.Name, "remittance A")
	}
}
