package cnab

import "testing"

func TestNewCPFValid(t *testing.T) {
	cases := []string{"11144477735", "111.444.777-35"}
	for _, raw := range cases {
		cpf, err := NewCPF(raw)
		if err != nil {
			t.Fatalf("NewCPF(%q) error = %v", raw, err)
		}
		if cpf.Digits() != "11144477735" {
			t.Fatalf("Digits() = %q, want %q", cpf.Digits(), "11144477735")
		}
		if cpf.Kind() != "CPF" {
			t.Fatalf("Kind() = %q, want CPF", cpf.Kind())
		}
	}
}

func TestNewCPFInvalid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"wrong length", "123456"},
		{"repeated digits", "11111111111"},
		{"wrong check digit", "11144477736"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewCPF(c.raw); err == nil {
				t.Fatalf("NewCPF(%q) error = nil, want an error", c.raw)
			}
		})
	}
}

func TestNewCNPJValid(t *testing.T) {
	cases := []string{"11222333000181", "11.222.333/0001-81"}
	for _, raw := range cases {
		cnpj, err := NewCNPJ(raw)
		if err != nil {
			t.Fatalf("NewCNPJ(%q) error = %v", raw, err)
		}
		if cnpj.Digits() != "11222333000181" {
			t.Fatalf("Digits() = %q, want %q", cnpj.Digits(), "11222333000181")
		}
		if cnpj.Kind() != "CNPJ" {
			t.Fatalf("Kind() = %q, want CNPJ", cnpj.Kind())
		}
	}
}

func TestNewCNPJInvalid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"wrong length", "123456"},
		{"repeated digits", "11111111111111"},
		{"wrong check digit", "11222333000182"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewCNPJ(c.raw); err == nil {
				t.Fatalf("NewCNPJ(%q) error = nil, want an error", c.raw)
			}
		})
	}
}

func TestDocumentKind(t *testing.T) {
	cnpj, _ := NewCNPJ("11222333000181")
	cpf, _ := NewCPF("11144477735")

	if got := documentKind(cnpj); got != "2" {
		t.Fatalf("documentKind(CNPJ) = %q, want \"2\"", got)
	}
	if got := documentKind(cpf); got != "1" {
		t.Fatalf("documentKind(CPF) = %q, want \"1\"", got)
	}
	if got := documentKind(nil); got != "1" {
		t.Fatalf("documentKind(nil) = %q, want \"1\"", got)
	}
}
