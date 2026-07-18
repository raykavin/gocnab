package layout

import (
	"fmt"
	"sort"
	"sync"
)

// Layout is implemented by a bank/product specific field layout. The
// engine (internal/engine) and the public cnab package consume a Layout
// through this interface only; neither has any hardcoded knowledge of a
// specific bank.
type Layout interface {
	// Name returns the layout identifier used to register and look it up,
	// e.g. "febraban240".
	Name() string
	// Version returns the CNAB layout version this Layout implements. It
	// must not be empty; the engine rejects a Layout with an empty
	// version at registration time.
	Version() string
	// Record returns the RecordSpec for the given key, and ok=false when
	// this layout does not support that record/segment at all (for
	// example a layout that never implements PIX support would return
	// ok=false for SegmentB when asked in a PIX context, though in
	// practice SegmentB is shared by several payment kinds so the check
	// that matters most in practice happens in the public cnab package
	// against the specific fields a payment kind needs).
	Record(key RecordKey) (spec RecordSpec, ok bool)
}

var (
	registryMu sync.RWMutex
	registry   = map[string]Layout{}
)

// Register makes a Layout available under name for later lookup with
// Lookup. It is meant to be called from a bank package's init function,
// mirroring the registration pattern used by database/sql drivers.
//
// Register panics if l is nil, if l.Name() is empty, if l.Version() is
// empty, or if a layout is already registered under the same name: all of
// these are programming errors detected at start-up time, not runtime
// conditions a caller should recover from.
func Register(name string, l Layout) {
	if l == nil {
		panic("cnab/layout: Register called with a nil Layout")
	}
	if name == "" {
		panic("cnab/layout: Register called with an empty name")
	}
	if l.Version() == "" {
		panic(fmt.Sprintf("cnab/layout: layout %q has an empty Version", name))
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("cnab/layout: Register called twice for layout %q", name))
	}
	registry[name] = l
}

// Lookup returns the layout previously registered under name, and
// ok=false when no layout is registered under that name.
func Lookup(name string) (Layout, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	l, ok := registry[name]
	return l, ok
}

// Names returns the names of every registered layout, sorted
// alphabetically. It is mainly useful for error messages that need to
// suggest valid alternatives.
func Names() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
