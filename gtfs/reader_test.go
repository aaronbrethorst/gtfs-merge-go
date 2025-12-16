package gtfs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFromDirectory(t *testing.T) {
	// Read the minimal test feed from directory
	feed, err := ReadFromPath("../testdata/minimal")
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	// Verify agency was read
	if len(feed.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed.Agencies))
	}
	agency, ok := feed.Agencies["agency1"]
	if !ok {
		t.Error("expected agency with ID 'agency1'")
	} else {
		if agency.Name != "Minimal Transit" {
			t.Errorf("expected agency name 'Minimal Transit', got '%s'", agency.Name)
		}
		if agency.Timezone != "America/Los_Angeles" {
			t.Errorf("expected timezone 'America/Los_Angeles', got '%s'", agency.Timezone)
		}
	}

	// Verify stop was read
	if len(feed.Stops) != 1 {
		t.Errorf("expected 1 stop, got %d", len(feed.Stops))
	}
	stop, ok := feed.Stops["stop1"]
	if !ok {
		t.Error("expected stop with ID 'stop1'")
	} else {
		if stop.Name != "Main Street Station" {
			t.Errorf("expected stop name 'Main Street Station', got '%s'", stop.Name)
		}
		if stop.Lat != 37.7749 {
			t.Errorf("expected lat 37.7749, got %f", stop.Lat)
		}
	}

	// Verify route was read
	if len(feed.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(feed.Routes))
	}
	route, ok := feed.Routes["route1"]
	if !ok {
		t.Error("expected route with ID 'route1'")
	} else {
		if route.ShortName != "1" {
			t.Errorf("expected route short name '1', got '%s'", route.ShortName)
		}
		if route.AgencyID != "agency1" {
			t.Errorf("expected agency_id 'agency1', got '%s'", route.AgencyID)
		}
	}

	// Verify trip was read
	if len(feed.Trips) != 1 {
		t.Errorf("expected 1 trip, got %d", len(feed.Trips))
	}
	trip, ok := feed.Trips["trip1"]
	if !ok {
		t.Error("expected trip with ID 'trip1'")
	} else {
		if trip.RouteID != "route1" {
			t.Errorf("expected route_id 'route1', got '%s'", trip.RouteID)
		}
		if trip.ServiceID != "service1" {
			t.Errorf("expected service_id 'service1', got '%s'", trip.ServiceID)
		}
	}

	// Verify stop_times was read
	if len(feed.StopTimes) != 1 {
		t.Errorf("expected 1 stop_time, got %d", len(feed.StopTimes))
	}
	if len(feed.StopTimes) > 0 {
		st := feed.StopTimes[0]
		if st.TripID != "trip1" {
			t.Errorf("expected trip_id 'trip1', got '%s'", st.TripID)
		}
		if st.ArrivalTime != "08:00:00" {
			t.Errorf("expected arrival_time '08:00:00', got '%s'", st.ArrivalTime)
		}
	}

	// Verify calendar was read
	if len(feed.Calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(feed.Calendars))
	}
	cal, ok := feed.Calendars["service1"]
	if !ok {
		t.Error("expected calendar with service_id 'service1'")
	} else {
		if !cal.Monday || !cal.Tuesday || !cal.Wednesday || !cal.Thursday || !cal.Friday {
			t.Error("expected weekdays to be true")
		}
		if cal.Saturday || cal.Sunday {
			t.Error("expected weekend to be false")
		}
	}
}

func TestReadFromDirectorySimpleA(t *testing.T) {
	// Read the simple_a test feed
	feed, err := ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	// Verify counts
	if len(feed.Agencies) != 2 {
		t.Errorf("expected 2 agencies, got %d", len(feed.Agencies))
	}
	if len(feed.Stops) != 5 {
		t.Errorf("expected 5 stops, got %d", len(feed.Stops))
	}
	if len(feed.Routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(feed.Routes))
	}
	if len(feed.Trips) != 4 {
		t.Errorf("expected 4 trips, got %d", len(feed.Trips))
	}
	if len(feed.StopTimes) != 10 {
		t.Errorf("expected 10 stop_times, got %d", len(feed.StopTimes))
	}
	if len(feed.Calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(feed.Calendars))
	}
}

func TestReadFromDirectoryMissingRequired(t *testing.T) {
	// Create a temp directory with missing required file
	tmpDir, err := os.MkdirTemp("", "gtfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create only agency.txt, missing other required files
	agencyPath := filepath.Join(tmpDir, "agency.txt")
	err = os.WriteFile(agencyPath, []byte("agency_id,agency_name,agency_url,agency_timezone\nagency1,Test,http://test.com,UTC\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write agency.txt: %v", err)
	}

	// Try to read - should fail due to missing required files
	_, err = ReadFromPath(tmpDir)
	if err == nil {
		t.Error("expected error for missing required files, got nil")
	}
}

func TestReadFromDirectoryOptionalFiles(t *testing.T) {
	// The minimal feed has no optional files - verify it still loads
	feed, err := ReadFromPath("../testdata/minimal")
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	// Optional fields should be empty/nil but not cause errors
	if len(feed.Shapes) != 0 {
		t.Errorf("expected 0 shapes, got %d", len(feed.Shapes))
	}
	if len(feed.Frequencies) != 0 {
		t.Errorf("expected 0 frequencies, got %d", len(feed.Frequencies))
	}
	if len(feed.Transfers) != 0 {
		t.Errorf("expected 0 transfers, got %d", len(feed.Transfers))
	}
	if len(feed.FareAttributes) != 0 {
		t.Errorf("expected 0 fare_attributes, got %d", len(feed.FareAttributes))
	}
	if len(feed.FareRules) != 0 {
		t.Errorf("expected 0 fare_rules, got %d", len(feed.FareRules))
	}
	if feed.FeedInfo != nil {
		t.Error("expected nil feed_info")
	}
	if len(feed.Areas) != 0 {
		t.Errorf("expected 0 areas, got %d", len(feed.Areas))
	}
	if len(feed.Pathways) != 0 {
		t.Errorf("expected 0 pathways, got %d", len(feed.Pathways))
	}
	if len(feed.CalendarDates) != 0 {
		t.Errorf("expected 0 calendar_dates entries, got %d", len(feed.CalendarDates))
	}
}

func TestReadFromDirectoryInvalidCSV(t *testing.T) {
	// Create a temp directory with invalid CSV
	tmpDir, err := os.MkdirTemp("", "gtfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create all required files but make one have invalid format
	// Use a quote that starts but never ends, spanning what looks like multiple lines
	files := map[string]string{
		"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\nagency1,Test,http://test.com,UTC\n",
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nstop1,Stop,0.0,0.0\n",
		"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\nroute1,agency1,1,Test,3\n",
		"trips.txt":      "route_id,service_id,trip_id\nroute1,service1,trip1\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n\"never ending quote\n",
		"calendar.txt":   "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nservice1,1,1,1,1,1,0,0,20240101,20241231\n",
	}

	for name, content := range files {
		err = os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// Try to read - should fail due to malformed CSV (unterminated quote at EOF)
	_, err = ReadFromPath(tmpDir)
	if err == nil {
		t.Error("expected error for invalid CSV, got nil")
	}
}

func TestReadFromDirectoryNotFound(t *testing.T) {
	_, err := ReadFromPath("/nonexistent/path/to/feed")
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}
}

func TestReadFromZipPath(t *testing.T) {
	// Create a temporary zip file from the minimal test data
	tmpFile, err := os.CreateTemp("", "gtfs-test-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	// Create the zip file
	if err := createTestZip(t, "../testdata/minimal", tmpPath); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Read from zip path
	feed, err := ReadFromPath(tmpPath)
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	// Verify data was read correctly
	if len(feed.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed.Agencies))
	}
	if len(feed.Stops) != 1 {
		t.Errorf("expected 1 stop, got %d", len(feed.Stops))
	}
	if len(feed.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(feed.Routes))
	}
	if len(feed.Trips) != 1 {
		t.Errorf("expected 1 trip, got %d", len(feed.Trips))
	}
	if len(feed.StopTimes) != 1 {
		t.Errorf("expected 1 stop_time, got %d", len(feed.StopTimes))
	}
	if len(feed.Calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(feed.Calendars))
	}
}

func TestReadFromZipReader(t *testing.T) {
	// Create a temporary zip file
	tmpFile, err := os.CreateTemp("", "gtfs-test-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	// Create the zip file
	if err := createTestZip(t, "../testdata/simple_a", tmpPath); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Open the file and read using ReadFromZip
	f, err := os.Open(tmpPath)
	if err != nil {
		t.Fatalf("failed to open zip file: %v", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	feed, err := ReadFromZip(f, info.Size())
	if err != nil {
		t.Fatalf("ReadFromZip failed: %v", err)
	}

	// Verify simple_a data
	if len(feed.Agencies) != 2 {
		t.Errorf("expected 2 agencies, got %d", len(feed.Agencies))
	}
	if len(feed.Stops) != 5 {
		t.Errorf("expected 5 stops, got %d", len(feed.Stops))
	}
	if len(feed.Routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(feed.Routes))
	}
	if len(feed.Trips) != 4 {
		t.Errorf("expected 4 trips, got %d", len(feed.Trips))
	}
}

func TestReadFromPathAutoDetect(t *testing.T) {
	// Test that ReadFromPath auto-detects directory
	feed, err := ReadFromPath("../testdata/minimal")
	if err != nil {
		t.Fatalf("ReadFromPath (directory) failed: %v", err)
	}
	if len(feed.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed.Agencies))
	}

	// Create a zip file
	tmpFile, err := os.CreateTemp("", "gtfs-test-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := createTestZip(t, "../testdata/minimal", tmpPath); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	// Test that ReadFromPath auto-detects zip
	feed2, err := ReadFromPath(tmpPath)
	if err != nil {
		t.Fatalf("ReadFromPath (zip) failed: %v", err)
	}
	if len(feed2.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed2.Agencies))
	}
}

func TestReadFromZipNestedDirectory(t *testing.T) {
	// Create a zip file with files inside a nested directory
	tmpFile, err := os.CreateTemp("", "gtfs-test-nested-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := createNestedTestZip(t, "../testdata/minimal", tmpPath, "gtfs_data"); err != nil {
		t.Fatalf("failed to create nested test zip: %v", err)
	}

	// Read from nested zip
	feed, err := ReadFromPath(tmpPath)
	if err != nil {
		t.Fatalf("ReadFromPath (nested zip) failed: %v", err)
	}

	// Verify data was read correctly
	if len(feed.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed.Agencies))
	}
	if len(feed.Stops) != 1 {
		t.Errorf("expected 1 stop, got %d", len(feed.Stops))
	}
}

func TestReadFromZipMissingRequired(t *testing.T) {
	// Create a zip without all required files
	tmpFile, err := os.CreateTemp("", "gtfs-test-incomplete-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	// Create zip with only agency.txt
	if err := createPartialZip(t, tmpPath); err != nil {
		t.Fatalf("failed to create partial zip: %v", err)
	}

	_, err = ReadFromPath(tmpPath)
	if err == nil {
		t.Error("expected error for missing required files, got nil")
	}
}

// Helper function to create a test zip file from a directory
func createTestZip(t *testing.T, srcDir, destPath string) error {
	t.Helper()

	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	files := []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt", "calendar.txt"}

	for _, filename := range files {
		srcPath := filepath.Join(srcDir, filename)
		content, err := os.ReadFile(srcPath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		w, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}
		if _, err := w.Write(content); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to create a test zip with files in a nested directory
func createNestedTestZip(t *testing.T, srcDir, destPath, nestedDir string) error {
	t.Helper()

	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	files := []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt", "calendar.txt"}

	for _, filename := range files {
		srcPath := filepath.Join(srcDir, filename)
		content, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}

		// Create file inside nested directory
		w, err := zipWriter.Create(nestedDir + "/" + filename)
		if err != nil {
			return err
		}
		if _, err := w.Write(content); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to create a zip with only some files
func createPartialZip(t *testing.T, destPath string) error {
	t.Helper()

	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// Only write agency.txt
	w, err := zipWriter.Create("agency.txt")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("agency_id,agency_name,agency_url,agency_timezone\nagency1,Test,http://test.com,UTC\n"))
	return err
}
