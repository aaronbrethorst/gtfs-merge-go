package gtfs

import (
	"bytes"
	"strings"
	"testing"
)

// TestWriteCSVHeader verifies that the CSV writer writes the correct header row
func TestWriteCSVHeader(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"agency_id", "agency_name", "agency_url", "agency_timezone"}
	err := w.WriteHeader(header)
	if err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expected := "agency_id,agency_name,agency_url,agency_timezone\n"
	if buf.String() != expected {
		t.Errorf("header mismatch:\ngot:  %q\nwant: %q", buf.String(), expected)
	}
}

// TestWriteCSVRecord verifies that the CSV writer writes data records correctly
func TestWriteCSVRecord(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_name", "stop_lat", "stop_lon"}
	_ = w.WriteHeader(header)

	err := w.WriteRecord([]string{"S1", "Main Street", "47.123", "-122.456"})
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expected := "stop_id,stop_name,stop_lat,stop_lon\nS1,Main Street,47.123,-122.456\n"
	if buf.String() != expected {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", buf.String(), expected)
	}
}

// TestWriteCSVEmptySlice verifies that the CSV writer handles an empty record set
func TestWriteCSVEmptySlice(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"route_id", "route_short_name", "route_long_name"}
	_ = w.WriteHeader(header)

	// No records written
	err := w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Should only have header
	expected := "route_id,route_short_name,route_long_name\n"
	if buf.String() != expected {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", buf.String(), expected)
	}
}

// TestWriteCSVEscapeCommas verifies that commas in values are properly escaped
func TestWriteCSVEscapeCommas(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_name"}
	_ = w.WriteHeader(header)

	// Name contains a comma
	err := w.WriteRecord([]string{"S1", "Main Street, Downtown"})
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Value with comma should be quoted
	output := buf.String()
	if !strings.Contains(output, `"Main Street, Downtown"`) {
		t.Errorf("comma in value should be quoted:\ngot: %q", output)
	}
}

// TestWriteCSVEscapeQuotes verifies that quotes in values are properly escaped
func TestWriteCSVEscapeQuotes(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_name"}
	_ = w.WriteHeader(header)

	// Name contains quotes
	err := w.WriteRecord([]string{"S1", `The "Main" Stop`})
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Quotes should be doubled and the value should be quoted
	output := buf.String()
	if !strings.Contains(output, `"The ""Main"" Stop"`) {
		t.Errorf("quotes in value should be escaped:\ngot: %q", output)
	}
}

// TestWriteCSVEscapeNewlines verifies that newlines in values are properly escaped
func TestWriteCSVEscapeNewlines(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_desc"}
	_ = w.WriteHeader(header)

	// Description contains a newline
	err := w.WriteRecord([]string{"S1", "Line 1\nLine 2"})
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Value with newline should be quoted
	output := buf.String()
	if !strings.Contains(output, `"Line 1`) {
		t.Errorf("newline in value should be quoted:\ngot: %q", output)
	}
}

// TestWriteCSVMultipleRecords verifies that multiple records are written correctly
func TestWriteCSVMultipleRecords(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_name"}
	_ = w.WriteHeader(header)

	_ = w.WriteRecord([]string{"S1", "First Stop"})
	_ = w.WriteRecord([]string{"S2", "Second Stop"})
	_ = w.WriteRecord([]string{"S3", "Third Stop"})

	err := w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 4 { // header + 3 records
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

// TestWriteCSVEmptyValues verifies that empty values are written correctly
func TestWriteCSVEmptyValues(t *testing.T) {
	var buf bytes.Buffer
	w := NewCSVWriter(&buf)

	header := []string{"stop_id", "stop_code", "stop_name"}
	_ = w.WriteHeader(header)

	// stop_code is empty
	err := w.WriteRecord([]string{"S1", "", "Main Stop"})
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	err = w.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expected := "stop_id,stop_code,stop_name\nS1,,Main Stop\n"
	if buf.String() != expected {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", buf.String(), expected)
	}
}
