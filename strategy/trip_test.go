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
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "service1",
		ShapeID:   "shape1",
		Headsign:  "Downtown Express",
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "a_")
	// Simulate that dependencies have already been merged
	ctx.RouteIDMapping[gtfs.RouteID("route1")] = gtfs.RouteID("a_route1")
	ctx.ServiceIDMapping[gtfs.ServiceID("service1")] = gtfs.ServiceID("a_service1")
	ctx.ShapeIDMapping[gtfs.ShapeID("shape1")] = gtfs.ShapeID("a_shape1")

	strategy := NewTripMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: trip should reference the mapped entities
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
