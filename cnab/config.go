package cnab

// Config holds everything NewRemittance needs to start a new remittance
// file.
type Config struct {
	// Layout is the name of a registered Layout, e.g. "febraban240".
	// Ignored when LayoutSpec is set.
	Layout string
	// LayoutSpec is a Layout instance to use directly, bypassing the
	// registry. Set it when the layout is not a process-wide constant but
	// data the caller resolved at runtime (loaded from a database, say),
	// which the name-based registry cannot express: Register is
	// single-shot per name by design and holds its entries for the life of
	// the process. When both are set, LayoutSpec wins.
	LayoutSpec Layout
	// Company is the payer sending the file.
	Company Company
	// Account is the bank account the file's payments are debited from.
	Account Account
	// NSA is the file sequence number ("Número Sequencial do Arquivo"),
	// a positive, incrementing number the caller controls across the
	// files it sends to a given bank.
	NSA int
	// FileDensity is the recording density some bank manuals require on
	// the file header (e.g. "01600" or "06250"). Optional: a Layout that
	// doesn't bind layout.KeyFileDensity simply never reads this field.
	FileDensity int
	// FileNameLayout is the file name convention File.FileName renders for
	// this file, overriding the process wide default installed by
	// SetRemittanceFileNameLayout. Set it (rather than the global) whenever
	// one process generates files for more than one bank, since each bank
	// names its files differently and the global has room for only one
	// convention. Left zero, the global applies; with no global either,
	// File.FileName keeps its historical "<LAYOUT>_<NSA>_<YYYYMMDD>.REM".
	FileNameLayout FileNameLayout
	// FileNameValues supplies the values this file's FileNameLayout draws
	// its Custom fields from, keyed by the name each Custom field was
	// created with. Optional: a layout with no Custom field never reads it.
	FileNameValues map[string]string
}
