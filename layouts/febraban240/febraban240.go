// Package febraban240 implements the standard FEBRABAN CNAB 240 layout.
// It self-registers as "febraban240" and models payment-specific segment
// variants as separate record keys.
package febraban240

import "github.com/raykavin/gocnab/cnab/layout"

// version is the CNAB layout version this package implements, written
// into every "número da versão do layout" field. It is also returned by
// Version.
const version = "081"

// bankCode is the compensation bank code rendered in every record. This
// reference layout does not represent a specific bank, so it uses a
// placeholder; a real bank layout derived from this package should
// override every occurrence with its own COMPE code.
const bankCode = "000"

type febraban240 struct{}

func (febraban240) Name() string    { return "febraban240" }
func (febraban240) Version() string { return version }

func (febraban240) Record(key layout.RecordKey) (layout.RecordSpec, bool) {
	switch key {
	case layout.FileHeader:
		return fileHeaderSpec, true
	case layout.FileTrailer:
		return fileTrailerSpec, true
	case layout.BatchHeader:
		return batchHeaderSpec, true
	case layout.BatchTrailer:
		return batchTrailerSpec, true
	case layout.SegmentA:
		return segmentASpec, true
	case layout.SegmentB:
		return segmentBSpec, true
	case layout.SegmentBPix:
		return segmentBPixSpec, true
	case layout.SegmentJ:
		return segmentJSpec, true
	case layout.SegmentJ52:
		return segmentJ52Spec, true
	case layout.SegmentO:
		return segmentOSpec, true
	case layout.SegmentN:
		return segmentNSpec, true
	case layout.SegmentNSimple:
		return segmentNSimpleSpec, true
	case layout.SegmentNSocial:
		return segmentNSocialSpec, true
	default:
		return layout.RecordSpec{}, false
	}
}

func init() {
	layout.Register("febraban240", febraban240{})
}

// --- shared field constructors, used by every record file in this package ---

func numeric(name string, start, end int, key layout.Key) layout.FieldSpec {
	return layout.FieldSpec{Name: name, Start: start, End: end, Kind: layout.KindNumeric, Key: key}
}

func numericDecimal(name string, start, end, decimals int, key layout.Key) layout.FieldSpec {
	return layout.FieldSpec{Name: name, Start: start, End: end, Kind: layout.KindNumeric, Decimals: decimals, Key: key}
}

func alpha(name string, start, end int, key layout.Key) layout.FieldSpec {
	return layout.FieldSpec{Name: name, Start: start, End: end, Kind: layout.KindAlphanumeric, Key: key}
}

func numericConst(name string, start, end int, value string) layout.FieldSpec {
	return layout.FieldSpec{Name: name, Start: start, End: end, Kind: layout.KindNumeric, Const: value}
}

func alphaConst(name string, start, end int, value string) layout.FieldSpec {
	return layout.FieldSpec{Name: name, Start: start, End: end, Kind: layout.KindAlphanumeric, Const: value}
}

func numericFiller(start, end int) layout.FieldSpec {
	return layout.FieldSpec{Name: "Filler", Start: start, End: end, Kind: layout.KindNumeric}
}

func alphaFiller(start, end int) layout.FieldSpec {
	return layout.FieldSpec{Name: "Filler", Start: start, End: end, Kind: layout.KindAlphanumeric}
}
