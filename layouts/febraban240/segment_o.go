package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentOSpec is the Segmento O detail record: a utility bill or tax
// payment identified by a barcode. Field positions per the FEBRABAN CNAB
// 240 standard.
var segmentOSpec = layout.RecordSpec{
	Name: "segment_o",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "O"),
		numeric("MovementType", 15, 15, layout.KeyMovementType),
		numeric("MovementInstructionCode", 16, 17, layout.KeyInstructionCode),
		alpha("BarCode", 18, 61, layout.KeyBarcode),
		alphaFiller(62, 91),
		numeric("DueDate", 92, 99, layout.KeyDueDate),
		numeric("PaymentDate", 100, 107, layout.KeyPaymentDate),
		numericDecimal("PaymentTotalAmount", 108, 122, 2, layout.KeyAmount),
		alpha("ClientDocumentNumber", 123, 142, layout.KeyYourNumber),
		alphaFiller(143, 162),
		alphaFiller(163, 230),
		alphaFiller(231, 240),
	},
}
