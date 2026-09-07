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
}
