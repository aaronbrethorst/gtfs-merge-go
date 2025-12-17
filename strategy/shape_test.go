package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestShapeMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping shape IDs
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
		{ShapeID: "shape1", Lat: 40.7580, Lon: -73.9855, Sequence: 2},
	}

	target := gtfs.NewFeed()
	target.Shapes[gtfs.ShapeID("shape2")] = []*gtfs.ShapePoint{
		{ShapeID: "shape2", Lat: 40.7831, Lon: -73.9712, Sequence: 1},
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both shapes should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Shapes) != 2 {
		t.Errorf("Expected 2 shapes, got %d", len(target.Shapes))
	}

	if _, ok := target.Shapes["shape1"]; !ok {
		t.Error("Expected shape1 to be in target")
	}
	if _, ok := target.Shapes["shape2"]; !ok {
		t.Error("Expected shape2 to be in target")
	}
}

func TestShapeMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have shape with ID "shape1"
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 39.7392, Lon: -104.9903, Sequence: 1},
	}

	target := gtfs.NewFeed()
	target.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
		{ShapeID: "shape1", Lat: 40.7580, Lon: -73.9855, Sequence: 2},
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one shape1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Shapes) != 1 {
		t.Errorf("Expected 1 shape, got %d", len(target.Shapes))
	}

	points := target.Shapes["shape1"]
	if len(points) != 2 {
		t.Errorf("Expected 2 points (existing shape kept), got %d", len(points))
	}

	// Check that the ID mapping points to the existing shape
	if ctx.ShapeIDMapping["shape1"] != "shape1" {
		t.Errorf("Expected ShapeIDMapping[shape1] = shape1, got %q", ctx.ShapeIDMapping["shape1"])
	}
}

func TestShapeMergeUpdatesTripRefs(t *testing.T) {
	// Given: source feed has a shape
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 39.7392, Lon: -104.9903, Sequence: 1},
	}

	target := gtfs.NewFeed()
	target.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: shapes are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing shape
	mappedID := ctx.ShapeIDMapping["shape1"]
	if mappedID != "shape1" {
		t.Errorf("Expected mapped ID to be shape1, got %q", mappedID)
	}
}

func TestShapeMergeWithPrefix(t *testing.T) {
	// Given: source feed has a shape that collides with target
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
		{ShapeID: "shape1", Lat: 40.7580, Lon: -73.9855, Sequence: 2},
	}

	target := gtfs.NewFeed()
	// Add colliding shape to force prefixing
	target.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 41.0, Lon: -75.0, Sequence: 1},
	}

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: should have 2 shapes (original + prefixed)
	if len(target.Shapes) != 2 {
		t.Errorf("Expected 2 shapes, got %d", len(target.Shapes))
	}

	if _, ok := target.Shapes["a_shape1"]; !ok {
		t.Error("Expected a_shape1 to be in target")
	}

	points := target.Shapes["a_shape1"]
	if len(points) != 2 {
		t.Errorf("Expected 2 points for a_shape1, got %d", len(points))
	}

	// Verify all points have the new shape ID
	for _, p := range points {
		if p.ShapeID != "a_shape1" {
			t.Errorf("Expected point ShapeID = a_shape1, got %q", p.ShapeID)
		}
	}

	if ctx.ShapeIDMapping["shape1"] != "a_shape1" {
		t.Errorf("Expected mapping shape1 -> a_shape1, got %q", ctx.ShapeIDMapping["shape1"])
	}
}

func TestShapeMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have shape with same ID and error logging enabled
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 39.7392, Lon: -104.9903, Sequence: 1},
	}

	target := gtfs.NewFeed()
	target.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogError)

	// When: merged with LogError
	err := strategy.Merge(ctx)

	// Then: should return an error
	if err == nil {
		t.Fatal("Expected error when duplicate detected with LogError")
	}
}

func TestShapeMergeDeterministicSequences(t *testing.T) {
	// This test verifies that shape points get deterministic sequence numbers
	// regardless of Go's map iteration order. This is critical for Java parity.
	//
	// We create multiple shapes and verify that they always get the same
	// sequence numbers in the same order across multiple runs.

	// Given: a source feed with multiple shapes (using IDs that would sort differently)
	createSourceFeed := func() *gtfs.Feed {
		source := gtfs.NewFeed()
		// Use shape IDs that might iterate in different orders in a map
		// to verify we're sorting them properly
		shapes := map[gtfs.ShapeID][]*gtfs.ShapePoint{
			"zebra": {
				{ShapeID: "zebra", Lat: 1.0, Lon: 1.0, Sequence: 1},
				{ShapeID: "zebra", Lat: 1.1, Lon: 1.1, Sequence: 2},
			},
			"alpha": {
				{ShapeID: "alpha", Lat: 2.0, Lon: 2.0, Sequence: 1},
			},
			"beta": {
				{ShapeID: "beta", Lat: 3.0, Lon: 3.0, Sequence: 1},
				{ShapeID: "beta", Lat: 3.1, Lon: 3.1, Sequence: 2},
				{ShapeID: "beta", Lat: 3.2, Lon: 3.2, Sequence: 3},
			},
		}
		for id, points := range shapes {
			source.Shapes[id] = points
		}
		return source
	}

	// Run merge multiple times and verify sequence assignment is deterministic
	var firstRunSequences map[gtfs.ShapeID][]int
	for run := 0; run < 10; run++ {
		source := createSourceFeed()
		target := gtfs.NewFeed()
		ctx := NewMergeContext(source, target, "")
		strategy := NewShapeMergeStrategy()

		err := strategy.Merge(ctx)
		if err != nil {
			t.Fatalf("Run %d: Merge failed: %v", run, err)
		}

		// Collect sequences for each shape
		sequences := make(map[gtfs.ShapeID][]int)
		for shapeID, points := range target.Shapes {
			for _, p := range points {
				sequences[shapeID] = append(sequences[shapeID], p.Sequence)
			}
		}

		if run == 0 {
			firstRunSequences = sequences
		} else {
			// Verify all sequences match the first run
			for shapeID, seqs := range sequences {
				expected := firstRunSequences[shapeID]
				if len(seqs) != len(expected) {
					t.Errorf("Run %d: Shape %q has %d points, expected %d",
						run, shapeID, len(seqs), len(expected))
					continue
				}
				for i, seq := range seqs {
					if seq != expected[i] {
						t.Errorf("Run %d: Shape %q point %d has sequence %d, expected %d",
							run, shapeID, i, seq, expected[i])
					}
				}
			}
		}
	}

	// Verify the shapes were processed in sorted order (alpha, beta, zebra)
	// Alpha should get sequence 1
	// Beta should get sequences 2, 3, 4
	// Zebra should get sequences 5, 6
	if firstRunSequences["alpha"][0] != 1 {
		t.Errorf("Expected alpha to get sequence 1, got %d", firstRunSequences["alpha"][0])
	}
	if firstRunSequences["beta"][0] != 2 {
		t.Errorf("Expected beta first point to get sequence 2, got %d", firstRunSequences["beta"][0])
	}
	if firstRunSequences["zebra"][0] != 5 {
		t.Errorf("Expected zebra first point to get sequence 5, got %d", firstRunSequences["zebra"][0])
	}
}

func TestShapeMergeSharedCounterAcrossFeeds(t *testing.T) {
	// This test verifies that when using a shared counter, shape sequences
	// continue incrementing across multiple feeds rather than resetting.
	// This is critical for matching Java's behavior in multi-feed merges.

	// Given: two feeds with shapes
	feed1 := gtfs.NewFeed()
	feed1.Shapes[gtfs.ShapeID("shape_a")] = []*gtfs.ShapePoint{
		{ShapeID: "shape_a", Lat: 1.0, Lon: 1.0, Sequence: 1},
		{ShapeID: "shape_a", Lat: 1.1, Lon: 1.1, Sequence: 2},
	}

	feed2 := gtfs.NewFeed()
	feed2.Shapes[gtfs.ShapeID("shape_b")] = []*gtfs.ShapePoint{
		{ShapeID: "shape_b", Lat: 2.0, Lon: 2.0, Sequence: 1},
		{ShapeID: "shape_b", Lat: 2.1, Lon: 2.1, Sequence: 2},
		{ShapeID: "shape_b", Lat: 2.2, Lon: 2.2, Sequence: 3},
	}

	target := gtfs.NewFeed()

	// Create shared counter
	sharedCounter := 0

	// Merge first feed
	ctx1 := NewMergeContext(feed1, target, "")
	ctx1.SetSharedShapeCounter(&sharedCounter)
	strategy1 := NewShapeMergeStrategy()
	err := strategy1.Merge(ctx1)
	if err != nil {
		t.Fatalf("First merge failed: %v", err)
	}

	// Merge second feed
	ctx2 := NewMergeContext(feed2, target, "")
	ctx2.SetSharedShapeCounter(&sharedCounter)
	strategy2 := NewShapeMergeStrategy()
	err = strategy2.Merge(ctx2)
	if err != nil {
		t.Fatalf("Second merge failed: %v", err)
	}

	// Verify both shapes are in target
	if len(target.Shapes) != 2 {
		t.Errorf("Expected 2 shapes, got %d", len(target.Shapes))
	}

	// Verify shape_a has sequences 1, 2
	shapeA := target.Shapes["shape_a"]
	if len(shapeA) != 2 {
		t.Fatalf("Expected 2 points for shape_a, got %d", len(shapeA))
	}
	if shapeA[0].Sequence != 1 || shapeA[1].Sequence != 2 {
		t.Errorf("Expected shape_a sequences [1, 2], got [%d, %d]", shapeA[0].Sequence, shapeA[1].Sequence)
	}

	// Verify shape_b has sequences 3, 4, 5 (continuing from where shape_a left off)
	shapeB := target.Shapes["shape_b"]
	if len(shapeB) != 3 {
		t.Fatalf("Expected 3 points for shape_b, got %d", len(shapeB))
	}
	if shapeB[0].Sequence != 3 || shapeB[1].Sequence != 4 || shapeB[2].Sequence != 5 {
		t.Errorf("Expected shape_b sequences [3, 4, 5], got [%d, %d, %d]",
			shapeB[0].Sequence, shapeB[1].Sequence, shapeB[2].Sequence)
	}

	// Verify the shared counter has the final value
	if sharedCounter != 5 {
		t.Errorf("Expected shared counter to be 5, got %d", sharedCounter)
	}
}
