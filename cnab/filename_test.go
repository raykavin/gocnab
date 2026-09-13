package cnab

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// day27 is the reference instant every date-dependent case below renders,
// chosen so the day (27) differs from the month (11) and from both year
// forms: a field wired to the wrong part of the date cannot pass by
// coincidence.
var day27 = time.Date(2026, 11, 27, 14, 30, 0, 0, time.UTC)

func sicrediInput() FileNameInput {
	return FileNameInput{
		AgreementCode: "6CBY",
		Branch:        "0804",
		AccountNumber: "55390",
		LayoutName:    "sicredi240",
		LayoutVersion: "082",
		NSA:           1,
		Now:           day27,
	}
}

// TestFileNameLayoutSicrediConvention is the convention this repository
// actually ships: agreement code, day of the month, two digit sequence.
func TestFileNameLayoutSicrediConvention(t *testing.T) {
	l := NewFileNameLayout(
		AgreementCode.Width(4),
		CurrentDay,
		NSASequence.Width(2),
	).WithExtension(".REM").WithMaxLength(12)

	name, err := l.Render(sicrediInput())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if name != "6CBY2701.REM" {
		t.Fatalf("Render() = %q, want %q", name, "6CBY2701.REM")
	}
}

// TestFileNameLayoutZeroPadsAndAligns covers the formatting rules each
// field kind defaults to: a number pads on the left (so a sequence keeps
// its magnitude), text pads on the right.
func TestFileNameLayoutZeroPadsAndAligns(t *testing.T) {
	cases := []struct {
		name  string
		field FileNameField
		nsa   int
		want  string
	}{
		{"numeric zero pads on the left", NSASequence.Width(6), 42, "000042"},
		{"numeric at exact width", NSASequence.Width(2), 42, "42"},
		{"text pads on the right", AgreementCode.Width(6), 1, "6CBY00"},
		{"text pads on the left when asked", AgreementCode.Width(6).PadLeft(), 1, "006CBY"},
		{"explicit padding rune", NSASequence.Width(4).Pad('-'), 7, "---7"},
		{"natural width when unset", NSASequence, 12345, "12345"},
		{"negative width is natural width", NSASequence.Width(-3), 8, "8"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := sicrediInput()
			in.NSA = c.nsa
			name, err := NewFileNameLayout(c.field).WithExtension(ExtensionNone).Render(in)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if name != c.want {
				t.Fatalf("Render() = %q, want %q", name, c.want)
			}
		})
	}
}

// TestFileNameLayoutDateFields checks every predefined date field draws
// the part of the instant it is named for.
func TestFileNameLayoutDateFields(t *testing.T) {
	cases := []struct {
		name  string
		field FileNameField
		want  string
	}{
		{"day", CurrentDay, "27"},
		{"month", CurrentMonth, "11"},
		{"year", CurrentYear, "2026"},
		{"short year", CurrentShortYear, "26"},
		{"full date", CurrentDate, "20261127"},
		{"custom date layout", Date("0201"), "2711"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, err := NewFileNameLayout(c.field).WithExtension(ExtensionNone).Render(sicrediInput())
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if name != c.want {
				t.Fatalf("Render() = %q, want %q", name, c.want)
			}
		})
	}
}

// TestFileNameLayoutResolvesOneInstant guards the rule that a name's date
// fields all describe the same instant: a file named as the day turns must
// not carry one field from before midnight and another from after.
func TestFileNameLayoutResolvesOneInstant(t *testing.T) {
	l := NewFileNameLayout(CurrentDate, Literal("-"), CurrentDay).WithExtension(ExtensionNone)

	in := sicrediInput()
	in.Now = time.Time{} // unset: Render resolves time.Now() once

	name, err := l.Render(in)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	// The last two characters (CurrentDay) must repeat the date's own day.
	if day := name[6:8]; day != name[len(name)-2:] {
		t.Fatalf("Render() = %q: date fields disagree on the day (%q vs %q)", name, day, name[len(name)-2:])
	}
}

// TestFileNameLayoutLiteralAndDynamicFieldsMix is the "future bank"
// case: a convention this package was never written for, assembled from
// literals, account data and a bank-specific value, with no code change.
func TestFileNameLayoutLiteralAndDynamicFieldsMix(t *testing.T) {
	l := NewFileNameLayout(
		Literal("PG"),
		BranchNumber.Width(5),
		Literal("-"),
		AccountNumber.Width(8),
		Literal("-"),
		CurrentDate,
		Custom("product").Upper(),
		NSASequence.Width(3),
	).WithExtension(".TXT")

	in := sicrediInput()
	in.Values = map[string]string{"product": "sup"}
	in.NSA = 7

	name, err := l.Render(in)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if want := "PG00804-00055390-20261127SUP007.TXT"; name != want {
		t.Fatalf("Render() = %q, want %q", name, want)
	}
}

// TestFileNameLayoutCustomFieldRequiresValue keeps a missing bank-specific
// value from silently rendering a gap in the name.
func TestFileNameLayoutCustomFieldRequiresValue(t *testing.T) {
	l := NewFileNameLayout(AgreementCode, Custom("daily_sequence"))

	_, err := l.Render(sicrediInput())
	if err == nil {
		t.Fatal("Render() error = nil, want an error for a custom field with no value")
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Render() error = %T, want *ValidationError", err)
	}
	if !strings.Contains(err.Error(), "daily_sequence") {
		t.Fatalf("Render() error = %v, want it to name the missing field", err)
	}
}

// TestFileNameLayoutRejectsOverflowByDefault is the guard that matters
// most for a sequence: a name whose NSA was silently trimmed would name a
// file the bank has already received.
func TestFileNameLayoutRejectsOverflowByDefault(t *testing.T) {
	in := sicrediInput()
	in.NSA = 100

	_, err := NewFileNameLayout(AgreementCode.Width(4), CurrentDay, NSASequence.Width(2)).Render(in)
	if err == nil {
		t.Fatal("Render() error = nil, want an error for an NSA that does not fit its width")
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Render() error = %T, want *ValidationError", err)
	}
}

// TestFileNameLayoutTruncateIsOptIn checks the escape hatch trims from the
// side each alignment implies.
func TestFileNameLayoutTruncateIsOptIn(t *testing.T) {
	in := sicrediInput()
	in.NSA = 1234
	in.AgreementCode = "CONVENIO"

	cases := []struct {
		name  string
		field FileNameField
		want  string
	}{
		{"numeric keeps the least significant digits", NSASequence.Width(2).Truncate(), "34"},
		{"text keeps the leading characters", AgreementCode.Width(4).Truncate(), "CONV"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, err := NewFileNameLayout(c.field).WithExtension(ExtensionNone).Render(in)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if name != c.want {
				t.Fatalf("Render() = %q, want %q", name, c.want)
			}
		})
	}
}

// TestFileNameLayoutRejectsInvalidCharacters covers the character
// validation: a space or an accent coming out of an agreement code must
// fail here, not at the bank's transmission client.
func TestFileNameLayoutRejectsInvalidCharacters(t *testing.T) {
	cases := []struct {
		name          string
		agreementCode string
	}{
		{"space", "6C BY"},
		{"accent", "6CBÝ"},
		{"path separator", "6C/BY"},
		{"dot", "6C.BY"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := sicrediInput()
			in.AgreementCode = c.agreementCode

			_, err := NewFileNameLayout(AgreementCode).Render(in)
			if err == nil {
				t.Fatalf("Render() error = nil for agreement code %q, want an error", c.agreementCode)
			}
			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("Render() error = %T, want *ValidationError", err)
			}
		})
	}
}

// TestFileNameLayoutEnforcesMaxLength covers the length validation a bank
// that documents an exact name width relies on.
func TestFileNameLayoutEnforcesMaxLength(t *testing.T) {
	l := NewFileNameLayout(AgreementCode, CurrentDate, NSASequence.Width(6)).WithMaxLength(12)

	_, err := l.Render(sicrediInput())
	if err == nil {
		t.Fatal("Render() error = nil, want an error for a name over MaxLength")
	}
	if !strings.Contains(err.Error(), "over the 12 allowed") {
		t.Fatalf("Render() error = %v, want it to report the length cap", err)
	}
}

// TestFileNameLayoutExtension covers the three extension states: default,
// explicit, and none.
func TestFileNameLayoutExtension(t *testing.T) {
	cases := []struct {
		name string
		ext  string
		want string
	}{
		{"defaults to .REM", "", "6CBY.REM"},
		{"explicit", ".TXT", "6CBY.TXT"},
		{"none", ExtensionNone, "6CBY"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, err := NewFileNameLayout(AgreementCode).WithExtension(c.ext).Render(sicrediInput())
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if name != c.want {
				t.Fatalf("Render() = %q, want %q", name, c.want)
			}
		})
	}
}

func TestFileNameLayoutValidate(t *testing.T) {
	cases := []struct {
		name   string
		layout FileNameLayout
	}{
		{"no field", FileNameLayout{}},
		{"extension without a dot", NewFileNameLayout(AgreementCode).WithExtension("REM")},
		{"extension with only a dot", NewFileNameLayout(AgreementCode).WithExtension(".")},
		{"extension with a separator", NewFileNameLayout(AgreementCode).WithExtension(".re/m")},
		{"negative max length", NewFileNameLayout(AgreementCode).WithMaxLength(-1)},
		{"max length over the hard cap", NewFileNameLayout(AgreementCode).WithMaxLength(maxFileNameLength + 1)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.layout.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want an error")
			}
			if _, err := c.layout.Render(sicrediInput()); err == nil {
				t.Fatal("Render() error = nil, want an error")
			}
		})
	}
}

// TestFileNameFieldsAreImmutable guards the promise that the exported
// prototypes are safe to share: formatting one for a layout must not
// change what another layout renders.
func TestFileNameFieldsAreImmutable(t *testing.T) {
	narrow := NSASequence.Width(2)
	wide := NSASequence.Width(6)

	in := sicrediInput()
	in.NSA = 3

	got, err := NewFileNameLayout(narrow).WithExtension(ExtensionNone).Render(in)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got != "03" {
		t.Fatalf("narrow field rendered %q, want %q", got, "03")
	}

	got, err = NewFileNameLayout(wide).WithExtension(ExtensionNone).Render(in)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got != "000003" {
		t.Fatalf("wide field rendered %q, want %q", got, "000003")
	}

	if got, err := NewFileNameLayout(NSASequence).WithExtension(ExtensionNone).Render(in); err != nil || got != "3" {
		t.Fatalf("prototype rendered %q (err %v), want %q — a modifier mutated it", got, err, "3")
	}
}

// TestSetRemittanceFileNameLayout covers the process wide default: it is
// installed only when valid, applies to a File that configures nothing,
// and can be cleared back to the built-in fallback.
func TestSetRemittanceFileNameLayout(t *testing.T) {
	t.Cleanup(func() { _ = SetRemittanceFileNameLayout() })

	if err := SetRemittanceFileNameLayout(AgreementCode.Width(4), CurrentDay, NSASequence.Width(2)); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}

	f := fileWithOneBatch(t, validConfig())
	name, err := f.FileName()
	if err != nil {
		t.Fatalf("FileName() error = %v", err)
	}
	// validConfig's agreement code is shorter than four characters, so the
	// day and sequence sit at the end of a padded stem.
	if !strings.HasSuffix(name, "01.REM") || len(name) != 12 {
		t.Fatalf("FileName() = %q, want a 12 character name ending in the NSA and .REM", name)
	}

	// An invalid layout leaves the installed one untouched.
	if err := SetRemittanceFileNameLayout(NewFileNameLayout(AgreementCode).WithExtension("REM").Fields...); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}
	if RemittanceFileNameLayout().IsZero() {
		t.Fatal("RemittanceFileNameLayout() is zero after installing a valid layout")
	}

	if err := SetRemittanceFileNameLayout(); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}
	if !RemittanceFileNameLayout().IsZero() {
		t.Fatal("SetRemittanceFileNameLayout() with no field did not clear the default")
	}
}

// TestSetRemittanceFileNameLayoutRejectsInvalidLayout checks a malformed
// convention is refused where it is declared, leaving the previously
// installed one in place, rather than failing when a file is finally
// named.
func TestSetRemittanceFileNameLayoutRejectsInvalidLayout(t *testing.T) {
	t.Cleanup(func() { _ = SetRemittanceFileNameLayout() })

	if err := SetRemittanceFileNameLayout(Literal("GOOD")); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}

	cases := []struct {
		name  string
		field FileNameField
	}{
		{"unknown source", FileNameField{source: fileNameSource(-1)}},
		{"empty literal", Literal("")},
		{"empty date layout", Date("")},
		{"unnamed custom value", Custom("")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := SetRemittanceFileNameLayout(c.field)
			if err == nil {
				t.Fatal("SetRemittanceFileNameLayout() error = nil, want an error")
			}
			if got := RemittanceFileNameLayout(); got.IsZero() || len(got.Fields) != 1 || got.Fields[0].literal != "GOOD" {
				t.Fatalf("a rejected layout replaced the installed one: %+v", got)
			}
		})
	}
}

// TestConfigFileNameLayoutOverridesDefault is the multi-bank rule: a
// process serving two banks sets the convention per file, and the file's
// own layout wins over whatever global happens to be installed.
func TestConfigFileNameLayoutOverridesDefault(t *testing.T) {
	t.Cleanup(func() { _ = SetRemittanceFileNameLayout() })

	if err := SetRemittanceFileNameLayout(Literal("GLOBAL")); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}

	cfg := validConfig()
	cfg.FileNameLayout = NewFileNameLayout(Literal("PERFILE"), NSASequence.Width(2))
	f := fileWithOneBatch(t, cfg)

	name, err := f.FileName()
	if err != nil {
		t.Fatalf("FileName() error = %v", err)
	}
	if name != "PERFILE01.REM" {
		t.Fatalf("FileName() = %q, want the file's own layout to win", name)
	}
}

// TestFileNameFallsBackToLegacyName pins the compatibility promise: with
// no layout configured anywhere, an existing integration keeps getting the
// name File.FileName has always produced.
func TestFileNameFallsBackToLegacyName(t *testing.T) {
	t.Cleanup(func() { _ = SetRemittanceFileNameLayout() })
	if err := SetRemittanceFileNameLayout(); err != nil {
		t.Fatalf("SetRemittanceFileNameLayout() error = %v", err)
	}

	f := fileWithOneBatch(t, validConfig())
	name, err := f.FileName()
	if err != nil {
		t.Fatalf("FileName() error = %v", err)
	}
	want := "FEBRABAN240_0001_" + time.Now().Format("20060102") + ".REM"
	if name != want {
		t.Fatalf("FileName() = %q, want %q", name, want)
	}
}

// TestFileNameCarriesConfigNSAAndValues checks File.FileName feeds the
// layout from the file's own configuration — including the header's NSA,
// which the name must never contradict.
func TestFileNameCarriesConfigNSAAndValues(t *testing.T) {
	cfg := validConfig()
	cfg.NSA = 41
	cfg.FileNameValues = map[string]string{"daily": "03"}
	cfg.FileNameLayout = NewFileNameLayout(
		Custom("daily"),
		Literal("N"),
		NSASequence.Width(4),
	)

	f := fileWithOneBatch(t, cfg)
	name, err := f.FileName()
	if err != nil {
		t.Fatalf("FileName() error = %v", err)
	}
	if name != "03N0041.REM" {
		t.Fatalf("FileName() = %q, want %q", name, "03N0041.REM")
	}
}

// fileWithOneBatch returns a generated-shaped file: FileName refuses a
// file with no batch, so every naming case needs one payment in place.
func fileWithOneBatch(t *testing.T, cfg Config) *File {
	t.Helper()

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
	return f
}
