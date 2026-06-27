package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentNSimpleSpec is the Segmento N detail record for a DARF Simples
// tax payment, carried as a single total amount. Field positions per the
// FEBRABAN CNAB 240 standard; the gross-revenue tracking fields are not
// modeled by DARFSimple and render as filler.
var segmentNSimpleSpec = layout.RecordSpec{
	Name: "segment_n_simple",
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
		numeric("RevenueCode", 111, 114, layout.KeyTaxCode),
		alphaFiller(115, 116),
		numeric("TaxpayerIdType", 117, 118, layout.KeyTaxpayerIdType),
		numeric("TaxpayerId", 119, 132, layout.KeyTaxpayerDocument),
		alphaConst("TaxIdentificationCode", 133, 134, "18"),
		numericConst("AssessmentPeriod", 135, 142, "0"),
		numericConst("AccumulatedGrossRevenueAmount", 143, 157, "0"),
		numericConst("GrossRevenuePercentage", 158, 164, "0"),
		numericDecimal("PrincipalAmount", 165, 179, 2, layout.KeyPrincipalAmount),
		numericConst("PenaltyAmount", 180, 194, "0"),
		numericConst("InterestChargesAmount", 195, 209, "0"),
		alphaFiller(210, 230),
		alphaFiller(231, 240),
	},
}
