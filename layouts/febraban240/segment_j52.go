package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentJ52Spec is the Segmento J-52 detail record: payer and assignor
// document data complementing Segmento J. Field positions per the
// FEBRABAN CNAB 240 standard. The drawer/guarantor ("sacador/avalista")
// fields are optional per the standard and are not modeled by
// BoletoPayment; they render as filler.
var segmentJ52Spec = layout.RecordSpec{
	Name: "segment_j52",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "J"),
		alphaFiller(15, 15),
		numericConst("MovementCode", 16, 17, "0"),
		numericConst("RecordIdentification", 18, 19, "52"),
		numeric("PayerRegistrationKind", 20, 20, layout.KeyPayerDocumentKind),
		numeric("PayerCPFCNPJ", 21, 35, layout.KeyPayerDocument),
		alpha("PayerName", 36, 75, layout.KeyPayerName),
		numeric("BeneficiaryRegistrationKind", 76, 76, layout.KeyAssignorDocumentKind),
		numeric("BeneficiaryCPFCNPJ", 77, 91, layout.KeyAssignorDocument),
		alpha("BeneficiaryName", 92, 131, layout.KeyAssignorName),
		numericConst("DrawerGuarantorRegistrationKind", 132, 132, "0"),
		numericConst("DrawerGuarantorCPFCNPJ", 133, 147, "0"),
		alphaFiller(148, 187),
		alphaFiller(188, 240),
	},
}
