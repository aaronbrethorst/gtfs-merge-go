package gtfs

import (
	"encoding/csv"
	"io"
)

// CSVWriter wraps the standard csv.Writer for GTFS output.
// It uses standard CSV formatting with CRLF line endings converted to LF.
type CSVWriter struct {
	writer *csv.Writer
}

// NewCSVWriter creates a new CSVWriter that writes to the given io.Writer.
func NewCSVWriter(w io.Writer) *CSVWriter {
	csvWriter := csv.NewWriter(w)
	// Use Unix-style line endings (LF) for consistency
	csvWriter.UseCRLF = false
	return &CSVWriter{
		writer: csvWriter,
	}
}

// WriteHeader writes the header row to the CSV.
func (c *CSVWriter) WriteHeader(header []string) error {
	return c.writer.Write(header)
}

// WriteRecord writes a data record to the CSV.
func (c *CSVWriter) WriteRecord(record []string) error {
	return c.writer.Write(record)
}

// Flush writes any buffered data to the underlying writer.
func (c *CSVWriter) Flush() error {
	c.writer.Flush()
	return c.writer.Error()
}
