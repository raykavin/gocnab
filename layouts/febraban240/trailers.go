package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// batchTrailerSpec is the batch trailer record (Registro Trailer de
// Lote, tipo 5). Field positions per the FEBRABAN CNAB 240 standard.
var batchTrailerSpec = layout.RecordSpec{
	Name: "batch_trailer",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "5"),
		alphaFiller(9, 17),
		numeric("BatchRecordCount", 18, 23, layout.KeyBatchRecordCount),
		numericDecimal("BatchTotalAmount", 24, 41, 2, layout.KeyBatchAmount),
		numericConst("BatchTotalCurrencyQuantity", 42, 59, "0"),
		numericConst("DebitNoticeNumber", 60, 65, "0"),
		alphaFiller(66, 230),
		alphaFiller(231, 240),
	},
}

// fileTrailerSpec is the file trailer record (Registro Trailer de
// Arquivo, tipo 9). Field positions per the FEBRABAN CNAB 240 standard.
var fileTrailerSpec = layout.RecordSpec{
	Name: "file_trailer",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numericConst("ServiceBatchNumber", 4, 7, "9999"),
		numericConst("RecordType", 8, 8, "9"),
		alphaFiller(9, 17),
		numeric("FileBatchCount", 18, 23, layout.KeyBatchCount),
		numeric("FileRecordCount", 24, 29, layout.KeyFileRecordCount),
		numericConst("ReconciliationAccountCount", 30, 35, "0"),
		alphaFiller(36, 240),
	},
}
