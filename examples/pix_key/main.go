// Example: generating a CNAB 240 remittance with a PIX transfer
// addressed by key (in this case, an e-mail key). Swap EmailKey for
// PhoneKey, CPFKey, CNPJKey or RandomKey to address a different key type.
//
// Run with: go run ./examples/pix_key
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/raykavin/gocnab/cnab"
)

func main() {
	companyRegistration, err := cnab.NewCNPJ("11222333000181")
	if err != nil {
		log.Fatalf("invalid company CNPJ: %v", err)
	}

	file, err := cnab.NewRemittance(cnab.Config{
		Layout: "febraban240",
		Company: cnab.Company{
			Name:         "ACME LTDA",
			Registration: companyRegistration,
			Agreement:    "1234",
		},
		Account: cnab.Account{Branch: "0116", Number: "75890", CheckDigit: "6"},
		NSA:     3,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.PixTransfer)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	payeeRegistration, err := cnab.NewCNPJ("12345678000195")
	if err != nil {
		log.Fatalf("invalid payee CNPJ: %v", err)
	}
	err = batch.AddPayment(cnab.Pix{
		Key:        cnab.EmailKey("fornecedor@exemplo.com"),
		Payee:      cnab.Payee{Name: "FORNECEDOR Z", Registration: payeeRegistration},
		Amount:     cnab.Cents(25200), // R$ 252,00
		Date:       time.Now().AddDate(0, 0, 1),
		YourNumber: "NF-3003",
	})
	if err != nil {
		log.Fatalf("AddPayment: %v", err)
	}

	content, err := file.Generate()
	if err != nil {
		log.Fatalf("Generate: %v", err)
	}
	name, err := file.FileName()
	if err != nil {
		log.Fatalf("FileName: %v", err)
	}

	path := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		log.Fatalf("WriteFile: %v", err)
	}

	fmt.Printf("generated %s (%d bytes)\n", path, len(content))
}
