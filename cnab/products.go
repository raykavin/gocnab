package cnab

// BatchProduct identifies the kind of payment product a batch carries
// (the FEBRABAN "tipo de serviço"). Use one of the predefined values;
// BatchProduct has no exported fields so a caller cannot construct an
// invalid one by accident.
type BatchProduct struct {
	code string
	name string
}

// String returns the product's descriptive name.
func (p BatchProduct) String() string { return p.name }

var (
	// SupplierPayment is the "Pagamento Fornecedor" product: credit in
	// account, DOC, OP, TED and PIX payments to suppliers. Some banks
	// (e.g. Sicredi) require boleto and tax/utility payments to use their
	// own distinct product codes (see BoletoCollection and TaxPayment)
	// rather than being nested under this one, even though they also
	// ultimately pay a supplier; confirm against your bank's manual
	// before assuming this code covers every supplier payment method.
	SupplierPayment = BatchProduct{code: "20", name: "supplier_payment"}
	// PayrollPayment is the "Pagamento de Salários" product.
	PayrollPayment = BatchProduct{code: "30", name: "payroll_payment"}
	// BoletoCollection is the "Boleto Eletrônico"/cobrança product,
	// required by some banks (e.g. Sicredi) for boleto payments
	// (Segmento J/J-52) instead of SupplierPayment.
	BoletoCollection = BatchProduct{code: "03", name: "boleto_collection"}
	// TaxPayment is the "Pagamento de Contas, Tributos e Impostos"
	// product, required by some banks (e.g. Sicredi) for utility bill and
	// barcoded tax payments (Segmento O) instead of SupplierPayment.
	TaxPayment = BatchProduct{code: "22", name: "tax_payment"}
)

// BatchService identifies how the payments in a batch are settled (the
// FEBRABAN "forma de lançamento"). Use one of the predefined values;
// BatchService has no exported fields so a caller cannot construct an
// invalid one by accident.
type BatchService struct {
	code string
	name string
}

// String returns the service's descriptive name.
func (s BatchService) String() string { return s.name }

var (
	// CreditInAccount settles payments as a same-bank credit in account.
	CreditInAccount = BatchService{code: "01", name: "credit_in_account"}
	// TEDTransfer settles payments as a TED wire transfer.
	TEDTransfer = BatchService{code: "41", name: "ted_transfer"}
	// PixTransfer settles payments as a PIX transfer, by key or by bank
	// account data.
	PixTransfer = BatchService{code: "45", name: "pix_transfer"}
	// BoletoService settles a boleto issued by the paying bank itself
	// ("Liquidação de Títulos do Próprio Banco"). Use OtherBankBoletoService
	// instead for a boleto issued by a different bank: some banks (e.g.
	// Sicredi) reject a boleto batch that uses the wrong one of the two,
	// even though BoletoPayment itself renders identically either way —
	// only the batch's service code differs; which one to use depends on
	// which bank issued the boleto being paid, not on any property of the
	// payment.
	BoletoService = BatchService{code: "30", name: "boleto_payment"}
	// OtherBankBoletoService settles a boleto issued by a bank other than
	// the one the remittance is sent to ("Pagamento de Títulos de Outros
	// Bancos"). See BoletoService.
	OtherBankBoletoService = BatchService{code: "31", name: "other_bank_boleto_payment"}
	// BarcodeTaxService settles payments as utility bill / barcoded tax
	// payments.
	BarcodeTaxService = BatchService{code: "22", name: "barcode_tax_payment"}
	// TaxWithoutBarcodeService settles payments as DARF/GPS tax payments
	// without a barcode.
	TaxWithoutBarcodeService = BatchService{code: "17", name: "tax_without_barcode_payment"}
)
