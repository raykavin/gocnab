package layout

import "testing"

func TestFieldSpecSize(t *testing.T) {
	cases := []struct {
		name  string
		field FieldSpec
		want  int
	}{
		{"single column", FieldSpec{Start: 1, End: 1}, 1},
		{"typical amount field", FieldSpec{Start: 1, End: 15}, 15},
		{"full 240 record", FieldSpec{Start: 1, End: 240}, 240},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.field.Size(); got != c.want {
				t.Fatalf("Size() = %d, want %d", got, c.want)
			}
		})
	}
}

func TestFieldSpecIsConst(t *testing.T) {
	cases := []struct {
		name  string
		field FieldSpec
		want  bool
	}{
		{"dynamic field", FieldSpec{Key: KeyAmount}, false},
		{"literal const", FieldSpec{Const: "0"}, true},
		{"filler blank", FieldSpec{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.field.IsConst(); got != c.want {
				t.Fatalf("IsConst() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestFieldKindString(t *testing.T) {
	if KindNumeric.String() != "numeric" {
		t.Fatalf("KindNumeric.String() = %q", KindNumeric.String())
	}
	if KindAlphanumeric.String() != "alphanumeric" {
		t.Fatalf("KindAlphanumeric.String() = %q", KindAlphanumeric.String())
	}
	if FieldKind(99).String() != "unknown" {
		t.Fatalf("FieldKind(99).String() = %q", FieldKind(99).String())
	}
}
