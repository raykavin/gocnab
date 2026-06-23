package engine

import "testing"

func TestSanitize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain ascii", "ACME LTDA", "ACME LTDA"},
		{"lowercase becomes upper", "acme ltda", "ACME LTDA"},
		{"accented vowels", "João Ação Ímã Óculos Úlcera", "JOAO ACAO IMA OCULOS ULCERA"},
		{"cedilla and tilde n", "AÇÃO ESPANHOL NIÑO", "ACAO ESPANHOL NINO"},
		{"apostrophe and hash dropped, at-sign kept", "D'AVILA #1 @HOME", "DAVILA 1 @HOME"},
		{"allowed symbols kept", "RUA A, N. 10-B (FUNDOS)", "RUA A, N. 10-B (FUNDOS)"},
		{"email address kept intact", "fornecedor@exemplo.com", "FORNECEDOR@EXEMPLO.COM"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.input); got != c.want {
				t.Fatalf("Sanitize(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestValidateCharset(t *testing.T) {
	if err := validateCharset("ACME LTDA 123.,/-:()+*"); err != nil {
		t.Fatalf("validateCharset() error = %v, want nil", err)
	}
	if err := validateCharset("ACME LTDA #"); err == nil {
		t.Fatal("validateCharset() error = nil, want an error for '#'")
	}
	if err := validateCharset("JOÃO"); err == nil {
		t.Fatal("validateCharset() error = nil, want an error for an accented character")
	}
}

func TestIsAllowedRune(t *testing.T) {
	allowed := []rune{'A', 'Z', '0', '9', ' ', '.', ',', '/', '-', ':', '(', ')', '+', '*', '@', '_'}
	for _, r := range allowed {
		if !isAllowedRune(r) {
			t.Fatalf("isAllowedRune(%q) = false, want true", r)
		}
	}
	disallowed := []rune{'ã', 'Ã', '#', '%', '\''}
	for _, r := range disallowed {
		if isAllowedRune(r) {
			t.Fatalf("isAllowedRune(%q) = true, want false", r)
		}
	}
}
