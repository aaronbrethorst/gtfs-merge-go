package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestTripMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping trip IDs
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Downtown",
	}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{
		ID:        "trip2",
		RouteID:   "route2",
		ServiceID: "service2",
		Headsign:  "Uptown",
	}

	ctx := NewMergeContext(source, target, "")
	ctx.RouteIDMapping[gtfs.RouteID("route1")] = gtfs.RouteID("route1")
	ctx.ServiceIDMapping[gtfs.ServiceID("service1")] = gtfs.ServiceID("service1")

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both trips should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Trips) != 2 {
		t.Errorf("Expected 2 trips, got %d", len(target.Trips))
	}

	if _, ok := target.Trips["trip1"]; !ok {
		t.Error("Expected trip1 to be in target")
	}
	if _, ok := target.Trips["trip2"]; !ok {
		t.Error("Expected trip2 to be in target")
	}
}

func TestTripMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have trip with ID "trip1"
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Different Destination",
	}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Downtown",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one trip1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Trips) != 1 {
		t.Errorf("Expected 1 trip, got %d", len(target.Trips))
	}

	trip := target.Trips["trip1"]
	if trip == nil {
		t.Fatal("Expected trip1 to be in target")
	}

	// The existing trip should be kept
	if trip.Headsign != "Downtown" {
		t.Errorf("Expected existing trip to be kept, got headsign %q", trip.Headsign)
	}

	// Check that the ID mapping points to the existing trip
	if ctx.TripIDMapping["trip1"] != "trip1" {
		t.Errorf("Expected TripIDMapping[trip1] = trip1, got %q", ctx.TripIDMapping["trip1"])
	}
}

func TestTripMergeUpdatesStopTimeRefs(t *testing.T) {
	// Given: source feed has a trip
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Source Trip",
	}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Target Trip",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: trips are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing trip
	mappedID := ctx.TripIDMapping["trip1"]
	if mappedID != "trip1" {
		t.Errorf("Expected mapped ID to be trip1, got %q", mappedID)
	}
}

func TestTripMergeUpdatesFrequencyRefs(t *testing.T) {
	// This test verifies the ID mapping is correct for frequency updates
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Source Trip",
	}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Target Trip",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Verify mapping is set up correctly for frequency updates
	if ctx.TripIDMapping["trip1"] != "trip1" {
		t.Errorf("Expected TripIDMapping[trip1] = trip1, got %q", ctx.TripIDMapping["trip1"])
	}
}

func TestTripMergeWithReferences(t *testing.T) {
	// Given: source trip references route, service, and shape that have been mapped
	// and target has a colliding trip ID to force prefixing
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		ShapeID:   "shape1",
		Headsign:  "Downtown Express",
	}

	target := gtfs.NewFeed()
	// Add colliding trip to force prefixing
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "other_route",
		ServiceID: "other_service",
		Headsign:  "Different",
	}

	ctx := NewMergeContext(source, target, "a_")
	// Simulate that dependencies have already been merged with collision
	ctx.RouteIDMapping[gtfs.RouteID("route1")] = gtfs.RouteID("a_route1")
	ctx.ServiceIDMapping[gtfs.ServiceID("service1")] = gtfs.ServiceID("a_service1")
	ctx.ShapeIDMapping[gtfs.ShapeID("shape1")] = gtfs.ShapeID("a_shape1")

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: trip should reference the mapped entities with prefixed ID
	trip := target.Trips["a_trip1"]
	if trip == nil {
		t.Fatal("Expected a_trip1 to be in target")
	}

	if trip.RouteID != "a_route1" {
		t.Errorf("Expected RouteID = a_route1, got %q", trip.RouteID)
	}
	if trip.ServiceID != "a_service1" {
		t.Errorf("Expected ServiceID = a_service1, got %q", trip.ServiceID)
	}
	if trip.ShapeID != "a_shape1" {
		t.Errorf("Expected ShapeID = a_shape1, got %q", trip.ShapeID)
	}
}

func TestTripMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have trip with same ID and error logging enabled
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Different",
	}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		Headsign:  "Downtown",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogError)

	// When: merged with LogError
	err := strategy.Merge(ctx)

	// Then: should return an error
	if err == nil {
		t.Fatal("Expected error when duplicate detected with LogError")
	}
}

// Fuzzy detection tests for Milestone 10

func TestTripMergeFuzzyByStops(t *testing.T) {
	// Given: trips with different IDs but same route, service, and all stops in common
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	source.Trips[gtfs.TripID("trip_a")] = &gtfs.Trip{
		ID:        "trip_a",
		RouteID:   "route1",
		ServiceID: "svc1",
		Headsign:  "Downtown",
	}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip_a", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "trip_a", StopID: "stop2", StopSequence: 2, ArrivalTime: "08:30:00", DepartureTime: "08:30:00"},
	)

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	target.Trips[gtfs.TripID("trip_b")] = &gtfs.Trip{
		ID:        "trip_b",
		RouteID:   "route1",
		ServiceID: "svc1",
		Headsign:  "Downtown",
	}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	// Same stop times as source for perfect overlap
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip_b", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "trip_b", StopID: "stop2", StopSequence: 2, ArrivalTime: "08:30:00", DepartureTime: "08:30:00"},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.RouteIDMapping["route1"] = "route1"
	ctx.ServiceIDMapping["svc1"] = "svc1"
	ctx.StopIDMapping["stop1"] = "stop1"
	ctx.StopIDMapping["stop2"] = "stop2"

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should only have one trip (the original target trip)
	if len(target.Trips) != 1 {
		t.Errorf("Expected 1 trip (fuzzy duplicate detected), got %d", len(target.Trips))
	}

	// Source ID should map to target ID
	if ctx.TripIDMapping["trip_a"] != "trip_b" {
		t.Errorf("Expected TripIDMapping[trip_a] = trip_b, got %q", ctx.TripIDMapping["trip_a"])
	}
}

func TestTripMergeFuzzyBySchedule(t *testing.T) {
	// Given: trips with overlapping schedule windows
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	source.Trips[gtfs.TripID("trip_a")] = &gtfs.Trip{
		ID:        "trip_a",
		RouteID:   "route1",
		ServiceID: "svc1",
		Headsign:  "Express",
	}
	source.Stops[gtfs.StopID("s1")] = &gtfs.Stop{ID: "s1", Name: "First"}
	source.Stops[gtfs.StopID("s2")] = &gtfs.Stop{ID: "s2", Name: "Last"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip_a", StopID: "s1", StopSequence: 1, ArrivalTime: "09:00:00", DepartureTime: "09:00:00"},
		&gtfs.StopTime{TripID: "trip_a", StopID: "s2", StopSequence: 2, ArrivalTime: "09:30:00", DepartureTime: "09:30:00"},
	)

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	target.Trips[gtfs.TripID("trip_b")] = &gtfs.Trip{
		ID:        "trip_b",
		RouteID:   "route1",
		ServiceID: "svc1",
		Headsign:  "Express",
	}
	target.Stops[gtfs.StopID("s1")] = &gtfs.Stop{ID: "s1", Name: "First"}
	target.Stops[gtfs.StopID("s2")] = &gtfs.Stop{ID: "s2", Name: "Last"}
	// Same time window for perfect schedule overlap
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip_b", StopID: "s1", StopSequence: 1, ArrivalTime: "09:00:00", DepartureTime: "09:00:00"},
		&gtfs.StopTime{TripID: "trip_b", StopID: "s2", StopSequence: 2, ArrivalTime: "09:30:00", DepartureTime: "09:30:00"},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.RouteIDMapping["route1"] = "route1"
	ctx.ServiceIDMapping["svc1"] = "svc1"
	ctx.StopIDMapping["s1"] = "s1"
	ctx.StopIDMapping["s2"] = "s2"

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates (same route, service, stops, schedule)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Trips) != 1 {
		t.Errorf("Expected 1 trip (fuzzy duplicate detected), got %d", len(target.Trips))
	}

	if ctx.TripIDMapping["trip_a"] != "trip_b" {
		t.Errorf("Expected TripIDMapping[trip_a] = trip_b, got %q", ctx.TripIDMapping["trip_a"])
	}
}

func TestTripMergeFuzzyNoMatch_DifferentRoute(t *testing.T) {
	// Given: trips with different routes - should not fuzzy match
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route_src")] = &gtfs.Route{ID: "route_src", ShortName: "1"}
	source.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	source.Trips[gtfs.TripID("trip_a")] = &gtfs.Trip{
		ID:        "trip_a",
		RouteID:   "route_src",
		ServiceID: "svc1",
	}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip_a", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
	)

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route_tgt")] = &gtfs.Route{ID: "route_tgt", ShortName: "2"}
	target.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	target.Trips[gtfs.TripID("trip_b")] = &gtfs.Trip{
		ID:        "trip_b",
		RouteID:   "route_tgt", // Different route
		ServiceID: "svc1",
	}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip_b", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
	)

	ctx := NewMergeContext(source, target, "")
	// Routes map to different targets (different routes)
	ctx.RouteIDMapping["route_src"] = "route_src"
	ctx.ServiceIDMapping["svc1"] = "svc1"
	ctx.StopIDMapping["stop1"] = "stop1"

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (different routes)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Trips) != 2 {
		t.Errorf("Expected 2 trips (no fuzzy match - different routes), got %d", len(target.Trips))
	}
}

func TestTripMergeFuzzyNoMatch_NoScheduleOverlap(t *testing.T) {
	// Given: trips with same route/service but completely different schedules
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	source.Trips[gtfs.TripID("trip_a")] = &gtfs.Trip{
		ID:        "trip_a",
		RouteID:   "route1",
		ServiceID: "svc1",
	}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	// Morning trip
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip_a", StopID: "stop1", StopSequence: 1, ArrivalTime: "06:00:00", DepartureTime: "06:00:00"},
		&gtfs.StopTime{TripID: "trip_a", StopID: "stop2", StopSequence: 2, ArrivalTime: "06:30:00", DepartureTime: "06:30:00"},
	)

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	target.Trips[gtfs.TripID("trip_b")] = &gtfs.Trip{
		ID:        "trip_b",
		RouteID:   "route1",
		ServiceID: "svc1",
	}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	// Evening trip - no schedule overlap with source
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip_b", StopID: "stop1", StopSequence: 1, ArrivalTime: "18:00:00", DepartureTime: "18:00:00"},
		&gtfs.StopTime{TripID: "trip_b", StopID: "stop2", StopSequence: 2, ArrivalTime: "18:30:00", DepartureTime: "18:30:00"},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.RouteIDMapping["route1"] = "route1"
	ctx.ServiceIDMapping["svc1"] = "svc1"
	ctx.StopIDMapping["stop1"] = "stop1"
	ctx.StopIDMapping["stop2"] = "stop2"

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (no schedule overlap)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Trips) != 2 {
		t.Errorf("Expected 2 trips (no fuzzy match - no schedule overlap), got %d", len(target.Trips))
	}
}

func TestTripMergeFuzzyRejectsOnStopTimeDiff(t *testing.T) {
	// Given: trips that would match on route/service/stops but have different stop sequences
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	source.Trips[gtfs.TripID("trip_a")] = &gtfs.Trip{
		ID:        "trip_a",
		RouteID:   "route1",
		ServiceID: "svc1",
	}
	source.Stops[gtfs.StopID("stopA")] = &gtfs.Stop{ID: "stopA", Name: "Stop A"}
	source.Stops[gtfs.StopID("stopB")] = &gtfs.Stop{ID: "stopB", Name: "Stop B"}
	source.Stops[gtfs.StopID("stopC")] = &gtfs.Stop{ID: "stopC", Name: "Stop C"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip_a", StopID: "stopA", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "trip_a", StopID: "stopB", StopSequence: 2, ArrivalTime: "08:10:00", DepartureTime: "08:10:00"},
		&gtfs.StopTime{TripID: "trip_a", StopID: "stopC", StopSequence: 3, ArrivalTime: "08:20:00", DepartureTime: "08:20:00"},
	)

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Calendars[gtfs.ServiceID("svc1")] = &gtfs.Calendar{ServiceID: "svc1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	target.Trips[gtfs.TripID("trip_b")] = &gtfs.Trip{
		ID:        "trip_b",
		RouteID:   "route1",
		ServiceID: "svc1",
	}
	target.Stops[gtfs.StopID("stopA")] = &gtfs.Stop{ID: "stopA", Name: "Stop A"}
	target.Stops[gtfs.StopID("stopB")] = &gtfs.Stop{ID: "stopB", Name: "Stop B"}
	// Different stop at sequence 3 (stopD instead of stopC)
	target.Stops[gtfs.StopID("stopD")] = &gtfs.Stop{ID: "stopD", Name: "Stop D"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip_b", StopID: "stopA", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"},
		&gtfs.StopTime{TripID: "trip_b", StopID: "stopB", StopSequence: 2, ArrivalTime: "08:10:00", DepartureTime: "08:10:00"},
		&gtfs.StopTime{TripID: "trip_b", StopID: "stopD", StopSequence: 3, ArrivalTime: "08:20:00", DepartureTime: "08:20:00"}, // Different stop!
	)

	ctx := NewMergeContext(source, target, "")
	ctx.RouteIDMapping["route1"] = "route1"
	ctx.ServiceIDMapping["svc1"] = "svc1"
	ctx.StopIDMapping["stopA"] = "stopA"
	ctx.StopIDMapping["stopB"] = "stopB"
	ctx.StopIDMapping["stopC"] = "stopC"
	ctx.StopIDMapping["stopD"] = "stopD"

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates due to different stop at same sequence
	// (validation should reject when stops differ at same sequence position)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should have 2 trips (no fuzzy match due to stop sequence difference)
	if len(target.Trips) != 2 {
		t.Errorf("Expected 2 trips (no fuzzy match - different stops), got %d", len(target.Trips))
	}
}
