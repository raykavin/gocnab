package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentNSocialSpec is the Segmento N detail record for a GPS (Guia da
// Previdência Social) payment. Field positions per the FEBRABAN CNAB 240
// standard. RevenueCode is fixed since GPS is not modeled with a
// caller-supplied revenue code.
var segmentNSocialSpec = layout.RecordSpec{
	Name: "segment_n_social",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "N"),
		numeric("MovementType", 15, 15, layout.KeyMovementType),
		numeric("MovementInstructionCode", 16, 17, layout.KeyInstructionCode),
		alpha("ClientDocumentNumber", 18, 37, layout.KeyYourNumber),
		alphaFiller(38, 57),
		alpha("TaxpayerName", 58, 87, layout.KeyTaxpayerName),
		numeric("PaymentDate", 88, 95, layout.KeyPaymentDate),
		numericDecimal("PaymentTotalAmount", 96, 110, 2, layout.KeyAmount),
		numericConst("RevenueCode", 111, 114, "1200"),
		alphaFiller(115, 116),
		numeric("TaxpayerIdType", 117, 118, layout.KeyTaxpayerIdType),
		numeric("TaxpayerId", 119, 132, layout.KeyTaxpayerDocument),
		alphaConst("TaxIdentificationCode", 133, 134, "17"),
		numeric("CompetenceMonthYear", 135, 140, layout.KeyPeriod),
		numericDecimal("INSSExpectedAmount", 141, 155, 2, layout.KeyPrincipalAmount),
		numericConst("OtherEntitiesAmount", 156, 170, "0"),
		numericConst("MonetaryUpdateAmount", 171, 185, "0"),
		alphaFiller(186, 230),
		alphaFiller(231, 240),
	},
}
