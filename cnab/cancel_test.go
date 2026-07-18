package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestCancelPaymentValidateAllowsPastDate(t *testing.T) {
	now := time.Now()
	original := CreditAccount{
		Payee:   validPayee(),
		Account: validAccount(),
		Amount:  1000,
		Date:    now.AddDate(0, 0, -10), // legitimately in the past
	}
	cancel := CancelPayment{Original: original}

	if err := cancel.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil (past date must be allowed for cancellations)", err)
	}
}

func TestCancelPaymentValidateMissingOriginal(t *testing.T) {
	cancel := CancelPayment{}
	if err := cancel.validate(time.Now(), validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for a missing Original")
	}
}

func TestCancelPaymentValidatePropagatesOtherErrors(t *testing.T) {
	now := time.Now()
	// Missing Payee, unrelated to the past-date exemption.
	original := CreditAccount{Account: validAccount(), Amount: 1000, Date: now.AddDate(0, 0, -10)}
	cancel := CancelPayment{Original: original}

	if err := cancel.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want the underlying ValidationError to propagate")
	}
}

func TestCancelPaymentToSegments(t *testing.T) {
	now := time.Now()
	original := CreditAccount{
		Payee:   validPayee(),
		Account: validAccount(),
		Amount:  1000,
		Date:    now.AddDate(0, 0, 1),
	}
	cancel := CancelPayment{Original: original}

	segments, err := cancel.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	for _, seg := range segments {
		if seg.Values[layout.KeyMovementType] != "9" {
			t.Fatalf("KeyMovementType = %v, want \"9\"", seg.Values[layout.KeyMovementType])
		}
		if seg.Values[layout.KeyInstructionCode] != "99" {
			t.Fatalf("KeyInstructionCode = %v, want \"99\"", seg.Values[layout.KeyInstructionCode])
		}
	}

	// The original segments must not be mutated in place.
	originalSegments, _ := original.toSegments(nil)
	if originalSegments[0].Values[layout.KeyMovementType] != "0" {
		t.Fatalf("original segment was mutated: KeyMovementType = %v", originalSegments[0].Values[layout.KeyMovementType])
	}
}
