// Example: generating a CNAB 240 remittance with a same-bank
// credit-in-account payment.
//
// Run with: go run ./examples/credit_account
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
	// 1. Describe the company sending the remittance (the payer) and the
	// bank account the payment will be debited from.
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
		NSA:     1,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	// 2. Start a batch for supplier payments settled as a same-bank
	// credit in account.
	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.CreditInAccount)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	// 3. Describe the beneficiary and add the payment to the batch.
	payeeRegistration, err := cnab.NewCNPJ("11444777000161")
	if err != nil {
		log.Fatalf("invalid payee CNPJ: %v", err)
	}
	err = batch.AddPayment(cnab.CreditAccount{
		Payee:      cnab.Payee{Name: "FORNECEDOR X", Registration: payeeRegistration},
		Account:    cnab.Account{Branch: "0116", Number: "12345", CheckDigit: "0"},
		Amount:     cnab.Cents(25200), // R$ 252,00
		Date:       time.Now().AddDate(0, 0, 1),
		YourNumber: "NF-1001",
	})
	if err != nil {
		log.Fatalf("AddPayment: %v", err)
	}

	// 4. Generate the file content and save it next to the OS temp
	// directory so running this example never writes inside the module.
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

	fmt.Printf("generated %s (%d bytes, %d records)\n", path, len(content), countLines(content))
}

func countLines(content []byte) int {
	count := 0
	for _, b := range content {
		if b == '\n' {
			count++
		}
	}
	return count
}
