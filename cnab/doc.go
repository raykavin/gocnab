// Package cnab is a Go SDK for generating CNAB 240 FEBRABAN remittance
// files. It exposes a small, domain oriented API (companies, accounts,
// payees and strongly typed payments) and hides every field position and
// FEBRABAN encoding detail behind a pluggable bank Layout.
//
// A minimal remittance with a single PIX transfer looks like this:
//
//	file, err := cnab.NewRemittance(cnab.Config{
//		Layout:  "febraban240",
//		Company: cnab.Company{Name: "ACME LTDA", Registration: cnpj, Agreement: "1234"},
//		Account: cnab.Account{Branch: "0116", Number: "75890", CheckDigit: "6"},
//		NSA:     42,
//	})
//	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.PixTransfer)
//	err = batch.AddPayment(cnab.Pix{
//		Key:    cnab.EmailKey("fornecedor@exemplo.com"),
//		Payee:  cnab.Payee{Name: "FORNECEDOR X", Registration: cnpj},
//		Amount: cnab.Cents(25200),
//		Date:   time.Now().AddDate(0, 0, 1),
//	})
//	content, err := file.Generate()
//
// See ./docs in the module root (in Portuguese) for the full architecture
// and API reference, and ./examples for a runnable program per payment
// kind.
package cnab
