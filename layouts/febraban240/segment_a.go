package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentASpec is the Segmento A detail record: the core payment
// instruction (beneficiary bank data, amount, payment date). Used by
// credit-in-account, TED and PIX payments. Field positions per the
// FEBRABAN CNAB 240 standard.
var segmentASpec = layout.RecordSpec{
	Name: "segment_a",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "A"),
		numeric("MovementType", 15, 15, layout.KeyMovementType),
		numeric("MovementInstructionCode", 16, 17, layout.KeyInstructionCode),
		numeric("ClearingCode", 18, 20, layout.KeyClearingCode),
		numeric("FavoredBankCode", 21, 23, layout.KeyBeneficiaryBankCode),
		numeric("FavoredBranch", 24, 28, layout.KeyBeneficiaryBranch),
		alphaFiller(29, 29),
		numeric("FavoredAccountNumber", 30, 41, layout.KeyBeneficiaryAccount),
		alpha("FavoredAccountCheckDigit", 42, 42, layout.KeyBeneficiaryCheckDigit),
		alphaFiller(43, 43),
		alpha("FavoredName", 44, 73, layout.KeyPayeeName),
		alpha("ClientDocumentNumber", 74, 93, layout.KeyYourNumber),
		numeric("PaymentDate", 94, 101, layout.KeyPaymentDate),
		alphaConst("CurrencyType", 102, 104, "BRL"),
		numericConst("CurrencyQuantity", 105, 119, "0"),
		numericDecimal("PaymentAmount", 120, 134, 2, layout.KeyAmount),
		alphaFiller(135, 154),
		numericConst("ActualPaymentDate", 155, 162, "0"),
		numericConst("ActualPaymentAmount", 163, 177, "0"),
		alphaFiller(178, 217),
		alphaFiller(218, 219),
		alpha("TEDPurposeCode", 220, 224, layout.KeyPurposeCode),
		alphaFiller(225, 226),
		alphaFiller(227, 229),
		alphaConst("NotifyFavored", 230, 230, "0"),
		alphaFiller(231, 240),
	},
}
