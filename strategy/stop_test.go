package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestStopMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping stop IDs
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{
		ID:   "stop2",
		Name: "Uptown Station",
		Lat:  40.7831,
		Lon:  -73.9712,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both stops should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Stops) != 2 {
		t.Errorf("Expected 2 stops, got %d", len(target.Stops))
	}

	if _, ok := target.Stops["stop1"]; !ok {
		t.Error("Expected stop1 to be in target")
	}
	if _, ok := target.Stops["stop2"]; !ok {
		t.Error("Expected stop2 to be in target")
	}
}

func TestStopMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have stop with ID "stop1"
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Different Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	}

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one stop1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Stops) != 1 {
		t.Errorf("Expected 1 stop, got %d", len(target.Stops))
	}

	stop := target.Stops["stop1"]
	if stop == nil {
		t.Fatal("Expected stop1 to be in target")
	}

	// The existing stop should be kept
	if stop.Name != "Downtown Station" {
		t.Errorf("Expected existing stop to be kept, got name %q", stop.Name)
	}

	// Check that the ID mapping points to the existing stop
	if ctx.StopIDMapping["stop1"] != "stop1" {
		t.Errorf("Expected StopIDMapping[stop1] = stop1, got %q", ctx.StopIDMapping["stop1"])
	}
}

func TestStopMergeUpdatesStopTimeRefs(t *testing.T) {
	// Given: source feed has a stop
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Source Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	}

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Target Stop",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: stops are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing stop
	mappedID := ctx.StopIDMapping["stop1"]
	if mappedID != "stop1" {
		t.Errorf("Expected mapped ID to be stop1, got %q", mappedID)
	}
}

func TestStopMergeUpdatesTransferRefs(t *testing.T) {
	// This test verifies the ID mapping is correct for transfer updates
	// Actual transfer updates happen in the TransferMergeStrategy
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Transfer Stop Source",
		Lat:  39.7392,
		Lon:  -104.9903,
	}

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Transfer Stop Target",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Verify mapping is set up correctly for transfer updates
	if ctx.StopIDMapping["stop1"] != "stop1" {
		t.Errorf("Expected StopIDMapping[stop1] = stop1, got %q", ctx.StopIDMapping["stop1"])
	}
}

func TestStopMergeUpdatesParentStation(t *testing.T) {
	// Given: source has a child stop referencing a parent
	// and target has colliding IDs to force prefixing
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("parent1")] = &gtfs.Stop{
		ID:           "parent1",
		Name:         "Parent Station",
		LocationType: 1,
		Lat:          40.7128,
		Lon:          -74.0060,
	}
	source.Stops[gtfs.StopID("child1")] = &gtfs.Stop{
		ID:            "child1",
		Name:          "Child Platform",
		LocationType:  0,
		ParentStation: "parent1",
		Lat:           40.7128,
		Lon:           -74.0060,
	}

	target := gtfs.NewFeed()
	// Add colliding stops to force prefixing
	target.Stops[gtfs.StopID("parent1")] = &gtfs.Stop{
		ID:           "parent1",
		Name:         "Different Parent",
		LocationType: 1,
		Lat:          41.0,
		Lon:          -75.0,
	}
	target.Stops[gtfs.StopID("child1")] = &gtfs.Stop{
		ID:            "child1",
		Name:          "Different Child",
		LocationType:  0,
		ParentStation: "parent1",
		Lat:           41.0,
		Lon:           -75.0,
	}

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with prefix (collision forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: child should reference the prefixed parent
	child := target.Stops["a_child1"]
	if child == nil {
		t.Fatal("Expected a_child1 to be in target")
	}

	if child.ParentStation != "a_parent1" {
		t.Errorf("Expected ParentStation = a_parent1, got %q", child.ParentStation)
	}
}

func TestStopMergeWithPrefix(t *testing.T) {
	// Given: source feed has a stop that collides with target
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	target := gtfs.NewFeed()
	// Add colliding stop to force prefixing
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Different Station",
		Lat:  41.0,
		Lon:  -75.0,
	}

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: stop should have prefixed ID (collision adds new prefixed stop)
	if len(target.Stops) != 2 {
		t.Errorf("Expected 2 stops, got %d", len(target.Stops))
	}

	if _, ok := target.Stops["a_stop1"]; !ok {
		t.Error("Expected a_stop1 to be in target")
	}

	if ctx.StopIDMapping["stop1"] != "a_stop1" {
		t.Errorf("Expected mapping stop1 -> a_stop1, got %q", ctx.StopIDMapping["stop1"])
	}
}

func TestStopMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have stop with same ID and error logging enabled
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Different Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	}

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogError)

	// When: merged with LogError
	err := strategy.Merge(ctx)

	// Then: should return an error
	if err == nil {
		t.Fatal("Expected error when duplicate detected with LogError")
	}
}
