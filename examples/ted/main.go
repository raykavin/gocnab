// Example: generating a CNAB 240 remittance with a TED (wire transfer)
// payment to a supplier at a different bank.
//
// Run with: go run ./examples/ted
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
		NSA:     2,
	})
	if err != nil {
		log.Fatalf("NewRemittance: %v", err)
	}

	// A TED batch settles payments as wire transfers, typically to a
	// beneficiary at a different bank.
	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.TEDTransfer)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	payeeRegistration, err := cnab.NewCNPJ("11222333000262")
	if err != nil {
		log.Fatalf("invalid payee CNPJ: %v", err)
	}
	err = batch.AddPayment(cnab.TED{
		Payee:      cnab.Payee{Name: "FORNECEDOR Y", Registration: payeeRegistration},
		BankCode:   "341", // beneficiary's bank (COMPE code)
		Account:    cnab.Account{Branch: "4001", Number: "998877", CheckDigit: "1"},
		Amount:     cnab.Cents(150000), // R$ 1.500,00
		Date:       time.Now().AddDate(0, 0, 1),
		Purpose:    cnab.PurposeSupplierPayment,
		YourNumber: "NF-2002",
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
