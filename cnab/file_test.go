package cnab

import (
	"strings"
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func validConfig() Config {
	return Config{
		Layout:  "febraban240",
		Company: validCompany(),
		Account: validAccount(),
		NSA:     1,
	}
}

func TestNewRemittanceValid(t *testing.T) {
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	if f == nil {
		t.Fatal("NewRemittance() returned a nil file with no error")
	}
}

func TestNewRemittanceValidation(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"invalid company", Config{Layout: "febraban240", Account: validAccount(), NSA: 1}},
		{"invalid account", Config{Layout: "febraban240", Company: validCompany(), NSA: 1}},
		{"zero NSA", Config{Layout: "febraban240", Company: validCompany(), Account: validAccount()}},
		{"missing layout", Config{Company: validCompany(), Account: validAccount(), NSA: 1}},
		{"unregistered layout", Config{Layout: "does-not-exist", Company: validCompany(), Account: validAccount(), NSA: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewRemittance(c.cfg); err == nil {
				t.Fatal("NewRemittance() error = nil, want an error")
			}
		})
	}
}

func TestNewBatchValidation(t *testing.T) {
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}

	if _, err := f.NewBatch(BatchProduct{}, PixTransfer); err == nil {
		t.Fatal("NewBatch() error = nil, want an error for a zero-value product")
	}
	if _, err := f.NewBatch(SupplierPayment, BatchService{}); err == nil {
		t.Fatal("NewBatch() error = nil, want an error for a zero-value service")
	}
	if _, err := f.NewBatch(SupplierPayment, PixTransfer); err != nil {
		t.Fatalf("NewBatch() error = %v, want nil", err)
	}
}

func TestGenerateRequiresAtLeastOneBatch(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	if _, err := f.Generate(); err == nil {
		t.Fatal("Generate() error = nil, want an error when there are no batches")
	}
}

func TestFileNameRequiresAtLeastOneBatch(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	if _, err := f.FileName(); err == nil {
		t.Fatal("FileName() error = nil, want an error when there are no batches")
	}
}

func TestGenerateEndToEnd(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}

	batch, err := f.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := batch.AddPayment(Pix{
		Key:    EmailKey("fornecedor@exemplo.com"),
		Payee:  validPayee(),
		Amount: 25200,
		Date:   now.AddDate(0, 0, 1),
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	content, err := f.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.HasSuffix(string(content), "\r\n") {
		t.Fatal("Generate() output does not end with CRLF")
	}
	lines := strings.Split(strings.TrimSuffix(string(content), "\r\n"), "\r\n")
	// file header + batch header + segment A + segment B (PIX) + batch trailer + file trailer
	if len(lines) != 6 {
		t.Fatalf("got %d lines, want 6", len(lines))
	}
	for i, line := range lines {
		if len(line) != 240 {
			t.Fatalf("line %d has length %d, want 240", i, len(line))
		}
	}

	name, err := f.FileName()
	if err != nil {
		t.Fatalf("FileName() error = %v", err)
	}
	if !strings.Contains(name, "FEBRABAN240") {
		t.Fatalf("FileName() = %q, want it to contain the layout name", name)
	}
}

func TestGenerateMultipleBatchesAndPayments(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}

	creditBatch, _ := f.NewBatch(SupplierPayment, CreditInAccount)
	if err := creditBatch.AddPayment(CreditAccount{
		Payee: validPayee(), Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, 1),
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	boletoBatch, _ := f.NewBatch(SupplierPayment, BoletoService)
	if err := boletoBatch.AddPayment(validBoleto(now)); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	content, err := f.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// file header + (batch1: header+A+B+trailer) + (batch2: header+J+J52+trailer) + file trailer
	want := 1 + 4 + 4 + 1
	got := strings.Count(string(content), "\r\n")
	if got != want {
		t.Fatalf("got %d lines, want %d", got, want)
	}
}

func TestGenerateRejectsPaymentThatBecamePastDue(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, _ := f.NewBatch(SupplierPayment, CreditInAccount)
	if err := batch.AddPayment(CreditAccount{
		Payee: validPayee(), Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, 1),
	}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	// Simulate time passing between AddPayment and Generate by rewriting
	// the stored segment's payment date to the past.
	batch.movements[0][0].Values[layout.KeyPaymentDate] = formatDate(now.AddDate(0, 0, -5))

	if _, err := f.Generate(); err == nil {
		t.Fatal("Generate() error = nil, want an error for a payment date that became past due")
	}
}

func TestGenerateSkipsPastDateCheckForCancellations(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, _ := f.NewBatch(SupplierPayment, CreditInAccount)
	original := CreditAccount{Payee: validPayee(), Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, -10)}
	if err := batch.AddPayment(CancelPayment{Original: original}); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	if _, err := f.Generate(); err != nil {
		t.Fatalf("Generate() error = %v, want nil for a cancellation of a past payment", err)
	}
}
