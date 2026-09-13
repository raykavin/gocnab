package cnab

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestSanitizeHandlesRealWorldNames covers the values that actually reach
// this function: Brazilian company and person names, which routinely carry
// accents, "&" and apostrophes — none of them renderable into a CNAB
// alphanumeric field as they stand.
func TestSanitizeHandlesRealWorldNames(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"accented letters fold to ASCII", "José Antônio Gonçalves", "JOSE ANTONIO GONCALVES"},
		{"cedilla and tilde", "Ação Serviços Ltda", "ACAO SERVICOS LTDA"},
		{"ampersand is dropped", "A & B Comércio", "A  B COMERCIO"},
		{"apostrophe is dropped", "O'Brien Ltda", "OBRIEN LTDA"},
		{"already valid text only changes case", "posto 24 horas - ltda", "POSTO 24 HORAS - LTDA"},
		{"allowed symbols survive", "EMPRESA (MATRIZ) 1.234/56-7", "EMPRESA (MATRIZ) 1.234/56-7"},
		{"e-mail characters survive", "fornecedor_1@exemplo.com", "FORNECEDOR_1@EXEMPLO.COM"},
		{"empty stays empty", "", ""},
		{"a value made only of rejected characters empties out", "###", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in); got != c.want {
				t.Fatalf("Sanitize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSanitizeOutputIsRenderable is the property that matters: whatever
// Sanitize returns must pass the very validation that rejected the input,
// so a caller never has to sanitize twice or guess.
func TestSanitizeOutputIsRenderable(t *testing.T) {
	inputs := []string{
		"José Antônio Gonçalves",
		"A & B Comércio #1",
		"O'Brien & Sons, Ltda.",
		"AÇÃO ~ SERVIÇOS ª º",
		"tabs\tand\nnewlines",
		"emoji 🙂 name",
	}

	for _, in := range inputs {
		sanitized := Sanitize(in)
		if !AllowedInField(sanitized) {
			t.Errorf("Sanitize(%q) = %q, which is still not renderable", in, sanitized)
		}
	}
}

func TestAllowedInField(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"upper case letters and digits", "EMPRESA 123", true},
		{"lower case is not a reason to reject", "empresa 123", true},
		{"allowed symbols", "1.234,56/78-9:()+*@_", true},
		{"empty", "", true},
		{"ampersand", "A & B", false},
		{"accent", "JOSÉ", false},
		{"hash", "EMPRESA #1", false},
		{"apostrophe", "O'BRIEN", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := AllowedInField(c.in); got != c.want {
				t.Fatalf("AllowedInField(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// TestRenderRejectsUnsanitizedNameNamingTheRemedy pins both halves of the
// contract a caller depends on: rendering does NOT sanitize on its own (a
// dropped character in a document or a barcode would change who gets
// paid), and the error it returns points at the field to fix and at an
// exported function the caller can actually call — engine.Sanitize is
// unreachable from outside this module.
func TestRenderRejectsUnsanitizedNameNamingTheRemedy(t *testing.T) {
	cfg := validConfig()
	cfg.Company.Name = "A & B COMERCIO"

	f, err := NewRemittance(cfg)
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := batch.AddPayment(Pix{
		Key:    EmailKey("fornecedor@exemplo.com"),
		Payee:  validPayee(),
		Amount: 25200,
		Date:   time.Now().AddDate(0, 0, 1),
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	_, err = f.Generate()
	if err == nil {
		t.Fatal("Generate() error = nil, want a charset error for an unsanitized company name")
	}

	var fieldErr *FieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("Generate() error = %T (%v), want *FieldError", err, err)
	}
	if fieldErr.Field == "" {
		t.Error("the charset error does not name the field to fix")
	}
	if !strings.Contains(err.Error(), "cnab.Sanitize") {
		t.Errorf("error = %v, want it to name the exported cnab.Sanitize", err)
	}
	if strings.Contains(err.Error(), "engine.Sanitize") {
		t.Errorf("error = %v, want it not to name engine.Sanitize, which callers cannot reach", err)
	}

	// The same file generates once the name has been through Sanitize,
	// which is the whole point of the remedy the error names.
	cfg.Company.Name = Sanitize(cfg.Company.Name)
	sane, err := NewRemittance(cfg)
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	saneBatch, err := sane.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := saneBatch.AddPayment(Pix{
		Key:    EmailKey("fornecedor@exemplo.com"),
		Payee:  validPayee(),
		Amount: 25200,
		Date:   time.Now().AddDate(0, 0, 1),
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}
	if _, err := sane.Generate(); err != nil {
		t.Fatalf("Generate() error = %v after sanitizing the name", err)
	}
}
