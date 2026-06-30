package cnab

import "strings"

// Payee identifies the beneficiary of a payment: who gets paid.
type Payee struct {
	// Name is the beneficiary's name.
	Name string
	// Registration is the beneficiary's CPF or CNPJ.
	Registration Document
}

func (p Payee) validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return &ValidationError{Context: "Payee", Reason: "Name is required"}
	}
	if p.Registration == nil {
		return &ValidationError{Context: "Payee", Reason: "Registration (CNPJ or CPF) is required"}
	}
	return nil
}
