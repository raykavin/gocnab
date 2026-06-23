package engine

import (
	"sort"

	"github.com/raykavin/gocnab/cnab/layout"
)

// recordWidth is the fixed width of every CNAB 240 line.
const recordWidth = 240

// compiledRecord is a RecordSpec whose fields have been validated to
// cover columns 1-240 with no gap and no overlap, and sorted by Start so
// rendering can proceed left to right.
type compiledRecord struct {
	name   string
	fields []layout.FieldSpec
}

// compileRecord validates spec and returns a compiledRecord ready to
// render lines. It is called once per record kind when an Engine is
// built, not on every render, so a malformed layout is caught at
// start-up rather than while generating a file.
func compileRecord(spec layout.RecordSpec) (*compiledRecord, error) {
	if err := spec.Validate(); err != nil {
		return nil, &SpecError{Record: spec.Name, Reason: err.Error()}
	}

	fields := make([]layout.FieldSpec, len(spec.Fields))
	copy(fields, spec.Fields)
	sort.Slice(fields, func(i, j int) bool { return fields[i].Start < fields[j].Start })

	return &compiledRecord{name: spec.Name, fields: fields}, nil
}

// render fills every field of the record from values and returns the
// resulting 240 character line.
func (r *compiledRecord) render(values layout.Values) (string, error) {
	var buf [recordWidth]byte
	for _, f := range r.fields {
		rendered, err := renderField(f, values)
		if err != nil {
			return "", err
		}
		copy(buf[f.Start-1:f.End], rendered)
	}
	return string(buf[:]), nil
}
