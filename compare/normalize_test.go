package compare

import (
	"strings"
	"testing"
)

func TestNormalizeRowOrder(t *testing.T) {
	// Given: CSV with rows in random order
	input := `stop_id,stop_name,stop_lat,stop_lon
S3,Stop C,47.6,122.3
S1,Stop A,47.5,122.1
S2,Stop B,47.55,122.2
`
	// When: normalized by primary key (stop_id)
	result, err := NormalizeCSV("stops.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: rows are sorted by stop_id
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (header + 3 rows), got %d", len(lines))
	}

	// Check order: S1, S2, S3
	if !strings.HasPrefix(lines[1], "S1,") {
		t.Errorf("expected first data row to start with S1, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "S2,") {
		t.Errorf("expected second data row to start with S2, got: %s", lines[2])
	}
	if !strings.HasPrefix(lines[3], "S3,") {
		t.Errorf("expected third data row to start with S3, got: %s", lines[3])
	}
}

func TestNormalizeColumnOrder(t *testing.T) {
	// Given: CSV with columns in non-standard order
	input := `stop_name,stop_id,stop_lon,stop_lat
Stop A,S1,122.1,47.5
`
	// When: normalized
	result, err := NormalizeCSV("stops.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: columns match GTFS specification order
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	header := lines[0]

	// stop_id should come before stop_name in canonical order
	stopIDIdx := strings.Index(header, "stop_id")
	stopNameIdx := strings.Index(header, "stop_name")
	if stopIDIdx > stopNameIdx {
		t.Errorf("stop_id should come before stop_name in canonical order, got header: %s", header)
	}
}

func TestNormalizeFloatPrecision(t *testing.T) {
	// Given: lat/lon with varying precision
	input := `stop_id,stop_name,stop_lat,stop_lon
S1,Stop A,47.6,122.3
S2,Stop B,47.600000000001,122.299999999999
`
	// When: normalized
	result, err := NormalizeCSV("stops.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: both have consistent precision
	content := string(result)
	// Should have normalized floats
	if !strings.Contains(content, "47.600000") {
		t.Errorf("expected normalized float 47.600000, got: %s", content)
	}
}

func TestNormalizeEmptyFields(t *testing.T) {
	// Given: CSV with some fields empty vs missing
	input := `agency_id,agency_name,agency_url,agency_timezone,agency_phone
A1,Agency One,http://example.com,America/Los_Angeles,
A2,Agency Two,http://example2.com,America/Los_Angeles,555-1234
`
	// When: normalized
	result, err := NormalizeCSV("agency.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: empty fields are consistently represented
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	// Just verify it parses without error
}

func TestNormalizeWhitespace(t *testing.T) {
	// Given: CSV with trailing whitespace and different line endings
	input := "stop_id,stop_name,stop_lat,stop_lon\r\nS1,Stop A,47.5,122.1  \r\nS2,Stop B,47.6,122.2\r\n"

	// When: normalized
	result, err := NormalizeCSV("stops.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: whitespace is normalized (LF line endings, no trailing spaces)
	content := string(result)
	if strings.Contains(content, "\r") {
		t.Errorf("result should not contain CR characters")
	}
	if strings.Contains(content, "  \n") {
		t.Errorf("result should not have trailing spaces before newlines")
	}
}

func TestNormalizeAgencyTxt(t *testing.T) {
	input := `agency_name,agency_id,agency_url,agency_timezone
Agency One,A1,http://example.com,America/Los_Angeles
`
	result, err := NormalizeCSV("agency.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Verify header has canonical column order
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if !strings.HasPrefix(lines[0], "agency_id,") {
		t.Errorf("agency_id should be first column, got: %s", lines[0])
	}
}

func TestNormalizeStopsTxt(t *testing.T) {
	input := `stop_name,stop_id,stop_lon,stop_lat
Stop A,S1,-122.1,47.5
Stop B,S2,-122.2,47.6
`
	result, err := NormalizeCSV("stops.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if !strings.HasPrefix(lines[0], "stop_id,") {
		t.Errorf("stop_id should be first column, got: %s", lines[0])
	}
}

func TestNormalizeStopTimesTxt(t *testing.T) {
	// stop_times has composite key: trip_id + stop_sequence
	input := `trip_id,stop_sequence,arrival_time,departure_time,stop_id
T1,2,08:05:00,08:05:00,S2
T1,1,08:00:00,08:00:00,S1
T2,1,09:00:00,09:00:00,S1
`
	result, err := NormalizeCSV("stop_times.txt", []byte(input))
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Verify sorted by trip_id, then stop_sequence
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// First data row should be T1,1 (trip T1, sequence 1)
	if !strings.Contains(lines[1], "T1") || !strings.Contains(lines[1], "08:00:00") {
		t.Errorf("first data row should be T1 sequence 1, got: %s", lines[1])
	}
}

func TestPrimaryKey(t *testing.T) {
	tests := []struct {
		filename string
		expected []string
	}{
		{"agency.txt", []string{"agency_id"}},
		{"stops.txt", []string{"stop_id"}},
		{"routes.txt", []string{"route_id"}},
		{"trips.txt", []string{"trip_id"}},
		{"stop_times.txt", []string{"trip_id", "stop_sequence"}},
		{"calendar.txt", []string{"service_id"}},
		{"calendar_dates.txt", []string{"service_id", "date"}},
		{"shapes.txt", []string{"shape_id", "shape_pt_sequence"}},
	}

	for _, tc := range tests {
		got := PrimaryKey(tc.filename)
		if len(got) != len(tc.expected) {
			t.Errorf("PrimaryKey(%s): expected %v, got %v", tc.filename, tc.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("PrimaryKey(%s): expected %v, got %v", tc.filename, tc.expected, got)
				break
			}
		}
	}
}

func TestGTFSColumnOrder(t *testing.T) {
	// Verify column order for agency.txt
	order := GTFSColumnOrder("agency.txt")
	if len(order) == 0 {
		t.Fatal("GTFSColumnOrder(agency.txt) returned empty")
	}
	if order[0] != "agency_id" {
		t.Errorf("expected agency_id first, got %s", order[0])
	}

	// Verify column order for stops.txt
	order = GTFSColumnOrder("stops.txt")
	if len(order) == 0 {
		t.Fatal("GTFSColumnOrder(stops.txt) returned empty")
	}
	if order[0] != "stop_id" {
		t.Errorf("expected stop_id first, got %s", order[0])
	}
}

func TestStripUTF8BOM(t *testing.T) {
	// Given: CSV with UTF-8 BOM
	bom := []byte{0xEF, 0xBB, 0xBF}
	input := append(bom, []byte(`stop_id,stop_name
S1,Stop A
`)...)

	// When: normalized
	result, err := NormalizeCSV("stops.txt", input)
	if err != nil {
		t.Fatalf("NormalizeCSV failed: %v", err)
	}

	// Then: BOM is stripped
	if len(result) >= 3 && result[0] == 0xEF && result[1] == 0xBB && result[2] == 0xBF {
		t.Error("BOM should be stripped from output")
	}
}
