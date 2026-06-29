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

// CNPJ is a validated Brazilian company registration number.
type CNPJ string

// NewCNPJ validates raw (punctuation is stripped automatically) and
// returns a CNPJ. It returns a *ValidationError when raw does not have 14
// digits, is a sequence of 14 repeated digits, or fails the standard
// modulo 11 check digit algorithm.
func NewCNPJ(raw string) (CNPJ, error) {
	digits := onlyDigits(raw)
	if len(digits) != 14 {
		return "", &ValidationError{Context: "CNPJ", Reason: "must have 14 digits, got " + strconv.Itoa(len(digits))}
	}
	if !validCNPJ(digits) {
		return "", &ValidationError{Context: "CNPJ", Reason: "invalid check digits for \"" + digits + "\""}
	}
	return CNPJ(digits), nil
}

// Digits returns the CNPJ as 14 decimal digits.
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

func allSameDigit(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func mod11CheckDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

func validCPF(d string) bool {
	if len(d) != 11 || allSameDigit(d) {
		return false
	}
	dv1 := mod11CheckDigit(d[:9], []int{10, 9, 8, 7, 6, 5, 4, 3, 2})
	if dv1 != int(d[9]-'0') {
		return false
	}
	dv2 := mod11CheckDigit(d[:10], []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})
	return dv2 == int(d[10]-'0')
}

func validCNPJ(d string) bool {
	if len(d) != 14 || allSameDigit(d) {
		return false
	}
	dv1 := mod11CheckDigit(d[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	if dv1 != int(d[12]-'0') {
		return false
	}
	dv2 := mod11CheckDigit(d[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return dv2 == int(d[13]-'0')
}
