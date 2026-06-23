// Package engine implements the CNAB 240 rendering mechanics: field
// padding and alignment, 240 column record assembly, batch/file
// sequencing, trailer computation and the FEBRABAN structural limits. It
// has no knowledge of any specific bank; every position and constant it
// works with comes from a layout.Layout supplied by the caller. This
// package is internal to the module: the public API in the cnab package
// is the only supported entry point.
package engine

import "github.com/raykavin/gocnab/cnab/layout"

// recordKeys lists every record/segment kind the engine knows how to
// render. A Layout does not need to implement all of them: New skips any
// key for which Layout.Record returns ok=false, and Supports reports
// which keys ended up available.
var recordKeys = layout.AllRecordKeys

// Engine renders CNAB 240 files for one Layout. It is safe for concurrent
// use once built, since building it is the only step that mutates state.
type Engine struct {
	layout   layout.Layout
	compiled map[layout.RecordKey]*compiledRecord
}

// New validates l and compiles every record/segment it defines, catching
// gaps, overlaps and a missing layout version up front. It returns a
// *SpecError describing exactly what is wrong when validation fails.
func New(l layout.Layout) (*Engine, error) {
	if l == nil {
		return nil, &SpecError{Reason: "layout must not be nil"}
	}
	if l.Version() == "" {
		return nil, &SpecError{Record: l.Name(), Reason: "layout version must not be empty"}
	}

	compiled := make(map[layout.RecordKey]*compiledRecord, len(recordKeys))
	for _, key := range recordKeys {
		spec, ok := l.Record(key)
		if !ok {
			continue
		}
		rec, err := compileRecord(spec)
		if err != nil {
			return nil, err
		}
		compiled[key] = rec
	}

	return &Engine{layout: l, compiled: compiled}, nil
}

// Supports reports whether the underlying Layout defines the given
// record/segment kind.
func (e *Engine) Supports(key layout.RecordKey) bool {
	_, ok := e.compiled[key]
	return ok
}

// LayoutName returns the name of the Layout this Engine was built from.
func (e *Engine) LayoutName() string {
	return e.layout.Name()
}

// LayoutVersion returns the version of the Layout this Engine was built
// from.
func (e *Engine) LayoutVersion() string {
	return e.layout.Version()
}

func (e *Engine) record(key layout.RecordKey) (*compiledRecord, error) {
	rec, ok := e.compiled[key]
	if !ok {
		return nil, &SpecError{Record: string(key), Reason: "layout \"" + e.layout.Name() + "\" does not define this record"}
	}
	return rec, nil
}
