package cnab

import "testing"

func validCompany() Company {
	cnpj, _ := NewCNPJ("11222333000181")
	return Company{Name: "ACME LTDA", Registration: cnpj, Agreement: "1234"}
}

func TestCompanyValidate(t *testing.T) {
	if err := validCompany().validate(); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}
}

func TestCompanyValidateMissingFields(t *testing.T) {
	cases := []struct {
		name    string
		company Company
	}{
		{"missing name", Company{Registration: validCompany().Registration, Agreement: "1234"}},
		{"missing registration", Company{Name: "ACME LTDA", Agreement: "1234"}},
		{"missing agreement", Company{Name: "ACME LTDA", Registration: validCompany().Registration}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.company.validate(); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}
