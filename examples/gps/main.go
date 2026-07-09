// Example: generating a CNAB 240 remittance that pays a GPS (Guia da
// Previdência Social) contribution.
//
// Run with: go run ./examples/gps
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
		NSA:     8,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.TaxWithoutBarcodeService)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	err = batch.AddPayment(cnab.GPS{
		Taxpayer:   cnab.Payee{Name: "ACME LTDA", Registration: companyRegistration},
		Period:     time.Now().AddDate(0, -1, 0), // "competência": last month
		DueDate:    time.Now().AddDate(0, 0, 7),
		Amount:     cnab.Cents(80000), // R$ 800,00
		Date:       time.Now().AddDate(0, 0, 1),
		YourNumber: "NF-8008",
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
