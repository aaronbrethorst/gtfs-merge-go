package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// createTestFeedWithTrip creates a feed with a trip and its stop times
func createTestFeedWithTrip(tripID string, stopIDs []string) *gtfs.Feed {
	feed := gtfs.NewFeed()

	// Add trip
	feed.Trips[gtfs.TripID(tripID)] = &gtfs.Trip{
		ID:      gtfs.TripID(tripID),
		RouteID: "R1",
	}

	// Add stops and stop times
	for i, stopID := range stopIDs {
		feed.Stops[gtfs.StopID(stopID)] = &gtfs.Stop{
			ID:   gtfs.StopID(stopID),
			Name: stopID,
		}
		feed.StopTimes = append(feed.StopTimes, &gtfs.StopTime{
			TripID:       gtfs.TripID(tripID),
			StopID:       gtfs.StopID(stopID),
			StopSequence: i + 1,
		})
	}

	return feed
}

// TestTripStopsAllInCommon verifies score is 1.0 when all stops are shared
func TestTripStopsAllInCommon(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip serves stops A, B, C
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B", "C"})

	// Target trip serves the same stops A, B, C
	targetFeed := createTestFeedWithTrip("T2", []string{"A", "B", "C"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 when all stops are shared, got %f", score)
	}
}

// TestTripStopsNoneInCommon verifies score is 0 when no stops are shared
func TestTripStopsNoneInCommon(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip serves stops A, B, C
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B", "C"})

	// Target trip serves different stops X, Y, Z
	targetFeed := createTestFeedWithTrip("T2", []string{"X", "Y", "Z"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when no stops are shared, got %f", score)
	}
}

// TestTripStopsPartialOverlap verifies proportional score for partial overlap
func TestTripStopsPartialOverlap(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip serves stops A, B, C (3 stops)
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B", "C"})

	// Target trip serves stops B, C, D (3 stops), 2 in common
	targetFeed := createTestFeedWithTrip("T2", []string{"B", "C", "D"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// common = 2, formula: (2/3 + 2/3) / 2 = 0.6667
	expected := (2.0/3.0 + 2.0/3.0) / 2.0
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f, got %f", expected, score)
	}
}

// TestTripStopsAsymmetricOverlap verifies asymmetric overlap scoring
func TestTripStopsAsymmetricOverlap(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip serves stops A, B (2 stops)
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B"})

	// Target trip serves stops A, B, C, D (4 stops), 2 in common
	targetFeed := createTestFeedWithTrip("T2", []string{"A", "B", "C", "D"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// common = 2, formula: (2/2 + 2/4) / 2 = (1.0 + 0.5) / 2 = 0.75
	expected := 0.75
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f, got %f", expected, score)
	}
}

// TestTripStopsEmptySource verifies score is 0 when source has no stops
func TestTripStopsEmptySource(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip has no stops
	sourceFeed := gtfs.NewFeed()
	sourceFeed.Trips["T1"] = &gtfs.Trip{ID: "T1", RouteID: "R1"}

	// Target trip serves stops A, B, C
	targetFeed := createTestFeedWithTrip("T2", []string{"A", "B", "C"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when source has no stops, got %f", score)
	}
}

// TestTripStopsEmptyTarget verifies score is 0 when target has no stops
func TestTripStopsEmptyTarget(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip serves stops A, B, C
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B", "C"})

	// Target trip has no stops
	targetFeed := gtfs.NewFeed()
	targetFeed.Trips["T2"] = &gtfs.Trip{ID: "T2", RouteID: "R1"}

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when target has no stops, got %f", score)
	}
}

// TestTripStopsDuplicateStops verifies stops are counted only once
func TestTripStopsDuplicateStops(t *testing.T) {
	scorer := &TripStopsInCommonScorer{}

	// Source trip visits A, B, A (returns to A)
	sourceFeed := createTestFeedWithTrip("T1", []string{"A", "B", "A"})

	// Target trip visits A, B
	targetFeed := createTestFeedWithTrip("T2", []string{"A", "B"})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// Both trips serve unique stops {A, B}, should be 1.0
	if score != 1.0 {
		t.Errorf("Expected score 1.0 with duplicate stops counted once, got %f", score)
	}
}
