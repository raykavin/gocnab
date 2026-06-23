package engine

import "fmt"

// SpecError reports a problem with a Layout's RecordSpec itself: a gap or
// overlap between fields, a field outside columns 1-240, or a missing
// layout version. These are programming errors in a layout definition,
// not runtime data problems, and are detected once when the Engine is
// built.
type SpecError struct {
	Record string
	Reason string
}

func (e *SpecError) Error() string {
	if e.Record == "" {
		return fmt.Sprintf("cnab: invalid layout: %s", e.Reason)
	}
	return fmt.Sprintf("cnab: invalid layout: record %q: %s", e.Record, e.Reason)
}

// FieldRenderError reports a problem filling a specific field with the
// value supplied at render time.
type FieldRenderError struct {
	Field  string
	Reason string
}

func (e *FieldRenderError) Error() string {
	return fmt.Sprintf("field %q: %s", e.Field, e.Reason)
}

// LimitError reports that a FEBRABAN structural limit (batches per file,
// movements per batch) was exceeded.
type LimitError struct {
	Limit string
	Max   int
	Got   int
	Batch int // 0 when the limit is file-level, not batch-level
}

func (e *LimitError) Error() string {
	if e.Batch > 0 {
		return fmt.Sprintf("cnab: limit %q exceeded in batch %d: got %d, max %d", e.Limit, e.Batch, e.Got, e.Max)
	}
	return fmt.Sprintf("cnab: limit %q exceeded: got %d, max %d", e.Limit, e.Got, e.Max)
}

// BatchError wraps an error that occurred while rendering a specific
// batch and record, adding context the lower level error cannot know
// about on its own.
type BatchError struct {
	Batch  int
	Record string
	Err    error
}

func (e *BatchError) Error() string {
	return fmt.Sprintf("batch %d, record %q: %s", e.Batch, e.Record, e.Err)
}

func (e *BatchError) Unwrap() error { return e.Err }

func wrapBatchError(err error, batch int, record string) error {
	if err == nil {
		return nil
	}
	return &BatchError{Batch: batch, Record: record, Err: err}
}
