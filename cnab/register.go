package cnab

import (
	"github.com/raykavin/gocnab/cnab/layout"

	// The bundled reference layout self-registers on import, so any
	// program that imports only cnab already has "febraban240" available,
	// with no extra import required.
	_ "github.com/raykavin/gocnab/layouts/febraban240"
)

// Layout is the contract a bank/product field layout must implement to
// be usable with this SDK. It is an alias for layout.Layout so that
// layout authors only need to depend on the small cnab/layout package,
// while SDK users can spell the type as cnab.Layout.
type Layout = layout.Layout

// RegisterLayout makes a Layout available under name for later use in
// Config.Layout. It is meant to be called once from a bank package's
// init function, mirroring the driver registration pattern used by
// database/sql. RegisterLayout panics if l is nil, if name is empty, if
// l.Version() is empty, or if a layout is already registered under name.
func RegisterLayout(name string, l Layout) {
	layout.Register(name, l)
}
