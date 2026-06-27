package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentNSpec is the Segmento N detail record for a normal DARF tax
// payment, with principal, fine and interest tracked separately. Field
// positions per the FEBRABAN CNAB 240 standard.
var segmentNSpec = layout.RecordSpec{
	Name: "segment_n",
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
		alphaConst("TaxIdentificationCode", 133, 134, "16"),
		numeric("AssessmentPeriod", 135, 142, layout.KeyPeriod),
		numeric("ReferenceNumber", 143, 159, layout.KeyReferenceNumber),
		numericDecimal("PrincipalAmount", 160, 174, 2, layout.KeyPrincipalAmount),
		numericDecimal("PenaltyAmount", 175, 189, 2, layout.KeyFineAmount),
		numericDecimal("InterestChargesAmount", 190, 204, 2, layout.KeyInterestAmount),
		numeric("DueDate", 205, 212, layout.KeyDueDate),
		alphaFiller(213, 230),
		alphaFiller(231, 240),
	},
}
