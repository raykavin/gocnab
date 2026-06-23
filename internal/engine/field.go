package engine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/raykavin/gocnab/cnab/layout"
)

// renderField produces the exact Size() characters a FieldSpec occupies,
// resolving its value from either a fixed Const or a lookup in values by
// Key.
func renderField(f layout.FieldSpec, values layout.Values) (string, error) {
	raw := resolveValue(f, values)
	switch f.Kind {
	case layout.KindNumeric:
		return renderNumeric(f, raw)
	case layout.KindAlphanumeric:
		return renderAlphanumeric(f, raw)
	default:
		return "", &FieldRenderError{Field: f.Name, Reason: "unknown field kind"}
	}
}

func resolveValue(f layout.FieldSpec, values layout.Values) any {
	if f.Key == "" {
		return f.Const
	}
	if values == nil {
		return nil
	}
	return values[f.Key]
}

// renderNumeric right-aligns and zero-pads a numeric field. A value that
// does not fit in Size() columns is a hard error: numeric fields carry
// amounts, dates and identifiers that must never be silently truncated.
func renderNumeric(f layout.FieldSpec, raw any) (string, error) {
	digits, err := numericDigits(f, raw)
	if err != nil {
		return "", err
	}
	if len(digits) > f.Size() {
		return "", &FieldRenderError{
			Field:  f.Name,
			Reason: "value \"" + digits + "\" needs " + strconv.Itoa(len(digits)) + " digits but the field only has " + strconv.Itoa(f.Size()),
		}
	}
	return strings.Repeat("0", f.Size()-len(digits)) + digits, nil
}

// numericDigits converts raw into a plain digit string, honoring
// FieldSpec.Decimals for decimal-string inputs (e.g. "12.5" with
// Decimals=2 becomes "1250"). Integer inputs (of any width, including
// named types with an integer underlying type such as a Cents amount) are
// treated as already expressed in the field's minor unit and are not
// rescaled: this is what lets money move through the engine as a plain
// integer, without floating point ever touching an amount.
func numericDigits(f layout.FieldSpec, raw any) (string, error) {
	switch v := raw.(type) {
	case nil:
		return "0", nil
	case string:
		if v == "" {
			return "0", nil
		}
		if strings.Contains(v, ".") {
			return scaleDecimalString(f, v)
		}
		if !isDigitsOnly(v) {
			return "", &FieldRenderError{Field: f.Name, Reason: "value \"" + v + "\" is not numeric"}
		}
		return v, nil
	}

	if n, ok := intValue(raw); ok {
		if n < 0 {
			return "", &FieldRenderError{Field: f.Name, Reason: "negative values are not supported in numeric fields"}
		}
		return strconv.FormatInt(n, 10), nil
	}

	return "", &FieldRenderError{Field: f.Name, Reason: "unsupported numeric value type " + reflect.TypeOf(raw).String()}
}

// intValue extracts an int64 from any value whose underlying kind is an
// integer type, which covers plain int/int64 as well as named types such
// as a domain Cents amount.
func intValue(raw any) (int64, bool) {
	rv := reflect.ValueOf(raw)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	default:
		return 0, false
	}
}

func scaleDecimalString(f layout.FieldSpec, v string) (string, error) {
	parts := strings.SplitN(v, ".", 2)
	intPart, fracPart := parts[0], parts[1]
	if intPart == "" {
		intPart = "0"
	}
	if len(fracPart) > f.Decimals {
		return "", &FieldRenderError{
			Field:  f.Name,
			Reason: "value \"" + v + "\" has more decimal places than the field's " + strconv.Itoa(f.Decimals) + " implicit decimals",
		}
	}
	fracPart += strings.Repeat("0", f.Decimals-len(fracPart))
	digits := intPart + fracPart
	if !isDigitsOnly(digits) {
		return "", &FieldRenderError{Field: f.Name, Reason: "value \"" + v + "\" is not a valid decimal number"}
	}
	return digits, nil
}

func isDigitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// renderAlphanumeric left-aligns and space-pads an alphanumeric field.
// Values are uppercased automatically, but accented letters and any
// character outside the CNAB-allowed set are rejected rather than
// silently dropped; call Sanitize explicitly beforehand to clean up
// free-form text such as names. A value longer than Size() is truncated,
// matching how banks handle overlong names in practice.
func renderAlphanumeric(f layout.FieldSpec, raw any) (string, error) {
	s := strings.ToUpper(alphaString(raw))
	if err := validateCharset(s); err != nil {
		return "", &FieldRenderError{Field: f.Name, Reason: err.Error()}
	}
	if len(s) > f.Size() {
		s = s[:f.Size()]
	}
	return s + strings.Repeat(" ", f.Size()-len(s)), nil
}

func alphaString(raw any) string {
	switch v := raw.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
