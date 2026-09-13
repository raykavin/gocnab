package cnab

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Bank file name conventions have nothing in common beyond the extension:
// Sicredi wants the agreement code, the day of the month and a two digit
// sequence ("6CBY2701.REM"), other banks want the CNPJ, a full date, a
// literal prefix, or a six digit NSA. Rather than hard-coding any of them,
// this file lets a caller describe the convention as an ordered list of
// fields, each drawing one value and formatting it on its own.
//
// Use it either per file (Config.FileNameLayout) or process wide
// (SetRemittanceFileNameLayout); see FileNameLayout for the precedence.

// maxFileNameLength is the hard cap on a rendered file name, extension
// included. It exists so a defective layout can never produce a name the
// bank's own transmission software (or the local filesystem) rejects.
const maxFileNameLength = 255

// defaultFileNameExtension is the extension a FileNameLayout renders when
// it does not name one. Practically every Brazilian bank uses it for
// remittance files.
const defaultFileNameExtension = ".REM"

// fileNameSource identifies which value a FileNameField draws.
type fileNameSource int

const (
	// srcLiteral renders fixed text, held in the field itself.
	srcLiteral fileNameSource = iota
	srcAgreementCode
	srcBranch
	srcAccountNumber
	srcLayoutName
	srcLayoutVersion
	srcNSA
	srcTime
	srcCustom
)

// fileNameAlign selects which side of a fixed width field the padding goes
// on, and therefore which end Truncate trims.
type fileNameAlign int

const (
	// alignRight pads (and truncates) on the left: the natural rule for a
	// number, where the least significant digits are the ones to keep.
	alignRight fileNameAlign = iota
	// alignLeft pads (and truncates) on the right: the natural rule for
	// text, where the leading characters are the ones to keep.
	alignLeft
)

// FileNameField is one component of a remittance file name: a literal, or
// one dynamic value drawn from the file's configuration or from the
// generation instant, together with how it is formatted.
//
// Fields are immutable values. Every modifier (Width, Pad, Truncate,
// Upper, Lower, PadLeft, PadRight) returns a copy, so the exported
// prototypes below are safe to share and to reuse across layouts:
//
//	cnab.AgreementCode.Width(4)          // "6CBY", failing if longer
//	cnab.NSASequence.Width(2)            // "01", "42"
//	cnab.CurrentDay                      // "27"
//	cnab.Literal("REM")                  // "REM"
type FileNameField struct {
	source fileNameSource
	// literal holds the text of a srcLiteral field, the time layout of a
	// srcTime field, and the value name of a srcCustom field.
	literal string
	// width is the exact rendered width. Zero means "whatever the value's
	// natural width is", with no padding and no length check.
	width int
	pad   rune
	// padSet distinguishes "no Pad call" from Pad(0), so a field can keep
	// the default padding rune of its own source kind.
	padSet   bool
	align    fileNameAlign
	alignSet bool
	truncate bool
	upper    bool
	lower    bool
}

// The dynamic fields a file name layout can be built from. Combine them
// with Literal, Date and Custom, and format each one with the modifier
// methods on FileNameField.
var (
	// AgreementCode renders the payer's agreement ("convênio") code, as
	// carried by Config.Company.Agreement.
	AgreementCode = FileNameField{source: srcAgreementCode, align: alignLeft, alignSet: true, upper: true}
	// BranchNumber renders the debited account's branch, as carried by
	// Config.Account.Branch.
	BranchNumber = FileNameField{source: srcBranch}
	// AccountNumber renders the debited account's number, as carried by
	// Config.Account.Number.
	AccountNumber = FileNameField{source: srcAccountNumber}
	// LayoutName renders the active layout's name, upper cased.
	LayoutName = FileNameField{source: srcLayoutName, align: alignLeft, alignSet: true, upper: true}
	// LayoutVersion renders the active layout's version.
	LayoutVersion = FileNameField{source: srcLayoutVersion, align: alignLeft, alignSet: true}
	// NSASequence renders the file sequence number carried by Config.NSA
	// the very number stamped into the file header so a name built from
	// it can never name a sequence the file itself does not carry. Give it
	// a Width to zero pad it: NSASequence.Width(2) renders NSA 1 as "01".
	//
	// A bank whose file name sequence is NOT the header's NSA (Sicredi's
	// name, for instance, carries a sequence that restarts every day while
	// the header's NSA keeps climbing) must pass that other sequence as a
	// Custom value instead: silently rendering the wrong one of the two
	// would produce a name that contradicts the file's own header.
	NSASequence = FileNameField{source: srcNSA}
	// CurrentDay renders the generation date's day of the month, zero
	// padded to two digits.
	CurrentDay = Date("02")
	// CurrentMonth renders the generation date's month, zero padded to two
	// digits.
	CurrentMonth = Date("01")
	// CurrentYear renders the generation date's four digit year.
	CurrentYear = Date("2006")
	// CurrentShortYear renders the generation date's two digit year.
	CurrentShortYear = Date("06")
	// CurrentDate renders the generation date as YYYYMMDD.
	CurrentDate = Date("20060102")
)

// Literal returns a field that always renders text, unchanged. Use it for
// the fixed prefixes and separators a bank's convention calls for.
func Literal(text string) FileNameField {
	return FileNameField{source: srcLiteral, literal: text, align: alignLeft, alignSet: true}
}

// Date returns a field that renders the generation instant through a Go
// reference time layout (see time.Time.Format), e.g. Date("0201") for
// DDMM. The predefined CurrentDay, CurrentMonth, CurrentYear,
// CurrentShortYear and CurrentDate cover the common cases.
func Date(timeLayout string) FileNameField {
	return FileNameField{source: srcTime, literal: timeLayout, align: alignLeft, alignSet: true}
}

// Custom returns a field that renders FileNameInput.Values[name] (which
// Config.FileNameValues supplies when the name is rendered through
// File.FileName). It is how a bank specific value this package knows
// nothing about a daily file sequence, a product code, an operator id
// takes part in a file name without gocnab growing a field for it.
//
// Rendering fails when name has no value, rather than silently leaving a
// gap in the file name.
func Custom(name string) FileNameField {
	return FileNameField{source: srcCustom, literal: name, align: alignLeft, alignSet: true}
}

// Width returns a copy of the field rendered in exactly n characters. A
// shorter value is padded (see Pad, PadLeft and PadRight); a longer one is
// an error unless Truncate was called. n <= 0 restores the field's natural
// width, padded and checked not at all.
func (f FileNameField) Width(n int) FileNameField {
	if n < 0 {
		n = 0
	}
	f.width = n
	return f
}

// Pad returns a copy of the field padded with r instead of the default for
// its kind ('0' for a number, '0' for text too, since a space is not a
// character a file name may carry).
func (f FileNameField) Pad(r rune) FileNameField {
	f.pad, f.padSet = r, true
	return f
}

// PadLeft returns a copy of the field padded on the left, so the value
// ends up right aligned in its Width. This is the default for numeric
// fields.
func (f FileNameField) PadLeft() FileNameField {
	f.align, f.alignSet = alignRight, true
	return f
}

// PadRight returns a copy of the field padded on the right, so the value
// ends up left aligned in its Width. This is the default for text fields.
func (f FileNameField) PadRight() FileNameField {
	f.align, f.alignSet = alignLeft, true
	return f
}

// Truncate returns a copy of the field that trims a value wider than its
// Width instead of failing: from the left for a left padded (right
// aligned) field, keeping the least significant digits of a number, and
// from the right for a right padded one, keeping the leading characters of
// a text value.
//
// Leave it off for any field that identifies the file an NSA silently
// trimmed from 100 to "00" names a file the bank has already received.
func (f FileNameField) Truncate() FileNameField {
	f.truncate = true
	return f
}

// Upper returns a copy of the field rendered in upper case.
func (f FileNameField) Upper() FileNameField {
	f.upper, f.lower = true, false
	return f
}

// Lower returns a copy of the field rendered in lower case.
func (f FileNameField) Lower() FileNameField {
	f.lower, f.upper = true, false
	return f
}

// padRune returns the character this field pads with.
func (f FileNameField) padRune() rune {
	if f.padSet {
		return f.pad
	}
	return '0'
}

// isNumeric reports whether the field's source is a number, which decides
// its default alignment.
func (f FileNameField) isNumeric() bool {
	switch f.source {
	case srcNSA, srcBranch, srcAccountNumber:
		return true
	default:
		return false
	}
}

// alignment returns the field's effective padding side.
func (f FileNameField) alignment() fileNameAlign {
	if f.alignSet {
		return f.align
	}
	if f.isNumeric() {
		return alignRight
	}
	return alignLeft
}

// describe names the field in error messages.
func (f FileNameField) describe() string {
	switch f.source {
	case srcLiteral:
		return fmt.Sprintf("literal %q", f.literal)
	case srcAgreementCode:
		return "agreement code"
	case srcBranch:
		return "branch number"
	case srcAccountNumber:
		return "account number"
	case srcLayoutName:
		return "layout name"
	case srcLayoutVersion:
		return "layout version"
	case srcNSA:
		return "NSA sequence"
	case srcTime:
		return fmt.Sprintf("date %q", f.literal)
	case srcCustom:
		return fmt.Sprintf("custom value %q", f.literal)
	default:
		return "unknown field"
	}
}

// validate reports whether the field can render at all, independently of
// any input: its source must be one this package knows, and the sources
// that carry a name in literal (a literal's text, a date's time layout, a
// custom value's key) must actually carry one.
func (f FileNameField) validate() error {
	switch f.source {
	case srcAgreementCode, srcBranch, srcAccountNumber, srcLayoutName, srcLayoutVersion, srcNSA:
		return nil
	case srcLiteral:
		if f.literal == "" {
			return &ValidationError{Context: "FileNameLayout", Reason: "a literal field must carry text"}
		}
		return nil
	case srcTime:
		if f.literal == "" {
			return &ValidationError{Context: "FileNameLayout", Reason: "a date field must carry a time layout"}
		}
		return nil
	case srcCustom:
		if f.literal == "" {
			return &ValidationError{Context: "FileNameLayout", Reason: "a custom field must carry a value name"}
		}
		return nil
	default:
		return &ValidationError{
			Context: "FileNameLayout",
			Reason:  fmt.Sprintf("unknown field source %d; build fields with the constructors of this package", f.source),
		}
	}
}

// rawValue draws the field's value from in, before formatting.
func (f FileNameField) rawValue(in FileNameInput) (string, error) {
	switch f.source {
	case srcLiteral:
		return f.literal, nil
	case srcAgreementCode:
		return strings.TrimSpace(in.AgreementCode), nil
	case srcBranch:
		return strings.TrimSpace(in.Branch), nil
	case srcAccountNumber:
		return strings.TrimSpace(in.AccountNumber), nil
	case srcLayoutName:
		return strings.TrimSpace(in.LayoutName), nil
	case srcLayoutVersion:
		return strings.TrimSpace(in.LayoutVersion), nil
	case srcNSA:
		return strconv.Itoa(in.NSA), nil
	case srcTime:
		return in.at().Format(f.literal), nil
	case srcCustom:
		value, ok := in.Values[f.literal]
		if !ok {
			return "", &ValidationError{
				Context: "FileNameLayout",
				Reason:  fmt.Sprintf("no value supplied for custom field %q", f.literal),
			}
		}
		return strings.TrimSpace(value), nil
	default:
		return "", &ValidationError{Context: "FileNameLayout", Reason: "unknown field source"}
	}
}

// render draws the field's value and formats it to its configured width.
func (f FileNameField) render(in FileNameInput) (string, error) {
	value, err := f.rawValue(in)
	if err != nil {
		return "", err
	}

	// A dynamic field that drew nothing is missing data, not an empty
	// component: padding it would hide the gap behind the field's own fill
	// character, naming a file "00002701.REM" for a payer whose agreement
	// code never reached the layout. Only a literal may be empty, and
	// Literal("") is refused when the layout is validated.
	if value == "" && f.source != srcLiteral {
		return "", &ValidationError{
			Context: "FileNameLayout",
			Reason:  fmt.Sprintf("%s has no value", f.describe()),
		}
	}

	switch {
	case f.upper:
		value = strings.ToUpper(value)
	case f.lower:
		value = strings.ToLower(value)
	}

	if f.width == 0 {
		return value, nil
	}

	// Widths are in characters, not bytes: an agreement code is normally
	// ASCII, but a layout is free to name a literal that is not.
	runes := []rune(value)
	switch {
	case len(runes) == f.width:
		return value, nil

	case len(runes) < f.width:
		padding := strings.Repeat(string(f.padRune()), f.width-len(runes))
		if f.alignment() == alignRight {
			return padding + value, nil
		}
		return value + padding, nil

	case !f.truncate:
		return "", &ValidationError{
			Context: "FileNameLayout",
			Reason: fmt.Sprintf(
				"%s renders %q, which does not fit its width of %d; widen the field or allow truncation",
				f.describe(), value, f.width,
			),
		}

	case f.alignment() == alignRight:
		// Keep the least significant characters of a number.
		return string(runes[len(runes)-f.width:]), nil

	default:
		return string(runes[:f.width]), nil
	}
}

// FileNameInput carries the values a FileNameLayout draws its dynamic
// fields from. File.FileName fills it from the file's own Config and the
// current instant; a caller that names a file without building one (an
// adapter that reserves the name before it has the payments, say) fills it
// directly and calls FileNameLayout.Render.
type FileNameInput struct {
	// AgreementCode feeds AgreementCode.
	AgreementCode string
	// Branch feeds BranchNumber.
	Branch string
	// AccountNumber feeds AccountNumber.
	AccountNumber string
	// LayoutName feeds LayoutName.
	LayoutName string
	// LayoutVersion feeds LayoutVersion.
	LayoutVersion string
	// NSA feeds NSASequence, and must be the same file sequence number the
	// file header carries.
	NSA int
	// Now is the generation instant every Date field formats. Zero means
	// time.Now(), resolved once per Render so two date fields of one name
	// can never straddle midnight.
	Now time.Time
	// Values feeds Custom fields, by name.
	Values map[string]string
}

// at returns the generation instant, defaulting to now.
func (in FileNameInput) at() time.Time {
	if in.Now.IsZero() {
		return time.Now()
	}
	return in.Now
}

// FileNameLayout describes a bank's remittance file name convention: the
// ordered fields the stem is built from, the extension, and the length the
// bank accepts.
//
// A file resolves its layout in this order: Config.FileNameLayout when
// set, otherwise the process wide default installed by
// SetRemittanceFileNameLayout, otherwise the built-in
// "<LAYOUT>_<NSA>_<YYYYMMDD>.REM" fallback that File.FileName has always
// produced. Prefer Config.FileNameLayout in a process that generates files
// for more than one bank: the default is a single global, so two banks
// with different conventions cannot both own it.
type FileNameLayout struct {
	// Fields are the stem's components, rendered in order.
	Fields []FileNameField
	// Extension is appended to the stem, dot included (e.g. ".REM"). Empty
	// renders defaultFileNameExtension; use ExtensionNone for no extension
	// at all.
	Extension string
	// MaxLength caps the rendered name, extension included. Zero means
	// maxFileNameLength. A bank that documents an exact name width should
	// set it, so a layout that renders too much fails here rather than at
	// the bank.
	MaxLength int
}

// ExtensionNone, as FileNameLayout.Extension, renders a name with no
// extension at all distinct from the empty string, which selects the
// ".REM" default.
const ExtensionNone = "-"

// NewFileNameLayout returns a layout rendering fields in order, with the
// default ".REM" extension.
func NewFileNameLayout(fields ...FileNameField) FileNameLayout {
	return FileNameLayout{Fields: fields}
}

// WithExtension returns a copy of the layout using ext (dot included, e.g.
// ".REM"), or no extension when ext is ExtensionNone.
func (l FileNameLayout) WithExtension(ext string) FileNameLayout {
	l.Extension = ext
	return l
}

// WithMaxLength returns a copy of the layout that rejects a rendered name
// longer than n characters, extension included.
func (l FileNameLayout) WithMaxLength(n int) FileNameLayout {
	l.MaxLength = n
	return l
}

// IsZero reports whether the layout names no field, and therefore cannot
// render anything.
func (l FileNameLayout) IsZero() bool { return len(l.Fields) == 0 }

// extension returns the effective extension, "" when the layout wants none.
func (l FileNameLayout) extension() string {
	switch l.Extension {
	case "":
		return defaultFileNameExtension
	case ExtensionNone:
		return ""
	default:
		return l.Extension
	}
}

// maxLength returns the effective length cap.
func (l FileNameLayout) maxLength() int {
	if l.MaxLength > 0 && l.MaxLength < maxFileNameLength {
		return l.MaxLength
	}
	return maxFileNameLength
}

// Validate reports whether the layout can render a name at all: it must
// name at least one field, its extension must be a dot followed by
// alphanumerics, and its MaxLength must be within the hard cap. It is
// called by Render and by SetRemittanceFileNameLayout, so a malformed
// layout is refused where it is installed rather than when a file is
// finally named.
func (l FileNameLayout) Validate() error {
	if l.IsZero() {
		return &ValidationError{Context: "FileNameLayout", Reason: "at least one field is required"}
	}
	if l.MaxLength < 0 || l.MaxLength > maxFileNameLength {
		return &ValidationError{
			Context: "FileNameLayout",
			Reason:  fmt.Sprintf("MaxLength must be between 0 and %d", maxFileNameLength),
		}
	}
	for i, field := range l.Fields {
		if err := field.validate(); err != nil {
			return fmt.Errorf("field %d: %w", i+1, err)
		}
	}
	if ext := l.extension(); ext != "" {
		if !strings.HasPrefix(ext, ".") || len(ext) < 2 {
			return &ValidationError{
				Context: "FileNameLayout",
				Reason:  fmt.Sprintf("extension %q must start with a dot and name at least one character", ext),
			}
		}
		for _, r := range ext[1:] {
			if !isFileNameAlphanumeric(r) {
				return &ValidationError{
					Context: "FileNameLayout",
					Reason:  fmt.Sprintf("extension %q may only contain letters and digits after the dot", ext),
				}
			}
		}
	}
	return nil
}

// Render builds the file name from in.
//
// It returns a *ValidationError when the layout itself is malformed, when
// a field's value does not fit the width it was given, when a Custom field
// has no value, when the stem comes out empty, when the name exceeds the
// layout's MaxLength, or when any field rendered a character a file name
// may not carry (see isFileNameChar): a name that reaches the bank's
// transmission software malformed is worse than a generation that fails
// with the reason.
func (l FileNameLayout) Render(in FileNameInput) (string, error) {
	if err := l.Validate(); err != nil {
		return "", err
	}

	// Resolve the instant once so every Date field of one name agrees,
	// even for a file generated as the day turns.
	in.Now = in.at()

	var stem strings.Builder
	for _, field := range l.Fields {
		part, err := field.render(in)
		if err != nil {
			return "", err
		}
		if invalid, ok := firstInvalidFileNameChar(part); ok {
			return "", &ValidationError{
				Context: "FileNameLayout",
				Reason: fmt.Sprintf(
					"%s renders %q, which contains the character %q that a file name may not carry",
					field.describe(), part, invalid,
				),
			}
		}
		stem.WriteString(part)
	}

	if stem.Len() == 0 {
		return "", &ValidationError{Context: "FileNameLayout", Reason: "rendered an empty file name"}
	}

	name := stem.String() + l.extension()
	if length := len([]rune(name)); length > l.maxLength() {
		return "", &ValidationError{
			Context: "FileNameLayout",
			Reason: fmt.Sprintf(
				"rendered name %q is %d characters long, over the %d allowed",
				name, length, l.maxLength(),
			),
		}
	}
	return name, nil
}

// isFileNameAlphanumeric reports whether r is an ASCII letter or digit.
func isFileNameAlphanumeric(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// isFileNameChar reports whether r may appear in a rendered file name
// stem. The set is deliberately narrow letters, digits, and the three
// separators every filesystem and every bank transmission client accepts
// so a stray space, accent, slash or dot from an agreement code or a
// company name cannot produce a name that is rejected, silently renamed,
// or read as a path.
func isFileNameChar(r rune) bool {
	return isFileNameAlphanumeric(r) || r == '_' || r == '-' || r == '$'
}

// firstInvalidFileNameChar returns the first character of s that a file
// name may not carry.
func firstInvalidFileNameChar(s string) (rune, bool) {
	for _, r := range s {
		if !isFileNameChar(r) {
			return r, true
		}
	}
	return 0, false
}

// defaultFileNameLayout is the process wide convention File.FileName uses
// when a file's Config does not carry one of its own. Guarded because
// SetRemittanceFileNameLayout may be called from an init or a config load
// while other goroutines are already generating files.
var (
	defaultFileNameLayoutMu sync.RWMutex
	defaultFileNameLayout   FileNameLayout
)

// SetRemittanceFileNameLayout installs layout as the process wide file
// name convention, replacing any previously installed one:
//
//	cnab.SetRemittanceFileNameLayout(
//	    cnab.AgreementCode.Width(4),
//	    cnab.CurrentDay,
//	    cnab.NSASequence.Width(2),
//	)
//
// It validates the fields before installing them and returns a
// *ValidationError describing what is wrong, leaving the previous layout
// in place, so an invalid convention is reported where it is declared
// rather than when the first file is named.
//
// Pass no field to restore the built-in fallback. In a process that
// generates files for several banks, set Config.FileNameLayout per file
// instead: this default is a single global and the last caller wins.
func SetRemittanceFileNameLayout(fields ...FileNameField) error {
	if len(fields) == 0 {
		defaultFileNameLayoutMu.Lock()
		defaultFileNameLayout = FileNameLayout{}
		defaultFileNameLayoutMu.Unlock()
		return nil
	}

	l := NewFileNameLayout(fields...)
	if err := l.Validate(); err != nil {
		return err
	}

	defaultFileNameLayoutMu.Lock()
	defaultFileNameLayout = l
	defaultFileNameLayoutMu.Unlock()
	return nil
}

// RemittanceFileNameLayout returns the process wide layout currently
// installed, zero when none is.
func RemittanceFileNameLayout() FileNameLayout {
	defaultFileNameLayoutMu.RLock()
	defer defaultFileNameLayoutMu.RUnlock()
	return defaultFileNameLayout
}
