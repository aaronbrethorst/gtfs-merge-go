package gtfs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Required GTFS files
var requiredFiles = []string{
	"agency.txt",
	"stops.txt",
	"routes.txt",
	"trips.txt",
	"stop_times.txt",
}

// At least one of these must be present
var calendarFiles = []string{
	"calendar.txt",
	"calendar_dates.txt",
}

// ErrMissingRequiredFile is returned when a required GTFS file is missing
var ErrMissingRequiredFile = errors.New("missing required GTFS file")

// ErrMissingCalendarFile is returned when neither calendar.txt nor calendar_dates.txt is present
var ErrMissingCalendarFile = errors.New("missing calendar.txt or calendar_dates.txt")

// ReadFromPath reads a GTFS feed from a file path (zip or directory)
func ReadFromPath(path string) (*Feed, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access path %s: %w", path, err)
	}

	if info.IsDir() {
		return readFromDirectory(path)
	}

	// Assume it's a zip file
	return readFromZipPath(path)
}

// readFromDirectory reads a GTFS feed from a directory
func readFromDirectory(dirPath string) (*Feed, error) {
	// Check for required files
	for _, filename := range requiredFiles {
		filePath := filepath.Join(dirPath, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrMissingRequiredFile, filename)
		}
	}

	// Check for at least one calendar file
	hasCalendar := false
	for _, filename := range calendarFiles {
		filePath := filepath.Join(dirPath, filename)
		if _, err := os.Stat(filePath); err == nil {
			hasCalendar = true
			break
		}
	}
	if !hasCalendar {
		return nil, ErrMissingCalendarFile
	}

	feed := NewFeed()

	// Read each file using an opener function
	opener := func(filename string) (io.ReadCloser, error) {
		filePath := filepath.Join(dirPath, filename)
		return os.Open(filePath)
	}

	if err := readFeedFiles(feed, opener); err != nil {
		return nil, err
	}

	return feed, nil
}

// readFromZipPath reads a GTFS feed from a zip file path
func readFromZipPath(zipPath string) (*Feed, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open zip file %s: %w", zipPath, err)
	}
	defer func() { _ = r.Close() }()

	return readFromZipReader(&r.Reader)
}

// ReadFromZip reads a GTFS feed from a zip reader
func ReadFromZip(r io.ReaderAt, size int64) (*Feed, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("cannot read zip: %w", err)
	}
	return readFromZipReader(zr)
}

// readFromZipReader reads a GTFS feed from a zip.Reader
func readFromZipReader(zr *zip.Reader) (*Feed, error) {
	// Build a map of file names to zip file entries
	// Handle nested directories by stripping the prefix
	fileMap := make(map[string]*zip.File)
	var prefix string

	for _, f := range zr.File {
		name := f.Name
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}
		// Detect if files are in a nested directory
		if prefix == "" && strings.Contains(name, "/") {
			parts := strings.SplitN(name, "/", 2)
			if len(parts) == 2 {
				// Check if this looks like a GTFS file
				if isGTFSFile(parts[1]) {
					prefix = parts[0] + "/"
				}
			}
		}
	}

	// Build file map with prefix stripped
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := f.Name
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		fileMap[name] = f
	}

	// Check for required files
	for _, filename := range requiredFiles {
		if _, ok := fileMap[filename]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingRequiredFile, filename)
		}
	}

	// Check for at least one calendar file
	hasCalendar := false
	for _, filename := range calendarFiles {
		if _, ok := fileMap[filename]; ok {
			hasCalendar = true
			break
		}
	}
	if !hasCalendar {
		return nil, ErrMissingCalendarFile
	}

	feed := NewFeed()

	// Read each file using an opener function
	opener := func(filename string) (io.ReadCloser, error) {
		f, ok := fileMap[filename]
		if !ok {
			return nil, os.ErrNotExist
		}
		return f.Open()
	}

	if err := readFeedFiles(feed, opener); err != nil {
		return nil, err
	}

	return feed, nil
}

// isGTFSFile returns true if the filename is a known GTFS file
func isGTFSFile(name string) bool {
	gtfsFiles := []string{
		"agency.txt", "stops.txt", "routes.txt", "trips.txt",
		"stop_times.txt", "calendar.txt", "calendar_dates.txt",
		"fare_attributes.txt", "fare_rules.txt", "shapes.txt",
		"frequencies.txt", "transfers.txt", "feed_info.txt",
		"areas.txt", "pathways.txt",
	}
	for _, f := range gtfsFiles {
		if name == f {
			return true
		}
	}
	return false
}

// readFeedFiles reads all GTFS files using the provided opener function
func readFeedFiles(feed *Feed, opener func(string) (io.ReadCloser, error)) error {
	// Read agencies
	if err := readFileIntoFeed(feed, opener, "agency.txt", func(row *CSVRow) {
		agency := ParseAgency(row)
		feed.Agencies[agency.ID] = agency
		feed.AgencyOrder = append(feed.AgencyOrder, agency.ID)
	}); err != nil {
		return fmt.Errorf("reading agency.txt: %w", err)
	}

	// Read stops
	if err := readFileIntoFeed(feed, opener, "stops.txt", func(row *CSVRow) {
		stop := ParseStop(row)
		feed.Stops[stop.ID] = stop
		feed.StopOrder = append(feed.StopOrder, stop.ID)
	}); err != nil {
		return fmt.Errorf("reading stops.txt: %w", err)
	}

	// Read routes
	if err := readFileIntoFeed(feed, opener, "routes.txt", func(row *CSVRow) {
		route := ParseRoute(row)
		feed.Routes[route.ID] = route
		feed.RouteOrder = append(feed.RouteOrder, route.ID)
	}); err != nil {
		return fmt.Errorf("reading routes.txt: %w", err)
	}

	// Read trips
	if err := readFileIntoFeed(feed, opener, "trips.txt", func(row *CSVRow) {
		trip := ParseTrip(row)
		feed.Trips[trip.ID] = trip
		feed.TripOrder = append(feed.TripOrder, trip.ID)
	}); err != nil {
		return fmt.Errorf("reading trips.txt: %w", err)
	}

	// Read stop_times
	if err := readFileIntoFeed(feed, opener, "stop_times.txt", func(row *CSVRow) {
		stopTime := ParseStopTime(row)
		feed.StopTimes = append(feed.StopTimes, stopTime)
	}); err != nil {
		return fmt.Errorf("reading stop_times.txt: %w", err)
	}

	// Read calendar (optional - but at least one of calendar/calendar_dates required)
	if err := readOptionalFileIntoFeed(feed, opener, "calendar.txt", func(row *CSVRow) {
		calendar := ParseCalendar(row)
		feed.Calendars[calendar.ServiceID] = calendar
		feed.CalendarOrder = append(feed.CalendarOrder, calendar.ServiceID)
	}); err != nil {
		return fmt.Errorf("reading calendar.txt: %w", err)
	}

	// Read calendar_dates (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "calendar_dates.txt", func(row *CSVRow) {
		calDate := ParseCalendarDate(row)
		// Only track order for first occurrence of each service_id
		if _, exists := feed.CalendarDates[calDate.ServiceID]; !exists {
			feed.CalendarDateOrder = append(feed.CalendarDateOrder, calDate.ServiceID)
		}
		feed.CalendarDates[calDate.ServiceID] = append(feed.CalendarDates[calDate.ServiceID], calDate)
	}); err != nil {
		return fmt.Errorf("reading calendar_dates.txt: %w", err)
	}

	// Read shapes (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "shapes.txt", func(row *CSVRow) {
		shapePoint := ParseShapePoint(row)
		// Only track order for first occurrence of each shape_id
		if _, exists := feed.Shapes[shapePoint.ShapeID]; !exists {
			feed.ShapeOrder = append(feed.ShapeOrder, shapePoint.ShapeID)
		}
		feed.Shapes[shapePoint.ShapeID] = append(feed.Shapes[shapePoint.ShapeID], shapePoint)
	}); err != nil {
		return fmt.Errorf("reading shapes.txt: %w", err)
	}

	// Read frequencies (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "frequencies.txt", func(row *CSVRow) {
		frequency := ParseFrequency(row)
		feed.Frequencies = append(feed.Frequencies, frequency)
	}); err != nil {
		return fmt.Errorf("reading frequencies.txt: %w", err)
	}

	// Read transfers (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "transfers.txt", func(row *CSVRow) {
		transfer := ParseTransfer(row)
		feed.Transfers = append(feed.Transfers, transfer)
	}); err != nil {
		return fmt.Errorf("reading transfers.txt: %w", err)
	}

	// Read fare_attributes (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "fare_attributes.txt", func(row *CSVRow) {
		fareAttr := ParseFareAttribute(row)
		feed.FareAttributes[fareAttr.FareID] = fareAttr
		feed.FareAttrOrder = append(feed.FareAttrOrder, fareAttr.FareID)
	}); err != nil {
		return fmt.Errorf("reading fare_attributes.txt: %w", err)
	}

	// Read fare_rules (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "fare_rules.txt", func(row *CSVRow) {
		fareRule := ParseFareRule(row)
		feed.FareRules = append(feed.FareRules, fareRule)
	}); err != nil {
		return fmt.Errorf("reading fare_rules.txt: %w", err)
	}

	// Read feed_info (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "feed_info.txt", func(row *CSVRow) {
		fi := ParseFeedInfo(row)
		if fi.FeedID == "" {
			fi.FeedID = "1" // Java assigns 1 to feeds without feed_id
		}
		feed.FeedInfos[fi.FeedID] = fi // overwrites if same id
		feed.FeedInfoOrder = append(feed.FeedInfoOrder, fi.FeedID)
	}); err != nil {
		return fmt.Errorf("reading feed_info.txt: %w", err)
	}

	// Read areas (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "areas.txt", func(row *CSVRow) {
		area := ParseArea(row)
		feed.Areas[area.ID] = area
		feed.AreaOrder = append(feed.AreaOrder, area.ID)
	}); err != nil {
		return fmt.Errorf("reading areas.txt: %w", err)
	}

	// Read pathways (optional)
	if err := readOptionalFileIntoFeed(feed, opener, "pathways.txt", func(row *CSVRow) {
		pathway := ParsePathway(row)
		feed.Pathways = append(feed.Pathways, pathway)
	}); err != nil {
		return fmt.Errorf("reading pathways.txt: %w", err)
	}

	return nil
}

// readFileIntoFeed reads a required GTFS file and processes each row
func readFileIntoFeed(feed *Feed, opener func(string) (io.ReadCloser, error), filename string, process func(*CSVRow)) error {
	rc, err := opener(filename)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	reader := NewCSVReader(rc)
	header, err := reader.ReadHeader()
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	// Track which columns were present in this file
	feed.AddColumnSet(filename, header)

	for {
		record, err := reader.ReadRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading record: %w", err)
		}
		row := NewCSVRow(header, record)
		process(row)
	}

	return nil
}

// readOptionalFileIntoFeed reads an optional GTFS file if it exists
func readOptionalFileIntoFeed(feed *Feed, opener func(string) (io.ReadCloser, error), filename string, process func(*CSVRow)) error {
	rc, err := opener(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Optional file not present, that's OK
		}
		return err
	}
	defer func() { _ = rc.Close() }()

	reader := NewCSVReader(rc)
	header, err := reader.ReadHeader()
	if err != nil {
		if err == io.EOF {
			return nil // Empty file is OK for optional files
		}
		return fmt.Errorf("reading header: %w", err)
	}

	// Track which columns were present in this file
	feed.AddColumnSet(filename, header)

	for {
		record, err := reader.ReadRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading record: %w", err)
		}
		row := NewCSVRow(header, record)
		process(row)
	}

	return nil
}
