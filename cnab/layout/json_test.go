package layout

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const validJSONLayout = `{
	"name": "test-json-layout",
	"version": "001",
	"records": {
		"file_header": {
			"fields": [
				{"name": "BankCode", "start": 1, "end": 3, "kind": "9", "const": "341"},
				{"name": "ServiceBatchNumber", "start": 4, "end": 7, "kind": "9", "const": "0000"},
				{"name": "RecordType", "start": 8, "end": 8, "kind": "9", "const": "0"},
				{"name": "CompanyName", "start": 9, "end": 38, "kind": "X", "key": "company_name"},
				{"name": "Filler", "start": 39, "end": 240, "kind": "X"}
			]
		},
		"segment_a": {
			"fields": [
				{"name": "Amount", "start": 1, "end": 15, "kind": "9", "decimals": 2, "key": "amount"},
				{"name": "Filler", "start": 16, "end": 240, "kind": "X"}
			]
		}
	}
}`

func TestNewFromJSONValid(t *testing.T) {
	l, err := NewFromJSON([]byte(validJSONLayout))
	if err != nil {
		t.Fatalf("NewFromJSON() error = %v", err)
	}
	if l.Name() != "test-json-layout" {
		t.Fatalf("Name() = %q, want %q", l.Name(), "test-json-layout")
	}
	if l.Version() != "001" {
		t.Fatalf("Version() = %q, want %q", l.Version(), "001")
	}

	spec, ok := l.Record(FileHeader)
	if !ok {
		t.Fatal("Record(FileHeader) ok = false, want true")
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("file_header Validate() error = %v", err)
	}

	if _, ok := l.Record(SegmentJ); ok {
		t.Fatal("Record(SegmentJ) ok = true, want false (not defined in the JSON)")
	}
}

func TestNewFromJSONInvalidJSON(t *testing.T) {
	if _, err := NewFromJSON([]byte("{not json")); err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for malformed JSON")
	}
}

func TestNewFromJSONMissingNameOrVersion(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"missing name", `{"version": "001", "records": {"file_header": {"fields": [{"start":1,"end":240,"kind":"X"}]}}}`},
		{"missing version", `{"name": "x", "records": {"file_header": {"fields": [{"start":1,"end":240,"kind":"X"}]}}}`},
		{"missing records", `{"name": "x", "version": "001"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewFromJSON([]byte(c.data)); err == nil {
				t.Fatal("NewFromJSON() error = nil, want an error")
			}
		})
	}
}

func TestNewFromJSONUnknownRecordKey(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "not_a_real_record": {"fields": [{"start":1,"end":240,"kind":"X"}]} }
	}`
	_, err := NewFromJSON([]byte(data))
	if err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for an unknown record key")
	}
	if !strings.Contains(err.Error(), "not_a_real_record") {
		t.Fatalf("error %q does not mention the offending record key", err)
	}
}

func TestNewFromJSONNoFieldsInRecord(t *testing.T) {
	data := `{"name":"x","version":"001","records":{"file_header":{"fields":[]}}}`
	if _, err := NewFromJSON([]byte(data)); err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for a record with no fields")
	}
}

func TestNewFromJSONInvalidKind(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [{"start":1,"end":240,"kind":"Z"}]} }
	}`
	_, err := NewFromJSON([]byte(data))
	if err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for an invalid kind")
	}
	if !strings.Contains(err.Error(), "kind") {
		t.Fatalf("error %q does not mention the invalid kind", err)
	}
}

func TestNewFromJSONInvalidColumnRange(t *testing.T) {
	cases := []struct {
		name  string
		start int
		end   int
	}{
		{"start before 1", 0, 10},
		{"end before start", 10, 5},
		{"end beyond 240", 200, 241},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data := fmtField(c.start, c.end)
			if _, err := NewFromJSON([]byte(data)); err == nil {
				t.Fatal("NewFromJSON() error = nil, want an error for an invalid column range")
			}
		})
	}
}

func fmtField(start, end int) string {
	return `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [{"start":` +
		strconv.Itoa(start) + `,"end":` + strconv.Itoa(end) + `,"kind":"X"}]} }
	}`
}

func TestNewFromJSONBothKeyAndConst(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [
			{"start":1,"end":240,"kind":"X","key":"company_name","const":"FOO"}
		]} }
	}`
	_, err := NewFromJSON([]byte(data))
	if err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error when both key and const are set")
	}
	if !strings.Contains(err.Error(), "both") {
		t.Fatalf("error %q does not mention the conflict", err)
	}
}

func TestNewFromJSONUnknownKey(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [
			{"start":1,"end":240,"kind":"X","key":"not_a_real_key"}
		]} }
	}`
	_, err := NewFromJSON([]byte(data))
	if err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for an unknown key")
	}
	if !strings.Contains(err.Error(), "not_a_real_key") {
		t.Fatalf("error %q does not mention the offending key", err)
	}
}

func TestNewFromJSONNegativeDecimals(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [
			{"start":1,"end":240,"kind":"9","key":"amount","decimals":-1}
		]} }
	}`
	if _, err := NewFromJSON([]byte(data)); err == nil {
		t.Fatal("NewFromJSON() error = nil, want an error for negative decimals")
	}
}

func TestNewFromJSONGapAndOverlap(t *testing.T) {
	cases := []struct {
		name   string
		fields string
	}{
		{"gap", `[{"start":1,"end":10,"kind":"X"},{"start":15,"end":240,"kind":"X"}]`},
		{"overlap", `[{"start":1,"end":10,"kind":"X"},{"start":5,"end":240,"kind":"X"}]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data := `{"name":"x","version":"001","records":{"file_header":{"fields":` + c.fields + `}}}`
			if _, err := NewFromJSON([]byte(data)); err == nil {
				t.Fatalf("NewFromJSON() error = nil, want a %s error", c.name)
			}
		})
	}
}

func TestNewFromJSONFieldKindAliases(t *testing.T) {
	data := `{
		"name": "x", "version": "001",
		"records": { "file_header": {"fields": [
			{"start":1,"end":100,"kind":"numeric"},
			{"start":101,"end":240,"kind":"alphanumeric"}
		]} }
	}`
	if _, err := NewFromJSON([]byte(data)); err != nil {
		t.Fatalf("NewFromJSON() error = %v, want nil for the numeric/alphanumeric aliases", err)
	}
}

func TestNewFromJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "layout.json")
	if err := os.WriteFile(path, []byte(validJSONLayout), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	l, err := NewFromJSONFile(path)
	if err != nil {
		t.Fatalf("NewFromJSONFile() error = %v", err)
	}
	if l.Name() != "test-json-layout" {
		t.Fatalf("Name() = %q, want %q", l.Name(), "test-json-layout")
	}
}

func TestNewFromJSONFileMissing(t *testing.T) {
	if _, err := NewFromJSONFile(filepath.Join(t.TempDir(), "does-not-exist.json")); err == nil {
		t.Fatal("NewFromJSONFile() error = nil, want an error for a missing file")
	}
}

func TestJSONLoadedLayoutWorksWithEngine(t *testing.T) {
	// A minimal but complete layout (every record the engine needs for a
	// full Build) confirms the JSON loader output is a genuinely usable
	// Layout, not just something that satisfies the interface signature.
	data := `{
		"name": "full-json-layout",
		"version": "001",
		"records": {
			"file_header":   {"fields": [{"start":1,"end":240,"kind":"X"}]},
			"file_trailer":  {"fields": [{"start":1,"end":240,"kind":"X"}]},
			"batch_header":  {"fields": [{"start":1,"end":240,"kind":"X"}]},
			"batch_trailer": {"fields": [{"start":1,"end":240,"kind":"X"}]}
		}
	}`
	l, err := NewFromJSON([]byte(data))
	if err != nil {
		t.Fatalf("NewFromJSON() error = %v", err)
	}
	if l.Version() == "" {
		t.Fatal("Version() is empty")
	}
	if _, ok := l.Record(FileHeader); !ok {
		t.Fatal("Record(FileHeader) ok = false, want true")
	}
}
