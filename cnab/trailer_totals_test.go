package cnab

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// The FEBRABAN batch trailer is the one place a bank re-checks the file's
// arithmetic against the records it summarizes: columns 18-23 carry the
// count of every record in the batch (header, details and the trailer
// itself) and columns 24-41 the sum of every payment amount in it. The
// engine derives both from the lines it actually rendered, and these tests
// pin that derivation over the cases where an accumulator bug would hide:
// several payments in one batch, several batches in one file, and amounts
// distinct enough that a total equal to any single payment is visible.

// batchTrailerTotals returns the record count and summed amount a batch
// trailer line carries.
func batchTrailerTotals(t *testing.T, line string) (records int, amount int64) {
	t.Helper()

	if len(line) != 240 {
		t.Fatalf("batch trailer has length %d, want 240", len(line))
	}
	if line[7] != '5' {
		t.Fatalf("line is not a batch trailer (record type %q)", line[7:8])
	}

	records, err := strconv.Atoi(line[17:23])
	if err != nil {
		t.Fatalf("record count %q is not numeric: %v", line[17:23], err)
	}
	amount, err = strconv.ParseInt(line[23:41], 10, 64)
	if err != nil {
		t.Fatalf("amount %q is not numeric: %v", line[23:41], err)
	}
	return records, amount
}

// fileTrailerTotals returns the batch count and record count a file
// trailer line carries.
func fileTrailerTotals(t *testing.T, line string) (batches, records int) {
	t.Helper()

	if line[7] != '9' {
		t.Fatalf("line is not a file trailer (record type %q)", line[7:8])
	}
	batches, err := strconv.Atoi(line[17:23])
	if err != nil {
		t.Fatalf("batch count %q is not numeric: %v", line[17:23], err)
	}
	records, err = strconv.Atoi(line[23:29])
	if err != nil {
		t.Fatalf("record count %q is not numeric: %v", line[23:29], err)
	}
	return batches, records
}

func generatedLines(t *testing.T, f *File) []string {
	t.Helper()

	content, err := f.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(content), "\r\n"), "\r\n")
}

// boletoOf returns a valid boleto payment for the given amount.
func boletoOf(now time.Time, amount Cents) BoletoPayment {
	b := validBoleto(now)
	b.Amount = amount
	b.DocumentAmount = amount
	return b
}

// TestBatchTrailerTotalsOneBoleto is the single-payment baseline: the
// trailer must carry the four records of the batch and that payment's
// amount, unrounded.
func TestBatchTrailerTotalsOneBoleto(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(BoletoCollection, BoletoService)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := batch.AddPayment(boletoOf(now, 12500)); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	lines := generatedLines(t, f)
	// file header + batch header + J + J52 + batch trailer + file trailer
	if len(lines) != 6 {
		t.Fatalf("got %d lines, want 6", len(lines))
	}

	records, amount := batchTrailerTotals(t, lines[4])
	if records != 4 {
		t.Errorf("batch trailer record count = %d, want 4 (header + J + J52 + trailer)", records)
	}
	if amount != 12500 {
		t.Errorf("batch trailer amount = %d, want 12500", amount)
	}
}

// TestBatchTrailerTotalsSumsEveryPaymentOfTheBatch is the multi-payment
// case: three boletos of deliberately different values, so a total that
// accidentally equals any single payment — the shape of an accumulator
// that is assigned rather than added to, or reset per movement — fails
// here.
func TestBatchTrailerTotalsSumsEveryPaymentOfTheBatch(t *testing.T) {
	now := time.Now()
	amounts := []Cents{12500, 214512, 7}
	const wantTotal = 12500 + 214512 + 7

	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(BoletoCollection, OtherBankBoletoService)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	for _, amount := range amounts {
		if err := batch.AddPayment(boletoOf(now, amount)); err != nil {
			t.Fatalf("AddPayment(%d) error = %v", amount, err)
		}
	}

	lines := generatedLines(t, f)
	// file header + batch header + 3x(J + J52) + batch trailer + file trailer
	if len(lines) != 1+1+6+1+1 {
		t.Fatalf("got %d lines, want %d", len(lines), 1+1+6+1+1)
	}

	records, amount := batchTrailerTotals(t, lines[len(lines)-2])
	if records != 8 {
		t.Errorf("batch trailer record count = %d, want 8 (header + 6 details + trailer)", records)
	}
	if amount != wantTotal {
		t.Errorf("batch trailer amount = %d, want %d", amount, wantTotal)
	}
	for _, single := range amounts {
		if amount == int64(single) {
			t.Errorf("batch trailer amount = %d, which is one payment's amount rather than the sum", amount)
		}
	}
}

// TestBatchTrailerTotalsAreIndependentPerBatch pins the rule that makes a
// multi-batch file's arithmetic readable: each batch trailer summarizes
// only its own payments (the accumulator restarts per batch, by design),
// and the file trailer counts lotes and records but no amount — FEBRABAN's
// registro 9 has no monetary total, so a file's value is the sum of its
// batch trailers and nothing else.
func TestBatchTrailerTotalsAreIndependentPerBatch(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}

	ownBank, err := f.NewBatch(BoletoCollection, BoletoService)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if err := ownBank.AddPayment(boletoOf(now, 12500)); err != nil {
		t.Fatalf("AddPayment() error = %v", err)
	}

	otherBank, err := f.NewBatch(BoletoCollection, OtherBankBoletoService)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	for _, amount := range []Cents{214512, 100000} {
		if err := otherBank.AddPayment(boletoOf(now, amount)); err != nil {
			t.Fatalf("AddPayment(%d) error = %v", amount, err)
		}
	}

	lines := generatedLines(t, f)
	// file header + (header+J+J52+trailer) + (header+2x(J+J52)+trailer) + file trailer
	if len(lines) != 1+4+6+1 {
		t.Fatalf("got %d lines, want %d", len(lines), 1+4+6+1)
	}

	firstRecords, firstAmount := batchTrailerTotals(t, lines[4])
	if firstRecords != 4 || firstAmount != 12500 {
		t.Errorf("batch 1 trailer = (%d records, %d), want (4, 12500)", firstRecords, firstAmount)
	}

	secondRecords, secondAmount := batchTrailerTotals(t, lines[10])
	if secondRecords != 6 || secondAmount != 314512 {
		t.Errorf("batch 2 trailer = (%d records, %d), want (6, 314512)", secondRecords, secondAmount)
	}

	if total := firstAmount + secondAmount; total != 327012 {
		t.Errorf("batch trailers sum to %d, want 327012", total)
	}

	batches, records := fileTrailerTotals(t, lines[11])
	if batches != 2 {
		t.Errorf("file trailer batch count = %d, want 2", batches)
	}
	if records != len(lines) {
		t.Errorf("file trailer record count = %d, want %d", records, len(lines))
	}
}

// TestBatchTrailerCountsSegmentsNotPayments guards the distinction the
// record count rests on: a boleto contributes two records (J and J-52)
// while a PIX contributes two as well (A and B) — the trailer counts
// rendered lines, never movements.
func TestBatchTrailerCountsSegmentsNotPayments(t *testing.T) {
	now := time.Now()
	f, err := NewRemittance(validConfig())
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(SupplierPayment, PixTransfer)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	for _, amount := range []Cents{25200, 1} {
		if err := batch.AddPayment(Pix{
			Key:    EmailKey("fornecedor@exemplo.com"),
			Payee:  validPayee(),
			Amount: amount,
			Date:   now.AddDate(0, 0, 1),
		}); err != nil {
			t.Fatalf("AddPayment(%d) error = %v", amount, err)
		}
	}

	lines := generatedLines(t, f)
	records, amount := batchTrailerTotals(t, lines[len(lines)-2])
	if records != 6 {
		t.Errorf("batch trailer record count = %d, want 6 (header + 2x(A+B) + trailer)", records)
	}
	if amount != 25201 {
		t.Errorf("batch trailer amount = %d, want 25201", amount)
	}
}
