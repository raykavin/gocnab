package engine

import (
	"strings"
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
)

// cents mimics a domain money type with an integer underlying type (like
// cnab.Cents) without importing the cnab package, keeping this test
// package dependency-free in the other direction.
type cents int64

func TestRenderFieldNumeric(t *testing.T) {
	cases := []struct {
		name  string
		field layout.FieldSpec
		value any
		want  string
	}{
		{
			name:  "const literal",
			field: layout.FieldSpec{Name: "RecordType", Start: 1, End: 1, Kind: layout.KindNumeric, Const: "0"},
			want:  "0",
		},
		{
			name:  "int value zero padded",
			field: layout.FieldSpec{Name: "Sequence", Start: 1, End: 5, Kind: layout.KindNumeric, Key: layout.KeySequence},
			value: 42,
			want:  "00042",
		},
		{
			name:  "int64 value exact fit",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 3, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: int64(999),
			want:  "999",
		},
		{
			name:  "named integer type (money-like)",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 10, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: cents(25200),
			want:  "0000025200",
		},
		{
			name:  "nil value renders as zero",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 4, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: nil,
			want:  "0000",
		},
		{
			name:  "digit string value",
			field: layout.FieldSpec{Name: "Document", Start: 1, End: 8, Kind: layout.KindNumeric, Key: layout.KeyPayeeDocument},
			value: "1234",
			want:  "00001234",
		},
		{
			name:  "decimal string scaled by implicit decimals",
			field: layout.FieldSpec{Name: "Rate", Start: 1, End: 6, Kind: layout.KindNumeric, Key: layout.KeyAmount, Decimals: 2},
			value: "12.5",
			want:  "001250",
		},
		{
			name:  "decimal string with no fractional digits",
			field: layout.FieldSpec{Name: "Rate", Start: 1, End: 6, Kind: layout.KindNumeric, Key: layout.KeyAmount, Decimals: 2},
			value: "12.",
			want:  "001200",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values := layout.Values{}
			if c.field.Key != "" {
				values[c.field.Key] = c.value
			}
			got, err := renderField(c.field, values)
			if err != nil {
				t.Fatalf("renderField() error = %v", err)
			}
			if got != c.want {
				t.Fatalf("renderField() = %q, want %q", got, c.want)
			}
			if len(got) != c.field.Size() {
				t.Fatalf("rendered length = %d, want %d", len(got), c.field.Size())
			}
		})
	}
}

func TestRenderFieldNumericErrors(t *testing.T) {
	cases := []struct {
		name  string
		field layout.FieldSpec
		value any
	}{
		{
			name:  "value exceeds field size",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 2, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: 12345,
		},
		{
			name:  "negative value",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 5, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: -1,
		},
		{
			name:  "non numeric string",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 5, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: "12a45",
		},
		{
			name:  "decimal string with too many fractional digits",
			field: layout.FieldSpec{Name: "Rate", Start: 1, End: 6, Kind: layout.KindNumeric, Key: layout.KeyAmount, Decimals: 2},
			value: "12.555",
		},
		{
			name:  "unsupported value type",
			field: layout.FieldSpec{Name: "Amount", Start: 1, End: 5, Kind: layout.KindNumeric, Key: layout.KeyAmount},
			value: 12.5,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values := layout.Values{c.field.Key: c.value}
			if _, err := renderField(c.field, values); err == nil {
				t.Fatal("renderField() error = nil, want an error")
			}
		})
	}
}

func TestRenderFieldAlphanumeric(t *testing.T) {
	cases := []struct {
		name  string
		field layout.FieldSpec
		value any
		want  string
	}{
		{
			name:  "left aligned space padded",
			field: layout.FieldSpec{Name: "Name", Start: 1, End: 10, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName},
			value: "ACME",
			want:  "ACME      ",
		},
		{
			name:  "uppercased automatically",
			field: layout.FieldSpec{Name: "Name", Start: 1, End: 6, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName},
			value: "acme",
			want:  "ACME  ",
		},
		{
			name:  "truncated when longer than the field",
			field: layout.FieldSpec{Name: "Name", Start: 1, End: 4, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName},
			value: "ACME LTDA",
			want:  "ACME",
		},
		{
			name:  "const literal segment code",
			field: layout.FieldSpec{Name: "SegmentCode", Start: 1, End: 1, Kind: layout.KindAlphanumeric, Const: "A"},
			want:  "A",
		},
		{
			name:  "filler blank field",
			field: layout.FieldSpec{Name: "Filler", Start: 1, End: 5, Kind: layout.KindAlphanumeric},
			want:  "     ",
		},
		{
			name:  "allowed symbols pass through",
			field: layout.FieldSpec{Name: "Address", Start: 1, End: 12, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeAddressStreet},
			value: "RUA A, N. 10",
			want:  "RUA A, N. 10",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values := layout.Values{}
			if c.field.Key != "" {
				values[c.field.Key] = c.value
			}
			got, err := renderField(c.field, values)
			if err != nil {
				t.Fatalf("renderField() error = %v", err)
			}
			if got != c.want {
				t.Fatalf("renderField() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestRenderFieldAlphanumericInvalidCharacter(t *testing.T) {
	field := layout.FieldSpec{Name: "Name", Start: 1, End: 10, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName}
	values := layout.Values{layout.KeyPayeeName: "JOÃO"}

	_, err := renderField(field, values)
	if err == nil {
		t.Fatal("renderField() error = nil, want an error for the accented character")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Fatalf("error %q does not mention the field name", err)
	}
}

func TestRenderFieldAlphanumericAfterSanitize(t *testing.T) {
	field := layout.FieldSpec{Name: "Name", Start: 1, End: 11, Kind: layout.KindAlphanumeric, Key: layout.KeyPayeeName}
	values := layout.Values{layout.KeyPayeeName: Sanitize("João D'Ávila")}

	got, err := renderField(field, values)
	if err != nil {
		t.Fatalf("renderField() error = %v", err)
	}
	if got != "JOAO DAVILA" {
		t.Fatalf("renderField() = %q, want %q", got, "JOAO DAVILA")
	}
}

func TestUnknownFieldKind(t *testing.T) {
	field := layout.FieldSpec{Name: "Broken", Start: 1, End: 5, Kind: layout.FieldKind(99)}
	if _, err := renderField(field, layout.Values{}); err == nil {
		t.Fatal("renderField() error = nil, want an error for an unknown field kind")
	}
}
