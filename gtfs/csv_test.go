package gtfs

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestParseCSVHeader(t *testing.T) {
	input := "agency_id,agency_name,agency_url,agency_timezone\n"
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"agency_id", "agency_name", "agency_url", "agency_timezone"}
	if len(header) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(header))
	}

	for i, col := range expected {
		if header[i] != col {
			t.Errorf("column %d: expected %q, got %q", i, col, header[i])
		}
	}
}

func TestParseCSVWithAllFields(t *testing.T) {
	input := `agency_id,agency_name,agency_url,agency_timezone
1,Test Agency,http://example.com,America/Los_Angeles
`
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}

	row := NewCSVRow(header, record)

	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}
	if got := row.Get("agency_name"); got != "Test Agency" {
		t.Errorf("agency_name: expected %q, got %q", "Test Agency", got)
	}
	if got := row.Get("agency_url"); got != "http://example.com" {
		t.Errorf("agency_url: expected %q, got %q", "http://example.com", got)
	}
	if got := row.Get("agency_timezone"); got != "America/Los_Angeles" {
		t.Errorf("agency_timezone: expected %q, got %q", "America/Los_Angeles", got)
	}
}

func TestParseCSVWithMissingOptionalFields(t *testing.T) {
	input := `agency_id,agency_name,agency_url
1,Test Agency,http://example.com
`
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}

	row := NewCSVRow(header, record)

	// Existing field should work
	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}

	// Missing field should return empty string
	if got := row.Get("agency_timezone"); got != "" {
		t.Errorf("agency_timezone: expected empty string, got %q", got)
	}
}

func TestParseCSVWithExtraFields(t *testing.T) {
	input := `agency_id,agency_name,unknown_field,agency_url
1,Test Agency,extra_value,http://example.com
`
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}

	row := NewCSVRow(header, record)

	// Should be able to access known fields
	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}
	if got := row.Get("agency_url"); got != "http://example.com" {
		t.Errorf("agency_url: expected %q, got %q", "http://example.com", got)
	}

	// Should also be able to access the extra field (we don't reject them)
	if got := row.Get("unknown_field"); got != "extra_value" {
		t.Errorf("unknown_field: expected %q, got %q", "extra_value", got)
	}
}

func TestParseCSVEmptyFile(t *testing.T) {
	input := "agency_id,agency_name,agency_url\n"
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	if len(header) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(header))
	}

	// Should return io.EOF for no more records
	record, err := reader.ReadRecord()
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got err=%v, record=%v", err, record)
	}
}

func TestParseCSVQuotedFields(t *testing.T) {
	input := `agency_id,agency_name,agency_url
1,"Test, Agency (with commas)",http://example.com
2,"Agency with ""quotes""",http://example2.com
`
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	// First record with comma in quoted field
	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}
	row := NewCSVRow(header, record)
	if got := row.Get("agency_name"); got != "Test, Agency (with commas)" {
		t.Errorf("agency_name: expected %q, got %q", "Test, Agency (with commas)", got)
	}

	// Second record with escaped quotes
	record, err = reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}
	row = NewCSVRow(header, record)
	if got := row.Get("agency_name"); got != `Agency with "quotes"` {
		t.Errorf("agency_name: expected %q, got %q", `Agency with "quotes"`, got)
	}
}

func TestParseCSVUTF8BOM(t *testing.T) {
	// UTF-8 BOM is 0xEF 0xBB 0xBF
	input := "\xEF\xBB\xBFagency_id,agency_name\n1,Test Agency\n"
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	// The BOM should be stripped from the first column
	if header[0] != "agency_id" {
		t.Errorf("expected first column to be %q, got %q (BOM not stripped)", "agency_id", header[0])
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}

	row := NewCSVRow(header, record)
	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}
}

func TestParseCSVTrailingNewline(t *testing.T) {
	// Multiple trailing newlines should not create empty records
	input := "agency_id,agency_name\n1,Test Agency\n\n\n"
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	// First record
	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}
	row := NewCSVRow(header, record)
	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}

	// Should return io.EOF for no more records (trailing newlines ignored)
	record, err = reader.ReadRecord()
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got err=%v, record=%v", err, record)
	}
}

func TestParseCSVReadRecordBeforeHeader(t *testing.T) {
	input := "agency_id,agency_name\n1,Test Agency\n"
	reader := NewCSVReader(strings.NewReader(input))

	// Should error when ReadRecord is called before ReadHeader
	_, err := reader.ReadRecord()
	if !errors.Is(err, ErrHeaderNotRead) {
		t.Errorf("expected ErrHeaderNotRead, got %v", err)
	}
}

func TestParseCSVWhitespaceInColumnNames(t *testing.T) {
	// Column names with leading/trailing whitespace should be trimmed
	input := " agency_id , agency_name ,agency_url\n1,Test Agency,http://example.com\n"
	reader := NewCSVReader(strings.NewReader(input))

	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	// All column names should be trimmed
	expected := []string{"agency_id", "agency_name", "agency_url"}
	for i, col := range expected {
		if header[i] != col {
			t.Errorf("column %d: expected %q, got %q", i, col, header[i])
		}
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("unexpected error reading record: %v", err)
	}

	row := NewCSVRow(header, record)
	if got := row.Get("agency_id"); got != "1" {
		t.Errorf("agency_id: expected %q, got %q", "1", got)
	}
}

func TestCSVRowGetIntInvalid(t *testing.T) {
	header := []string{"valid", "invalid", "float"}
	record := []string{"123", "abc", "123.45"}
	row := NewCSVRow(header, record)

	// Valid int
	if got := row.GetInt("valid"); got != 123 {
		t.Errorf("expected 123, got %d", got)
	}
	// Invalid string should return 0
	if got := row.GetInt("invalid"); got != 0 {
		t.Errorf("expected 0 for invalid string, got %d", got)
	}
	// Float string should return 0 (not parsed as int)
	if got := row.GetInt("float"); got != 0 {
		t.Errorf("expected 0 for float string, got %d", got)
	}
}

func TestCSVRowGetFloatInvalid(t *testing.T) {
	header := []string{"valid", "invalid"}
	record := []string{"123.45", "abc"}
	row := NewCSVRow(header, record)

	// Valid float
	if got := row.GetFloat("valid"); got != 123.45 {
		t.Errorf("expected 123.45, got %f", got)
	}
	// Invalid string should return 0.0
	if got := row.GetFloat("invalid"); got != 0.0 {
		t.Errorf("expected 0.0 for invalid string, got %f", got)
	}
}

func TestCSVRowGetInt(t *testing.T) {
	header := []string{"id", "value", "empty"}
	record := []string{"123", "456", ""}
	row := NewCSVRow(header, record)

	if got := row.GetInt("id"); got != 123 {
		t.Errorf("expected 123, got %d", got)
	}
	if got := row.GetInt("value"); got != 456 {
		t.Errorf("expected 456, got %d", got)
	}
	// Empty or missing should return 0
	if got := row.GetInt("empty"); got != 0 {
		t.Errorf("expected 0 for empty, got %d", got)
	}
	if got := row.GetInt("missing"); got != 0 {
		t.Errorf("expected 0 for missing, got %d", got)
	}
}

func TestCSVRowGetFloat(t *testing.T) {
	header := []string{"lat", "lon", "empty"}
	record := []string{"47.6062", "-122.3321", ""}
	row := NewCSVRow(header, record)

	if got := row.GetFloat("lat"); got != 47.6062 {
		t.Errorf("expected 47.6062, got %f", got)
	}
	if got := row.GetFloat("lon"); got != -122.3321 {
		t.Errorf("expected -122.3321, got %f", got)
	}
	// Empty or missing should return 0.0
	if got := row.GetFloat("empty"); got != 0.0 {
		t.Errorf("expected 0.0 for empty, got %f", got)
	}
	if got := row.GetFloat("missing"); got != 0.0 {
		t.Errorf("expected 0.0 for missing, got %f", got)
	}
}

func TestCSVRowGetBool(t *testing.T) {
	header := []string{"mon", "tue", "wed", "thu", "empty"}
	record := []string{"1", "0", "true", "false", ""}
	row := NewCSVRow(header, record)

	if got := row.GetBool("mon"); !got {
		t.Errorf("expected true for '1'")
	}
	if got := row.GetBool("tue"); got {
		t.Errorf("expected false for '0'")
	}
	if got := row.GetBool("wed"); !got {
		t.Errorf("expected true for 'true'")
	}
	if got := row.GetBool("thu"); got {
		t.Errorf("expected false for 'false'")
	}
	// Empty should return false
	if got := row.GetBool("empty"); got {
		t.Errorf("expected false for empty")
	}
}
