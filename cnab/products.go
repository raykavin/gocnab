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
	// SupplierPayment is the "Pagamento a Fornecedores" product: credit
	// in account, TED, PIX, boleto and tax payments to suppliers.
	SupplierPayment = BatchProduct{code: "20", name: "supplier_payment"}
	// PayrollPayment is the "Pagamento de Salários" product.
	PayrollPayment = BatchProduct{code: "30", name: "payroll_payment"}
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
	// BoletoService settles payments as boleto payments.
	BoletoService = BatchService{code: "30", name: "boleto_payment"}
	// BarcodeTaxService settles payments as utility bill / barcoded tax
	// payments.
	BarcodeTaxService = BatchService{code: "22", name: "barcode_tax_payment"}
	// TaxWithoutBarcodeService settles payments as DARF/GPS tax payments
	// without a barcode.
	TaxWithoutBarcodeService = BatchService{code: "17", name: "tax_without_barcode_payment"}
)
