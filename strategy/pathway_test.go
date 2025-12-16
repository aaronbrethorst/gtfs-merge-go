package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestPathwayMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping pathways
	source := gtfs.NewFeed()
	source.Pathways = append(source.Pathways, &gtfs.Pathway{
		ID:              "pathway1",
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		PathwayMode:     1,
		IsBidirectional: 1,
	})

	target := gtfs.NewFeed()
	target.Pathways = append(target.Pathways, &gtfs.Pathway{
		ID:              "pathway2",
		FromStopID:      "stop3",
		ToStopID:        "stop4",
		PathwayMode:     2,
		IsBidirectional: 0,
	})

	ctx := NewMergeContext(source, target, "")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("stop2")

	strategy := NewPathwayMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both pathways should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Pathways) != 2 {
		t.Errorf("Expected 2 pathways, got %d", len(target.Pathways))
	}
}

func TestPathwayMergeWithMappedStops(t *testing.T) {
	// Given: source pathway references stops that have been mapped
	source := gtfs.NewFeed()
	source.Pathways = append(source.Pathways, &gtfs.Pathway{
		ID:              "pathway1",
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		PathwayMode:     1,
		IsBidirectional: 1,
	})

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "a_")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("a_stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("a_stop2")

	strategy := NewPathwayMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: pathway should reference mapped stops
	if len(target.Pathways) != 1 {
		t.Fatalf("Expected 1 pathway, got %d", len(target.Pathways))
	}

	if target.Pathways[0].ID != "a_pathway1" {
		t.Errorf("Expected ID = a_pathway1, got %q", target.Pathways[0].ID)
	}
	if target.Pathways[0].FromStopID != "a_stop1" {
		t.Errorf("Expected FromStopID = a_stop1, got %q", target.Pathways[0].FromStopID)
	}
	if target.Pathways[0].ToStopID != "a_stop2" {
		t.Errorf("Expected ToStopID = a_stop2, got %q", target.Pathways[0].ToStopID)
	}
}
