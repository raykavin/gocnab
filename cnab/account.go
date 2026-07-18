package cnab

import "strings"

// Account identifies the bank account a remittance is debited from (the
// company's account) or, embedded in a payment, credited to (the
// beneficiary's account).
type Account struct {
	// Branch is the bank branch ("agência") number.
	Branch string
	// Number is the account number, without the check digit.
	Number string
	// CheckDigit is the account check digit.
	CheckDigit string
}

func (a Account) validate() error {
	if strings.TrimSpace(a.Branch) == "" {
		return &ValidationError{Context: "Account", Reason: "Branch is required"}
	}
	if strings.TrimSpace(a.Number) == "" {
		return &ValidationError{Context: "Account", Reason: "Number is required"}
	}
	if strings.TrimSpace(a.CheckDigit) == "" {
		return &ValidationError{Context: "Account", Reason: "CheckDigit is required"}
	}
	return nil
}
