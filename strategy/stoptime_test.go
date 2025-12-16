package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestStopTimeMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping stop times
	source := gtfs.NewFeed()
	source.StopTimes = append(source.StopTimes, &gtfs.StopTime{
		TripID:        "trip1",
		StopID:        "stop1",
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:01:00",
	})

	target := gtfs.NewFeed()
	target.StopTimes = append(target.StopTimes, &gtfs.StopTime{
		TripID:        "trip2",
		StopID:        "stop2",
		StopSequence:  1,
		ArrivalTime:   "09:00:00",
		DepartureTime: "09:01:00",
	})

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("trip1")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("stop1")

	strategy := NewStopTimeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both stop times should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.StopTimes) != 2 {
		t.Errorf("Expected 2 stop times, got %d", len(target.StopTimes))
	}
}

func TestStopTimeMergeIdentical(t *testing.T) {
	// Given: both feeds have stop time for same trip+sequence
	source := gtfs.NewFeed()
	source.StopTimes = append(source.StopTimes, &gtfs.StopTime{
		TripID:        "trip1",
		StopID:        "stop1",
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:01:00",
	})

	target := gtfs.NewFeed()
	target.StopTimes = append(target.StopTimes, &gtfs.StopTime{
		TripID:        "trip1",
		StopID:        "stop1",
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:01:00",
	})

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("trip1")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("stop1")

	strategy := NewStopTimeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one stop time in output
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.StopTimes) != 1 {
		t.Errorf("Expected 1 stop time (duplicate skipped), got %d", len(target.StopTimes))
	}
}

func TestStopTimeMergeMappedRefs(t *testing.T) {
	// Given: source stop time references trip and stop that have been mapped
	source := gtfs.NewFeed()
	source.StopTimes = append(source.StopTimes, &gtfs.StopTime{
		TripID:        "trip1",
		StopID:        "stop1",
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:01:00",
	})

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	ctx.TripIDMapping[gtfs.TripID("trip1")] = gtfs.TripID("a_trip1")
	ctx.StopIDMapping[gtfs.StopID("stop1")] = gtfs.StopID("a_stop1")

	strategy := NewStopTimeMergeStrategy()

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: stop time should reference mapped entities
	if len(target.StopTimes) != 1 {
		t.Fatalf("Expected 1 stop time, got %d", len(target.StopTimes))
	}

	if target.StopTimes[0].TripID != "a_trip1" {
		t.Errorf("Expected TripID = a_trip1, got %q", target.StopTimes[0].TripID)
	}
	if target.StopTimes[0].StopID != "a_stop1" {
		t.Errorf("Expected StopID = a_stop1, got %q", target.StopTimes[0].StopID)
	}
}
