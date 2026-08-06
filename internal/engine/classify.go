package engine

import "strconv"

// recordTypeColumn and segmentCodeColumn are fixed by the FEBRABAN CNAB 240
// standard itself, the same way recordWidth is: every bank layout, remessa
// or retorno, puts the record type marker at column 8 and, for a type "3"
// detail record, the segment letter at column 14. A Layout never needs to
// declare these positions (they are already baked into every RecordSpec's
// RecordType/SegmentCode Const field); ClassifyLine reads them directly so
// a caller can tell which RecordKey a line plays before it has any other
// way to know.
const (
	recordTypeColumn  = 8
	segmentCodeColumn = 14
)

// Record type markers, FEBRABAN CNAB 240 column 8.
const (
	RecordTypeFileHeader   = "0"
	RecordTypeBatchHeader  = "1"
	RecordTypeDetail       = "3"
	RecordTypeBatchTrailer = "5"
	RecordTypeFileTrailer  = "9"
)

// ClassifyLine reads line's record type (column 8) and, for a detail
// record, its segment code (column 14, e.g. "A", "B", "J"), without
// consulting any Layout both positions are FEBRABAN-standard constants,
// not something a bank layout can move. segmentCode is empty for anything
// other than a detail record.
//
// It returns a *RecordParseError if line is shorter than column
// segmentCodeColumn; it does not validate that recordType is one of the
// five known markers, since an unrecognized marker is exactly the signal a
// caller needs to detect a return file it does not understand reporting
// it as data lets the caller decide, rather than ClassifyLine failing on
// their behalf.
func ClassifyLine(line string) (recordType string, segmentCode string, err error) {
	if len(line) < segmentCodeColumn {
		return "", "", &RecordParseError{Reason: "line has fewer than " + strconv.Itoa(segmentCodeColumn) + " characters"}
	}
	recordType = line[recordTypeColumn-1 : recordTypeColumn]
	if recordType == RecordTypeDetail {
		segmentCode = line[segmentCodeColumn-1 : segmentCodeColumn]
	}
	return recordType, segmentCode, nil
}
