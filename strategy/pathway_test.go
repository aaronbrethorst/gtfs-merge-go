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
	// and target has a colliding pathway ID
	source := gtfs.NewFeed()
	source.Pathways = append(source.Pathways, &gtfs.Pathway{
		ID:              "pathway1",
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		PathwayMode:     1,
		IsBidirectional: 1,
	})

	target := gtfs.NewFeed()
	// Add colliding pathway to force prefixing
	target.Pathways = append(target.Pathways, &gtfs.Pathway{
		ID:              "pathway1",
		FromStopID:      "other_stop1",
		ToStopID:        "other_stop2",
		PathwayMode:     2,
		IsBidirectional: 0,
	})

	ctx := NewMergeContext(source, target, "a_")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("a_stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("a_stop2")

	strategy := NewPathwayMergeStrategy()

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: should have 2 pathways (original + prefixed)
	if len(target.Pathways) != 2 {
		t.Fatalf("Expected 2 pathways, got %d", len(target.Pathways))
	}

	// Find the new prefixed pathway
	var newPathway *gtfs.Pathway
	for _, p := range target.Pathways {
		if p.ID == "a_pathway1" {
			newPathway = p
			break
		}
	}
	if newPathway == nil {
		t.Fatal("Expected a_pathway1 to be in target")
	}

	if newPathway.FromStopID != "a_stop1" {
		t.Errorf("Expected FromStopID = a_stop1, got %q", newPathway.FromStopID)
	}
	if newPathway.ToStopID != "a_stop2" {
		t.Errorf("Expected ToStopID = a_stop2, got %q", newPathway.ToStopID)
	}
}
