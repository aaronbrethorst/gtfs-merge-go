package merge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

func TestMergeTwoSimpleFeeds(t *testing.T) {
	// Given: two simple feeds with no overlapping IDs
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}
	feedB, err := gtfs.ReadFromPath("../testdata/simple_b")
	if err != nil {
		t.Fatalf("failed to read simple_b: %v", err)
	}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output contains all entities from both feeds
	totalAgencies := len(feedA.Agencies) + len(feedB.Agencies)
	if len(merged.Agencies) != totalAgencies {
		t.Errorf("expected %d agencies, got %d", totalAgencies, len(merged.Agencies))
	}

	totalStops := len(feedA.Stops) + len(feedB.Stops)
	if len(merged.Stops) != totalStops {
		t.Errorf("expected %d stops, got %d", totalStops, len(merged.Stops))
	}

	totalRoutes := len(feedA.Routes) + len(feedB.Routes)
	if len(merged.Routes) != totalRoutes {
		t.Errorf("expected %d routes, got %d", totalRoutes, len(merged.Routes))
	}

	totalTrips := len(feedA.Trips) + len(feedB.Trips)
	if len(merged.Trips) != totalTrips {
		t.Errorf("expected %d trips, got %d", totalTrips, len(merged.Trips))
	}
}

func TestMergeTwoFeedsAgencies(t *testing.T) {
	// Given: feed A has agencies [A1, A2], feed B has agency [B1]
	feedA := gtfs.NewFeed()
	feedA.Agencies["A1"] = &gtfs.Agency{ID: "A1", Name: "Agency A1", URL: "http://a1.com", Timezone: "America/Los_Angeles"}
	feedA.Agencies["A2"] = &gtfs.Agency{ID: "A2", Name: "Agency A2", URL: "http://a2.com", Timezone: "America/Los_Angeles"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["B1"] = &gtfs.Agency{ID: "B1", Name: "Agency B1", URL: "http://b1.com", Timezone: "America/New_York"}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has agencies [A1, A2, B1]
	if len(merged.Agencies) != 3 {
		t.Errorf("expected 3 agencies, got %d", len(merged.Agencies))
	}

	// Verify all agencies are present (may have prefixed IDs if there were collisions)
	agencyNames := make(map[string]bool)
	for _, a := range merged.Agencies {
		agencyNames[a.Name] = true
	}
	for _, name := range []string{"Agency A1", "Agency A2", "Agency B1"} {
		if !agencyNames[name] {
			t.Errorf("expected agency with name %q in merged feed", name)
		}
	}
}

func TestMergeTwoFeedsStops(t *testing.T) {
	// Given: feed A has stops [S1, S2], feed B has stops [S3, S4]
	feedA := gtfs.NewFeed()
	feedA.Stops["S1"] = &gtfs.Stop{ID: "S1", Name: "Stop 1", Lat: 47.6, Lon: -122.3}
	feedA.Stops["S2"] = &gtfs.Stop{ID: "S2", Name: "Stop 2", Lat: 47.7, Lon: -122.4}

	feedB := gtfs.NewFeed()
	feedB.Stops["S3"] = &gtfs.Stop{ID: "S3", Name: "Stop 3", Lat: 40.7, Lon: -74.0}
	feedB.Stops["S4"] = &gtfs.Stop{ID: "S4", Name: "Stop 4", Lat: 40.8, Lon: -74.1}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has stops [S1, S2, S3, S4]
	if len(merged.Stops) != 4 {
		t.Errorf("expected 4 stops, got %d", len(merged.Stops))
	}

	// Verify all stops are present by name
	stopNames := make(map[string]bool)
	for _, s := range merged.Stops {
		stopNames[s.Name] = true
	}
	for _, name := range []string{"Stop 1", "Stop 2", "Stop 3", "Stop 4"} {
		if !stopNames[name] {
			t.Errorf("expected stop with name %q in merged feed", name)
		}
	}
}

func TestMergeTwoFeedsRoutes(t *testing.T) {
	// Given: feed A has routes [R1], feed B has routes [R2]
	feedA := gtfs.NewFeed()
	feedA.Agencies["A1"] = &gtfs.Agency{ID: "A1", Name: "Agency A1", URL: "http://a1.com", Timezone: "UTC"}
	feedA.Routes["R1"] = &gtfs.Route{ID: "R1", AgencyID: "A1", ShortName: "Route 1", Type: 3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["B1"] = &gtfs.Agency{ID: "B1", Name: "Agency B1", URL: "http://b1.com", Timezone: "UTC"}
	feedB.Routes["R2"] = &gtfs.Route{ID: "R2", AgencyID: "B1", ShortName: "Route 2", Type: 3}

	// When: merged with agency references updated
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has routes [R1, R2] with correct agency refs
	if len(merged.Routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(merged.Routes))
	}

	// Verify routes exist and reference valid agencies
	for _, route := range merged.Routes {
		if _, ok := merged.Agencies[route.AgencyID]; !ok {
			t.Errorf("route %s references non-existent agency %s", route.ID, route.AgencyID)
		}
	}
}

func TestMergeTwoFeedsTrips(t *testing.T) {
	// Given: feed A has trips [T1, T2], feed B has trips [T3]
	feedA := gtfs.NewFeed()
	feedA.Agencies["A1"] = &gtfs.Agency{ID: "A1", Name: "Agency A1", URL: "http://a1.com", Timezone: "UTC"}
	feedA.Routes["R1"] = &gtfs.Route{ID: "R1", AgencyID: "A1", ShortName: "R1", Type: 3}
	feedA.Calendars["SVC1"] = &gtfs.Calendar{ServiceID: "SVC1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedA.Trips["T1"] = &gtfs.Trip{ID: "T1", RouteID: "R1", ServiceID: "SVC1"}
	feedA.Trips["T2"] = &gtfs.Trip{ID: "T2", RouteID: "R1", ServiceID: "SVC1"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["B1"] = &gtfs.Agency{ID: "B1", Name: "Agency B1", URL: "http://b1.com", Timezone: "UTC"}
	feedB.Routes["R2"] = &gtfs.Route{ID: "R2", AgencyID: "B1", ShortName: "R2", Type: 3}
	feedB.Calendars["SVC2"] = &gtfs.Calendar{ServiceID: "SVC2", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedB.Trips["T3"] = &gtfs.Trip{ID: "T3", RouteID: "R2", ServiceID: "SVC2"}

	// When: merged with route/service references updated
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has trips [T1, T2, T3] with correct refs
	if len(merged.Trips) != 3 {
		t.Errorf("expected 3 trips, got %d", len(merged.Trips))
	}

	// Verify trips reference valid routes and services
	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references non-existent route %s", trip.ID, trip.RouteID)
		}
		if _, ok := merged.Calendars[trip.ServiceID]; !ok {
			t.Errorf("trip %s references non-existent service %s", trip.ID, trip.ServiceID)
		}
	}
}

func TestMergeTwoFeedsStopTimes(t *testing.T) {
	// Given: each feed has stop_times for their trips
	feedA := gtfs.NewFeed()
	feedA.Agencies["A1"] = &gtfs.Agency{ID: "A1", Name: "Agency A1", URL: "http://a1.com", Timezone: "UTC"}
	feedA.Stops["S1"] = &gtfs.Stop{ID: "S1", Name: "Stop 1", Lat: 47.6, Lon: -122.3}
	feedA.Stops["S2"] = &gtfs.Stop{ID: "S2", Name: "Stop 2", Lat: 47.7, Lon: -122.4}
	feedA.Routes["R1"] = &gtfs.Route{ID: "R1", AgencyID: "A1", ShortName: "R1", Type: 3}
	feedA.Calendars["SVC1"] = &gtfs.Calendar{ServiceID: "SVC1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedA.Trips["T1"] = &gtfs.Trip{ID: "T1", RouteID: "R1", ServiceID: "SVC1"}
	feedA.StopTimes = append(feedA.StopTimes,
		&gtfs.StopTime{TripID: "T1", StopID: "S1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "T1", StopID: "S2", StopSequence: 2, ArrivalTime: "08:10:00", DepartureTime: "08:10:00"},
	)

	feedB := gtfs.NewFeed()
	feedB.Agencies["B1"] = &gtfs.Agency{ID: "B1", Name: "Agency B1", URL: "http://b1.com", Timezone: "UTC"}
	feedB.Stops["S3"] = &gtfs.Stop{ID: "S3", Name: "Stop 3", Lat: 40.7, Lon: -74.0}
	feedB.Routes["R2"] = &gtfs.Route{ID: "R2", AgencyID: "B1", ShortName: "R2", Type: 3}
	feedB.Calendars["SVC2"] = &gtfs.Calendar{ServiceID: "SVC2", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedB.Trips["T2"] = &gtfs.Trip{ID: "T2", RouteID: "R2", ServiceID: "SVC2"}
	feedB.StopTimes = append(feedB.StopTimes,
		&gtfs.StopTime{TripID: "T2", StopID: "S3", StopSequence: 1, ArrivalTime: "09:00:00", DepartureTime: "09:00:00"},
	)

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has all stop_times with correct trip/stop refs
	if len(merged.StopTimes) != 3 {
		t.Errorf("expected 3 stop_times, got %d", len(merged.StopTimes))
	}

	// Verify stop times reference valid trips and stops
	for _, st := range merged.StopTimes {
		if _, ok := merged.Trips[st.TripID]; !ok {
			t.Errorf("stop_time references non-existent trip %s", st.TripID)
		}
		if _, ok := merged.Stops[st.StopID]; !ok {
			t.Errorf("stop_time references non-existent stop %s", st.StopID)
		}
	}
}

func TestMergeTwoFeedsCalendars(t *testing.T) {
	// Given: feed A has service [SVC1], feed B has service [SVC2]
	feedA := gtfs.NewFeed()
	feedA.Calendars["SVC1"] = &gtfs.Calendar{ServiceID: "SVC1", Monday: true, Tuesday: true, StartDate: "20240101", EndDate: "20241231"}

	feedB := gtfs.NewFeed()
	feedB.Calendars["SVC2"] = &gtfs.Calendar{ServiceID: "SVC2", Wednesday: true, Thursday: true, StartDate: "20240101", EndDate: "20241231"}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output has services [SVC1, SVC2]
	if len(merged.Calendars) != 2 {
		t.Errorf("expected 2 calendars, got %d", len(merged.Calendars))
	}
}

func TestMergeTwoFeedsPreservesReferentialIntegrity(t *testing.T) {
	// Given: two complete feeds
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}
	feedB, err := gtfs.ReadFromPath("../testdata/simple_b")
	if err != nil {
		t.Fatalf("failed to read simple_b: %v", err)
	}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: all foreign key references are valid in output

	// Routes reference valid agencies
	for _, route := range merged.Routes {
		if route.AgencyID != "" {
			if _, ok := merged.Agencies[route.AgencyID]; !ok {
				t.Errorf("route %s references non-existent agency %s", route.ID, route.AgencyID)
			}
		}
	}

	// Trips reference valid routes and services
	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references non-existent route %s", trip.ID, trip.RouteID)
		}
		if _, ok := merged.Calendars[trip.ServiceID]; !ok {
			// May be defined in calendar_dates only
			if _, ok := merged.CalendarDates[trip.ServiceID]; !ok {
				t.Errorf("trip %s references non-existent service %s", trip.ID, trip.ServiceID)
			}
		}
	}

	// Stop times reference valid trips and stops
	for _, st := range merged.StopTimes {
		if _, ok := merged.Trips[st.TripID]; !ok {
			t.Errorf("stop_time references non-existent trip %s", st.TripID)
		}
		if _, ok := merged.Stops[st.StopID]; !ok {
			t.Errorf("stop_time references non-existent stop %s", st.StopID)
		}
	}

	// Stops with parent_station reference valid stops
	for _, stop := range merged.Stops {
		if stop.ParentStation != "" {
			if _, ok := merged.Stops[stop.ParentStation]; !ok {
				t.Errorf("stop %s references non-existent parent_station %s", stop.ID, stop.ParentStation)
			}
		}
	}
}

func TestMergeProducesValidGTFS(t *testing.T) {
	// Given: two valid feeds
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}
	feedB, err := gtfs.ReadFromPath("../testdata/simple_b")
	if err != nil {
		t.Fatalf("failed to read simple_b: %v", err)
	}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: output can be written and read back
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "merged.zip")

	err = gtfs.WriteToPath(merged, outPath)
	if err != nil {
		t.Fatalf("failed to write merged feed: %v", err)
	}

	// Read back and verify
	readBack, err := gtfs.ReadFromPath(outPath)
	if err != nil {
		t.Fatalf("failed to read back merged feed: %v", err)
	}

	// Verify counts match
	if len(readBack.Agencies) != len(merged.Agencies) {
		t.Errorf("agency count mismatch after round-trip: got %d, want %d", len(readBack.Agencies), len(merged.Agencies))
	}
	if len(readBack.Stops) != len(merged.Stops) {
		t.Errorf("stop count mismatch after round-trip: got %d, want %d", len(readBack.Stops), len(merged.Stops))
	}
	if len(readBack.Routes) != len(merged.Routes) {
		t.Errorf("route count mismatch after round-trip: got %d, want %d", len(readBack.Routes), len(merged.Routes))
	}
	if len(readBack.Trips) != len(merged.Trips) {
		t.Errorf("trip count mismatch after round-trip: got %d, want %d", len(readBack.Trips), len(merged.Trips))
	}
}

func TestMergeFilesEndToEnd(t *testing.T) {
	// Given: two zip files (we'll create them from directories)
	tmpDir := t.TempDir()

	// Create zip files from test directories
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}
	feedAPath := filepath.Join(tmpDir, "feed_a.zip")
	if err := gtfs.WriteToPath(feedA, feedAPath); err != nil {
		t.Fatalf("failed to create feed_a.zip: %v", err)
	}

	feedB, err := gtfs.ReadFromPath("../testdata/simple_b")
	if err != nil {
		t.Fatalf("failed to read simple_b: %v", err)
	}
	feedBPath := filepath.Join(tmpDir, "feed_b.zip")
	if err := gtfs.WriteToPath(feedB, feedBPath); err != nil {
		t.Fatalf("failed to create feed_b.zip: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "merged.zip")

	// When: MergeFiles() called
	merger := New()
	err = merger.MergeFiles([]string{feedAPath, feedBPath}, outputPath)
	if err != nil {
		t.Fatalf("MergeFiles failed: %v", err)
	}

	// Then: output zip is valid and contains merged data
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	merged, err := gtfs.ReadFromPath(outputPath)
	if err != nil {
		t.Fatalf("failed to read merged output: %v", err)
	}

	// Verify merged feed has data from both inputs
	expectedAgencies := len(feedA.Agencies) + len(feedB.Agencies)
	if len(merged.Agencies) != expectedAgencies {
		t.Errorf("expected %d agencies in merged output, got %d", expectedAgencies, len(merged.Agencies))
	}
}

func TestMergeNoInputFeeds(t *testing.T) {
	merger := New()
	_, err := merger.MergeFeeds([]*gtfs.Feed{})
	if err == nil {
		t.Error("expected error when merging empty feed list")
	}
}

func TestMergeSingleFeed(t *testing.T) {
	// Single feed should just return a copy
	feed := gtfs.NewFeed()
	feed.Agencies["A1"] = &gtfs.Agency{ID: "A1", Name: "Test", URL: "http://test.com", Timezone: "UTC"}

	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feed})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(merged.Agencies))
	}
}

// Tests for 5.3 - ID Prefixing

func TestMergeAppliesPrefixToSecondFeed(t *testing.T) {
	// Given: feeds with same IDs
	feedA := gtfs.NewFeed()
	feedA.Agencies["shared_id"] = &gtfs.Agency{ID: "shared_id", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop A", Lat: 47.6, Lon: -122.3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared_id"] = &gtfs.Agency{ID: "shared_id", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop B", Lat: 40.7, Lon: -74.0}

	// When: merged (with DetectionNone by default - no duplicate detection)
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: entities from first-processed feed have no prefix
	// Java processes in REVERSE order: feedB is processed first (no prefix)
	// and feedA is processed second (gets "b-" prefix)
	if len(merged.Agencies) != 2 {
		t.Errorf("expected 2 agencies (both preserved with different IDs), got %d", len(merged.Agencies))
	}

	// feedB was processed first (last in array) -> no prefix
	if _, ok := merged.Agencies["shared_id"]; !ok {
		t.Error("expected agency with original ID 'shared_id' from feedB (processed first in reverse order)")
	}

	// feedA was processed second (first in array) -> "b-" prefix
	if _, ok := merged.Agencies["b-shared_id"]; !ok {
		t.Error("expected agency with prefixed ID 'b-shared_id' from feedA (processed second)")
	}

	// Same for stops
	if len(merged.Stops) != 2 {
		t.Errorf("expected 2 stops, got %d", len(merged.Stops))
	}
}

func TestMergePrefixUpdatesAllReferences(t *testing.T) {
	// Given: feed B entity IDs will be prefixed
	feedA := gtfs.NewFeed()
	feedA.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop A1", Lat: 47.6, Lon: -122.3}
	feedA.Stops["stop2"] = &gtfs.Stop{ID: "stop2", Name: "Stop A2", Lat: 47.7, Lon: -122.4}
	feedA.Routes["route1"] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feedA.Calendars["svc1"] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedA.Trips["trip1"] = &gtfs.Trip{ID: "trip1", RouteID: "route1", ServiceID: "svc1"}
	feedA.StopTimes = append(feedA.StopTimes,
		&gtfs.StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "trip1", StopID: "stop2", StopSequence: 2, ArrivalTime: "08:10:00", DepartureTime: "08:10:00"},
	)

	// feedB has same structure but different data
	feedB := gtfs.NewFeed()
	feedB.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop B1", Lat: 40.7, Lon: -74.0}
	feedB.Routes["route1"] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1B", Type: 3}
	feedB.Calendars["svc1"] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedB.Trips["trip1"] = &gtfs.Trip{ID: "trip1", RouteID: "route1", ServiceID: "svc1"}
	feedB.StopTimes = append(feedB.StopTimes,
		&gtfs.StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "09:00:00", DepartureTime: "09:00:00"},
	)

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: all references to prefixed IDs are also prefixed
	// Check routes reference the correct (possibly prefixed) agency
	for _, route := range merged.Routes {
		if _, ok := merged.Agencies[route.AgencyID]; !ok {
			t.Errorf("route %s references invalid agency %s", route.ID, route.AgencyID)
		}
	}

	// Check trips reference the correct routes and services
	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references invalid route %s", trip.ID, trip.RouteID)
		}
		if _, ok := merged.Calendars[trip.ServiceID]; !ok {
			t.Errorf("trip %s references invalid service %s", trip.ID, trip.ServiceID)
		}
	}

	// Check stop times reference the correct trips and stops
	for _, st := range merged.StopTimes {
		if _, ok := merged.Trips[st.TripID]; !ok {
			t.Errorf("stop_time references invalid trip %s", st.TripID)
		}
		if _, ok := merged.Stops[st.StopID]; !ok {
			t.Errorf("stop_time references invalid stop %s", st.StopID)
		}
	}
}

func TestMergePrefixSequence(t *testing.T) {
	// Given: three feeds with overlapping IDs
	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

	feedC := gtfs.NewFeed()
	feedC.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency C", URL: "http://c.com", Timezone: "UTC"}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB, feedC})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: prefixes are applied correctly (Java behavior: reverse processing order)
	// Feeds processed in REVERSE order: C first, B second, A third
	// - C (index 2, process order 0) processed first: no collision → "shared"
	// - B (index 1, process order 1) collision → prefix "b-" → "b-shared"
	// - A (index 0, process order 2) collision → prefix "c-" → "c-shared"
	if len(merged.Agencies) != 3 {
		t.Errorf("expected 3 agencies, got %d", len(merged.Agencies))
	}

	// Check expected IDs exist
	expectedIDs := []gtfs.AgencyID{"shared", "b-shared", "c-shared"}
	for _, id := range expectedIDs {
		if _, ok := merged.Agencies[id]; !ok {
			t.Errorf("expected agency with ID %s", id)
		}
	}

	// Verify the names match the expected prefix order (Java-compatible: reverse processing)
	if merged.Agencies["shared"].Name != "Agency C" {
		t.Errorf("expected 'shared' to be Agency C (processed first, no prefix), got %s", merged.Agencies["shared"].Name)
	}
	if merged.Agencies["b-shared"].Name != "Agency B" {
		t.Errorf("expected 'b-shared' to be Agency B (process order 1), got %s", merged.Agencies["b-shared"].Name)
	}
	if merged.Agencies["c-shared"].Name != "Agency A" {
		t.Errorf("expected 'c-shared' to be Agency A (process order 2), got %s", merged.Agencies["c-shared"].Name)
	}
}

func TestMergeWithOverlapTestData(t *testing.T) {
	// Test with actual overlap test data
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}

	feedOverlap, err := gtfs.ReadFromPath("../testdata/overlap")
	if err != nil {
		t.Fatalf("failed to read overlap: %v", err)
	}

	// When: merged
	merger := New()
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedOverlap})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Both feeds have agency_a1, so we should have both with prefixing
	totalAgencies := len(feedA.Agencies) + len(feedOverlap.Agencies)
	if len(merged.Agencies) != totalAgencies {
		t.Errorf("expected %d agencies, got %d", totalAgencies, len(merged.Agencies))
	}

	// Verify referential integrity is maintained
	for _, route := range merged.Routes {
		if route.AgencyID != "" {
			if _, ok := merged.Agencies[route.AgencyID]; !ok {
				t.Errorf("route %s references invalid agency %s", route.ID, route.AgencyID)
			}
		}
	}

	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references invalid route %s", trip.ID, trip.RouteID)
		}
	}

	for _, st := range merged.StopTimes {
		if _, ok := merged.Trips[st.TripID]; !ok {
			t.Errorf("stop_time references invalid trip %s", st.TripID)
		}
		if _, ok := merged.Stops[st.StopID]; !ok {
			t.Errorf("stop_time references invalid stop %s", st.StopID)
		}
	}
}

// Tests for Milestone 8 - Identity-Based Duplicate Detection

func TestMergeWithIdentityDetection(t *testing.T) {
	// Given: feeds with same IDs
	feedA := gtfs.NewFeed()
	feedA.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop A", Lat: 47.6, Lon: -122.3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop B", Lat: 40.7, Lon: -74.0}

	// When: merged with DetectionIdentity
	merger := New(WithDefaultDetection(strategy.DetectionIdentity))
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: duplicates are detected and only one copy is kept
	// Java processes in REVERSE order: Feed B is processed first, Feed A second
	// Since identity detection is on, when Feed A is processed, it detects the duplicate
	// and maps to the existing entity instead of creating a prefixed one
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency with identity detection, got %d", len(merged.Agencies))
	}
	if len(merged.Stops) != 1 {
		t.Errorf("expected 1 stop with identity detection, got %d", len(merged.Stops))
	}

	// The first processed feed (B, last in array) should be kept
	if merged.Agencies["agency1"].Name != "Agency B" {
		t.Errorf("expected Agency B to be kept (processed first in reverse order), got %s", merged.Agencies["agency1"].Name)
	}
}

func TestMergeWithIdentityDetectionPreservesReferences(t *testing.T) {
	// Given: two feeds with overlapping IDs and dependencies
	feedA := gtfs.NewFeed()
	feedA.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Routes["route1"] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1A", Type: 3}
	feedA.Calendars["svc1"] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedA.Trips["trip1"] = &gtfs.Trip{ID: "trip1", RouteID: "route1", ServiceID: "svc1"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["agency1"] = &gtfs.Agency{ID: "agency1", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Routes["route1"] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1B", Type: 3}
	feedB.Calendars["svc1"] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feedB.Trips["trip1"] = &gtfs.Trip{ID: "trip1", RouteID: "route1", ServiceID: "svc1"}

	// When: merged with DetectionIdentity
	merger := New(WithDefaultDetection(strategy.DetectionIdentity))
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: only one of each entity type and references are valid
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(merged.Agencies))
	}
	if len(merged.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(merged.Routes))
	}
	if len(merged.Trips) != 1 {
		t.Errorf("expected 1 trip, got %d", len(merged.Trips))
	}
	if len(merged.Calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(merged.Calendars))
	}

	// Verify all references are valid
	for _, route := range merged.Routes {
		if _, ok := merged.Agencies[route.AgencyID]; !ok {
			t.Errorf("route %s references non-existent agency %s", route.ID, route.AgencyID)
		}
	}
	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references non-existent route %s", trip.ID, trip.RouteID)
		}
		if _, ok := merged.Calendars[trip.ServiceID]; !ok {
			t.Errorf("trip %s references non-existent service %s", trip.ID, trip.ServiceID)
		}
	}
}

func TestMergeWithIdentityDetectionAndNoOverlap(t *testing.T) {
	// Given: feeds with different IDs
	feedA := gtfs.NewFeed()
	feedA.Agencies["agency_a"] = &gtfs.Agency{ID: "agency_a", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["agency_b"] = &gtfs.Agency{ID: "agency_b", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

	// When: merged with DetectionIdentity
	merger := New(WithDefaultDetection(strategy.DetectionIdentity))
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Then: both agencies are present (no duplicates to merge)
	if len(merged.Agencies) != 2 {
		t.Errorf("expected 2 agencies, got %d", len(merged.Agencies))
	}
}

func TestMergeSetDuplicateDetectionForAll(t *testing.T) {
	// Test that SetDuplicateDetectionForAll works
	merger := New()
	merger.SetDuplicateDetectionForAll(strategy.DetectionIdentity)

	// Verify by merging feeds with overlap
	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "A", URL: "http://a.com", Timezone: "UTC"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "B", URL: "http://b.com", Timezone: "UTC"}

	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// With identity detection, duplicates should be merged
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency after identity merge, got %d", len(merged.Agencies))
	}
}

func TestMergeGetStrategyForFile(t *testing.T) {
	merger := New()

	tests := []struct {
		filename string
		wantNil  bool
	}{
		{"agency.txt", false},
		{"stops.txt", false},
		{"routes.txt", false},
		{"trips.txt", false},
		{"calendar.txt", false},
		{"calendar_dates.txt", false},
		{"shapes.txt", false},
		{"stop_times.txt", false},
		{"frequencies.txt", false},
		{"transfers.txt", false},
		{"pathways.txt", false},
		{"fare_attributes.txt", false},
		{"fare_rules.txt", false},
		{"feed_info.txt", false},
		{"areas.txt", false},
		{"unknown.txt", true},
	}

	for _, tt := range tests {
		s := merger.GetStrategyForFile(tt.filename)
		if tt.wantNil && s != nil {
			t.Errorf("GetStrategyForFile(%q) should return nil", tt.filename)
		}
		if !tt.wantNil && s == nil {
			t.Errorf("GetStrategyForFile(%q) should not return nil", tt.filename)
		}
	}
}

func TestMergeWithOverlapAndIdentityDetection(t *testing.T) {
	// Test with actual overlap test data and identity detection
	feedA, err := gtfs.ReadFromPath("../testdata/simple_a")
	if err != nil {
		t.Fatalf("failed to read simple_a: %v", err)
	}

	feedOverlap, err := gtfs.ReadFromPath("../testdata/overlap")
	if err != nil {
		t.Fatalf("failed to read overlap: %v", err)
	}

	// When: merged with DetectionIdentity
	merger := New(WithDefaultDetection(strategy.DetectionIdentity))
	merged, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedOverlap})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// The overlap feed has agency_a1 which conflicts with simple_a
	// With identity detection, the duplicate should be merged
	// simple_a has 2 agencies (agency_a1, agency_a2), overlap has 1 (agency_a1)
	// Result should have 2 agencies total (agency_a1 deduplicated, agency_a2 unique)
	if len(merged.Agencies) != 2 {
		t.Errorf("expected 2 agencies with identity detection, got %d", len(merged.Agencies))
	}

	// Verify referential integrity is maintained
	for _, route := range merged.Routes {
		if route.AgencyID != "" {
			if _, ok := merged.Agencies[route.AgencyID]; !ok {
				t.Errorf("route %s references invalid agency %s", route.ID, route.AgencyID)
			}
		}
	}

	for _, trip := range merged.Trips {
		if _, ok := merged.Routes[trip.RouteID]; !ok {
			t.Errorf("trip %s references invalid route %s", trip.ID, trip.RouteID)
		}
	}

	for _, st := range merged.StopTimes {
		if _, ok := merged.Trips[st.TripID]; !ok {
			t.Errorf("stop_time references invalid trip %s", st.TripID)
		}
		if _, ok := merged.Stops[st.StopID]; !ok {
			t.Errorf("stop_time references invalid stop %s", st.StopID)
		}
	}
}
