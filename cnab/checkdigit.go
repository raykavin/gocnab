package cnab

// mod10 computes the FEBRABAN "módulo 10" check digit for a digit string:
// walking right to left, digits are alternately multiplied by 2 and 1
// (starting with 2 on the rightmost digit); any product of 10 or more has
// its digits summed (equivalent to subtracting 9, since the maximum
// product is 18). The check digit is (10 - (sum mod 10)) mod 10.
func mod10(digits string) byte {
	sum := 0
	weight := 2
	for i := len(digits) - 1; i >= 0; i-- {
		product := int(digits[i]-'0') * weight
		if product > 9 {
			product -= 9
		}
		sum += product
		if weight == 2 {
			weight = 1
		} else {
			weight = 2
		}
	}
	remainder := sum % 10
	if remainder == 0 {
		return '0'
	}
	return byte('0' + (10 - remainder))
}

// generalMod11 computes the FEBRABAN "módulo 11" general check digit used
// by a bank slip barcode: walking right to left, digits are multiplied by
// cyclically repeating weights 2..9, summed, and reduced mod 11. A
// remainder of 0 or 1 maps to check digit 1 (avoiding the invalid,
// two-digit results 11-0=11 and 11-1=10); any other remainder maps to
// 11 - remainder.
func generalMod11(digits string) byte {
	sum := 0
	weight := 2
	for i := len(digits) - 1; i >= 0; i-- {
		sum += int(digits[i]-'0') * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}
	remainder := sum % 11
	if remainder == 0 || remainder == 1 {
		return '1'
	}
	return byte('0' + (11 - remainder))
}

// validateGeneralCheckDigit verifies a 44 digit barcode's own general check
// digit against the FEBRABAN rule for its segment: position 5 (0-indexed
// 4), módulo 11 over the other 43 digits, for a bank slip; position 4
// (0-indexed 3), módulo 10 over the other 43 digits, for the common
// utility-bill/tax variant (value-type digit '6' or '8' at position 3,
// 0-indexed 2). The far less common módulo 11 utility-bill variant
// (value-type digit '7' or '9') is not implemented and is rejected
// explicitly rather than silently accepted or misvalidated.
func validateGeneralCheckDigit(barcode string) error {
	if len(barcode) != 44 {
		return &ValidationError{Context: "ConvertToBarcode", Reason: "barcode must have 44 digits"}
	}

	if detectBarcodeSegment(barcode) == SegmentFeesOrTaxes {
		valueTypeDigit := barcode[2]
		if valueTypeDigit != '6' && valueTypeDigit != '8' {
			return &ValidationError{
				Context: "ConvertToBarcode",
				Reason:  "módulo 11 utility bill/tax variant is not supported",
			}
		}
		want := barcode[3]
		got := mod10(barcode[0:3] + barcode[4:44])
		if got != want {
			return &ValidationError{Context: "ConvertToBarcode", Reason: "general check digit does not match the barcode"}
		}
		return nil
	}

	want := barcode[4]
	got := generalMod11(barcode[0:4] + barcode[5:44])
	if got != want {
		return &ValidationError{Context: "ConvertToBarcode", Reason: "general check digit does not match the barcode"}
	}
	return nil
}
