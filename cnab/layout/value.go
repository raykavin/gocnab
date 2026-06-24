package layout

// Key identifies a semantic value shared between the public cnab package
// (which knows about companies, payees and payments) and a bank Layout
// (which knows only about column positions). A Layout binds a FieldSpec to
// a Key; any code that produces Values for that Layout must use the same
// vocabulary. This indirection is what lets new bank layouts be added
// without any change to the engine or to existing payment types: a new
// layout is just a new set of FieldSpec.Key bindings over the same Key
// vocabulary, extended additively when a genuinely new concept appears.
type Key string

// Values holds the semantic data for a single record, keyed by Key. It is
// the only data structure that crosses the boundary between the public
// domain API and a bank Layout.
type Values map[Key]any

// Structural keys are written by the engine itself while rendering a
// batch: NSA, the running batch number, the record sequence inside a
// batch, and the trailer totals. Code outside internal/engine must never
// set these keys; any value supplied for them is overwritten before
// rendering.
const (
	// KeyBatchNumber is the running "lote de serviço" number: 0001 for the
	// first batch, incrementing by one per batch. The file header and file
	// trailer use fixed 0000/9999 constants instead of this key.
	KeyBatchNumber Key = "sys.batch_number"
	// KeySequence is the record sequence number inside its batch,
	// restarting at 1 on every new batch header.
	KeySequence Key = "sys.sequence"
	// KeyBatchRecordCount is the total number of records in a batch
	// (header + details + trailer), written on the batch trailer.
	KeyBatchRecordCount Key = "sys.batch_record_count"
	// KeyBatchAmount is the sum of every KeyAmount value found in a
	// batch's detail segments, written on the batch trailer.
	KeyBatchAmount Key = "sys.batch_amount"
	// KeyBatchCount is the total number of batches in the file, written on
	// the file trailer.
	KeyBatchCount Key = "sys.batch_count"
	// KeyFileRecordCount is the total number of records in the file
	// (file header + every batch header/detail/trailer + file trailer),
	// written on the file trailer.
	KeyFileRecordCount Key = "sys.file_record_count"
)

// Semantic keys are written by the public cnab package from domain values
// (Company, Account, Payee, Payment) and read by a Layout's FieldSpec.Key
// bindings. The list below is the vocabulary the bundled febraban240
// reference layout binds to; a new bank layout should reuse these keys and
// only introduce a new one when no existing key carries the right meaning.
const (
	// KeyFileSequenceNumber is the file sequence number (NSA) supplied by
	// the caller in Config.
	KeyFileSequenceNumber Key = "file_sequence_number"
	// KeyFileGenerationDate is the date the file was generated.
	KeyFileGenerationDate Key = "file_generation_date"
	// KeyFileGenerationTime is the time of day the file was generated.
	KeyFileGenerationTime Key = "file_generation_time"

	// KeyCompanyRegistrationKind is "1" for CPF or "2" for CNPJ.
	KeyCompanyRegistrationKind Key = "company_registration_kind"
	// KeyCompanyRegistration is the company CPF/CNPJ, digits only.
	KeyCompanyRegistration Key = "company_registration"
	// KeyCompanyName is the company name.
	KeyCompanyName Key = "company_name"
	// KeyAgreement is the company's agreement ("convênio") code with the
	// bank.
	KeyAgreement Key = "agreement"
	// KeyBranch is the bank branch ("agência") number.
	KeyBranch Key = "branch"
	// KeyAccountNumber is the bank account number.
	KeyAccountNumber Key = "account_number"
	// KeyAccountCheckDigit is the bank account check digit.
	KeyAccountCheckDigit Key = "account_check_digit"

	// KeyBatchProductCode is the "tipo de serviço" code of a batch (e.g.
	// supplier payments, payroll).
	KeyBatchProductCode Key = "batch_product_code"
	// KeyBatchServiceCode is the "forma de lançamento" code of a batch
	// (e.g. credit in account, TED, PIX transfer).
	KeyBatchServiceCode Key = "batch_service_code"

	// KeyMovementType is "0" for a normal inclusion or "9" to cancel a
	// previously sent detail entry.
	KeyMovementType Key = "movement_type"
	// KeyInstructionCode is the movement instruction code; cancellations
	// use "99".
	KeyInstructionCode Key = "instruction_code"
	// KeyClearingCode is the settlement clearing house code on Segmento A
	// ("Código da Câmara de Compensação"): "000" for a same-bank credit,
	// "018" for TED, "009" for PIX. Each payment type that uses Segmento A
	// sets this to the value matching how it settles.
	KeyClearingCode Key = "clearing_code"

	// KeyBeneficiaryBankCode is the beneficiary's bank code (used for TED
	// and PIX by bank data, when the beneficiary is at a different bank).
	KeyBeneficiaryBankCode Key = "beneficiary_bank_code"
	// KeyBeneficiaryBranch is the beneficiary's bank branch.
	KeyBeneficiaryBranch Key = "beneficiary_branch"
	// KeyBeneficiaryAccount is the beneficiary's bank account number.
	KeyBeneficiaryAccount Key = "beneficiary_account"
	// KeyBeneficiaryCheckDigit is the beneficiary's account check digit.
	KeyBeneficiaryCheckDigit Key = "beneficiary_check_digit"
	// KeyPayeeName is the payee/beneficiary name.
	KeyPayeeName Key = "payee_name"
	// KeyPayeeDocumentKind is "1" for CPF or "2" for CNPJ.
	KeyPayeeDocumentKind Key = "payee_document_kind"
	// KeyPayeeDocument is the payee CPF/CNPJ, digits only.
	KeyPayeeDocument Key = "payee_document"
	// KeyYourNumber is the payer's own document/reference number for the
	// payment ("seu número").
	KeyYourNumber Key = "your_number"
	// KeyAmount is the payment amount, in Cents.
	KeyAmount Key = "amount"
	// KeyPaymentDate is the date the payment should be settled.
	KeyPaymentDate Key = "payment_date"
	// KeyPurposeCode is the TED purpose ("finalidade") code.
	KeyPurposeCode Key = "purpose_code"

	// KeyPixKeyType identifies which PIX key variant KeyPixKeyValue holds
	// ("phone", "email", "cpf", "cnpj" or "random").
	KeyPixKeyType Key = "pix_key_type"
	// KeyPixKeyValue is the PIX key value itself.
	KeyPixKeyValue Key = "pix_key_value"

	// KeyPayeeAddressStreet, KeyPayeeAddressNumber, KeyPayeeAddressDistrict,
	// KeyPayeeAddressCity, KeyPayeeAddressState and KeyPayeeAddressZipCode
	// are the payee address fields carried on Segment B.
	KeyPayeeAddressStreet   Key = "payee_address_street"
	KeyPayeeAddressNumber   Key = "payee_address_number"
	KeyPayeeAddressDistrict Key = "payee_address_district"
	KeyPayeeAddressCity     Key = "payee_address_city"
	KeyPayeeAddressState    Key = "payee_address_state"
	KeyPayeeAddressZipCode  Key = "payee_address_zip_code"

	// KeyBarcode is the 44 digit barcode ("código de barras"/"linha
	// digitável") of a boleto or utility bill.
	KeyBarcode Key = "barcode"
	// KeyDueDate is the document due date.
	KeyDueDate Key = "due_date"
	// KeyDocumentAmount is the nominal ("de face") amount of a boleto or
	// bill, before discounts and additions.
	KeyDocumentAmount Key = "document_amount"
	// KeyDiscountAmount is a discount applied to a boleto payment.
	KeyDiscountAmount Key = "discount_amount"
	// KeyAdditionAmount is interest/fine added to a boleto payment.
	KeyAdditionAmount Key = "addition_amount"
	// KeyPayerDocumentKind is "1" for CPF or "2" for CNPJ, for the boleto
	// payer.
	KeyPayerDocumentKind Key = "payer_document_kind"
	// KeyPayerDocument is the boleto payer CPF/CNPJ, digits only.
	KeyPayerDocument Key = "payer_document"
	// KeyPayerName is the boleto payer name.
	KeyPayerName Key = "payer_name"
	// KeyAssignorDocumentKind is "1" for CPF or "2" for CNPJ, for the
	// boleto assignor ("cedente").
	KeyAssignorDocumentKind Key = "assignor_document_kind"
	// KeyAssignorDocument is the boleto assignor CPF/CNPJ, digits only.
	KeyAssignorDocument Key = "assignor_document"
	// KeyAssignorName is the boleto assignor ("cedente") name.
	KeyAssignorName Key = "assignor_name"

	// KeyTaxCode is the revenue code ("código da receita") of a DARF/GPS
	// payment.
	KeyTaxCode Key = "tax_code"
	// KeyTaxpayerDocumentKind is "1" for CPF or "2" for CNPJ, for the
	// taxpayer.
	KeyTaxpayerDocumentKind Key = "taxpayer_document_kind"
	// KeyTaxpayerIdType is the tax segment's own taxpayer identification
	// type code, a different domain than KeyTaxpayerDocumentKind ("1" for
	// CNPJ or "2" for CPF, per the DARF/GPS segment convention).
	KeyTaxpayerIdType Key = "taxpayer_id_type"
	// KeyTaxpayerDocument is the taxpayer CPF/CNPJ, digits only.
	KeyTaxpayerDocument Key = "taxpayer_document"
	// KeyTaxpayerName is the taxpayer name.
	KeyTaxpayerName Key = "taxpayer_name"
	// KeyReferenceNumber is the tax reference number ("número de
	// referência"), used by DARF Simples and GPS.
	KeyReferenceNumber Key = "reference_number"
	// KeyPeriod is the assessment period ("período de apuração" /
	// "competência") of a tax payment.
	KeyPeriod Key = "period"
	// KeyPrincipalAmount is the principal tax amount.
	KeyPrincipalAmount Key = "principal_amount"
	// KeyFineAmount is the fine ("multa") portion of a tax payment.
	KeyFineAmount Key = "fine_amount"
	// KeyInterestAmount is the interest ("juros") portion of a tax
	// payment.
	KeyInterestAmount Key = "interest_amount"
)

// AllKeys lists every Key constant defined in this file, structural and
// semantic alike. Keep it in sync when adding a new Key: it backs
// IsKnownKey, which a config-driven Layout loader (see NewFromJSON) uses
// to catch a typo in a hand-written descriptor at load time instead of
// silently accepting an unrecognized key.
var AllKeys = []Key{
	KeyBatchNumber, KeySequence, KeyBatchRecordCount, KeyBatchAmount, KeyBatchCount, KeyFileRecordCount,

	KeyFileSequenceNumber, KeyFileGenerationDate, KeyFileGenerationTime,
	KeyCompanyRegistrationKind, KeyCompanyRegistration, KeyCompanyName, KeyAgreement,
	KeyBranch, KeyAccountNumber, KeyAccountCheckDigit,
	KeyBatchProductCode, KeyBatchServiceCode,
	KeyMovementType, KeyInstructionCode, KeyClearingCode,
	KeyBeneficiaryBankCode, KeyBeneficiaryBranch, KeyBeneficiaryAccount, KeyBeneficiaryCheckDigit,
	KeyPayeeName, KeyPayeeDocumentKind, KeyPayeeDocument, KeyYourNumber, KeyAmount, KeyPaymentDate, KeyPurposeCode,
	KeyPixKeyType, KeyPixKeyValue,
	KeyPayeeAddressStreet, KeyPayeeAddressNumber, KeyPayeeAddressDistrict,
	KeyPayeeAddressCity, KeyPayeeAddressState, KeyPayeeAddressZipCode,
	KeyBarcode, KeyDueDate, KeyDocumentAmount, KeyDiscountAmount, KeyAdditionAmount,
	KeyPayerDocumentKind, KeyPayerDocument, KeyPayerName,
	KeyAssignorDocumentKind, KeyAssignorDocument, KeyAssignorName,
	KeyTaxCode, KeyTaxpayerDocumentKind, KeyTaxpayerIdType, KeyTaxpayerDocument, KeyTaxpayerName,
	KeyReferenceNumber, KeyPeriod, KeyPrincipalAmount, KeyFineAmount, KeyInterestAmount,
}

// IsKnownKey reports whether k is one of the values in AllKeys.
func IsKnownKey(k Key) bool {
	for _, known := range AllKeys {
		if known == k {
			return true
		}
	}
	return false
}
