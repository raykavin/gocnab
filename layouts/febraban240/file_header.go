package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// fileHeaderSpec is the file header record (Registro Header de Arquivo,
// tipo 0). Field positions per the FEBRABAN CNAB 240 standard.
var fileHeaderSpec = layout.RecordSpec{
	Name: "file_header",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numericConst("ServiceBatchNumber", 4, 7, "0000"),
		numericConst("RecordType", 8, 8, "0"),
		alphaFiller(9, 17),
		numeric("CompanyRegistrationKind", 18, 18, layout.KeyCompanyRegistrationKind),
		numeric("CompanyRegistrationNumber", 19, 32, layout.KeyCompanyRegistration),
		alpha("AgreementCode", 33, 52, layout.KeyAgreement),
		numeric("AccountBranch", 53, 57, layout.KeyBranch),
		alphaFiller(58, 58),
		numeric("AccountNumber", 59, 70, layout.KeyAccountNumber),
		alpha("AccountCheckDigit", 71, 71, layout.KeyAccountCheckDigit),
		alphaFiller(72, 72),
		alpha("CompanyName", 73, 102, layout.KeyCompanyName),
		alphaFiller(103, 132),
		alphaFiller(133, 142),
		numericConst("FileCode", 143, 143, "1"),
		numeric("FileGenerationDate", 144, 151, layout.KeyFileGenerationDate),
		numeric("FileGenerationTime", 152, 157, layout.KeyFileGenerationTime),
		numeric("FileSequenceNumber", 158, 163, layout.KeyFileSequenceNumber),
		numericConst("LayoutVersionNumber", 164, 166, version),
		numericConst("FileDensity", 167, 171, "0"),
		alphaFiller(172, 191),
		alphaFiller(192, 211),
		alphaFiller(212, 240),
	},
}
