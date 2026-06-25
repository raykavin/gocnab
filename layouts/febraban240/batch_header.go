package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// batchHeaderSpec is the batch header record (Registro Header de Lote,
// tipo 1), generic across credit-in-account, TED and PIX batches. Field
// positions per the FEBRABAN CNAB 240 standard.
var batchHeaderSpec = layout.RecordSpec{
	Name: "batch_header",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "1"),
		alphaConst("OperationType", 9, 9, "C"),
		numeric("ServiceType", 10, 11, layout.KeyBatchProductCode),
		numeric("PaymentMethod", 12, 13, layout.KeyBatchServiceCode),
		numericConst("BatchLayoutVersion", 14, 16, version),
		alphaFiller(17, 17),
		numeric("CompanyRegistrationKind", 18, 18, layout.KeyCompanyRegistrationKind),
		numeric("CompanyRegistrationNumber", 19, 32, layout.KeyCompanyRegistration),
		alpha("AgreementCode", 33, 52, layout.KeyAgreement),
		numeric("AccountBranch", 53, 57, layout.KeyBranch),
		alphaFiller(58, 58),
		numeric("AccountNumber", 59, 70, layout.KeyAccountNumber),
		alpha("AccountCheckDigit", 71, 71, layout.KeyAccountCheckDigit),
		alphaFiller(72, 72),
		alpha("CompanyName", 73, 102, layout.KeyCompanyName),
		alphaFiller(103, 142),
		alphaFiller(143, 172),
		numericFiller(173, 177),
		alphaFiller(178, 192),
		alphaFiller(193, 212),
		numericFiller(213, 217),
		numericFiller(218, 220),
		alphaFiller(221, 222),
		alphaFiller(223, 230),
		alphaFiller(231, 240),
	},
}
