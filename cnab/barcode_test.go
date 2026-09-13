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
// (value-type digit '6' at position 3, 1-indexed, index 2).
func buildTaxBarcode(value int64, freeField string) string {
	return buildTaxBarcodeVariant('6', value, freeField)
}

// buildTaxBarcodeVariant assembles a 44-digit collection barcode for a
// given value-type digit, check-digited with whichever módulo that digit
// selects: productID(1) + segment(1, unvalidated by this package) +
// valueType(1) + generalDV(1) + value(11) + freeField(29). '6' is the
// módulo 10 variant, '8' the módulo 11 one.
func buildTaxBarcodeVariant(valueType byte, value int64, freeField string) string {
	if len(freeField) != 29 {
		panic("buildTaxBarcodeVariant: freeField must have 29 digits")
	}
	checkDigit, ok := collectionCheckDigit(valueType)
	if !ok {
		panic("buildTaxBarcodeVariant: unsupported value type")
	}
	head := "8" + "1" + string(valueType)          // index 0 (product), 1 (segment), 2 (value type)
	tail := zeroPad(itoa64(value), 11) + freeField // index 4..43
	dv := checkDigit(head + tail)
	return head + string(dv) + tail
}

// buildTaxLine converts a valid 44-digit tax barcode into its 48-digit
// typeable line: four 12-character fields, each 11 barcode data digits
// followed by its own check digit, under the rule the barcode's value-type
// digit selects.
func buildTaxLine(t *testing.T, barcode string) string {
	t.Helper()
	if len(barcode) != 44 {
		t.Fatalf("bad test input: barcode=%q", barcode)
	}
	checkDigit, ok := collectionCheckDigit(barcode[2])
	if !ok {
		t.Fatalf("bad test input: unsupported value type %q", barcode[2])
	}
	line := make([]byte, 0, 48)
	for i := 0; i < 4; i++ {
		data := barcode[i*11 : i*11+11]
		line = append(line, data...)
		line = append(line, checkDigit(data))
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

// TestConvertToBarcode_TaxBarcode_UnsupportedValueTypes_Rejected pins the
// value-type digits this package does not decode. '7' and '9' are the
// "quantidade de moeda" variants, whose value field is not centavos.
//
// This test used to call '7' "the unimplemented módulo 11 variant", which
// had FEBRABAN's table backwards: the módulo is selected by 6/7 against
// 8/9, so '8' is a módulo 11 document — and it was being validated with
// módulo 10, which rejected well-formed tax guides.
func TestConvertToBarcode_TaxBarcode_UnsupportedValueTypes_Rejected(t *testing.T) {
	barcode := buildTaxBarcode(1000, "00000000000000000000000000000"[:29])
	for _, valueType := range []byte{'0', '5', '7', '9'} {
		t.Run(string(valueType), func(t *testing.T) {
			tampered := []byte(barcode)
			tampered[2] = valueType
			if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
				t.Fatalf("expected an error for value type %q, got nil", valueType)
			}
		})
	}
}

// TestConvertToBarcode_TaxMod11_RealGPSDocument is the regression test for
// the reported failure, on the CNAB side: the same GPS guide that
// pkg/boleto rejected also had to reach Segmento O generation, and
// taxLineToBarcode carried an identical copy of the same défaut.
func TestConvertToBarcode_TaxMod11_RealGPSDocument(t *testing.T) {
	const (
		line    = "858000001239061603852620610716262474385997310225"
		barcode = "85800000123061603852626107162624738599731022"
	)

	t.Run("typeable line", func(t *testing.T) {
		segment, got, err := ConvertToBarcode(line)
		if err != nil {
			t.Fatalf("ConvertToBarcode(line) error = %v, want success", err)
		}
		if segment != SegmentFeesOrTaxes {
			t.Errorf("segment = %v, want SegmentFeesOrTaxes", segment)
		}
		if got != barcode {
			t.Errorf("barcode = %s, want %s", got, barcode)
		}
	})

	t.Run("barcode", func(t *testing.T) {
		segment, got, err := ConvertToBarcode(barcode)
		if err != nil {
			t.Fatalf("ConvertToBarcode(barcode) error = %v, want success", err)
		}
		if segment != SegmentFeesOrTaxes {
			t.Errorf("segment = %v, want SegmentFeesOrTaxes", segment)
		}
		if got != barcode {
			t.Errorf("barcode = %s, want %s", got, barcode)
		}
	})
}

// TestConvertToBarcode_TaxMod11_RoundTripAndTampering exercises the módulo
// 11 variant through the same shape the módulo 10 one uses, and keeps the
// fix from degenerating into accepting anything.
func TestConvertToBarcode_TaxMod11_RoundTripAndTampering(t *testing.T) {
	barcode := buildTaxBarcodeVariant('8', 4599, "00000000000000000000000000000"[:29])
	line := buildTaxLine(t, barcode)

	segment, got, err := ConvertToBarcode(line)
	if err != nil {
		t.Fatalf("ConvertToBarcode() error = %v", err)
	}
	if segment != SegmentFeesOrTaxes || got != barcode {
		t.Errorf("got (%v, %s), want (SegmentFeesOrTaxes, %s)", segment, got, barcode)
	}

	tampered := []byte(line)
	tampered[11] = '0' + (tampered[11]-'0'+1)%10
	if _, _, err := ConvertToBarcode(string(tampered)); err == nil {
		t.Error("expected an error for a tampered módulo 11 field check digit, got nil")
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
