package cnab

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
	"github.com/raykavin/gocnab/internal/engine"
)

// ReturnMovement is one payment's outcome as reported by a bank on a
// CNAB 240 return file, decoded from its Segmento A, J or O one movement
// per credit-in-account/TED/PIX payment (Segmento A), boleto payment
// (Segmento J its J-52 continuation is skipped) or utility/barcoded tax
// payment (Segmento O).
//
// A return file's other segments (B, BPix, J-52, ...) carry the
// beneficiary's document/address/PIX key back, unchanged from what the
// remittance sent data a caller already has from its own records.
// ParseReturn does not decode them: everything needed to reconcile a
// payment (whether it settled, when, for how much, why not if it did not,
// and YourNumber to match it back to the caller's own payment) lives on
// the primary segment (A, J or O) alone. Segmento N (DARF/GPS) is not
// decoded either; add support for it the same way if your bank layout
// needs it.
type ReturnMovement struct {
	// YourNumber is the payer's own reference ("seu número") echoed back
	// unchanged from the remittance, the same value AddPayment received
	// via a Payment's YourNumber field. Use it to match a movement back to
	// a specific payment.
	YourNumber string

	// Amount is the amount the remittance instructed the bank to pay.
	Amount Cents

	// SettlementDate is the date the bank actually settled the payment
	// ("data real da efetivação do pagamento"). Zero if the bank left it
	// unset, which normally means the payment was rejected before
	// settlement see OccurrenceCodes.
	SettlementDate time.Time

	// SettlementAmount is the amount the bank actually settled ("valor
	// real da efetivação do pagamento"). It can differ from Amount for a
	// bank-specific reason (e.g. a partial settlement); it is zero
	// whenever SettlementDate is zero.
	SettlementAmount Cents

	// OccurrenceCodes lists every non-zero return/rejection occurrence
	// code the bank reported (up to five, per the FEBRABAN standard), in
	// the order they appear. Empty means the bank reported no occurrence,
	// i.e. Accepted reports true. The table mapping a code to what it
	// means is bank-specific; confirm it against your bank's manual this
	// SDK only extracts the raw codes, matching how it treats a TED
	// PurposeCode.
	OccurrenceCodes []string

	// Authentication and BankControl carry the payment authentication/
	// protocol data from a trailing Segmento Z, when the bank layout
	// implements it and the bank included one after this movement's
	// primary segment. Both are empty when the layout does not support
	// Segmento Z, or the bank did not send one for this movement (it is
	// documented as optional in every manual this SDK has been built
	// against).
	Authentication string
	BankControl    string
}

// Accepted reports whether the bank reported no occurrence code for this
// movement.
func (m ReturnMovement) Accepted() bool {
	return len(m.OccurrenceCodes) == 0
}

// ReturnFile is the structured result of parsing a CNAB 240 return file.
type ReturnFile struct {
	// Movements lists one ReturnMovement per Segmento A found in the file,
	// in the order they appear.
	Movements []ReturnMovement
}

// ParseReturn decodes a CNAB 240 return file previously received from a
// bank, using the layout registered under layoutName normally the same
// layout the corresponding remittance was generated with, since a bank
// returns data in the same physical positions it received it in.
//
// It splits content into 240 character lines (accepting either CRLF or
// bare LF line endings, and tolerating a trailing blank line), classifies
// each by its FEBRABAN-standard record type and segment code, and decodes
// every Segmento A, J or O it finds into a ReturnMovement (a Segmento J
// whose "Código Reg. Opcional" field, columns 18-19, reads "52" is a
// Segmento J-52 continuation record, not a primary Segmento J movement,
// and is skipped like every other continuation segment). Every other
// record kind (file/batch headers and trailers, Segmento B/BPix/J-52/N
// and any other detail segment) is skipped: see ReturnMovement for why
// the primary segment alone is enough to reconcile a payment.
//
// It returns a *ValidationError if layoutName is not registered or if the
// layout itself is malformed (the same as NewRemittance), and a
// *ReturnParseError (a type this package exports, unlike the internal
// engine errors ParseReturn's lower-level calls actually produce) if a
// line is not exactly 240 characters, has an unrecognized record type
// marker, or (classified as a Segmento A, J or O) has a const field whose
// content does not match what the layout expects there.
func ParseReturn(layoutName string, content []byte) (*ReturnFile, error) {
	l, ok := layout.Lookup(layoutName)
	if !ok {
		return nil, &ValidationError{
			Context: "ParseReturn",
			Reason:  fmt.Sprintf("layout %q is not registered (available: %v)", layoutName, layout.Names()),
		}
	}
	return ParseReturnWithLayout(l, content)
}

// ParseReturnWithLayout is ParseReturn against a Layout instance instead
// of a registered name, for a caller whose layout is runtime data rather
// than a process-wide constant (see Config.LayoutSpec). It returns a
// *ValidationError when l is nil.
func ParseReturnWithLayout(l Layout, content []byte) (*ReturnFile, error) {
	if l == nil {
		return nil, &ValidationError{Context: "ParseReturn", Reason: "Layout is required"}
	}

	eng, err := engine.New(l)
	if err != nil {
		return nil, err
	}

	lines := splitLines(content)

	result := &ReturnFile{}
	for i, line := range lines {
		lineNumber := i + 1
		if len(line) != 240 {
			return nil, &ReturnParseError{Line: lineNumber, Reason: fmt.Sprintf("line has %d characters, want 240", len(line))}
		}

		recordType, segmentCode, err := engine.ClassifyLine(line)
		if err != nil {
			return nil, &ReturnParseError{Line: lineNumber, Reason: err.Error()}
		}

		var recordKey layout.RecordKey
		switch recordType {
		case engine.RecordTypeFileHeader, engine.RecordTypeBatchHeader, engine.RecordTypeBatchTrailer, engine.RecordTypeFileTrailer:
			continue
		case engine.RecordTypeDetail:
			switch {
			case segmentCode == "A":
				recordKey = layout.SegmentA
			case segmentCode == "O":
				recordKey = layout.SegmentO
			case segmentCode == "J" && !isSegmentJ52(line):
				recordKey = layout.SegmentJ
			case segmentCode == "Z":
				if err := attachSegmentZ(eng, line, result); err != nil {
					return nil, returnParseErrorFor(lineNumber, err)
				}
				continue
			default:
				// Segmento J-52, B, BPix, N and any other detail segment:
				// a continuation record, not a primary movement.
				continue
			}
		default:
			return nil, &ReturnParseError{Line: lineNumber, Reason: fmt.Sprintf("unrecognized record type %q", recordType)}
		}

		movement, err := parseReturnMovement(eng, recordKey, line)
		if err != nil {
			return nil, returnParseErrorFor(lineNumber, err)
		}
		result.Movements = append(result.Movements, movement)
	}

	return result, nil
}

// returnParseErrorFor translates an error from Engine.ParseRecord (an
// *engine.RecordMismatchError, or occasionally a *ValidationError raised by
// parseReturnMovement itself while decoding a value) into a
// *ReturnParseError, so ParseReturn never returns a type from
// internal/engine which a caller outside this module cannot name to
// match against with errors.As, even though the error value would
// otherwise genuinely be of that type.
func returnParseErrorFor(line int, err error) error {
	var mismatch *engine.RecordMismatchError
	if errors.As(err, &mismatch) {
		return &ReturnParseError{Line: line, Field: mismatch.Field, Expected: mismatch.Expected, Got: mismatch.Got}
	}
	return &ReturnParseError{Line: line, Reason: err.Error()}
}

// segmentJ52IdentifierColumn is the 1-based, inclusive column range of the
// "Código Reg. Opcional" field FEBRABAN uses to tell a Segmento J-52
// continuation record apart from a primary Segmento J one: both share
// segment code "J" at column 14 (see engine.ClassifyLine), so this second
// marker is the only way to distinguish them before deciding which
// RecordSpec to parse the line with. Fixed by the FEBRABAN standard
// itself, the same way engine.ClassifyLine's own column positions are.
const (
	segmentJ52IdentifierStart = 18
	segmentJ52IdentifierEnd   = 19
	segmentJ52Identifier      = "52"
)

// isSegmentJ52 reports whether line is a Segmento J-52 continuation
// record rather than a primary Segmento J movement. It is only meaningful
// for a line engine.ClassifyLine already reported as record type "3"
// (detail) with segment code "J".
func isSegmentJ52(line string) bool {
	if len(line) < segmentJ52IdentifierEnd {
		return false
	}
	return line[segmentJ52IdentifierStart-1:segmentJ52IdentifierEnd] == segmentJ52Identifier
}

// attachSegmentZ decodes a Segmento Z line and attaches its authentication/
// bank control data to the most recently appended movement in result:
// per every manual this SDK has been built against, a bank appends Z
// right after the primary segment of the movement it authenticates, so it
// never starts a movement of its own. It is a no-op (not an error) when
// the active layout does not implement Segmento Z, or when result has no
// movement yet to attach to (a malformed or truncated file parsing
// continues rather than failing on what is, per every manual, optional
// data).
func attachSegmentZ(eng *engine.Engine, line string, result *ReturnFile) error {
	if !eng.Supports(layout.SegmentZ) || len(result.Movements) == 0 {
		return nil
	}

	values, err := eng.ParseRecord(layout.SegmentZ, line)
	if err != nil {
		return err
	}

	last := &result.Movements[len(result.Movements)-1]
	last.Authentication = stringValue(values, layout.KeyAuthentication)
	last.BankControl = stringValue(values, layout.KeyBankControl)
	return nil
}

func parseReturnMovement(eng *engine.Engine, key layout.RecordKey, line string) (ReturnMovement, error) {
	values, err := eng.ParseRecord(key, line)
	if err != nil {
		return ReturnMovement{}, err
	}

	movement := ReturnMovement{
		YourNumber:      stringValue(values, layout.KeyYourNumber),
		OccurrenceCodes: splitOccurrenceCodes(stringValue(values, layout.KeyOccurrenceCodes)),
	}

	if raw := stringValue(values, layout.KeyAmount); raw != "" {
		amount, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return ReturnMovement{}, &ValidationError{Context: "ParseReturn", Reason: "amount: " + err.Error()}
		}
		movement.Amount = Cents(amount)
	}

	if raw := stringValue(values, layout.KeySettlementAmount); raw != "" {
		amount, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return ReturnMovement{}, &ValidationError{Context: "ParseReturn", Reason: "settlement amount: " + err.Error()}
		}
		movement.SettlementAmount = Cents(amount)
	}

	if raw := stringValue(values, layout.KeySettlementDate); raw != "" && raw != "00000000" {
		d, err := time.Parse("02012006", raw)
		if err != nil {
			return ReturnMovement{}, &ValidationError{Context: "ParseReturn", Reason: "settlement date: " + err.Error()}
		}
		movement.SettlementDate = d
	}

	return movement, nil
}

// stringValue reads key from values as a string, defaulting to "" when the
// key is absent every value engine.ParseRecord produces is a string, by
// construction (see engine.parseField), so this never needs to report a
// type-assertion failure.
func stringValue(values layout.Values, key layout.Key) string {
	s, _ := values[key].(string)
	return s
}

// splitOccurrenceCodes chunks the 10 character occurrence codes field into
// up to five 2 character codes, dropping "00" and blank chunks both mean
// "no occurrence" per the FEBRABAN standard.
func splitOccurrenceCodes(raw string) []string {
	var codes []string
	for i := 0; i+2 <= len(raw); i += 2 {
		code := raw[i : i+2]
		if code == "00" || strings.TrimSpace(code) == "" {
			continue
		}
		codes = append(codes, code)
	}
	return codes
}

// splitLines splits content into lines on "\n", stripping a trailing "\r"
// from each (so either CRLF or bare LF input works) and dropping a
// trailing empty line (the one after content's final line terminator, if
// it has one).
func splitLines(content []byte) []string {
	raw := strings.Split(string(content), "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	lines := make([]string, len(raw))
	for i, l := range raw {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}