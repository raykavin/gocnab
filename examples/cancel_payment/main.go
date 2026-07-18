// Example: generating a CNAB 240 remittance that cancels a payment sent
// in an earlier file. A cancellation re-sends the original payment's
// data wrapped in cnab.CancelPayment, which sets the FEBRABAN movement
// type to "9" and the instruction code to "99" automatically.
//
// Run with: go run ./examples/cancel_payment
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
		NSA:     9,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.CreditInAccount)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	payeeRegistration, err := cnab.NewCNPJ("11444777000161")
	if err != nil {
		log.Fatalf("invalid payee CNPJ: %v", err)
	}

	// This is the same payment data that was sent (and accepted) in a
	// previous remittance file; the original payment date is legitimately
	// in the past by the time a cancellation is sent.
	original := cnab.CreditAccount{
		Payee:      cnab.Payee{Name: "FORNECEDOR X", Registration: payeeRegistration},
		Account:    cnab.Account{Branch: "0116", Number: "12345", CheckDigit: "0"},
		Amount:     cnab.Cents(25200),
		Date:       time.Now().AddDate(0, 0, -3),
		YourNumber: "NF-1001",
	}

	err = batch.AddPayment(cnab.CancelPayment{Original: original})
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
