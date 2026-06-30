package cnab

import "testing"

func validAccount() Account {
	return Account{Branch: "0116", Number: "75890", CheckDigit: "6"}
}

func TestAccountValidate(t *testing.T) {
	if err := validAccount().validate(); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}
}

func TestAccountValidateMissingFields(t *testing.T) {
	cases := []struct {
		name    string
		account Account
	}{
		{"missing branch", Account{Number: "75890", CheckDigit: "6"}},
		{"missing number", Account{Branch: "0116", CheckDigit: "6"}},
		{"missing check digit", Account{Branch: "0116", Number: "75890"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.account.validate(); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}
