package cnab

import "github.com/raykavin/gocnab/internal/engine"

// Sanitize prepares free-form text for a CNAB alphanumeric field: it
// uppercases it, replaces accented letters with their unaccented ASCII
// equivalent ("JOSÉ" becomes "JOSE", "AÇÃO" becomes "ACAO") and drops
// every remaining character CNAB does not accept ("A & B" becomes
// "A  B").
//
// Rendering never does this on its own: a character a field cannot carry
// is reported as a *FieldError naming the field, because for a document,
// a barcode, an account number or a PIX key, silently dropping part of
// the value would change who gets paid. So sanitizing is opt-in, and the
// caller decides which of its values are free-form text whose exact
// spelling does not matter — a payer or payee name, typically, which the
// bank only reproduces on a statement — and which are identifiers that
// must reach the bank exactly as given or not at all.
//
// Sanitize can return an empty (or blank) string, for a value made up
// entirely of characters CNAB rejects. Callers that require a value
// should check the result rather than pass it on: Company and Payee both
// reject a blank Name, which is the error to prefer over a file naming
// nobody.
func Sanitize(s string) string { return engine.Sanitize(s) }

// AllowedInField reports whether every character of s can be rendered
// into a CNAB alphanumeric field as it stands, i.e. whether Sanitize
// would leave it unchanged apart from case. Use it to validate an
// identifier that must not be rewritten — an agreement code, say — early,
// where the error can name the field the operator has to fix, rather than
// at render time.
func AllowedInField(s string) bool { return engine.AllowedInField(s) }
