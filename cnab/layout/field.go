// Package layout defines the contract between the generic CNAB 240 engine
// and a bank or product specific field layout. It has no dependency on the
// engine or on the public cnab package, so a bank layout can be implemented
// in a separate module without importing anything except this package.
package layout

// FieldKind identifies how a field is rendered: right-aligned zero-padded
// numeric digits, or left-aligned space-padded text.
type FieldKind int

const (
	// KindNumeric marks a field composed only of decimal digits ("9" in the
	// FEBRABAN picture notation). It is right-aligned and zero-padded.
	KindNumeric FieldKind = iota
	// KindAlphanumeric marks a free-form text field ("X" in the FEBRABAN
	// picture notation). It is left-aligned and space-padded.
	KindAlphanumeric
)

// String returns a human readable name for the field kind, used in error
// messages.
func (k FieldKind) String() string {
	switch k {
	case KindNumeric:
		return "numeric"
	case KindAlphanumeric:
		return "alphanumeric"
	default:
		return "unknown"
	}
}

// FieldSpec describes a single column range of a 240 character CNAB record.
//
// Start and End are 1-based and inclusive, matching the column numbering
// used in FEBRABAN manuals. Exactly one of Key or Const must be set: Key
// names the semantic value the field is filled with at render time (see the
// Key* constants in value.go), Const holds a literal value that never
// changes (record type markers, segment codes, layout versions, filler
// blanks).
type FieldSpec struct {
	// Name is a descriptive English identifier for documentation and error
	// messages, e.g. "PaymentAmount".
	Name string
	// Start is the 1-based inclusive starting column.
	Start int
	// End is the 1-based inclusive ending column.
	End int
	// Kind selects the fill and alignment rule.
	Kind FieldKind
	// Decimals is the number of implicit decimal places for numeric fields
	// following the "9(n)V(d)" notation. Zero means the field has no
	// implicit decimal point.
	Decimals int
	// Key is the semantic value key looked up in Values at render time.
	// Leave empty when Const is set.
	Key Key
	// Const is a fixed literal value rendered regardless of the supplied
	// Values. Leave empty when Key is set.
	Const string
}

// Size returns the number of columns the field occupies.
func (f FieldSpec) Size() int {
	return f.End - f.Start + 1
}

// IsConst reports whether the field always renders a fixed literal value.
func (f FieldSpec) IsConst() bool {
	return f.Key == "" && f.Const != ""
}
