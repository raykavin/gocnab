package cnab

import "fmt"

// FieldError reports that a specific field of a specific record, inside a
// specific batch, could not be rendered. It is returned when a lower
// level rendering error is translated at the cnab package boundary, where
// batch/record context becomes available.
type FieldError struct {
	Batch  int
	Record string
	Field  string
	Reason string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("cnab: batch %d, record %q, field %q: %s", e.Batch, e.Record, e.Field, e.Reason)
}

// ValidationError reports that a domain value (Company, Account, Payee, a
// Payment, or the SDK call sequence itself) failed a business rule.
type ValidationError struct {
	Context string
	Reason  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("cnab: %s: %s", e.Context, e.Reason)
}

// LimitExceededError reports that a FEBRABAN structural limit (at most 70
// batches per file, at most 10,000 movements per batch) was exceeded.
type LimitExceededError struct {
	Limit     string
	Max       int
	Attempted int
	Batch     int // 0 when the limit is file-level, not batch-level
}

func (e *LimitExceededError) Error() string {
	if e.Batch > 0 {
		return fmt.Sprintf("cnab: limit %q exceeded in batch %d: attempted %d, max %d", e.Limit, e.Batch, e.Attempted, e.Max)
	}
	return fmt.Sprintf("cnab: limit %q exceeded: attempted %d, max %d", e.Limit, e.Attempted, e.Max)
}

// SequenceError reports that a record sequence number ended up out of
// order. The engine alone computes sequence numbers, so this should only
// ever be observed if a Layout implementation is defective.
type SequenceError struct {
	Context  string
	Expected int
	Got      int
}

func (e *SequenceError) Error() string {
	return fmt.Sprintf("cnab: sequence error in %s: expected %d, got %d", e.Context, e.Expected, e.Got)
}

// TrailerMismatchError reports that a computed trailer total disagreed
// with the records it is supposed to summarize. The engine alone computes
// trailer totals from the records it renders, so this should only ever be
// observed if a Layout implementation is defective; Generate runs this
// check as a defensive guard rather than trusting the computation blindly.
type TrailerMismatchError struct {
	Batch    int // 0 for the file trailer
	Field    string
	Expected string
	Got      string
}

func (e *TrailerMismatchError) Error() string {
	if e.Batch > 0 {
		return fmt.Sprintf("cnab: trailer mismatch in batch %d, field %q: expected %s, got %s", e.Batch, e.Field, e.Expected, e.Got)
	}
	return fmt.Sprintf("cnab: trailer mismatch, field %q: expected %s, got %s", e.Field, e.Expected, e.Got)
}
