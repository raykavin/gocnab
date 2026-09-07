package cnab

import (
	"strconv"
	"strings"
)

// Document is a Brazilian taxpayer registration number: either a CNPJ
// (company) or a CPF (individual). Payment.Registration and
// Payee.Registration fields accept any Document.
type Document interface {
	// Digits returns the registration number as 11 (CPF) or 14 (CNPJ)
	// decimal digits, with no punctuation.
	Digits() string
	// Kind returns "CPF" or "CNPJ".
	Kind() string
}

// CNPJ is a validated Brazilian company registration number. Since the
// Receita Federal's alphanumeric CNPJ rule (in effect since July 2026),
// this may hold 12 alphanumeric "root+order" characters followed by 2
// numeric check digits, not just 14 decimal digits — see NewCNPJ.
type CNPJ string

// NewCNPJ validates raw (punctuation is stripped automatically) and
// returns a CNPJ. It returns a *ValidationError when raw does not have 14
// characters, is a sequence of 14 repeated characters, or fails the
// modulo 11 check digit algorithm. raw may be the legacy all-numeric
// format or the Receita Federal alphanumeric format (12 alphanumeric
// "root+order" characters, uppercase letters or digits, followed by 2
// numeric check digits); either way it is never reduced to a numeric
// type, so a valid alphanumeric CNPJ is preserved exactly as issued.
func NewCNPJ(raw string) (CNPJ, error) {
	chars := alphanumericChars(raw)
	if len(chars) != 14 {
		return "", &ValidationError{Context: "CNPJ", Reason: "must have 14 characters, got " + strconv.Itoa(len(chars))}
	}
	if !validCNPJ(chars) {
		return "", &ValidationError{Context: "CNPJ", Reason: "invalid check digits for \"" + chars + "\""}
	}
	return CNPJ(chars), nil
}

// Digits returns the CNPJ as its 14 characters (digits, or — for an
// alphanumeric CNPJ — 12 alphanumeric characters followed by 2 decimal
// check digits).
func (c CNPJ) Digits() string { return string(c) }

// Kind returns "CNPJ".
func (c CNPJ) Kind() string { return "CNPJ" }

// CPF is a validated Brazilian individual registration number.
type CPF string

// NewCPF validates raw (punctuation is stripped automatically) and
// returns a CPF. It returns a *ValidationError when raw does not have 11
// digits, is a sequence of 11 repeated digits, or fails the standard
// modulo 11 check digit algorithm.
func NewCPF(raw string) (CPF, error) {
	digits := onlyDigits(raw)
	if len(digits) != 11 {
		return "", &ValidationError{Context: "CPF", Reason: "must have 11 digits, got " + strconv.Itoa(len(digits))}
	}
	if !validCPF(digits) {
		return "", &ValidationError{Context: "CPF", Reason: "invalid check digits for \"" + digits + "\""}
	}
	return CPF(digits), nil
}

// Digits returns the CPF as 11 decimal digits.
func (c CPF) Digits() string { return string(c) }

// Kind returns "CPF".
func (c CPF) Kind() string { return "CPF" }

// documentKind returns the FEBRABAN registration type digit for d: "1"
// for CPF, "2" for CNPJ.
func documentKind(d Document) string {
	if d != nil && d.Kind() == "CNPJ" {
		return "2"
	}
	return "1"
}

func onlyDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// alphanumericChars strips separator/punctuation characters, uppercasing
// letters, but — unlike onlyDigits — keeps A-Z intact: the normalization a
// Receita Federal alphanumeric CNPJ needs, since stripping its letters
// would silently corrupt it into a shorter, meaningless digit string.
func alphanumericChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - ('a' - 'A'))
		}
	}
	return b.String()
}

func allSameChar(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

// mod11CheckDigit computes a modulo-11 check digit. Each character's value
// is its ASCII code minus 48 ('0'): for a decimal digit ('0'-'9', codes
// 48-57) that is the digit's numeric value, unchanged from before; for an
// alphanumeric CNPJ's uppercase letters (codes 65-90) it is 17-42, the
// substitution Receita Federal defines for the alphanumeric CNPJ check
// digit algorithm. CPF is always plain digits, so this is a no-op change
// for it.
func mod11CheckDigit(chars string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += (int(chars[i]) - '0') * w
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

func validCPF(d string) bool {
	if len(d) != 11 || allSameChar(d) {
		return false
	}
	dv1 := mod11CheckDigit(d[:9], []int{10, 9, 8, 7, 6, 5, 4, 3, 2})
	if dv1 != int(d[9]-'0') {
		return false
	}
	dv2 := mod11CheckDigit(d[:10], []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})
	return dv2 == int(d[10]-'0')
}

// validCNPJ validates chars (14 characters: 12 alphanumeric "root+order"
// characters — digits for a legacy CNPJ, possibly uppercase letters for a
// Receita Federal alphanumeric CNPJ — followed by 2 numeric check digits)
// against the modulo 11 algorithm.
func validCNPJ(chars string) bool {
	if len(chars) != 14 || allSameChar(chars) {
		return false
	}
	// The 2 check digits themselves are always decimal, regardless of the
	// root+order characters preceding them.
	if !isDigitsOnly(chars[12:]) {
		return false
	}
	dv1 := mod11CheckDigit(chars[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	if dv1 != int(chars[12]-'0') {
		return false
	}
	dv2 := mod11CheckDigit(chars[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return dv2 == int(chars[13]-'0')
}

func isDigitsOnly(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
