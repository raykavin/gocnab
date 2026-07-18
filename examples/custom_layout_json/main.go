// Example: defining a custom bank layout from a JSON config file instead
// of Go structs, then using it exactly like the bundled "febraban240"
// reference layout to generate a remittance.
//
// See layout.json next to this file for the descriptor, and
// NOVO-BANCO.md ("Alternativa: descrevendo o layout em JSON em vez de
// Go") for the full format reference.
//
// Run with: go run ./examples/custom_layout_json
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/raykavin/gocnab/cnab"
	"github.com/raykavin/gocnab/cnab/layout"
)

//go:embed layout.json
var layoutJSON []byte

func main() {
	// 1. Parse the JSON descriptor into a Layout. NewFromJSON validates
	// the whole file right here: unknown record keys, invalid field
	// kinds, out-of-range columns, an unknown "key" value, or a record
	// that does not cover columns 1-240 all fail at this line, with an
	// error naming the exact record/field at fault.
	l, err := layout.NewFromJSON(layoutJSON)
	if err != nil {
		log.Fatalf("NewFromJSON: %v", err)
	}

	// 2. Register it like any other layout. From here on it behaves
	// exactly like "febraban240": NewRemittance looks it up by name.
	cnab.RegisterLayout("example-json-240", l)

	companyRegistration, err := cnab.NewCNPJ("11222333000181")
	if err != nil {
		log.Fatalf("invalid company CNPJ: %v", err)
	}

	file, err := cnab.NewRemittance(cnab.Config{
		Layout: "example-json-240",
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

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.CreditInAccount)
	if err != nil {
		log.Fatalf("NewBatch: %v", err)
	}

	payeeRegistration, err := cnab.NewCNPJ("11444777000161")
	if err != nil {
		log.Fatalf("invalid payee CNPJ: %v", err)
	}
	err = batch.AddPayment(cnab.CreditAccount{
		Payee:      cnab.Payee{Name: "FORNECEDOR X", Registration: payeeRegistration},
		Account:    cnab.Account{Branch: "0116", Number: "12345", CheckDigit: "0"},
		Amount:     cnab.Cents(25200),
		Date:       time.Now().AddDate(0, 0, 1),
		YourNumber: "NF-1001",
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

	fmt.Printf("generated %s (%d bytes) using the JSON-loaded layout %q\n", path, len(content), l.Name())
}
