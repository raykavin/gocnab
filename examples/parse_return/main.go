// Example: parsing a CNAB 240 return file with cnab.ParseReturn.
//
// Banks send return files, not this SDK, so there is nothing to generate
// here the way every other example does. Instead, this example builds a
// remittance exactly as examples/pix_key does, then patches the three
// column ranges of its Segmento A that a bank fills in on the way back
// (see docs/ARQUITETURA.md, "Processando retorno") to simulate what that
// remittance's return might look like: one settled payment, one rejected.
// A real integration skips straight to ParseReturn, on bytes read from the
// file a bank actually sent.
//
// Run with: go run ./examples/parse_return
package main

import (
	"fmt"
	"log"
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
		NSA:     4,
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
	tomorrow := time.Now().AddDate(0, 0, 1)
	if err := batch.AddPayment(cnab.Pix{
		Key:        cnab.EmailKey("fornecedor@exemplo.com"),
		Payee:      cnab.Payee{Name: "FORNECEDOR Z", Registration: payeeRegistration},
		Amount:     cnab.Cents(25200), // R$ 252,00
		Date:       tomorrow,
		YourNumber: "NF-4001",
	}); err != nil {
		log.Fatalf("AddPayment (settled): %v", err)
	}
	if err := batch.AddPayment(cnab.Pix{
		Key:        cnab.EmailKey("outro@exemplo.com"),
		Payee:      cnab.Payee{Name: "FORNECEDOR W", Registration: payeeRegistration},
		Amount:     cnab.Cents(9900), // R$ 99,00
		Date:       tomorrow,
		YourNumber: "NF-4002",
	}); err != nil {
		log.Fatalf("AddPayment (rejected): %v", err)
	}

	content, err := file.Generate()
	if err != nil {
		log.Fatalf("Generate: %v", err)
	}

	simulateBankReturn(content)

	result, err := cnab.ParseReturn("febraban240", content)
	if err != nil {
		log.Fatalf("ParseReturn: %v", err)
	}

	for _, m := range result.Movements {
		if m.Accepted() {
			fmt.Printf("%s: settled R$ %d,%02d on %s\n", m.YourNumber, m.SettlementAmount/100, m.SettlementAmount%100, m.SettlementDate.Format("2006-01-02"))
		} else {
			fmt.Printf("%s: rejected, occurrence codes %v\n", m.YourNumber, m.OccurrenceCodes)
		}
	}
}

// simulateBankReturn patches content's two Segmento A lines in place, the
// way a bank's return would have them: the first payment settled for its
// full amount on 2026-01-05, the second rejected with occurrence code "02"
// ("saldo insuficiente" in some bank manuals; the exact meaning of a code
// is bank-specific see ReturnMovement.OccurrenceCodes).
//
// A real integration never does this: it calls ParseReturn directly on the
// bytes a bank sent. This function only exists so this example has
// something return-shaped to parse.
func simulateBankReturn(content []byte) {
	const firstSegmentALine = 2  // 0: file header, 1: batch header, 2: first Segmento A
	const secondSegmentALine = 4 // 3: Segmento B (Pix) for the first payment

	patch(content, firstSegmentALine, 155, "02012026")        // ActualPaymentDate, DDMMYYYY
	patch(content, firstSegmentALine, 163, "000000000025200") // ActualPaymentAmount, matches Cents(25200)
	patch(content, secondSegmentALine, 231, "02        ")     // OccurrenceCodes
}

func patch(content []byte, line, startCol int, value string) {
	const lineStride = 242
	offset := line*lineStride + (startCol - 1)
	copy(content[offset:offset+len(value)], value)
}
