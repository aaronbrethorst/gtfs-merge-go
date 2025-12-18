package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestStopMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping stop IDs
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop2",
		Name: "Uptown Station",
		Lat:  40.7831,
		Lon:  -73.9712,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Different Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Source Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Target Stop",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Transfer Stop Source",
		Lat:  39.7392,
		Lon:  -104.9903,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Transfer Stop Target",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:           "parent1",
		Name:         "Parent Station",
		LocationType: 1,
		Lat:          40.7128,
		Lon:          -74.0060,
	})
	source.AddStop(&gtfs.Stop{
		ID:            "child1",
		Name:          "Child Platform",
		LocationType:  0,
		ParentStation: "parent1",
		Lat:           40.7128,
		Lon:           -74.0060,
	})

	target := gtfs.NewFeed()
	// Add colliding stops to force prefixing
	target.AddStop(&gtfs.Stop{
		ID:           "parent1",
		Name:         "Different Parent",
		LocationType: 1,
		Lat:          41.0,
		Lon:          -75.0,
	})
	target.AddStop(&gtfs.Stop{
		ID:            "child1",
		Name:          "Different Child",
		LocationType:  0,
		ParentStation: "parent1",
		Lat:           41.0,
		Lon:           -75.0,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	// Add colliding stop to force prefixing
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Different Station",
		Lat:  41.0,
		Lon:  -75.0,
	})

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
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Different Stop",
		Lat:  39.7392,
		Lon:  -104.9903,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

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

// Fuzzy detection tests for Milestone 10

func TestStopMergeFuzzyByName(t *testing.T) {
	// Given: stops with different IDs but same name and nearby location
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_a",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop_b",
		Name: "Downtown Station",
		Lat:  40.7128, // Same location (0m apart)
		Lon:  -74.0060,
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates - source stop should map to target stop
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should only have one stop (the original target stop)
	if len(target.Stops) != 1 {
		t.Errorf("Expected 1 stop (fuzzy duplicate detected), got %d", len(target.Stops))
	}

	// Source ID should map to target ID
	if ctx.StopIDMapping["stop_a"] != "stop_b" {
		t.Errorf("Expected StopIDMapping[stop_a] = stop_b, got %q", ctx.StopIDMapping["stop_a"])
	}
}

func TestStopMergeFuzzyByDistance(t *testing.T) {
	// Given: stops with different IDs, same name, but within threshold distance
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_a",
		Name: "Central Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop_b",
		Name: "Central Station",
		// ~30m away (well within 50m threshold for score 1.0)
		Lat: 40.7130,
		Lon: -74.0062,
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: detected as duplicates
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Stops) != 1 {
		t.Errorf("Expected 1 stop (fuzzy duplicate detected), got %d", len(target.Stops))
	}

	if ctx.StopIDMapping["stop_a"] != "stop_b" {
		t.Errorf("Expected StopIDMapping[stop_a] = stop_b, got %q", ctx.StopIDMapping["stop_a"])
	}
}

func TestStopMergeFuzzyNoMatch_DifferentName(t *testing.T) {
	// Given: stops with different IDs and different names, even if same location
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_a",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop_b",
		Name: "Uptown Station", // Different name
		Lat:  40.7128,          // Same location
		Lon:  -74.0060,
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (name doesn't match)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should have two stops (no fuzzy match due to different names)
	if len(target.Stops) != 2 {
		t.Errorf("Expected 2 stops (no fuzzy match), got %d", len(target.Stops))
	}
}

func TestStopMergeFuzzyNoMatch_TooFarApart(t *testing.T) {
	// Given: stops with same name but too far apart (>500m)
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_a",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop_b",
		Name: "Downtown Station", // Same name
		Lat:  40.7228,            // ~1.1km away (beyond 500m threshold)
		Lon:  -74.0160,
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy
	err := strategy.Merge(ctx)

	// Then: NOT detected as duplicates (too far apart)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should have two stops (no fuzzy match due to distance)
	if len(target.Stops) != 2 {
		t.Errorf("Expected 2 stops (no fuzzy match - too far), got %d", len(target.Stops))
	}
}

func TestStopMergeFuzzyWithPrefix(t *testing.T) {
	// Given: stops with different names, no match expected, collision should add prefix
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Source Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	target.AddStop(&gtfs.Stop{
		ID:   "stop1",
		Name: "Target Station", // Different name
		Lat:  41.0,             // Different location
		Lon:  -75.0,
	})

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	// When: merged with DetectionFuzzy and collision
	err := strategy.Merge(ctx)

	// Then: no fuzzy match, but collision should add prefix
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Stops) != 2 {
		t.Errorf("Expected 2 stops, got %d", len(target.Stops))
	}

	if _, ok := target.Stops["a_stop1"]; !ok {
		t.Error("Expected a_stop1 to be in target (prefixed due to collision)")
	}

	if ctx.StopIDMapping["stop1"] != "a_stop1" {
		t.Errorf("Expected StopIDMapping[stop1] = a_stop1, got %q", ctx.StopIDMapping["stop1"])
	}
}

// Concurrent fuzzy matching tests for Milestone 15

func TestStopMergeFuzzyConcurrent(t *testing.T) {
	// Given: stops with same name and nearby location, concurrent processing enabled
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_a",
		Name: "Downtown Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	target := gtfs.NewFeed()
	// Add many stops to trigger concurrent processing
	for i := 0; i < 150; i++ {
		id := gtfs.StopID(string(rune('A'+i/10)) + string(rune('0'+i%10)))
		target.AddStop(&gtfs.Stop{
			ID:   id,
			Name: "Other Station " + string(rune('A'+i)),
			Lat:  41.0 + float64(i)*0.01,
			Lon:  -75.0 + float64(i)*0.01,
		})
	}
	// Add the matching stop
	target.AddStop(&gtfs.Stop{
		ID:   "stop_b",
		Name: "Downtown Station",
		Lat:  40.7128, // Same location (0m apart)
		Lon:  -74.0060,
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)
	strategy.SetConcurrent(true)
	strategy.Concurrent.MinItemsForConcurrency = 100 // Ensure concurrent is triggered

	// When: merged with DetectionFuzzy and concurrent enabled
	err := strategy.Merge(ctx)

	// Then: detected as duplicates - source stop should map to target stop
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Source ID should map to target ID
	if ctx.StopIDMapping["stop_a"] != "stop_b" {
		t.Errorf("Expected StopIDMapping[stop_a] = stop_b, got %q", ctx.StopIDMapping["stop_a"])
	}
}

func TestStopMergeFuzzyConcurrentCorrectness(t *testing.T) {
	// Test that concurrent and sequential produce the same results
	source := gtfs.NewFeed()
	source.AddStop(&gtfs.Stop{
		ID:   "stop_src",
		Name: "Test Station",
		Lat:  40.7128,
		Lon:  -74.0060,
	})

	// Create target with many stops and one matching
	createTarget := func() *gtfs.Feed {
		target := gtfs.NewFeed()
		for i := 0; i < 200; i++ {
			id := gtfs.StopID(string(rune('A'+i/26)) + string(rune('0'+i%26)))
			target.AddStop(&gtfs.Stop{
				ID:   id,
				Name: "Different Station " + string(id),
				Lat:  42.0 + float64(i)*0.001,
				Lon:  -76.0 + float64(i)*0.001,
			})
		}
		// Add matching stop
		target.AddStop(&gtfs.Stop{
			ID:   "target_match",
			Name: "Test Station",
			Lat:  40.7128,
			Lon:  -74.0060,
		})
		return target
	}

	// Sequential test
	targetSeq := createTarget()
	ctxSeq := NewMergeContext(source, targetSeq, "")
	strategySeq := NewStopMergeStrategy()
	strategySeq.SetDuplicateDetection(DetectionFuzzy)
	strategySeq.SetConcurrent(false)
	if err := strategySeq.Merge(ctxSeq); err != nil {
		t.Fatalf("Sequential merge failed: %v", err)
	}

	// Concurrent test
	targetConc := createTarget()
	ctxConc := NewMergeContext(source, targetConc, "")
	strategyConc := NewStopMergeStrategy()
	strategyConc.SetDuplicateDetection(DetectionFuzzy)
	strategyConc.SetConcurrent(true)
	strategyConc.Concurrent.MinItemsForConcurrency = 10 // Force concurrent
	if err := strategyConc.Merge(ctxConc); err != nil {
		t.Fatalf("Concurrent merge failed: %v", err)
	}

	// Both should have same mapping
	if ctxSeq.StopIDMapping["stop_src"] != ctxConc.StopIDMapping["stop_src"] {
		t.Errorf("Sequential result %q does not match concurrent result %q",
			ctxSeq.StopIDMapping["stop_src"], ctxConc.StopIDMapping["stop_src"])
	}
}

func TestStopMergeSetConcurrentWorkers(t *testing.T) {
	strategy := NewStopMergeStrategy()

	// Default workers should be runtime.NumCPU()
	if strategy.Concurrent.NumWorkers <= 0 {
		t.Error("Expected NumWorkers to be positive")
	}

	// Set workers
	strategy.SetConcurrentWorkers(8)
	if strategy.Concurrent.NumWorkers != 8 {
		t.Errorf("Expected NumWorkers to be 8, got %d", strategy.Concurrent.NumWorkers)
	}

	// Invalid value should not change
	strategy.SetConcurrentWorkers(-1)
	if strategy.Concurrent.NumWorkers != 8 {
		t.Errorf("Expected NumWorkers to remain 8, got %d", strategy.Concurrent.NumWorkers)
	}
}
