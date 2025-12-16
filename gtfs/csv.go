package gtfs

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
)

// CSVReader wraps the standard csv.Reader with GTFS-specific handling.
// It handles UTF-8 BOM, trailing newlines, and provides convenient record access.
type CSVReader struct {
	reader     *csv.Reader
	headerRead bool
}

// NewCSVReader creates a new CSVReader from an io.Reader.
func NewCSVReader(r io.Reader) *CSVReader {
	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	csvReader.LazyQuotes = true    // Be lenient with quote handling
	csvReader.TrimLeadingSpace = true
	return &CSVReader{
		reader: csvReader,
	}
}

// ReadHeader reads the header row from the CSV.
// It strips the UTF-8 BOM if present on the first field and trims whitespace
// from all column names. Must be called before ReadRecord.
func (c *CSVReader) ReadHeader() ([]string, error) {
	record, err := c.reader.Read()
	if err != nil {
		return nil, err
	}
	c.headerRead = true

	// Strip UTF-8 BOM from first field if present
	if len(record) > 0 {
		record[0] = stripBOM(record[0])
	}

	// Trim whitespace from all column names
	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	return record, nil
}

// ErrHeaderNotRead is returned when ReadRecord is called before ReadHeader.
var ErrHeaderNotRead = errors.New("must call ReadHeader before ReadRecord")

// ReadRecord reads the next record from the CSV.
// It skips empty lines and returns io.EOF when no more records are available.
// ReadHeader must be called before ReadRecord.
func (c *CSVReader) ReadRecord() ([]string, error) {
	if !c.headerRead {
		return nil, ErrHeaderNotRead
	}

	for {
		record, err := c.reader.Read()
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}

		// Skip empty records (lines with only whitespace or empty fields)
		if isEmptyRecord(record) {
			continue
		}

		return record, nil
	}
}

// stripBOM removes the UTF-8 BOM (Byte Order Mark) from the beginning of a string.
func stripBOM(s string) string {
	const bom = "\xEF\xBB\xBF"
	if strings.HasPrefix(s, bom) {
		return strings.TrimPrefix(s, bom)
	}
	return s
}

// isEmptyRecord returns true if the record contains only empty or whitespace fields.
func isEmptyRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

// CSVRow provides convenient access to CSV record fields by column name.
type CSVRow struct {
	header  []string
	record  []string
	indices map[string]int
}

// NewCSVRow creates a new CSVRow from a header and record.
func NewCSVRow(header, record []string) *CSVRow {
	indices := make(map[string]int, len(header))
	for i, col := range header {
		indices[col] = i
	}
	return &CSVRow{
		header:  header,
		record:  record,
		indices: indices,
	}
}

// Get returns the value of the field with the given column name.
// Returns an empty string if the column doesn't exist.
func (r *CSVRow) Get(column string) string {
	idx, ok := r.indices[column]
	if !ok || idx >= len(r.record) {
		return ""
	}
	return r.record[idx]
}

// GetInt returns the value of the field as an int.
// Returns 0 if the field is empty, missing, or not a valid integer.
func (r *CSVRow) GetInt(column string) int {
	s := r.Get(column)
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// GetFloat returns the value of the field as a float64.
// Returns 0.0 if the field is empty, missing, or not a valid float.
func (r *CSVRow) GetFloat(column string) float64 {
	s := r.Get(column)
	if s == "" {
		return 0.0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// GetBool returns the value of the field as a bool.
// Returns true for "1" or "true" (case-insensitive), false otherwise.
func (r *CSVRow) GetBool(column string) bool {
	s := strings.ToLower(r.Get(column))
	return s == "1" || s == "true"
}
