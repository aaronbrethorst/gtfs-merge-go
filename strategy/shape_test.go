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
	// Given: source feed has a shape and we're using a prefix
	source := gtfs.NewFeed()
	source.Shapes[gtfs.ShapeID("shape1")] = []*gtfs.ShapePoint{
		{ShapeID: "shape1", Lat: 40.7128, Lon: -74.0060, Sequence: 1},
		{ShapeID: "shape1", Lat: 40.7580, Lon: -73.9855, Sequence: 2},
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewShapeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with prefix
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: shape should have prefixed ID
	if len(target.Shapes) != 1 {
		t.Errorf("Expected 1 shape, got %d", len(target.Shapes))
	}

	if _, ok := target.Shapes["a_shape1"]; !ok {
		t.Error("Expected a_shape1 to be in target")
	}

	points := target.Shapes["a_shape1"]
	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
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
