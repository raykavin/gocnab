package engine

import "github.com/raykavin/gocnab/cnab/layout"

// batchAccumulator tracks the running totals needed to fill a batch
// trailer: the FEBRABAN standard requires the trailer to carry the count
// of every record in the batch (header, details and the trailer itself)
// and the sum of every payment amount in it. Totals are derived from the
// lines actually rendered, which is what makes a trailer/record-count
// mismatch impossible to produce through the public API.
type batchAccumulator struct {
	recordCount int
	amount      int64
}

// addRecord accounts for one more rendered line (header or detail).
func (a *batchAccumulator) addRecord() {
	a.recordCount++
}

// addAmount adds the KeyAmount value found in a detail line's values, if
// any, to the running batch amount.
func (a *batchAccumulator) addAmount(values layout.Values) error {
	raw, ok := values[layout.KeyAmount]
	if !ok {
		return nil
	}
	n, ok := intValue(raw)
	if !ok {
		return &FieldRenderError{Field: string(layout.KeyAmount), Reason: "amount must be an integer number of cents"}
	}
	a.amount += n
	return nil
}
