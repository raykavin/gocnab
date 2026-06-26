package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// segmentBPixSpec is the Segmento B detail record used together with
// Segmento A for a PIX-by-key payment: the beneficiary's document and
// PIX key. Field positions per the FEBRABAN CNAB 240 standard. For a key
// type of CPF/CNPJ the key value is already present in FavoredCPFCNPJ;
// this reference layout still (harmlessly) repeats it into PixKeyValue
// rather than leaving that zone blank as strict fidelity would have it,
// which keeps the payment types simpler with no loss of information.
var segmentBPixSpec = layout.RecordSpec{
	Name: "segment_b_pix",
	Fields: []layout.FieldSpec{
		numericConst("BankCode", 1, 3, bankCode),
		numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
		numericConst("RecordType", 8, 8, "3"),
		numeric("SequentialNumberInBatch", 9, 13, layout.KeySequence),
		alphaConst("SegmentCode", 14, 14, "B"),
		alpha("PixKeyType", 15, 16, layout.KeyPixKeyType),
		alphaFiller(17, 17),
		numeric("FavoredRegistrationKind", 18, 18, layout.KeyPayeeDocumentKind),
		numeric("FavoredCPFCNPJ", 19, 32, layout.KeyPayeeDocument),
		alphaFiller(33, 62),
		alphaFiller(63, 127),
		alpha("PixKeyValue", 128, 226, layout.KeyPixKeyValue),
		alphaFiller(227, 232),
		alphaFiller(233, 240),
	},
}
