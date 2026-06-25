package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// jsonLayoutFile is the on-disk shape a bank layout descriptor is parsed
// from. Its "records" map is keyed by the RecordKey string value (e.g.
// "file_header", "segment_a"), so it can be copied verbatim from the
// RecordKey constants documented in this package.
type jsonLayoutFile struct {
	Name    string                    `json:"name"`
	Version string                    `json:"version"`
	Records map[string]jsonRecordSpec `json:"records"`
}

type jsonRecordSpec struct {
	Name   string          `json:"name,omitempty"`
	Fields []jsonFieldSpec `json:"fields"`
}

type jsonFieldSpec struct {
	Name     string `json:"name,omitempty"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Kind     string `json:"kind"`
	Decimals int    `json:"decimals,omitempty"`
	Key      string `json:"key,omitempty"`
	Const    string `json:"const,omitempty"`
}

// jsonLayout is the Layout implementation NewFromJSON builds.
type jsonLayout struct {
	name    string
	version string
	records map[RecordKey]RecordSpec
}

func (l *jsonLayout) Name() string    { return l.name }
func (l *jsonLayout) Version() string { return l.version }
func (l *jsonLayout) Record(key RecordKey) (RecordSpec, bool) {
	spec, ok := l.records[key]
	return spec, ok
}

// NewFromJSON parses a bank layout descriptor from JSON and returns a
// ready to use Layout. It does not register the result; call Register
// (or cnab.RegisterLayout) explicitly once it succeeds.
//
// Every structural problem is reported here, pinpointing which record
// and field caused it, instead of surfacing later as a generic
// engine.New or Register error:
//
//   - "name" and "version" must be present.
//   - each key of "records" must be one of the RecordKey values in
//     AllRecordKeys (e.g. "file_header", "segment_a").
//   - each field's "kind" must be "9"/"numeric" or "X"/"alphanumeric".
//   - each field's "start"/"end" must describe a valid 1-240 column
//     range (1 <= start <= end <= 240).
//   - a field sets at most one of "key" or "const", never both.
//   - a field's "key", when set, must be one of the values in AllKeys.
//   - every record's fields must cover columns 1-240 with no gap and no
//     overlap (the same check RecordSpec.Validate and the engine run).
//
// Example descriptor (abbreviated):
//
//	{
//	  "name": "meubanco240",
//	  "version": "081",
//	  "records": {
//	    "file_header": {
//	      "fields": [
//	        {"name": "BankCode", "start": 1, "end": 3, "kind": "9", "const": "341"},
//	        {"name": "ServiceBatchNumber", "start": 4, "end": 7, "kind": "9", "const": "0000"},
//	        {"name": "RecordType", "start": 8, "end": 8, "kind": "9", "const": "0"},
//	        {"name": "CompanyName", "start": 73, "end": 102, "kind": "X", "key": "company_name"}
//	      ]
//	    }
//	  }
//	}
func NewFromJSON(data []byte) (Layout, error) {
	var file jsonLayoutFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("cnab/layout: invalid JSON layout: %w", err)
	}

	if strings.TrimSpace(file.Name) == "" {
		return nil, fmt.Errorf("cnab/layout: JSON layout is missing \"name\"")
	}
	if strings.TrimSpace(file.Version) == "" {
		return nil, fmt.Errorf("cnab/layout: JSON layout %q is missing \"version\"", file.Name)
	}
	if len(file.Records) == 0 {
		return nil, fmt.Errorf("cnab/layout: JSON layout %q has no \"records\"", file.Name)
	}

	records := make(map[RecordKey]RecordSpec, len(file.Records))
	for rawKey, rawSpec := range file.Records {
		key := RecordKey(rawKey)
		if !IsKnownRecordKey(key) {
			return nil, fmt.Errorf(
				"cnab/layout: JSON layout %q: unknown record key %q (valid keys: %s)",
				file.Name, rawKey, joinRecordKeys(AllRecordKeys),
			)
		}

		spec, err := parseJSONRecordSpec(rawKey, rawSpec)
		if err != nil {
			return nil, fmt.Errorf("cnab/layout: JSON layout %q: %w", file.Name, err)
		}
		records[key] = spec
	}

	return &jsonLayout{name: file.Name, version: file.Version, records: records}, nil
}

// NewFromJSONFile reads path and calls NewFromJSON on its contents.
func NewFromJSONFile(path string) (Layout, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cnab/layout: reading %s: %w", path, err)
	}
	l, err := NewFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return l, nil
}

func parseJSONRecordSpec(recordKey string, raw jsonRecordSpec) (RecordSpec, error) {
	name := raw.Name
	if name == "" {
		name = recordKey
	}
	if len(raw.Fields) == 0 {
		return RecordSpec{}, fmt.Errorf("record %q has no \"fields\"", recordKey)
	}

	fields := make([]FieldSpec, len(raw.Fields))
	for i, rf := range raw.Fields {
		f, err := parseJSONFieldSpec(rf)
		if err != nil {
			return RecordSpec{}, fmt.Errorf("record %q, field #%d: %w", recordKey, i+1, err)
		}
		fields[i] = f
	}

	spec := RecordSpec{Name: name, Fields: fields}
	if err := spec.Validate(); err != nil {
		return RecordSpec{}, fmt.Errorf("record %q: %w", recordKey, err)
	}
	return spec, nil
}

func parseJSONFieldSpec(rf jsonFieldSpec) (FieldSpec, error) {
	kind, err := parseJSONFieldKind(rf.Kind)
	if err != nil {
		return FieldSpec{}, err
	}
	if rf.Start < 1 || rf.End < rf.Start || rf.End > 240 {
		return FieldSpec{}, fmt.Errorf("invalid column range %d-%d (want 1 <= start <= end <= 240)", rf.Start, rf.End)
	}
	if rf.Key != "" && rf.Const != "" {
		return FieldSpec{}, fmt.Errorf("field %q at %d-%d sets both \"key\" and \"const\"; use only one", fieldLabel(rf), rf.Start, rf.End)
	}
	key := Key(rf.Key)
	if key != "" && !IsKnownKey(key) {
		return FieldSpec{}, fmt.Errorf("field %q at %d-%d: unknown key %q", fieldLabel(rf), rf.Start, rf.End, rf.Key)
	}
	if rf.Decimals < 0 {
		return FieldSpec{}, fmt.Errorf("field %q at %d-%d: \"decimals\" must not be negative", fieldLabel(rf), rf.Start, rf.End)
	}

	return FieldSpec{
		Name:     fieldLabel(rf),
		Start:    rf.Start,
		End:      rf.End,
		Kind:     kind,
		Decimals: rf.Decimals,
		Key:      key,
		Const:    rf.Const,
	}, nil
}

func fieldLabel(rf jsonFieldSpec) string {
	if rf.Name != "" {
		return rf.Name
	}
	return fmt.Sprintf("Field_%d_%d", rf.Start, rf.End)
}

func parseJSONFieldKind(s string) (FieldKind, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "9", "NUMERIC":
		return KindNumeric, nil
	case "X", "ALPHANUMERIC":
		return KindAlphanumeric, nil
	default:
		return 0, fmt.Errorf("invalid field kind %q (want \"9\"/\"numeric\" or \"X\"/\"alphanumeric\")", s)
	}
}

func joinRecordKeys(keys []RecordKey) string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = string(k)
	}
	return strings.Join(names, ", ")
}
