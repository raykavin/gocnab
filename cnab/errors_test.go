package cnab

import "testing"

func TestErrorMessages(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"FieldError", &FieldError{Batch: 1, Record: "segment_a", Field: "Amount", Reason: "too big"}},
		{"ValidationError", &ValidationError{Context: "Payee", Reason: "Name is required"}},
		{"LimitExceededError file", &LimitExceededError{Limit: "batches_per_file", Max: 70, Attempted: 71}},
		{"LimitExceededError batch", &LimitExceededError{Limit: "movements_per_batch", Max: 10000, Attempted: 10001, Batch: 3}},
		{"SequenceError", &SequenceError{Context: "batch 1", Expected: 1, Got: 2}},
		{"TrailerMismatchError file", &TrailerMismatchError{Field: "record_count", Expected: "5", Got: "4"}},
		{"TrailerMismatchError batch", &TrailerMismatchError{Batch: 2, Field: "amount", Expected: "100", Got: "90"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err.Error() == "" {
				t.Fatal("Error() returned an empty string")
			}
		})
	}
}
