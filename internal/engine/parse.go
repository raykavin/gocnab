package engine

import (
	"strconv"
	"strings"

	"github.com/raykavin/gocnab/cnab/layout"
)

// parseField extracts f's column range from line and decodes it, the exact
// inverse of renderField: a numeric field yields its raw zero-padded digit
// string unchanged (renderNumeric produced exactly Size() digit characters,
// so slicing them back out needs no further decoding the digit string
// already is the value, leading zeros and all; a caller that wants an
// integer parses it with strconv, a caller that wants a date parses it
// with the same "DDMMYYYY" layout used to render it), and an alphanumeric
// field yields its content with the trailing space padding trimmed.
func parseField(f layout.FieldSpec, line string) (string, error) {
	raw := line[f.Start-1 : f.End]
	switch f.Kind {
	case layout.KindNumeric:
		if !isDigitsOnly(raw) {
			return "", &FieldParseError{Field: f.Name, Reason: "value \"" + raw + "\" is not numeric"}
		}
		return raw, nil
	case layout.KindAlphanumeric:
		return strings.TrimRight(raw, " "), nil
	default:
		return "", &FieldParseError{Field: f.Name, Reason: "unknown field kind"}
	}
}

// parse decodes a 240 character line into a Values map, the inverse of
// render. Const fields are not exposed as data (they carry no Key to store
// a value under) but their column content is checked against the literal
// value the layout expects there; a mismatch means either a corrupt line or
// a line that does not actually belong to this record kind, and is reported
// as a *RecordMismatchError rather than silently ignored.
func (r *compiledRecord) parse(line string) (layout.Values, error) {
	if len(line) != recordWidth {
		return nil, &RecordParseError{Record: r.name, Reason: "line has " + strconv.Itoa(len(line)) + " characters, want " + strconv.Itoa(recordWidth)}
	}

	values := make(layout.Values, len(r.fields))
	for _, f := range r.fields {
		raw, err := parseField(f, line)
		if err != nil {
			return nil, err
		}
		if f.Key == "" {
			if f.IsConst() {
				// Compare against what rendering this same const field
				// produces (zero-padded for numeric, space-padded for
				// alphanumeric) rather than re-deriving that padding rule
				// here a second time: renderField already knows it, and a
				// const field ignores values entirely (see resolveValue),
				// so passing nil is safe.
				expected, err := renderField(f, nil)
				if err != nil {
					return nil, err
				}
				if raw != expected {
					return nil, &RecordMismatchError{Record: r.name, Field: f.Name, Expected: expected, Got: raw}
				}
			}
			continue
		}
		values[f.Key] = raw
	}
	return values, nil
}
