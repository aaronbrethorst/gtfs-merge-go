package compare

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// DiffType represents the type of difference found
type DiffType int

const (
	// RowMissing indicates a row exists in expected but not actual
	RowMissing DiffType = iota
	// RowExtra indicates a row exists in actual but not expected
	RowExtra
	// RowDifferent indicates a row exists in both but has different content
	RowDifferent
	// ColumnMissing indicates a column exists in expected but not actual
	ColumnMissing
)

func (d DiffType) String() string {
	switch d {
	case RowMissing:
		return "RowMissing"
	case RowExtra:
		return "RowExtra"
	case RowDifferent:
		return "RowDifferent"
	case ColumnMissing:
		return "ColumnMissing"
	default:
		return fmt.Sprintf("DiffType(%d)", d)
	}
}

// Difference represents a single difference between two files
type Difference struct {
	Type     DiffType
	Location string // Row key or line number
	Expected string
	Actual   string
}

// DiffResult represents differences between two GTFS files
type DiffResult struct {
	File        string
	Differences []Difference
}

// CompareGTFS compares two GTFS outputs and returns differences
// Both paths should be zip files or directories containing GTFS data
func CompareGTFS(expectedPath, actualPath string) ([]DiffResult, error) {
	// Read both GTFS archives
	expectedFiles, err := readGTFSFiles(expectedPath)
	if err != nil {
		return nil, fmt.Errorf("reading expected GTFS: %w", err)
	}

	actualFiles, err := readGTFSFiles(actualPath)
	if err != nil {
		return nil, fmt.Errorf("reading actual GTFS: %w", err)
	}

	var results []DiffResult

	// Compare all files from expected
	for filename, expectedContent := range expectedFiles {
		actualContent, exists := actualFiles[filename]
		if !exists {
			// File missing in actual
			results = append(results, DiffResult{
				File: filename,
				Differences: []Difference{{
					Type:     RowMissing,
					Location: "file",
					Expected: fmt.Sprintf("file exists (%d bytes)", len(expectedContent)),
					Actual:   "file missing",
				}},
			})
			continue
		}

		// Normalize both files
		normalizedExpected, err := NormalizeCSV(filename, expectedContent)
		if err != nil {
			return nil, fmt.Errorf("normalizing expected %s: %w", filename, err)
		}

		normalizedActual, err := NormalizeCSV(filename, actualContent)
		if err != nil {
			return nil, fmt.Errorf("normalizing actual %s: %w", filename, err)
		}

		// Compare normalized content
		diff, err := CompareCSV(filename, normalizedExpected, normalizedActual)
		if err != nil {
			return nil, fmt.Errorf("comparing %s: %w", filename, err)
		}

		if diff != nil && len(diff.Differences) > 0 {
			results = append(results, *diff)
		}
	}

	// Check for extra files in actual
	for filename := range actualFiles {
		if _, exists := expectedFiles[filename]; !exists {
			results = append(results, DiffResult{
				File: filename,
				Differences: []Difference{{
					Type:     RowExtra,
					Location: "file",
					Expected: "file missing",
					Actual:   fmt.Sprintf("file exists (%d bytes)", len(actualFiles[filename])),
				}},
			})
		}
	}

	return results, nil
}

// CompareCSV compares two normalized CSV contents
func CompareCSV(filename string, expected, actual []byte) (*DiffResult, error) {
	expectedLines := splitLines(expected)
	actualLines := splitLines(actual)

	if len(expectedLines) == 0 && len(actualLines) == 0 {
		return nil, nil
	}

	var diffs []Difference

	// Compare headers
	if len(expectedLines) > 0 && len(actualLines) > 0 {
		if expectedLines[0] != actualLines[0] {
			diffs = append(diffs, Difference{
				Type:     RowDifferent,
				Location: "header",
				Expected: expectedLines[0],
				Actual:   actualLines[0],
			})
		}
	}

	// Build maps of rows by their primary key
	primaryKey := PrimaryKey(filename)
	expectedRows := buildRowMap(expectedLines[1:], expectedLines[0], primaryKey)
	actualRows := buildRowMap(actualLines[1:], actualLines[0], primaryKey)

	// Find missing and different rows
	for key, expectedRow := range expectedRows {
		actualRow, exists := actualRows[key]
		if !exists {
			diffs = append(diffs, Difference{
				Type:     RowMissing,
				Location: key,
				Expected: expectedRow,
				Actual:   "",
			})
		} else if expectedRow != actualRow {
			diffs = append(diffs, Difference{
				Type:     RowDifferent,
				Location: key,
				Expected: expectedRow,
				Actual:   actualRow,
			})
		}
	}

	// Find extra rows
	for key, actualRow := range actualRows {
		if _, exists := expectedRows[key]; !exists {
			diffs = append(diffs, Difference{
				Type:     RowExtra,
				Location: key,
				Expected: "",
				Actual:   actualRow,
			})
		}
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return &DiffResult{
		File:        filename,
		Differences: diffs,
	}, nil
}

// readGTFSFiles reads all CSV files from a GTFS zip or directory
func readGTFSFiles(path string) (map[string][]byte, error) {
	// Try to open as zip file
	r, err := zip.OpenReader(path)
	if err == nil {
		defer func() { _ = r.Close() }()
		return readFromZip(r)
	}

	// TODO: Add directory reading support if needed
	return nil, fmt.Errorf("could not read GTFS from %s: %w", path, err)
}

// readFromZip reads all CSV files from a zip archive
func readFromZip(r *zip.ReadCloser) (map[string][]byte, error) {
	files := make(map[string][]byte)

	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Get base filename (handle nested directories)
		filename := filepath.Base(f.Name)

		// Only read .txt files
		if !strings.HasSuffix(filename, ".txt") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", f.Name, err)
		}

		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", f.Name, err)
		}

		files[filename] = content
	}

	return files, nil
}

// splitLines splits content into lines, trimming empty trailing lines
func splitLines(content []byte) []string {
	lines := strings.Split(string(content), "\n")

	// Remove trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// buildRowMap builds a map of rows keyed by their primary key values
func buildRowMap(rows []string, header string, primaryKey []string) map[string]string {
	result := make(map[string]string)

	if len(primaryKey) == 0 {
		// No primary key, use line number
		for i, row := range rows {
			result[fmt.Sprintf("line_%d", i+1)] = row
		}
		return result
	}

	// Parse header to find primary key column indices
	headerCols := parseCSVLine(header)
	keyIndices := make([]int, len(primaryKey))
	for i, key := range primaryKey {
		keyIndices[i] = -1
		for j, col := range headerCols {
			if col == key {
				keyIndices[i] = j
				break
			}
		}
	}

	// Build map using primary key
	for _, row := range rows {
		cols := parseCSVLine(row)
		var keyParts []string
		for _, idx := range keyIndices {
			if idx >= 0 && idx < len(cols) {
				keyParts = append(keyParts, cols[idx])
			} else {
				keyParts = append(keyParts, "")
			}
		}
		key := strings.Join(keyParts, "|")
		result[key] = row
	}

	return result
}

// parseCSVLine parses a single CSV line into fields
// This is a simple parser that handles basic quoting
func parseCSVLine(line string) []string {
	var fields []string
	var current bytes.Buffer
	inQuote := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			if inQuote && i+1 < len(line) && line[i+1] == '"' {
				// Escaped quote
				current.WriteByte('"')
				i++
			} else {
				inQuote = !inQuote
			}
		case c == ',' && !inQuote:
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	fields = append(fields, current.String())

	return fields
}

// FormatDiffStyleOutput formats differences in a unified diff-like format
// for clear terminal output when tests fail
func FormatDiffStyleOutput(results []DiffResult) string {
	var buf bytes.Buffer

	for _, result := range results {
		if len(result.Differences) == 0 {
			continue
		}

		// Header similar to unified diff
		buf.WriteString(fmt.Sprintf("--- expected/%s\n", result.File))
		buf.WriteString(fmt.Sprintf("+++ actual/%s\n", result.File))
		buf.WriteString(fmt.Sprintf("@@ %d difference(s) @@\n", len(result.Differences)))

		for _, diff := range result.Differences {
			switch diff.Type {
			case RowMissing:
				buf.WriteString(fmt.Sprintf("- [%s] %s\n", diff.Location, diff.Expected))
			case RowExtra:
				buf.WriteString(fmt.Sprintf("+ [%s] %s\n", diff.Location, diff.Actual))
			case RowDifferent:
				buf.WriteString(fmt.Sprintf("! [%s]\n", diff.Location))
				buf.WriteString(fmt.Sprintf("-   %s\n", diff.Expected))
				buf.WriteString(fmt.Sprintf("+   %s\n", diff.Actual))
			case ColumnMissing:
				buf.WriteString(fmt.Sprintf("! column missing: %s\n", diff.Expected))
			}
		}
		buf.WriteString("\n")
	}

	return buf.String()
}
