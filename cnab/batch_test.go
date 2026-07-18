package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
	"github.com/raykavin/gocnab/internal/engine"
)

// minimalLayout only supports the records a plain credit-in-account
// payment needs, used to test that AddPayment rejects a payment kind the
// active layout does not support.
type minimalLayout struct{}

func (minimalLayout) Name() string    { return "minimal-test-layout" }
func (minimalLayout) Version() string { return "001" }
func (minimalLayout) Record(key layout.RecordKey) (layout.RecordSpec, bool) {
	full := 240
	basic := []layout.FieldSpec{{Name: "Filler", Start: 1, End: full, Kind: layout.KindAlphanumeric}}
	switch key {
	case layout.FileHeader, layout.FileTrailer, layout.BatchHeader, layout.BatchTrailer, layout.SegmentA, layout.SegmentB:
		return layout.RecordSpec{Name: string(key), Fields: basic}, true
	default:
		return layout.RecordSpec{}, false
	}
}

func init() {
	RegisterLayout("minimal-test-layout", minimalLayout{})
}

func TestAddPaymentRejectsUnsupportedSegment(t *testing.T) {
	f, err := NewRemittance(Config{
		Layout:  "minimal-test-layout",
		Company: validCompany(),
		Account: validAccount(),
		NSA:     1,
	})
	if err != nil {
		t.Fatalf("NewRemittance() error = %v", err)
	}
	batch, err := f.NewBatch(SupplierPayment, BoletoService)
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}

	if err := batch.AddPayment(validBoleto(time.Now())); err == nil {
		t.Fatal("AddPayment() error = nil, want an error since the layout has no SegmentJ")
	}
}

func TestAddPaymentRejectsNilPayment(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	batch, _ := f.NewBatch(SupplierPayment, PixTransfer)
	if err := batch.AddPayment(nil); err == nil {
		t.Fatal("AddPayment(nil) error = nil, want an error")
	}
}

func TestAddPaymentRejectsInvalidPayment(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	batch, _ := f.NewBatch(SupplierPayment, PixTransfer)
	if err := batch.AddPayment(Pix{}); err == nil {
		t.Fatal("AddPayment() error = nil, want an error for a zero-value Pix")
	}
}

func TestAddPaymentMovementsLimit(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	batch, _ := f.NewBatch(SupplierPayment, PixTransfer)

	now := time.Now()
	payment := func() Payment {
		return Pix{Key: EmailKey("a@b.com"), Payee: validPayee(), Amount: 1, Date: now.AddDate(0, 0, 1)}
	}
	for i := 0; i < engine.MaxMovementsPerBatch; i++ {
		if err := batch.AddPayment(payment()); err != nil {
			t.Fatalf("AddPayment() #%d error = %v", i, err)
		}
	}
	if err := batch.AddPayment(payment()); err == nil {
		t.Fatal("AddPayment() error = nil, want a *LimitExceededError once the batch is full")
	}
}

func TestNewBatchesLimit(t *testing.T) {
	f, _ := NewRemittance(validConfig())
	for i := 0; i < engine.MaxBatchesPerFile; i++ {
		if _, err := f.NewBatch(SupplierPayment, PixTransfer); err != nil {
			t.Fatalf("NewBatch() #%d error = %v", i, err)
		}
	}
	if _, err := f.NewBatch(SupplierPayment, PixTransfer); err == nil {
		t.Fatal("NewBatch() error = nil, want a *LimitExceededError once the file is full")
	}
}
