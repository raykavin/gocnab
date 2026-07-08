package cnab

// Config holds everything NewRemittance needs to start a new remittance
// file.
type Config struct {
	// Layout is the name of a registered Layout, e.g. "febraban240".
	Layout string
	// Company is the payer sending the file.
	Company Company
	// Account is the bank account the file's payments are debited from.
	Account Account
	// NSA is the file sequence number ("Número Sequencial do Arquivo"),
	// a positive, incrementing number the caller controls across the
	// files it sends to a given bank.
	NSA int
}
