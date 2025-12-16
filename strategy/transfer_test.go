package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestTransferMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping transfers
	source := gtfs.NewFeed()
	source.Transfers = append(source.Transfers, &gtfs.Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		TransferType:    0,
		MinTransferTime: 120,
	})

	target := gtfs.NewFeed()
	target.Transfers = append(target.Transfers, &gtfs.Transfer{
		FromStopID:      "stop3",
		ToStopID:        "stop4",
		TransferType:    1,
		MinTransferTime: 180,
	})

	ctx := NewMergeContext(source, target, "")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("stop2")

	strategy := NewTransferMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both transfers should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Transfers) != 2 {
		t.Errorf("Expected 2 transfers, got %d", len(target.Transfers))
	}
}

func TestTransferMergeIdentical(t *testing.T) {
	// Given: both feeds have identical transfer
	source := gtfs.NewFeed()
	source.Transfers = append(source.Transfers, &gtfs.Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		TransferType:    0,
		MinTransferTime: 120,
	})

	target := gtfs.NewFeed()
	target.Transfers = append(target.Transfers, &gtfs.Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		TransferType:    0,
		MinTransferTime: 120,
	})

	ctx := NewMergeContext(source, target, "")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("stop2")

	strategy := NewTransferMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one transfer in output
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Transfers) != 1 {
		t.Errorf("Expected 1 transfer (duplicate skipped), got %d", len(target.Transfers))
	}
}

func TestTransferMergeMappedStops(t *testing.T) {
	// Given: source transfer references stops that have been mapped
	source := gtfs.NewFeed()
	source.Transfers = append(source.Transfers, &gtfs.Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		TransferType:    0,
		MinTransferTime: 120,
	})

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("a_stop1")
	ctx.StopIDMapping[gtfs.StopID("stop2")] = gtfs.StopID("a_stop2")

	strategy := NewTransferMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: transfer should reference mapped stops
	if len(target.Transfers) != 1 {
		t.Fatalf("Expected 1 transfer, got %d", len(target.Transfers))
	}

	if target.Transfers[0].FromStopID != "a_stop1" {
		t.Errorf("Expected FromStopID = a_stop1, got %q", target.Transfers[0].FromStopID)
	}
	if target.Transfers[0].ToStopID != "a_stop2" {
		t.Errorf("Expected ToStopID = a_stop2, got %q", target.Transfers[0].ToStopID)
	}
}
