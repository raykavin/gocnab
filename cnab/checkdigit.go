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

// collectionMod11 computes the "módulo 11" check digit a collection
// document ("arrecadação") uses, for both its four linha digitável field
// digits and its barcode's own general digit.
//
// Same cyclic 2..9 weighting as generalMod11, different remainder mapping:
// a remainder of 0 or 1 yields check digit 0 here, where a bank slip yields
// 1. The two rules really do differ, so generalMod11 cannot stand in.
func collectionMod11(digits string) byte {
	sum := 0
	weight := 2
	for i := len(digits) - 1; i >= 0; i-- {
		sum += int(digits[i]-'0') * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}
	if remainder := sum % 11; remainder > 1 {
		return byte('0' + (11 - remainder))
	}
	return '0'
}

// collectionCheckDigit returns the check digit rule a collection document's
// "identificador de valor efetivo ou referência" selects — barcode position
// 3 (0-indexed 2) — and whether this package supports that variant.
//
// FEBRABAN pairs the four values by módulo as 6/7 (módulo 10) against 8/9
// (módulo 11), not as 6/8 against 7/9. This package had the table the wrong
// way round and validated every módulo 11 document with módulo 10, which
// made a well-formed tax or social security guide look like a corrupted
// one. The "quantidade de moeda" variants ('7' and '9') stay unsupported
// because their value field is not centavos, which is a different reason
// from the one they used to be refused for.
func collectionCheckDigit(valueTypeDigit byte) (func(string) byte, bool) {
	switch valueTypeDigit {
	case '6':
		return mod10, true
	case '8':
		return collectionMod11, true
	default:
		return nil, false
	}
}

// validateGeneralCheckDigit verifies a 44 digit barcode's own general check
// digit against the FEBRABAN rule for its segment: position 5 (0-indexed
// 4), módulo 11 over the other 43 digits, for a bank slip; position 4
// (0-indexed 3) for a collection document, over the other 43 digits, under
// whichever módulo its value-type digit selects (see collectionCheckDigit).
// A variant this package does not decode is rejected explicitly rather than
// silently accepted or misvalidated.
func validateGeneralCheckDigit(barcode string) error {
	if len(barcode) != 44 {
		return &ValidationError{Context: "ConvertToBarcode", Reason: "barcode must have 44 digits"}
	}

	if detectBarcodeSegment(barcode) == SegmentFeesOrTaxes {
		checkDigit, ok := collectionCheckDigit(barcode[2])
		if !ok {
			return &ValidationError{
				Context: "ConvertToBarcode",
				Reason:  "utility bill/tax value type is not supported",
			}
		}
		want := barcode[3]
		got := checkDigit(barcode[0:3] + barcode[4:44])
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
