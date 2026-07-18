package cnab

import (
	"testing"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestFebraban240AutoRegistered(t *testing.T) {
	l, ok := layout.Lookup("febraban240")
	if !ok {
		t.Fatal("febraban240 is not registered even though cnab imports it")
	}
	if l.Version() == "" {
		t.Fatal("febraban240 layout has an empty Version")
	}
}

func TestRegisterLayoutIsLayoutRegister(t *testing.T) {
	RegisterLayout("test-register-layout-wrapper", minimalLayout{})
	if _, ok := layout.Lookup("test-register-layout-wrapper"); !ok {
		t.Fatal("RegisterLayout did not register under cnab/layout's registry")
	}
}
