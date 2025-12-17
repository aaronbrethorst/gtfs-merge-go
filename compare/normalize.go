package compare

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// gtfsColumnOrders defines the canonical column order for each GTFS file
// based on the GTFS specification
var gtfsColumnOrders = map[string][]string{
	"agency.txt": {
		"agency_id", "agency_name", "agency_url", "agency_timezone",
		"agency_lang", "agency_phone", "agency_fare_url", "agency_email",
	},
	"stops.txt": {
		"stop_id", "stop_code", "stop_name", "stop_desc", "stop_lat", "stop_lon",
		"zone_id", "stop_url", "location_type", "parent_station", "stop_timezone",
		"wheelchair_boarding", "level_id", "platform_code",
	},
	"routes.txt": {
		"route_id", "agency_id", "route_short_name", "route_long_name", "route_desc",
		"route_type", "route_url", "route_color", "route_text_color", "route_sort_order",
		"continuous_pickup", "continuous_drop_off",
	},
	"trips.txt": {
		"route_id", "service_id", "trip_id", "trip_headsign", "trip_short_name",
		"direction_id", "block_id", "shape_id", "wheelchair_accessible", "bikes_allowed",
	},
	"stop_times.txt": {
		"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence",
		"stop_headsign", "pickup_type", "drop_off_type", "continuous_pickup",
		"continuous_drop_off", "shape_dist_traveled", "timepoint",
	},
	"calendar.txt": {
		"service_id", "monday", "tuesday", "wednesday", "thursday", "friday",
		"saturday", "sunday", "start_date", "end_date",
	},
	"calendar_dates.txt": {
		"service_id", "date", "exception_type",
	},
	"shapes.txt": {
		"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence", "shape_dist_traveled",
	},
	"frequencies.txt": {
		"trip_id", "start_time", "end_time", "headway_secs", "exact_times",
	},
	"transfers.txt": {
		"from_stop_id", "to_stop_id", "transfer_type", "min_transfer_time",
		"from_route_id", "to_route_id", "from_trip_id", "to_trip_id",
	},
	"fare_attributes.txt": {
		"fare_id", "price", "currency_type", "payment_method", "transfers",
		"agency_id", "transfer_duration", "youth_price", "senior_price",
	},
	"fare_rules.txt": {
		"fare_id", "route_id", "origin_id", "destination_id", "contains_id",
	},
	"feed_info.txt": {
		"feed_publisher_name", "feed_publisher_url", "feed_lang", "default_lang",
		"feed_start_date", "feed_end_date", "feed_version", "feed_contact_email",
		"feed_contact_url", "feed_id",
	},
	"areas.txt": {
		"area_id", "area_name",
	},
	"pathways.txt": {
		"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional",
		"length", "traversal_time", "stair_count", "max_slope", "min_width",
		"signposted_as", "reversed_signposted_as",
	},
}

// gtfsPrimaryKeys defines the primary key columns for each GTFS file
var gtfsPrimaryKeys = map[string][]string{
	"agency.txt":          {"agency_id"},
	"stops.txt":           {"stop_id"},
	"routes.txt":          {"route_id"},
	"trips.txt":           {"trip_id"},
	"stop_times.txt":      {"trip_id", "stop_sequence"},
	"calendar.txt":        {"service_id"},
	"calendar_dates.txt":  {"service_id", "date"},
	"shapes.txt":          {"shape_id", "shape_pt_sequence"},
	"frequencies.txt":     {"trip_id", "start_time"},
	"transfers.txt":       {"from_stop_id", "to_stop_id"},
	"fare_attributes.txt": {"fare_id"},
	"fare_rules.txt":      {"fare_id", "route_id", "origin_id", "destination_id"},
	"feed_info.txt":       {"feed_publisher_name"},
	"areas.txt":           {"area_id"},
	"pathways.txt":        {"pathway_id"},
}

// floatColumns lists columns that should have normalized float precision
var floatColumns = map[string]bool{
	"stop_lat":            true,
	"stop_lon":            true,
	"shape_pt_lat":        true,
	"shape_pt_lon":        true,
	"shape_dist_traveled": true,
	"length":              true,
	"max_slope":           true,
	"min_width":           true,
	"price":               true,
	"youth_price":         true,
	"senior_price":        true,
}

// GTFSColumnOrder returns the canonical column order for a GTFS file
func GTFSColumnOrder(filename string) []string {
	if order, ok := gtfsColumnOrders[filename]; ok {
		return order
	}
	return nil
}

// PrimaryKey returns the primary key columns for a GTFS file
func PrimaryKey(filename string) []string {
	if key, ok := gtfsPrimaryKeys[filename]; ok {
		return key
	}
	return nil
}

// NormalizeCSV normalizes a GTFS CSV file for comparison
// It performs the following normalizations:
// - Strips UTF-8 BOM
// - Normalizes line endings to LF
// - Trims trailing whitespace
// - Reorders columns to canonical GTFS order
// - Sorts rows by primary key
// - Normalizes float precision to 6 decimal places
func NormalizeCSV(filename string, content []byte) ([]byte, error) {
	// Strip UTF-8 BOM if present
	content = stripBOM(content)

	// Normalize line endings
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	content = bytes.ReplaceAll(content, []byte("\r"), []byte("\n"))

	// Parse CSV
	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Get header and create column index map
	header := records[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.TrimSpace(col)] = i
	}

	// Determine output column order
	canonicalOrder := GTFSColumnOrder(filename)
	var outputCols []string
	var outputIndices []int

	if canonicalOrder != nil {
		// Use canonical order, but only include columns that exist in input
		for _, col := range canonicalOrder {
			if idx, ok := colIndex[col]; ok {
				outputCols = append(outputCols, col)
				outputIndices = append(outputIndices, idx)
			}
		}
		// Add any extra columns not in canonical order
		for _, col := range header {
			col = strings.TrimSpace(col)
			found := false
			for _, c := range outputCols {
				if c == col {
					found = true
					break
				}
			}
			if !found {
				outputCols = append(outputCols, col)
				outputIndices = append(outputIndices, colIndex[col])
			}
		}
	} else {
		// No canonical order, use input order
		for i, col := range header {
			outputCols = append(outputCols, strings.TrimSpace(col))
			outputIndices = append(outputIndices, i)
		}
	}

	// Build output records with reordered columns
	outputRecords := make([][]string, len(records))
	outputRecords[0] = outputCols

	for i := 1; i < len(records); i++ {
		row := make([]string, len(outputCols))
		for j, srcIdx := range outputIndices {
			if srcIdx < len(records[i]) {
				val := strings.TrimSpace(records[i][srcIdx])
				// Normalize float columns
				if floatColumns[outputCols[j]] && val != "" {
					val = normalizeFloat(val)
				}
				row[j] = val
			}
		}
		outputRecords[i] = row
	}

	// Normalize shape sequences for shapes.txt before sorting
	// This is needed because Java and Go use global sequence counters during merge
	// but process shapes in different orders, resulting in different sequence values.
	if filename == "shapes.txt" && len(outputRecords) > 1 {
		normalizeShapeSequences(outputRecords[1:], outputCols)
	}

	// Normalize symmetric transfers (where from_stop_id == to_stop_id)
	// to ensure route/trip IDs are in consistent order
	if filename == "transfers.txt" && len(outputRecords) > 1 {
		normalizeSymmetricTransfers(outputRecords[1:], outputCols)
	}

	// Sort rows by primary key
	primaryKey := PrimaryKey(filename)
	if primaryKey != nil && len(outputRecords) > 1 {
		// Get indices of primary key columns in output
		keyIndices := make([]int, len(primaryKey))
		for i, key := range primaryKey {
			for j, col := range outputCols {
				if col == key {
					keyIndices[i] = j
					break
				}
			}
		}

		// Sort data rows (skip header)
		dataRows := outputRecords[1:]
		sort.SliceStable(dataRows, func(i, j int) bool {
			for _, idx := range keyIndices {
				// Try numeric comparison first
				vi, ei := strconv.Atoi(dataRows[i][idx])
				vj, ej := strconv.Atoi(dataRows[j][idx])
				if ei == nil && ej == nil {
					if vi != vj {
						return vi < vj
					}
				} else {
					// Fall back to string comparison
					if dataRows[i][idx] != dataRows[j][idx] {
						return dataRows[i][idx] < dataRows[j][idx]
					}
				}
			}
			return false
		})
	}

	// Write output CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.UseCRLF = false // Use LF only

	for _, record := range outputRecords {
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("writing CSV: %w", err)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flushing CSV: %w", err)
	}

	return buf.Bytes(), nil
}

// stripBOM removes UTF-8 BOM if present
func stripBOM(content []byte) []byte {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}
	return content
}

// normalizeFloat normalizes a float string to 6 decimal places
func normalizeFloat(s string) string {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s // Return original if not a valid float
	}
	return strconv.FormatFloat(f, 'f', 6, 64)
}

// normalizeSymmetricTransfers normalizes transfers where from_stop_id == to_stop_id
// by ensuring route IDs are in a consistent order (lexicographically smaller first).
// This is needed because symmetric transfers (same origin and destination stop) may have
// their route/trip IDs in different orders between Java and Go outputs.
func normalizeSymmetricTransfers(records [][]string, header []string) {
	// Find column indices
	fromStopIdx := -1
	toStopIdx := -1
	fromRouteIdx := -1
	toRouteIdx := -1
	fromTripIdx := -1
	toTripIdx := -1

	for i, col := range header {
		switch col {
		case "from_stop_id":
			fromStopIdx = i
		case "to_stop_id":
			toStopIdx = i
		case "from_route_id":
			fromRouteIdx = i
		case "to_route_id":
			toRouteIdx = i
		case "from_trip_id":
			fromTripIdx = i
		case "to_trip_id":
			toTripIdx = i
		}
	}

	if fromStopIdx < 0 || toStopIdx < 0 {
		return
	}

	for _, row := range records {
		if len(row) <= fromStopIdx || len(row) <= toStopIdx {
			continue
		}

		// Only normalize if from_stop_id == to_stop_id (symmetric transfer)
		if row[fromStopIdx] != row[toStopIdx] {
			continue
		}

		// Normalize route IDs: ensure from_route_id <= to_route_id
		if fromRouteIdx >= 0 && toRouteIdx >= 0 && fromRouteIdx < len(row) && toRouteIdx < len(row) {
			if row[fromRouteIdx] > row[toRouteIdx] {
				row[fromRouteIdx], row[toRouteIdx] = row[toRouteIdx], row[fromRouteIdx]
			}
		}

		// Normalize trip IDs: ensure from_trip_id <= to_trip_id
		if fromTripIdx >= 0 && toTripIdx >= 0 && fromTripIdx < len(row) && toTripIdx < len(row) {
			if row[fromTripIdx] > row[toTripIdx] {
				row[fromTripIdx], row[toTripIdx] = row[toTripIdx], row[fromTripIdx]
			}
		}
	}
}

// normalizeShapeSequences re-assigns shape_pt_sequence values within each shape_id
// to ensure consistent comparison regardless of original sequence values.
// Both Java and Go use global sequence counters during merge but process shapes in
// different orders, resulting in different sequence values. By normalizing to 1, 2, 3, ...
// within each shape, we can compare the shape geometry regardless of sequence assignment.
func normalizeShapeSequences(records [][]string, header []string) {
	// Find column indices
	shapeIDIdx := -1
	seqIdx := -1
	for i, col := range header {
		switch col {
		case "shape_id":
			shapeIDIdx = i
		case "shape_pt_sequence":
			seqIdx = i
		}
	}

	if shapeIDIdx < 0 || seqIdx < 0 {
		return
	}

	// Group row indices by shape_id
	shapeRows := make(map[string][]int)
	for i, row := range records {
		if len(row) > shapeIDIdx {
			shapeID := row[shapeIDIdx]
			shapeRows[shapeID] = append(shapeRows[shapeID], i)
		}
	}

	// For each shape, sort rows by current sequence and assign new sequences 1, 2, 3, ...
	for _, rowIndices := range shapeRows {
		// Sort row indices by current sequence value
		sort.Slice(rowIndices, func(i, j int) bool {
			ri, rj := records[rowIndices[i]], records[rowIndices[j]]
			if seqIdx < len(ri) && seqIdx < len(rj) {
				vi, ei := strconv.Atoi(ri[seqIdx])
				vj, ej := strconv.Atoi(rj[seqIdx])
				if ei == nil && ej == nil {
					return vi < vj
				}
				return ri[seqIdx] < rj[seqIdx]
			}
			return false
		})

		// Assign new sequences 1, 2, 3, ...
		for newSeq, rowIdx := range rowIndices {
			if seqIdx < len(records[rowIdx]) {
				records[rowIdx][seqIdx] = strconv.Itoa(newSeq + 1)
			}
		}
	}
}
