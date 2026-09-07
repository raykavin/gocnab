package cnab

import (
	"strconv"
)

// BarcodeSegment identifies which detail segment a barcode-based payment
// belongs to: a bank slip (Segmento J) or a utility bill/tax payment
// (Segmento O). ConvertToBarcode reports it alongside the normalized
// barcode, since the two are decided by the same digits.
type BarcodeSegment int8

const (
	// SegmentBankSlip is a bank slip ("boleto"), rendered as a BoletoPayment
	// (Segmento J).
	SegmentBankSlip BarcodeSegment = iota + 1
	// SegmentFeesOrTaxes is a utility bill or tax payment, rendered as a
	// BarcodeTax (Segmento O).
	SegmentFeesOrTaxes
)

// ConvertToBarcode normalizes a Brazilian bank slip's "linha digitável"
// (typeable line, 47 digits) or a utility/tax bill's typeable line (48
// digits) into the 44 digit barcode BoletoPayment.Barcode and
// BarcodeTax.Barcode expect, and reports which of the two segments the
// result belongs to. A value that is already 44 digits is accepted as-is
// and only classified.
//
// Every path validates the FEBRABAN check digits before returning: a 47 or
// 48 digit typeable line has each of its field check digits (módulo 10)
// verified while being reduced to a barcode, and every barcode however
// it arrived, typed in directly or just assembled has its own general
// check digit (módulo 10 or 11, depending on segment) verified against the
// other 43 digits. A tampered or mistyped input a single digit flipped by
// an OCR misread or a fat-fingered manual entry is rejected here, never
// silently turned into a structurally-valid-looking but wrong barcode that
// would go on to be paid.
//
// It returns a *ValidationError if raw contains characters other than
// digits and the conventional '.', ' ', '-' separators, if its digit count
// does not match one of the three known cases, or if any check digit
// (field or general) does not match.
func ConvertToBarcode(raw string) (BarcodeSegment, string, error) {
	if !isDigitsOrConventionalSeparators(raw) {
		return 0, "", &ValidationError{
			Context: "ConvertToBarcode",
			Reason:  "typeable line contains characters other than digits and the conventional '.', ' ', '-' separators",
		}
	}

	digits := onlyDigits(raw)

	switch len(digits) {
	case 44:
		if err := validateGeneralCheckDigit(digits); err != nil {
			return 0, "", err
		}
		return detectBarcodeSegment(digits), digits, nil
	case 47:
		barcode, err := bankSlipLineToBarcode(digits)
		if err != nil {
			return 0, "", err
		}
		if err := validateGeneralCheckDigit(barcode); err != nil {
			return 0, "", err
		}
		return SegmentBankSlip, barcode, nil
	case 48:
		barcode, err := taxLineToBarcode(digits)
		if err != nil {
			return 0, "", err
		}
		if err := validateGeneralCheckDigit(barcode); err != nil {
			return 0, "", err
		}
		return SegmentFeesOrTaxes, barcode, nil
	default:
		return 0, "", &ValidationError{
			Context: "ConvertToBarcode",
			Reason:  "typeable line must have 44, 47 or 48 digits, got " + strconv.Itoa(len(digits)),
		}
	}
}

// isDigitsOrConventionalSeparators reports whether s contains only ASCII
// digits and the separators conventionally used to typeset a linha
// digitável for humans ('.', ' ', '-'). onlyDigits silently discards
// anything that is not a digit, so this must run first: without it, a
// mistyped or OCR-corrupted character (a letter, a stray symbol) would
// simply vanish instead of being rejected, and the digits on either side of
// it could still happen to add up to a valid length.
func isDigitsOrConventionalSeparators(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '.' || r == ' ' || r == '-':
		default:
			return false
		}
	}
	return true
}

// detectBarcodeSegment classifies an already-44-digit barcode: position 1
// (the currency/product code) reads "8" for a utility bill or tax payment,
// and the bank's own COMPE code otherwise.
func detectBarcodeSegment(barcode string) BarcodeSegment {
	if len(barcode) > 0 && barcode[0] == '8' {
		return SegmentFeesOrTaxes
	}
	return SegmentBankSlip
}

// bankSlipLineToBarcode validates and rearranges a 47 digit bank slip
// typeable line into its 44 digit barcode, per the FEBRABAN layout
// (0-indexed):
//
//	Campo 1 (10): data = line[0:9]  (bankCode[3]+currency[1]+freeField[0:5]), DV1 = line[9]  (módulo 10)
//	Campo 2 (11): data = line[10:20] (freeField[5:15]),                       DV2 = line[20] (módulo 10)
//	Campo 3 (11): data = line[21:31] (freeField[15:25]),                      DV3 = line[31] (módulo 10)
//	Campo 4 (1):  general check digit (barcode position 5)  = line[32]
//	Campo 5 (14): fator de vencimento (4) + valor (10)       = line[33:47]
//
// The line interleaves these three check-digit blocks into the barcode's
// contiguous free-field/due-date/amount block; this verifies each block's
// own check digit against its data before reordering them back out. A
// mismatch here is the exact failure mode a mistyped or OCR-misread
// digitable line produces, and previously went undetected: the digits were
// simply reordered into a barcode that looked structurally valid.
func bankSlipLineToBarcode(line string) (string, error) {
	data1, dv1 := line[0:9], line[9]
	data2, dv2 := line[10:20], line[20]
	data3, dv3 := line[21:31], line[31]

	if mod10(data1) != dv1 || mod10(data2) != dv2 || mod10(data3) != dv3 {
		return "", &ValidationError{Context: "ConvertToBarcode", Reason: "bank slip typeable line field check digit does not match"}
	}

	return line[0:4] + line[32:47] + line[4:9] + line[10:20] + line[21:31], nil
}

// taxLineToBarcode validates and strips the four block check digits (one
// every 12th position, módulo 10 over the preceding 11 data digits) a 48
// digit utility/tax typeable line carries on top of its 44 digit barcode.
func taxLineToBarcode(line string) (string, error) {
	if len(line) != 48 {
		return "", &ValidationError{Context: "ConvertToBarcode", Reason: "tax typeable line must have 48 digits"}
	}

	barcode := make([]byte, 0, 44)
	for i := 0; i < 4; i++ {
		segment := line[i*12 : i*12+12]
		data, dv := segment[0:11], segment[11]
		if mod10(data) != dv {
			return "", &ValidationError{
				Context: "ConvertToBarcode",
				Reason:  "tax typeable line field check digit does not match",
			}
		}
		barcode = append(barcode, data...)
	}

	return string(barcode), nil
}
