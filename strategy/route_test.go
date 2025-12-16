package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestRouteMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping route IDs
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Downtown Express",
		Type:     3,
	}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route2")] = &gtfs.Route{
		ID:       "route2",
		LongName: "Uptown Local",
		Type:     3,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both routes should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(target.Routes))
	}

	if _, ok := target.Routes["route1"]; !ok {
		t.Error("Expected route1 to be in target")
	}
	if _, ok := target.Routes["route2"]; !ok {
		t.Error("Expected route2 to be in target")
	}
}

func TestRouteMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have route with ID "route1"
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Different Route",
		Type:     1,
	}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Downtown Express",
		Type:     3,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one route1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(target.Routes))
	}

	route := target.Routes["route1"]
	if route == nil {
		t.Fatal("Expected route1 to be in target")
	}

	// The existing route should be kept
	if route.LongName != "Downtown Express" {
		t.Errorf("Expected existing route to be kept, got name %q", route.LongName)
	}

	// Check that the ID mapping points to the existing route
	if ctx.RouteIDMapping["route1"] != "route1" {
		t.Errorf("Expected RouteIDMapping[route1] = route1, got %q", ctx.RouteIDMapping["route1"])
	}
}

func TestRouteMergeUpdatesTripRefs(t *testing.T) {
	// Given: source feed has a route
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Source Route",
		Type:     3,
	}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Target Route",
		Type:     3,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: routes are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing route
	mappedID := ctx.RouteIDMapping["route1"]
	if mappedID != "route1" {
		t.Errorf("Expected mapped ID to be route1, got %q", mappedID)
	}
}

func TestRouteMergeUpdatesFareRuleRefs(t *testing.T) {
	// This test verifies the ID mapping is correct for fare rule updates
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Source Route",
		Type:     3,
	}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Target Route",
		Type:     3,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Verify mapping is set up correctly for fare rule updates
	if ctx.RouteIDMapping["route1"] != "route1" {
		t.Errorf("Expected RouteIDMapping[route1] = route1, got %q", ctx.RouteIDMapping["route1"])
	}
}

func TestRouteMergeWithAgencyRef(t *testing.T) {
	// Given: source route references an agency that has been mapped
	// and target has a colliding route to force prefixing
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		AgencyID: "agency1",
		LongName: "Downtown Express",
		Type:     3,
	}

	target := gtfs.NewFeed()
	// Add colliding route to force prefixing
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Different Route",
		Type:     1,
	}

	ctx := NewMergeContext(source, target, "a_")
	// Simulate that agency has already been merged with collision
	ctx.AgencyIDMapping[gtfs.AgencyID("agency1")] = gtfs.AgencyID("a_agency1")

	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: route should reference the mapped agency with prefixed ID
	route := target.Routes["a_route1"]
	if route == nil {
		t.Fatal("Expected a_route1 to be in target")
	}

	if route.AgencyID != "a_agency1" {
		t.Errorf("Expected AgencyID = a_agency1, got %q", route.AgencyID)
	}
}

func TestRouteMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have route with same ID and error logging enabled
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Different Route",
		Type:     1,
	}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{
		ID:       "route1",
		LongName: "Downtown Express",
		Type:     3,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewRouteMergeStrategy()
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

func TestRouteMergeFuzzyByName(t *testing.T) {
	// Given: routes with different IDs but same short_name and long_name
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	source.Routes[gtfs.RouteID("route_a")] = &gtfs.Route{
		ID:        "route_a",
		AgencyID:  "agency1",
		ShortName: "1",
		LongName:  "Downtown Express",
		Type:      3,
	}
	// Add trip and stop times for this route
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{
		ID:        "trip1",
		RouteID:   "route_a",
		ServiceID: "svc1",
	}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1},
		&gtfs.StopTime{TripID: "trip1", StopID: "stop2", StopSequence: 2},
	)

	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	target.Routes[gtfs.RouteID("route_b")] = &gtfs.Route{
		ID:        "route_b",
		AgencyID:  "agency1",
		ShortName: "1",
		LongName:  "Downtown Express",
		Type:      3,
	}
	// Add trip and stop times with same stops
	target.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{
		ID:        "trip2",
		RouteID:   "route_b",
		ServiceID: "svc1",
	}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop 2"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip2", StopID: "stop1", StopSequence: 1},
		&gtfs.StopTime{TripID: "trip2", StopID: "stop2", StopSequence: 2},
	)

	ctx := NewMergeContext(source, target, "")
	// Map agencies so fuzzy matching works on agency reference
	ctx.AgencyIDMapping["agency1"] = "agency1"
	// Map stops so fuzzy matching works on stop references
	ctx.StopIDMapping["stop1"] = "stop1"
	ctx.StopIDMapping["stop2"] = "stop2"

	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should only have one route (the original target route)
	if len(target.Routes) != 1 {
		t.Errorf("Expected 1 route (fuzzy duplicate detected), got %d", len(target.Routes))
	}

	// Source ID should map to target ID
	if ctx.RouteIDMapping["route_a"] != "route_b" {
		t.Errorf("Expected RouteIDMapping[route_a] = route_b, got %q", ctx.RouteIDMapping["route_a"])
	}
}

func TestRouteMergeFuzzyByStops(t *testing.T) {
	// Given: routes with same names and shared stops across trips
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	source.Routes[gtfs.RouteID("route_a")] = &gtfs.Route{
		ID:        "route_a",
		AgencyID:  "agency1",
		ShortName: "Express",
		LongName:  "City Express",
		Type:      3,
	}
	source.Trips[gtfs.TripID("src_trip")] = &gtfs.Trip{
		ID:        "src_trip",
		RouteID:   "route_a",
		ServiceID: "svc1",
	}
	source.Stops[gtfs.StopID("s1")] = &gtfs.Stop{ID: "s1", Name: "First Stop"}
	source.Stops[gtfs.StopID("s2")] = &gtfs.Stop{ID: "s2", Name: "Second Stop"}
	source.Stops[gtfs.StopID("s3")] = &gtfs.Stop{ID: "s3", Name: "Third Stop"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "src_trip", StopID: "s1", StopSequence: 1},
		&gtfs.StopTime{TripID: "src_trip", StopID: "s2", StopSequence: 2},
		&gtfs.StopTime{TripID: "src_trip", StopID: "s3", StopSequence: 3},
	)

	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	target.Routes[gtfs.RouteID("route_b")] = &gtfs.Route{
		ID:        "route_b",
		AgencyID:  "agency1",
		ShortName: "Express",
		LongName:  "City Express",
		Type:      3,
	}
	target.Trips[gtfs.TripID("tgt_trip")] = &gtfs.Trip{
		ID:        "tgt_trip",
		RouteID:   "route_b",
		ServiceID: "svc1",
	}
	// Same stops as source (will have high overlap score)
	target.Stops[gtfs.StopID("s1")] = &gtfs.Stop{ID: "s1", Name: "First Stop"}
	target.Stops[gtfs.StopID("s2")] = &gtfs.Stop{ID: "s2", Name: "Second Stop"}
	target.Stops[gtfs.StopID("s3")] = &gtfs.Stop{ID: "s3", Name: "Third Stop"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "tgt_trip", StopID: "s1", StopSequence: 1},
		&gtfs.StopTime{TripID: "tgt_trip", StopID: "s2", StopSequence: 2},
		&gtfs.StopTime{TripID: "tgt_trip", StopID: "s3", StopSequence: 3},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.AgencyIDMapping["agency1"] = "agency1"
	ctx.StopIDMapping["s1"] = "s1"
	ctx.StopIDMapping["s2"] = "s2"
	ctx.StopIDMapping["s3"] = "s3"

	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates (same names and all stops in common)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Routes) != 1 {
		t.Errorf("Expected 1 route (fuzzy duplicate detected), got %d", len(target.Routes))
	}

	if ctx.RouteIDMapping["route_a"] != "route_b" {
		t.Errorf("Expected RouteIDMapping[route_a] = route_b, got %q", ctx.RouteIDMapping["route_a"])
	}
}

func TestRouteMergeFuzzyNoMatch_DifferentNames(t *testing.T) {
	// Given: routes with different names - should not fuzzy match
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	source.Routes[gtfs.RouteID("route_a")] = &gtfs.Route{
		ID:        "route_a",
		AgencyID:  "agency1",
		ShortName: "1",
		LongName:  "Downtown Express", // Different long name
		Type:      3,
	}
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{ID: "trip1", RouteID: "route_a", ServiceID: "svc1"}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1},
	)

	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	target.Routes[gtfs.RouteID("route_b")] = &gtfs.Route{
		ID:        "route_b",
		AgencyID:  "agency1",
		ShortName: "1",
		LongName:  "Uptown Local", // Different long name
		Type:      3,
	}
	target.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{ID: "trip2", RouteID: "route_b", ServiceID: "svc1"}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop 1"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip2", StopID: "stop1", StopSequence: 1},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.AgencyIDMapping["agency1"] = "agency1"
	ctx.StopIDMapping["stop1"] = "stop1"

	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (different long names)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Routes) != 2 {
		t.Errorf("Expected 2 routes (no fuzzy match - different names), got %d", len(target.Routes))
	}
}

func TestRouteMergeFuzzyNoMatch_NoSharedStops(t *testing.T) {
	// Given: routes with same names but no shared stops
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	source.Routes[gtfs.RouteID("route_a")] = &gtfs.Route{
		ID:        "route_a",
		AgencyID:  "agency1",
		ShortName: "Express",
		LongName:  "City Express",
		Type:      3,
	}
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{ID: "trip1", RouteID: "route_a", ServiceID: "svc1"}
	source.Stops[gtfs.StopID("stop_src_1")] = &gtfs.Stop{ID: "stop_src_1", Name: "Source Stop 1"}
	source.Stops[gtfs.StopID("stop_src_2")] = &gtfs.Stop{ID: "stop_src_2", Name: "Source Stop 2"}
	source.StopTimes = append(source.StopTimes,
		&gtfs.StopTime{TripID: "trip1", StopID: "stop_src_1", StopSequence: 1},
		&gtfs.StopTime{TripID: "trip1", StopID: "stop_src_2", StopSequence: 2},
	)

	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	target.Routes[gtfs.RouteID("route_b")] = &gtfs.Route{
		ID:        "route_b",
		AgencyID:  "agency1",
		ShortName: "Express",
		LongName:  "City Express",
		Type:      3,
	}
	target.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{ID: "trip2", RouteID: "route_b", ServiceID: "svc1"}
	target.Stops[gtfs.StopID("stop_tgt_1")] = &gtfs.Stop{ID: "stop_tgt_1", Name: "Target Stop 1"}
	target.Stops[gtfs.StopID("stop_tgt_2")] = &gtfs.Stop{ID: "stop_tgt_2", Name: "Target Stop 2"}
	target.StopTimes = append(target.StopTimes,
		&gtfs.StopTime{TripID: "trip2", StopID: "stop_tgt_1", StopSequence: 1},
		&gtfs.StopTime{TripID: "trip2", StopID: "stop_tgt_2", StopSequence: 2},
	)

	ctx := NewMergeContext(source, target, "")
	ctx.AgencyIDMapping["agency1"] = "agency1"
	// No stop mappings - stops are completely different

	strategy := NewRouteMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (no shared stops gives score 0)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Routes) != 2 {
		t.Errorf("Expected 2 routes (no fuzzy match - no shared stops), got %d", len(target.Routes))
	}
}
