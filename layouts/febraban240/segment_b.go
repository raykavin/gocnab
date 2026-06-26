package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentBSpec is the Segmento B detail record used together with
// Segmento A for a plain credit-in-account, TED or PIX-by-bank-data
// payment: the beneficiary's document and address. Field positions per
// the FEBRABAN CNAB 240 standard. Most address fields are optional per
// the standard and are not modeled by this SDK's Payee type yet; they
// render as filler until a caller has a reason to populate them.
var segmentBSpec = layout.RecordSpec{
	Name: "segment_b",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "B"),
		alphaFiller(15, 17),
		numeric("FavoredRegistrationKind", 18, 18, layout.KeyPayeeDocumentKind),
		numeric("FavoredCPFCNPJ", 19, 32, layout.KeyPayeeDocument),
		alpha("AddressStreet", 33, 62, layout.KeyPayeeAddressStreet),
		numeric("AddressNumber", 63, 67, layout.KeyPayeeAddressNumber),
		alphaFiller(68, 82),
		alpha("Neighborhood", 83, 97, layout.KeyPayeeAddressDistrict),
		alpha("City", 98, 117, layout.KeyPayeeAddressCity),
		numeric("ZipCode", 118, 125, layout.KeyPayeeAddressZipCode),
		alpha("State", 126, 127, layout.KeyPayeeAddressState),
		numericConst("DueDate", 128, 135, "0"),
		numericConst("DocumentAmount", 136, 150, "0"),
		numericConst("RebateAmount", 151, 165, "0"),
		numericConst("DiscountAmount", 166, 180, "0"),
		numericConst("LateFeeAmount", 181, 195, "0"),
		numericConst("PenaltyAmount", 196, 210, "0"),
		numericFiller(211, 225),
		alphaConst("NotifyFavored", 226, 226, "0"),
		numericFiller(227, 232), // "uso exclusivo para o SIAPE" in some bank manuals
		numericFiller(233, 240), // "código ISPB" in some bank manuals
	},
}
