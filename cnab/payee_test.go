package cnab

import "testing"

func validPayee() Payee {
	cnpj, _ := NewCNPJ("11222333000181")
	return Payee{Name: "FORNECEDOR X", Registration: cnpj}
}

func TestPayeeValidate(t *testing.T) {
	if err := validPayee().validate(); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}
}

func TestPayeeValidateMissingFields(t *testing.T) {
	cases := []struct {
		name  string
		payee Payee
	}{
		{"missing name", Payee{Registration: validPayee().Registration}},
		{"missing registration", Payee{Name: "FORNECEDOR X"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.payee.validate(); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}
