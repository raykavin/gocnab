package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentJSpec is the Segmento J detail record: a boleto payment
// (barcode, due date, amount). Field positions per the FEBRABAN CNAB 240
// standard.
var segmentJSpec = layout.RecordSpec{
	Name: "segment_j",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "J"),
		numeric("MovementType", 15, 15, layout.KeyMovementType),
		numeric("MovementInstructionCode", 16, 17, layout.KeyInstructionCode),
		alpha("BarCode", 18, 61, layout.KeyBarcode),
		alpha("BeneficiaryName", 62, 91, layout.KeyAssignorName),
		numeric("DueDate", 92, 99, layout.KeyDueDate),
		numericDecimal("NominalAmount", 100, 114, 2, layout.KeyDocumentAmount),
		numericDecimal("DiscountRebateAmount", 115, 129, 2, layout.KeyDiscountAmount),
		numericDecimal("PenaltyInterestAmount", 130, 144, 2, layout.KeyAdditionAmount),
		numeric("PaymentDate", 145, 152, layout.KeyPaymentDate),
		numericDecimal("PaymentAmount", 153, 167, 2, layout.KeyAmount),
		numericConst("CurrencyQuantity", 168, 182, "0"),
		alpha("ClientDocumentNumber", 183, 202, layout.KeyYourNumber),
		alphaFiller(203, 222),
		numericConst("CurrencyCode", 223, 224, "9"),
		alphaFiller(225, 230),
		alphaFiller(231, 240),
	},
}
