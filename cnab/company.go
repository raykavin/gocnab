package cnab

import "strings"

// Company identifies the entity sending the remittance file: the payer.
type Company struct {
	// Name is the company's legal name.
	Name string
	// Registration is the company's CNPJ (or, less commonly, CPF for an
	// individual acting as the payer).
	Registration Document
	// Agreement is the company's agreement ("convênio") code with the
	// bank.
	Agreement string
}

func (c Company) validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return &ValidationError{Context: "Company", Reason: "Name is required"}
	}
	if c.Registration == nil {
		return &ValidationError{Context: "Company", Reason: "Registration (CNPJ or CPF) is required"}
	}
	if strings.TrimSpace(c.Agreement) == "" {
		return &ValidationError{Context: "Company", Reason: "Agreement is required"}
	}
	return nil
}
