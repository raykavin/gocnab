package cnab

// Cents is a monetary amount expressed as an integer number of the
// currency's minor unit (e.g. R$ 252,00 is Cents(25200)). Amounts are
// never represented as floating point anywhere in this SDK, so a payment
// value can never suffer binary floating point rounding error.
type Cents int64
