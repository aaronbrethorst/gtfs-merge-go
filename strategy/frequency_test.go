package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestFrequencyMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping frequencies
	source := gtfs.NewFeed()
	source.Frequencies = append(source.Frequencies, &gtfs.Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "09:00:00",
		HeadwaySecs: 600,
	})

	target := gtfs.NewFeed()
	target.Frequencies = append(target.Frequencies, &gtfs.Frequency{
		TripID:      "trip2",
		StartTime:   "06:00:00",
		EndTime:     "09:00:00",
		HeadwaySecs: 900,
	})

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("trip1")

	strategy := NewFrequencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both frequencies should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Frequencies) != 2 {
		t.Errorf("Expected 2 frequencies, got %d", len(target.Frequencies))
	}
}

func TestFrequencyMergeIdentical(t *testing.T) {
	// Given: both feeds have identical frequency
	source := gtfs.NewFeed()
	source.Frequencies = append(source.Frequencies, &gtfs.Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "09:00:00",
		HeadwaySecs: 600,
	})

	target := gtfs.NewFeed()
	target.Frequencies = append(target.Frequencies, &gtfs.Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "09:00:00",
		HeadwaySecs: 600,
	})

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("trip1")

	strategy := NewFrequencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one frequency in output
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Frequencies) != 1 {
		t.Errorf("Expected 1 frequency (duplicate skipped), got %d", len(target.Frequencies))
	}
}

func TestFrequencyMergeMappedTripID(t *testing.T) {
	// Given: source frequency references a trip that has been mapped
	source := gtfs.NewFeed()
	source.Frequencies = append(source.Frequencies, &gtfs.Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "09:00:00",
		HeadwaySecs: 600,
	})

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("a_trip1")

	strategy := NewFrequencyMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: frequency should reference mapped trip
	if len(target.Frequencies) != 1 {
		t.Fatalf("Expected 1 frequency, got %d", len(target.Frequencies))
	}

	if target.Frequencies[0].TripID != "a_trip1" {
		t.Errorf("Expected TripID = a_trip1, got %q", target.Frequencies[0].TripID)
	}
}
