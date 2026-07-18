package cnab

import (
	"fmt"
	"time"

	"github.com/raykavin/gocnab/internal/engine"
)

// Batch is one batch ("lote") inside a remittance File, grouping
// payments that share the same product and settlement service. Create a
// Batch with File.NewBatch.
type Batch struct {
	file      *File
	product   BatchProduct
	service   BatchService
	movements [][]DetailSegment
}

// AddPayment validates p and appends it to the batch as a new movement.
// Validation covers every field required by p's own kind (see the
// specific payment type for details) and confirms the batch's active
// Layout actually supports every segment p needs; a payment kind not
// compatible with the active layout is rejected immediately rather than
// only failing later at File.Generate.
//
// AddPayment returns a *ValidationError for a missing/invalid field, or a
// *LimitExceededError once the batch already holds
// internal/engine.MaxMovementsPerBatch payments.
func (b *Batch) AddPayment(p Payment) error {
	if p == nil {
		return &ValidationError{Context: "AddPayment", Reason: "payment must not be nil"}
	}
	if len(b.movements) >= engine.MaxMovementsPerBatch {
		return &LimitExceededError{
			Limit:     engine.LimitMovementsPerBatch,
			Max:       engine.MaxMovementsPerBatch,
			Attempted: len(b.movements) + 1,
		}
	}

	if err := p.validate(time.Now(), validateOptions{}); err != nil {
		return err
	}

	segments, err := p.toSegments(b.file.layout)
	if err != nil {
		return err
	}
	for _, seg := range segments {
		if !b.file.engine.Supports(seg.Key) {
			return &ValidationError{
				Context: "AddPayment",
				Reason: fmt.Sprintf(
					"layout %q does not support segment %q required by this payment type",
					b.file.engine.LayoutName(), seg.Key,
				),
			}
		}
	}

	b.movements = append(b.movements, segments)
	return nil
}
