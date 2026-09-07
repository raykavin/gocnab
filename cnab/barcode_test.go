package cnab

import "testing"

// buildBankSlipBarcode assembles a structurally valid, correctly
// check-digited 44-digit bank slip barcode: bankCode+"9"+generalDV(mod11)+
// factor(4)+value(10)+freeField(25).
func buildBankSlipBarcode(t *testing.T, bankCode string, factor int, value int64, freeField string) string {
	t.Helper()
	if len(bankCode) != 3 || len(freeField) != 25 {
		t.Fatalf("bad test input: bankCode=%q freeField=%q", bankCode, freeField)
	}
	factorStr := zeroPad(itoa(factor), 4)
	valueStr := zeroPad(itoa64(value), 10)
	withoutDV := bankCode + "9" + "?" + factorStr + valueStr + freeField
	dv := generalMod11(withoutDV[0:4] + withoutDV[5:44])
	return bankCode + "9" + string(dv) + factorStr + valueStr + freeField
}

// buildBankSlipLine converts a valid 44-digit bank slip barcode into its
// 47-digit typeable line, computing the three field check digits (módulo
// 10) the same way a bank's own system would the encoding counterpart to
// bankSlipLineToBarcode.
func buildBankSlipLine(t *testing.T, barcode string) string {
	t.Helper()
	if len(barcode) != 44 {
		t.Fatalf("bad test input: barcode=%q", barcode)
	}
	bankAndCurrency := barcode[0:4]
	generalDV := barcode[4]
	factorAndValue := barcode[5:19]
	freeField := barcode[19:44]

	data1 := bankAndCurrency + freeField[0:5]
	data2 := freeField[5:15]
	data3 := freeField[15:25]

	return data1 + string(mod10(data1)) +
		data2 + string(mod10(data2)) +
		data3 + string(mod10(data3)) +
		string(generalDV) + factorAndValue
}

// buildTaxBarcode assembles a structurally valid, correctly check-digited
// 44-digit utility-bill/tax barcode using the common módulo 10 variant
// (value-type digit '6' at position 3, 1-indexed, index 2): productID(1) +
// segment(1, unvalidated by this package) + valueType(1) + generalDV(1,
// módulo 10) + value(11) + freeField(29).
func buildTaxBarcode(value int64, freeField string) string {
	if len(freeField) != 29 {
		panic("buildTaxBarcode: freeField must have 29 digits")
	}
	head := "8" + "1" + "6"                        // index 0 (product), 1 (segment), 2 (value type)
	tail := zeroPad(itoa64(value), 11) + freeField // index 4..43
	dv := mod10(head + tail)
	return head + string(dv) + tail
}

// buildTaxLine converts a valid 44-digit tax barcode into its 48-digit
// typeable line: four 12-character fields, each 11 barcode data digits
// followed by their own módulo 10 check digit.
func buildTaxLine(t *testing.T, barcode string) string {
	t.Helper()
	if len(barcode) != 44 {
		t.Fatalf("bad test input: barcode=%q", barcode)
	}
	line := make([]byte, 0, 48)
	for i := 0; i < 4; i++ {
		data := barcode[i*11 : i*11+11]
		line = append(line, data...)
		line = append(line, mod10(data))
	}
	return string(line)
}

func zeroPad(s string, n int) string {
	for len(s) < n {
		s = "0" + s
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func itoa64(n int64) string { return itoa(int(n)) }

func TestConvertToBarcode_AlreadyBarcode(t *testing.T) {
	bankSlip := buildBankSlipBarcode(t, "748", 5000, 123456, "1234567890123456789012345")
	tax := buildTaxBarcode(987654321, "00000000000000000000000000000"[:29])

	tests := []struct {
		name    string
		barcode string
		want    BarcodeSegment
	}{
		{"bank slip", bankSlip, SegmentBankSlip},
		{"tax/utility", tax, SegmentFeesOrTaxes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segment, barcode, err := ConvertToBarcode(tt.barcode)
			if err != nil {
				t.Fatalf("ConvertToBarcode() error = %v", err)
			}
			if segment != tt.want {
				t.Errorf("segment = %v, want %v", segment, tt.want)
			}
			if barcode != tt.barcode {
				t.Errorf("barcode = %q, want %q (44-digit input must pass through unchanged)", barcode, tt.barcode)
			}
		})
	}
}

func TestConvertToBarcode_AlreadyBarcode_TamperedGeneralCheckDigit_Rejected(t *testing.T) {
	barcode := buildBankSlipBarcode(t, "748", 5000, 123456, "1234567890123456789012345")
	tampered := []byte(barcode)
	tampered[4] = flipDigit(tampered[4])

	if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
		t.Fatal("expected an error for a barcode with a tampered general check digit, got nil")
	}
}

func TestConvertToBarcode_BankSlipTypeableLine(t *testing.T) {
	barcode := buildBankSlipBarcode(t, "341", 1234, 987654, "0000000000000000000000001")
	line := buildBankSlipLine(t, barcode)

	segment, got, err := ConvertToBarcode(line)
	if err != nil {
		t.Fatalf("ConvertToBarcode() error = %v", err)
	}
	if segment != SegmentBankSlip {
		t.Errorf("segment = %v, want SegmentBankSlip", segment)
	}
	if got != barcode {
		t.Errorf("barcode = %q, want %q", got, barcode)
	}
}

func TestConvertToBarcode_BankSlipTypeableLine_TamperedFieldCheckDigit_Rejected(t *testing.T) {
	barcode := buildBankSlipBarcode(t, "341", 1234, 987654, "0000000000000000000000001")
	line := buildBankSlipLine(t, barcode)
	tampered := []byte(line)
	tampered[9] = flipDigit(tampered[9]) // DV1

	if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
		t.Fatal("expected an error for a typeable line with a tampered field check digit, got nil")
	}
}

func TestConvertToBarcode_TaxTypeableLine(t *testing.T) {
	barcode := buildTaxBarcode(555566, "11111111111111111111111111111"[:29])
	line := buildTaxLine(t, barcode)

	segment, got, err := ConvertToBarcode(line)
	if err != nil {
		t.Fatalf("ConvertToBarcode() error = %v", err)
	}
	if segment != SegmentFeesOrTaxes {
		t.Errorf("segment = %v, want SegmentFeesOrTaxes", segment)
	}
	if got != barcode {
		t.Errorf("barcode = %q, want %q", got, barcode)
	}
}

func TestConvertToBarcode_TaxTypeableLine_TamperedFieldCheckDigit_Rejected(t *testing.T) {
	barcode := buildTaxBarcode(555566, "11111111111111111111111111111"[:29])
	line := buildTaxLine(t, barcode)
	tampered := []byte(line)
	tampered[11] = flipDigit(tampered[11]) // first block's check digit

	if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
		t.Fatal("expected an error for a tax typeable line with a tampered field check digit, got nil")
	}
}

func TestConvertToBarcode_TaxBarcode_UnsupportedMod11Variant_Rejected(t *testing.T) {
	barcode := buildTaxBarcode(1000, "00000000000000000000000000000"[:29])
	tampered := []byte(barcode)
	tampered[2] = '7' // selects the unimplemented módulo 11 variant

	if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
		t.Fatal("expected an error for the unsupported módulo 11 utility-bill variant, got nil")
	}
}

func TestConvertToBarcode_InvalidLength(t *testing.T) {
	if _, _, err := ConvertToBarcode("12345"); err == nil {
		t.Fatal("ConvertToBarcode() error = nil, want error for a line that is not 44, 47 or 48 digits")
	}
}

func TestConvertToBarcode_InvalidCharacters_Rejected(t *testing.T) {
	barcode := buildBankSlipBarcode(t, "748", 5000, 123456, "1234567890123456789012345")
	withLetter := barcode[:5] + "A" + barcode[6:]

	if _, _, err := ConvertToBarcode(withLetter); err == nil {
		t.Fatal("expected an error for input containing a non-digit, non-separator character, got nil")
	}
}

func TestConvertToBarcode_EmptyInput_Rejected(t *testing.T) {
	if _, _, err := ConvertToBarcode(""); err == nil {
		t.Fatal("expected an error for empty input, got nil")
	}
}

func TestConvertToBarcode_StripsConventionalSeparators(t *testing.T) {
	barcode := buildBankSlipBarcode(t, "748", 5000, 123456, "1234567890123456789012345")
	spaced := barcode[:5] + "." + barcode[5:10] + " " + barcode[10:15] + "-" + barcode[15:]

	_, got, err := ConvertToBarcode(spaced)
	if err != nil {
		t.Fatalf("ConvertToBarcode() error = %v", err)
	}
	if got != barcode {
		t.Errorf("barcode = %q, want %q", got, barcode)
	}
}

// flipDigit returns a different digit from d, wrapping 9 to 0.
func flipDigit(d byte) byte {
	if d == '9' {
		return '0'
	}
	return d + 1
}
