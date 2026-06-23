package engine

import (
	"strings"

	"github.com/raykavin/gocnab/cnab/layout"
)

// crlf is the line terminator every CNAB 240 record ends with (hex
// 0D0A), as required by the FEBRABAN standard.
const crlf = "\r\n"

// DetailLine is one 240 character line inside a movement: a movement
// (payment) contributes one or more DetailLine values, one per segment it
// uses (for example a credit-in-account payment contributes a SegmentA
// and a SegmentB line).
type DetailLine struct {
	Key    layout.RecordKey
	Values layout.Values
}

// BatchInput is everything Build needs to render one batch ("lote"):
// the header values, the ordered list of movements (each a slice of
// DetailLine), and the trailer values. Build fills in the structural
// values the engine itself owns (batch number, record sequence, record
// count, amount total); the caller must never set those keys.
type BatchInput struct {
	Header    layout.Values
	Movements [][]DetailLine
	Trailer   layout.Values
}

// FileInput is everything Build needs to render a whole file: the file
// header values, the ordered list of batches, and the file trailer
// values.
type FileInput struct {
	Header  layout.Values
	Batches []BatchInput
	Trailer layout.Values
}

// Build renders a complete CNAB 240 file: the file header, every batch
// (header, movements and trailer) and the file trailer, each line
// terminated with CRLF. It enforces the FEBRABAN structural limits (at
// most MaxBatchesPerFile batches, at most MaxMovementsPerBatch movements
// per batch) and computes every sequential number and trailer total
// itself, so a caller can never produce an out-of-order sequence or a
// trailer that disagrees with the records it summarizes.
func (e *Engine) Build(in FileInput) ([]byte, error) {
	if len(in.Batches) > MaxBatchesPerFile {
		return nil, &LimitError{Limit: LimitBatchesPerFile, Max: MaxBatchesPerFile, Got: len(in.Batches)}
	}

	fileHeaderRec, err := e.record(layout.FileHeader)
	if err != nil {
		return nil, err
	}
	fileTrailerRec, err := e.record(layout.FileTrailer)
	if err != nil {
		return nil, err
	}
	batchHeaderRec, err := e.record(layout.BatchHeader)
	if err != nil {
		return nil, err
	}
	batchTrailerRec, err := e.record(layout.BatchTrailer)
	if err != nil {
		return nil, err
	}

	var out strings.Builder

	headerLine, err := fileHeaderRec.render(in.Header)
	if err != nil {
		return nil, err
	}
	writeLine(&out, headerLine)
	fileRecordCount := 1

	for i, batchIn := range in.Batches {
		batchNumber := i + 1
		if len(batchIn.Movements) > MaxMovementsPerBatch {
			return nil, &LimitError{
				Limit: LimitMovementsPerBatch,
				Max:   MaxMovementsPerBatch,
				Got:   len(batchIn.Movements),
				Batch: batchNumber,
			}
		}

		acc := &batchAccumulator{}

		headerLine, err := batchHeaderRec.render(withBatchNumber(batchIn.Header, batchNumber))
		if err != nil {
			return nil, wrapBatchError(err, batchNumber, "batch header")
		}
		writeLine(&out, headerLine)
		acc.addRecord()

		seq := 0
		for _, movement := range batchIn.Movements {
			for _, line := range movement {
				seq++
				rec, err := e.record(line.Key)
				if err != nil {
					return nil, wrapBatchError(err, batchNumber, string(line.Key))
				}
				values := withSequence(line.Values, batchNumber, seq)
				rendered, err := rec.render(values)
				if err != nil {
					return nil, wrapBatchError(err, batchNumber, string(line.Key))
				}
				writeLine(&out, rendered)
				acc.addRecord()
				if err := acc.addAmount(line.Values); err != nil {
					return nil, wrapBatchError(err, batchNumber, string(line.Key))
				}
			}
		}

		trailerValues := withBatchTotals(batchIn.Trailer, batchNumber, acc.recordCount+1, acc.amount)
		trailerLine, err := batchTrailerRec.render(trailerValues)
		if err != nil {
			return nil, wrapBatchError(err, batchNumber, "batch trailer")
		}
		writeLine(&out, trailerLine)

		fileRecordCount += acc.recordCount + 1
	}

	fileTrailerValues := withFileTotals(in.Trailer, len(in.Batches), fileRecordCount+1)
	trailerLine, err := fileTrailerRec.render(fileTrailerValues)
	if err != nil {
		return nil, err
	}
	writeLine(&out, trailerLine)

	return []byte(out.String()), nil
}

func writeLine(out *strings.Builder, line string) {
	out.WriteString(line)
	out.WriteString(crlf)
}

func cloneValues(v layout.Values) layout.Values {
	out := make(layout.Values, len(v)+4)
	for k, val := range v {
		out[k] = val
	}
	return out
}

func withBatchNumber(v layout.Values, batchNumber int) layout.Values {
	out := cloneValues(v)
	out[layout.KeyBatchNumber] = batchNumber
	return out
}

func withSequence(v layout.Values, batchNumber, seq int) layout.Values {
	out := cloneValues(v)
	out[layout.KeyBatchNumber] = batchNumber
	out[layout.KeySequence] = seq
	return out
}

func withBatchTotals(v layout.Values, batchNumber, recordCount int, amount int64) layout.Values {
	out := cloneValues(v)
	out[layout.KeyBatchNumber] = batchNumber
	out[layout.KeyBatchRecordCount] = recordCount
	out[layout.KeyBatchAmount] = amount
	return out
}

func withFileTotals(v layout.Values, batchCount, recordCount int) layout.Values {
	out := cloneValues(v)
	out[layout.KeyBatchCount] = batchCount
	out[layout.KeyFileRecordCount] = recordCount
	return out
}
