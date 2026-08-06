package cnab

import (
	"errors"
	"fmt"

	"github.com/raykavin/gocnab/internal/engine"
)

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

// ReturnParseError reports that a specific line of a return file, passed to
// ParseReturn, could not be decoded. Field is empty for a problem that is
// not about one specific field (currently only a line with the wrong
// length); when Field is set, Expected/Got describe a const field whose
// content did not match what the layout expects there the file is either
// corrupt or was parsed against the wrong layout.
type ReturnParseError struct {
	Line     int
	Field    string
	Expected string
	Got      string
	Reason   string
}

func (e *ReturnParseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("cnab: parse return, line %d, field %q: expected %q, got %q", e.Line, e.Field, e.Expected, e.Got)
	}
	return fmt.Sprintf("cnab: parse return, line %d: %s", e.Line, e.Reason)
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

// translateEngineError maps an error coming out of internal/engine (via
// engine.New or Engine.Build) onto the exported error types above, so a
// *engine.FieldRenderError, *engine.LimitError or *engine.SpecError never
// reaches a caller unwrapped. None of those types can be named outside
// this module to begin with (internal/engine is
// exactly that, internal), so a caller could previously only ever match on
// the error's string form; this gives them a real type to use with
// errors.As instead, the same promise every other error path in this
// package already keeps (see ReturnParseError, and the FieldError/
// LimitExceededError already constructed directly by Batch.AddPayment and
// File.NewBatch for the limit checks made before ever calling the engine).
func translateEngineError(err error) error {
	if err == nil {
		return nil
	}

	batch, record, cause := 0, "", err
	var batchErr *engine.BatchError
	if errors.As(err, &batchErr) {
		batch, record, cause = batchErr.Batch, batchErr.Record, batchErr.Err
	}

	var fieldErr *engine.FieldRenderError
	if errors.As(cause, &fieldErr) {
		return &FieldError{Batch: batch, Record: record, Field: fieldErr.Field, Reason: fieldErr.Reason}
	}

	var limitErr *engine.LimitError
	if errors.As(cause, &limitErr) {
		return &LimitExceededError{Limit: limitErr.Limit, Max: limitErr.Max, Attempted: limitErr.Got, Batch: limitErr.Batch}
	}

	// *engine.SpecError (a malformed Layout) or anything else not
	// specifically handled above: still a real business-rule failure from
	// the caller's point of view, just without a more specific shape to
	// preserve.
	context := "Generate"
	if record != "" {
		context = record
	}
	return &ValidationError{Context: context, Reason: cause.Error()}
}
