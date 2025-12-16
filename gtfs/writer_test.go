package gtfs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteToPath verifies that a feed can be written to a zip file
func TestWriteToPath(t *testing.T) {
	// Create a minimal feed
	feed := NewFeed()
	feed.Agencies[AgencyID("agency1")] = &Agency{
		ID:       AgencyID("agency1"),
		Name:     "Test Agency",
		URL:      "http://example.com",
		Timezone: "America/Los_Angeles",
	}
	feed.Stops[StopID("stop1")] = &Stop{
		ID:   StopID("stop1"),
		Name: "Test Stop",
		Lat:  47.123,
		Lon:  -122.456,
	}
	feed.Routes[RouteID("route1")] = &Route{
		ID:        RouteID("route1"),
		AgencyID:  AgencyID("agency1"),
		ShortName: "1",
		LongName:  "Route 1",
		Type:      3,
	}
	feed.Trips[TripID("trip1")] = &Trip{
		ID:        TripID("trip1"),
		RouteID:   RouteID("route1"),
		ServiceID: ServiceID("service1"),
	}
	feed.StopTimes = append(feed.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop1"),
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
	})
	feed.Calendars[ServiceID("service1")] = &Calendar{
		ServiceID: ServiceID("service1"),
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    false,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	// Write to a temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.zip")

	err := WriteToPath(feed, outputPath)
	if err != nil {
		t.Fatalf("WriteToPath failed: %v", err)
	}

	// Verify file exists
	_, err = os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	// Verify it's a valid zip file with expected files
	zr, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("cannot open output zip: %v", err)
	}
	defer func() { _ = zr.Close() }()

	expectedFiles := map[string]bool{
		"agency.txt":     false,
		"stops.txt":      false,
		"routes.txt":     false,
		"trips.txt":      false,
		"stop_times.txt": false,
		"calendar.txt":   false,
	}

	for _, f := range zr.File {
		if _, ok := expectedFiles[f.Name]; ok {
			expectedFiles[f.Name] = true
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("expected file %s not found in zip", name)
		}
	}
}

// TestWriteAndReadRoundTrip verifies that writing and reading produces the same data
func TestWriteAndReadRoundTrip(t *testing.T) {
	// Create a feed with various entities
	original := NewFeed()
	original.Agencies[AgencyID("agency1")] = &Agency{
		ID:       AgencyID("agency1"),
		Name:     "Test Agency",
		URL:      "http://example.com",
		Timezone: "America/Los_Angeles",
		Lang:     "en",
		Phone:    "555-1234",
	}
	original.Stops[StopID("stop1")] = &Stop{
		ID:   StopID("stop1"),
		Name: "Test Stop",
		Lat:  47.123456,
		Lon:  -122.456789,
	}
	original.Stops[StopID("stop2")] = &Stop{
		ID:   StopID("stop2"),
		Name: "Another Stop",
		Lat:  47.234567,
		Lon:  -122.567890,
	}
	original.Routes[RouteID("route1")] = &Route{
		ID:        RouteID("route1"),
		AgencyID:  AgencyID("agency1"),
		ShortName: "1",
		LongName:  "Route One",
		Type:      3,
		Color:     "FF0000",
		TextColor: "FFFFFF",
	}
	original.Trips[TripID("trip1")] = &Trip{
		ID:        TripID("trip1"),
		RouteID:   RouteID("route1"),
		ServiceID: ServiceID("service1"),
		Headsign:  "Downtown",
	}
	original.StopTimes = append(original.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop1"),
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
	})
	original.StopTimes = append(original.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop2"),
		StopSequence:  2,
		ArrivalTime:   "08:10:00",
		DepartureTime: "08:10:00",
	})
	original.Calendars[ServiceID("service1")] = &Calendar{
		ServiceID: ServiceID("service1"),
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    false,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	// Write to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "roundtrip.zip")

	err := WriteToPath(original, outputPath)
	if err != nil {
		t.Fatalf("WriteToPath failed: %v", err)
	}

	// Read it back
	roundTrip, err := ReadFromPath(outputPath)
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	// Verify agencies
	if len(roundTrip.Agencies) != len(original.Agencies) {
		t.Errorf("agency count mismatch: got %d, want %d", len(roundTrip.Agencies), len(original.Agencies))
	}
	if agency, ok := roundTrip.Agencies[AgencyID("agency1")]; !ok {
		t.Error("agency1 not found")
	} else {
		if agency.Name != "Test Agency" {
			t.Errorf("agency name mismatch: got %q, want %q", agency.Name, "Test Agency")
		}
	}

	// Verify stops
	if len(roundTrip.Stops) != len(original.Stops) {
		t.Errorf("stop count mismatch: got %d, want %d", len(roundTrip.Stops), len(original.Stops))
	}

	// Verify routes
	if len(roundTrip.Routes) != len(original.Routes) {
		t.Errorf("route count mismatch: got %d, want %d", len(roundTrip.Routes), len(original.Routes))
	}
	if route, ok := roundTrip.Routes[RouteID("route1")]; !ok {
		t.Error("route1 not found")
	} else {
		if route.ShortName != "1" {
			t.Errorf("route short name mismatch: got %q, want %q", route.ShortName, "1")
		}
	}

	// Verify trips
	if len(roundTrip.Trips) != len(original.Trips) {
		t.Errorf("trip count mismatch: got %d, want %d", len(roundTrip.Trips), len(original.Trips))
	}

	// Verify stop times
	if len(roundTrip.StopTimes) != len(original.StopTimes) {
		t.Errorf("stop_time count mismatch: got %d, want %d", len(roundTrip.StopTimes), len(original.StopTimes))
	}

	// Verify calendars
	if len(roundTrip.Calendars) != len(original.Calendars) {
		t.Errorf("calendar count mismatch: got %d, want %d", len(roundTrip.Calendars), len(original.Calendars))
	}
}

// TestWriteEmptyFeed verifies that a feed with only required files can be written
func TestWriteEmptyFeed(t *testing.T) {
	// Create a minimal feed (no optional files)
	feed := NewFeed()
	feed.Agencies[AgencyID("agency1")] = &Agency{
		ID:       AgencyID("agency1"),
		Name:     "Test Agency",
		URL:      "http://example.com",
		Timezone: "America/Los_Angeles",
	}
	feed.Stops[StopID("stop1")] = &Stop{
		ID:   StopID("stop1"),
		Name: "Test Stop",
		Lat:  47.0,
		Lon:  -122.0,
	}
	feed.Routes[RouteID("route1")] = &Route{
		ID:        RouteID("route1"),
		AgencyID:  AgencyID("agency1"),
		ShortName: "1",
		Type:      3,
	}
	feed.Trips[TripID("trip1")] = &Trip{
		ID:        TripID("trip1"),
		RouteID:   RouteID("route1"),
		ServiceID: ServiceID("service1"),
	}
	feed.StopTimes = append(feed.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop1"),
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
	})
	feed.Calendars[ServiceID("service1")] = &Calendar{
		ServiceID: ServiceID("service1"),
		Monday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	// Write to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "minimal.zip")

	err := WriteToPath(feed, outputPath)
	if err != nil {
		t.Fatalf("WriteToPath failed: %v", err)
	}

	// Read it back and verify
	readBack, err := ReadFromPath(outputPath)
	if err != nil {
		t.Fatalf("ReadFromPath failed: %v", err)
	}

	if len(readBack.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(readBack.Agencies))
	}
}

// TestWriteAllOptionalFiles verifies that optional files are written when present
func TestWriteAllOptionalFiles(t *testing.T) {
	feed := NewFeed()
	// Required entities
	feed.Agencies[AgencyID("agency1")] = &Agency{
		ID:       AgencyID("agency1"),
		Name:     "Test Agency",
		URL:      "http://example.com",
		Timezone: "America/Los_Angeles",
	}
	feed.Stops[StopID("stop1")] = &Stop{
		ID:   StopID("stop1"),
		Name: "Test Stop",
		Lat:  47.0,
		Lon:  -122.0,
	}
	feed.Stops[StopID("stop2")] = &Stop{
		ID:   StopID("stop2"),
		Name: "Stop 2",
		Lat:  47.1,
		Lon:  -122.1,
	}
	feed.Routes[RouteID("route1")] = &Route{
		ID:        RouteID("route1"),
		AgencyID:  AgencyID("agency1"),
		ShortName: "1",
		Type:      3,
	}
	feed.Trips[TripID("trip1")] = &Trip{
		ID:        TripID("trip1"),
		RouteID:   RouteID("route1"),
		ServiceID: ServiceID("service1"),
		ShapeID:   ShapeID("shape1"),
	}
	feed.StopTimes = append(feed.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop1"),
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
	})
	feed.Calendars[ServiceID("service1")] = &Calendar{
		ServiceID: ServiceID("service1"),
		Monday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	// Optional entities
	feed.CalendarDates[ServiceID("service1")] = append(
		feed.CalendarDates[ServiceID("service1")],
		&CalendarDate{ServiceID: ServiceID("service1"), Date: "20240704", ExceptionType: 2},
	)
	feed.Shapes[ShapeID("shape1")] = append(
		feed.Shapes[ShapeID("shape1")],
		&ShapePoint{ShapeID: ShapeID("shape1"), Lat: 47.0, Lon: -122.0, Sequence: 1},
	)
	feed.Frequencies = append(feed.Frequencies, &Frequency{
		TripID:      TripID("trip1"),
		StartTime:   "06:00:00",
		EndTime:     "22:00:00",
		HeadwaySecs: 600,
	})
	feed.Transfers = append(feed.Transfers, &Transfer{
		FromStopID:   StopID("stop1"),
		ToStopID:     StopID("stop2"),
		TransferType: 2,
	})
	feed.FareAttributes[FareID("fare1")] = &FareAttribute{
		FareID:       FareID("fare1"),
		Price:        2.50,
		CurrencyType: "USD",
	}
	feed.FareRules = append(feed.FareRules, &FareRule{
		FareID:  FareID("fare1"),
		RouteID: RouteID("route1"),
	})
	feed.FeedInfo = &FeedInfo{
		PublisherName: "Test Publisher",
		PublisherURL:  "http://example.com",
		Lang:          "en",
	}
	feed.Areas[AreaID("area1")] = &Area{
		ID:   AreaID("area1"),
		Name: "Downtown",
	}
	feed.Pathways = append(feed.Pathways, &Pathway{
		ID:              "pathway1",
		FromStopID:      StopID("stop1"),
		ToStopID:        StopID("stop2"),
		PathwayMode:     1,
		IsBidirectional: 1,
	})

	// Write to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "full.zip")

	err := WriteToPath(feed, outputPath)
	if err != nil {
		t.Fatalf("WriteToPath failed: %v", err)
	}

	// Verify all expected files are in the zip
	zr, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("cannot open output zip: %v", err)
	}
	defer func() { _ = zr.Close() }()

	expectedFiles := []string{
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",
		"calendar.txt",
		"calendar_dates.txt",
		"shapes.txt",
		"frequencies.txt",
		"transfers.txt",
		"fare_attributes.txt",
		"fare_rules.txt",
		"feed_info.txt",
		"areas.txt",
		"pathways.txt",
	}

	fileSet := make(map[string]bool)
	for _, f := range zr.File {
		fileSet[f.Name] = true
	}

	for _, expected := range expectedFiles {
		if !fileSet[expected] {
			t.Errorf("expected file %s not found in zip", expected)
		}
	}
}

// TestWriteSkipsEmptyOptionalFiles verifies that empty optional files are not written
func TestWriteSkipsEmptyOptionalFiles(t *testing.T) {
	// Create a minimal feed (no optional entities)
	feed := NewFeed()
	feed.Agencies[AgencyID("agency1")] = &Agency{
		ID:       AgencyID("agency1"),
		Name:     "Test Agency",
		URL:      "http://example.com",
		Timezone: "America/Los_Angeles",
	}
	feed.Stops[StopID("stop1")] = &Stop{
		ID:   StopID("stop1"),
		Name: "Test Stop",
		Lat:  47.0,
		Lon:  -122.0,
	}
	feed.Routes[RouteID("route1")] = &Route{
		ID:        RouteID("route1"),
		AgencyID:  AgencyID("agency1"),
		ShortName: "1",
		Type:      3,
	}
	feed.Trips[TripID("trip1")] = &Trip{
		ID:        TripID("trip1"),
		RouteID:   RouteID("route1"),
		ServiceID: ServiceID("service1"),
	}
	feed.StopTimes = append(feed.StopTimes, &StopTime{
		TripID:        TripID("trip1"),
		StopID:        StopID("stop1"),
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
	})
	feed.Calendars[ServiceID("service1")] = &Calendar{
		ServiceID: ServiceID("service1"),
		Monday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	// Write to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "minimal_optional.zip")

	err := WriteToPath(feed, outputPath)
	if err != nil {
		t.Fatalf("WriteToPath failed: %v", err)
	}

	// Verify optional files are NOT in the zip
	zr, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("cannot open output zip: %v", err)
	}
	defer func() { _ = zr.Close() }()

	// These optional files should not exist
	optionalFiles := []string{
		"calendar_dates.txt",
		"shapes.txt",
		"frequencies.txt",
		"transfers.txt",
		"fare_attributes.txt",
		"fare_rules.txt",
		"feed_info.txt",
		"areas.txt",
		"pathways.txt",
	}

	fileSet := make(map[string]bool)
	for _, f := range zr.File {
		fileSet[f.Name] = true
	}

	for _, optional := range optionalFiles {
		if fileSet[optional] {
			t.Errorf("optional file %s should not be in zip when empty", optional)
		}
	}

	// But required files should exist
	requiredFiles := []string{
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",
		"calendar.txt",
	}

	for _, required := range requiredFiles {
		if !fileSet[required] {
			t.Errorf("required file %s should be in zip", required)
		}
	}
}
