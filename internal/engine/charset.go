package engine

import "strings"

// allowedSymbols lists the non-alphanumeric characters accepted in
// alphanumeric fields, besides the space character. This starts from the
// conservative classic CNAB set and adds "@" and "_", without which a
// PIX e-mail key (a required feature of this SDK) could never pass
// validation; extend it further if a bank layout needs another symbol.
const allowedSymbols = ".,/-:()+*@_"

// accentReplacements maps common Portuguese/Spanish accented letters to
// their unaccented ASCII equivalent, covering the characters most likely
// to appear in Brazilian company and payee names.
var accentReplacements = map[rune]rune{
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A',
	'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'Ç': 'C', 'Ñ': 'N',
}

// Sanitize uppercases s, replaces accented letters with their unaccented
// equivalent and drops every rune that is still not allowed in a CNAB
// alphanumeric field afterwards. Sanitization is an explicit, opt-in step:
// rendering a field never sanitizes silently, so that an invalid
// character reaching a field is always reported as an error rather than
// quietly discarded.
func Sanitize(s string) string {
	s = strings.ToUpper(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if replacement, ok := accentReplacements[r]; ok {
			r = replacement
		}
		if isAllowedRune(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isAllowedRune(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r == ' ':
		return true
	default:
		return strings.ContainsRune(allowedSymbols, r)
	}
}

func validateCharset(s string) error {
	for _, r := range s {
		if !isAllowedRune(r) {
			return &charsetError{rune: r}
		}
	}
	return nil
}

type charsetError struct {
	rune rune
}

func (e *charsetError) Error() string {
	return "character " + string(e.rune) + " is not allowed in a CNAB field; call engine.Sanitize (or cnab.Sanitize) first if this is expected input"
}
