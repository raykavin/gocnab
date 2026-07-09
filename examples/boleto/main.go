// Example: generating a CNAB 240 remittance that pays a boleto.
//
// Run with: go run ./examples/boleto
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
		NSA:     5,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.BoletoService)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	// The assignor ("cedente") issued the boleto; the payer ("sacado") is
	// normally the company itself.
	assignorRegistration, err := cnab.NewCNPJ("11122233000183")
	if err != nil {
		log.Fatalf("invalid assignor CNPJ: %v", err)
	}
	err = batch.AddPayment(cnab.BoletoPayment{
		Barcode:        "34191924500025200001570004013540025876327000",
		Assignor:       cnab.Payee{Name: "CEDENTE COMERCIO LTDA", Registration: assignorRegistration},
		Payer:          cnab.Payee{Name: "ACME LTDA", Registration: companyRegistration},
		DueDate:        time.Now().AddDate(0, 0, 10),
		DocumentAmount: cnab.Cents(50000), // R$ 500,00 face value
		Amount:         cnab.Cents(50000), // R$ 500,00 actually paid
		Date:           time.Now().AddDate(0, 0, 1),
		YourNumber:     "NF-5005",
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
