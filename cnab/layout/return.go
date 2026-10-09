package layout

// ReturnLayout is implemented by a Layout whose return files are not just
// its remittance records read backwards.
//
// A bank answers in the same 240 columns it received, but on the way back
// some of them are the bank's: a column the remittance fills with a fixed
// value can come back blank, or carry something only the bank knows. Read
// with the remittance record, such a column fails the const check every
// return parse runs (a const describes what the payer writes, not what the
// bank must echo). ReturnRecord describes the record as the bank sends it,
// so the check compares the line against the bank's own constants instead.
//
// A Layout only declares the records that differ. Every other record is
// read with the same RecordSpec Record returns, which is what a Layout that
// does not implement ReturnLayout at all gets for every record.
type ReturnLayout interface {
	Layout
	// ReturnRecord returns the RecordSpec a return file uses for key, and
	// ok=false when the return carries key exactly as Record describes it.
	// A record that only ever appears on a return file (SegmentZ) belongs
	// here rather than in Record.
	ReturnRecord(key RecordKey) (spec RecordSpec, ok bool)
}

// ForReturn returns the Layout a return file is read with: for each
// RecordKey, l's ReturnRecord when l implements ReturnLayout and declares
// one, and l's Record otherwise. A Layout that does not implement
// ReturnLayout reads its return files with the records it writes its
// remittances with, so ForReturn returns it unchanged. Name and Version are
// always l's own.
//
// cnab.ParseReturn applies ForReturn itself; call it directly only to read
// a return file's records some other way (a file header, say) with the same
// RecordSpec ParseReturn uses.
func ForReturn(l Layout) Layout {
	rl, ok := l.(ReturnLayout)
	if !ok {
		return l
	}
	return returnView{layout: rl}
}

// returnView is the Layout ForReturn builds over a ReturnLayout. It
// deliberately does not implement ReturnLayout itself, so it never offers a
// second, contradicting answer for a record.
type returnView struct {
	layout ReturnLayout
}

func (v returnView) Name() string    { return v.layout.Name() }
func (v returnView) Version() string { return v.layout.Version() }

func (v returnView) Record(key RecordKey) (RecordSpec, bool) {
	if spec, ok := v.layout.ReturnRecord(key); ok {
		return spec, true
	}
	return v.layout.Record(key)
}
