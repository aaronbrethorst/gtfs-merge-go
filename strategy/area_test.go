package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestAreaMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping area IDs
	source := gtfs.NewFeed()
	source.Areas[gtfs.AreaID("area1")] = &gtfs.Area{
		ID:   "area1",
		Name: "Downtown",
	}

	target := gtfs.NewFeed()
	target.Areas[gtfs.AreaID("area2")] = &gtfs.Area{
		ID:   "area2",
		Name: "Uptown",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewAreaMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both areas should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Areas) != 2 {
		t.Errorf("Expected 2 areas, got %d", len(target.Areas))
	}
}

func TestAreaMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have area with ID "area1"
	source := gtfs.NewFeed()
	source.Areas[gtfs.AreaID("area1")] = &gtfs.Area{
		ID:   "area1",
		Name: "Different Area",
	}

	target := gtfs.NewFeed()
	target.Areas[gtfs.AreaID("area1")] = &gtfs.Area{
		ID:   "area1",
		Name: "Downtown",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewAreaMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one area1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Areas) != 1 {
		t.Errorf("Expected 1 area, got %d", len(target.Areas))
	}

	area := target.Areas["area1"]
	if area.Name != "Downtown" {
		t.Errorf("Expected existing area to be kept, got name %q", area.Name)
	}

	if ctx.AreaIDMapping["area1"] != "area1" {
		t.Errorf("Expected AreaIDMapping[area1] = area1, got %q", ctx.AreaIDMapping["area1"])
	}
}

func TestAreaMergeWithPrefix(t *testing.T) {
	// Given: source feed has an area and we're using a prefix
	source := gtfs.NewFeed()
	source.Areas[gtfs.AreaID("area1")] = &gtfs.Area{
		ID:   "area1",
		Name: "Downtown",
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewAreaMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with prefix
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: area should have prefixed ID
	if len(target.Areas) != 1 {
		t.Errorf("Expected 1 area, got %d", len(target.Areas))
	}

	if _, ok := target.Areas["a_area1"]; !ok {
		t.Error("Expected a_area1 to be in target")
	}

	if ctx.AreaIDMapping["area1"] != "a_area1" {
		t.Errorf("Expected mapping area1 -> a_area1, got %q", ctx.AreaIDMapping["area1"])
	}
}
