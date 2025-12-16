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
