package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// createTestFeedWithTripSchedule creates a feed with a trip and its stop times with schedules
func createTestFeedWithTripSchedule(tripID string, stopTimes []struct {
	stopID    string
	arrival   string
	departure string
}) *gtfs.Feed {
	feed := gtfs.NewFeed()

	// Add trip
	feed.Trips[gtfs.TripID(tripID)] = &gtfs.Trip{
		ID:      gtfs.TripID(tripID),
		RouteID: "R1",
	}

	// Add stops and stop times
	for i, st := range stopTimes {
		feed.Stops[gtfs.StopID(st.stopID)] = &gtfs.Stop{
			ID:   gtfs.StopID(st.stopID),
			Name: st.stopID,
		}
		feed.StopTimes = append(feed.StopTimes, &gtfs.StopTime{
			TripID:        gtfs.TripID(tripID),
			StopID:        gtfs.StopID(st.stopID),
			StopSequence:  i + 1,
			ArrivalTime:   st.arrival,
			DepartureTime: st.departure,
		})
	}

	return feed
}

// TestTripScheduleExactMatch verifies score is 1.0 for exact schedule match
func TestTripScheduleExactMatch(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 08:00 to 09:00
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:05:00"},
		{"B", "08:30:00", "08:35:00"},
		{"C", "09:00:00", "09:00:00"},
	})

	// Target trip: 08:00 to 09:00 (same schedule)
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:05:00"},
		{"B", "08:30:00", "08:35:00"},
		{"C", "09:00:00", "09:00:00"},
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 for exact schedule match, got %f", score)
	}
}

// TestTripSchedulePartialOverlap verifies proportional score for partial schedule overlap
func TestTripSchedulePartialOverlap(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 08:00 to 10:00 (2 hours)
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "10:00:00", "10:00:00"},
	})

	// Target trip: 09:00 to 11:00 (2 hours), 1 hour overlap
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "09:00:00", "09:00:00"},
		{"B", "11:00:00", "11:00:00"},
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// overlap = 1 hour, formula: (1/2 + 1/2) / 2 = 0.5
	expected := 0.5
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f for partial overlap, got %f", expected, score)
	}
}

// TestTripScheduleNoOverlap verifies score is 0 for non-overlapping schedules
func TestTripScheduleNoOverlap(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 08:00 to 09:00
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "09:00:00", "09:00:00"},
	})

	// Target trip: 10:00 to 11:00 (no overlap)
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "10:00:00", "10:00:00"},
		{"B", "11:00:00", "11:00:00"},
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 for non-overlapping schedules, got %f", score)
	}
}

// TestTripScheduleAsymmetricOverlap verifies asymmetric overlap scoring
func TestTripScheduleAsymmetricOverlap(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 08:00 to 09:00 (1 hour)
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "09:00:00", "09:00:00"},
	})

	// Target trip: 08:00 to 12:00 (4 hours), full overlap of source
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "12:00:00", "12:00:00"},
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// overlap = 1 hour, formula: (1/1 + 1/4) / 2 = (1.0 + 0.25) / 2 = 0.625
	expected := 0.625
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f for asymmetric overlap, got %f", expected, score)
	}
}

// TestTripScheduleEmptySource verifies score is 0 when source has no stop times
func TestTripScheduleEmptySource(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip has no stop times
	sourceFeed := gtfs.NewFeed()
	sourceFeed.Trips["T1"] = &gtfs.Trip{ID: "T1", RouteID: "R1"}

	// Target trip: 08:00 to 09:00
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "09:00:00", "09:00:00"},
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when source has no stop times, got %f", score)
	}
}

// TestTripScheduleEmptyTarget verifies score is 0 when target has no stop times
func TestTripScheduleEmptyTarget(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 08:00 to 09:00
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "08:00:00", "08:00:00"},
		{"B", "09:00:00", "09:00:00"},
	})

	// Target trip has no stop times
	targetFeed := gtfs.NewFeed()
	targetFeed.Trips["T2"] = &gtfs.Trip{ID: "T2", RouteID: "R1"}

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when target has no stop times, got %f", score)
	}
}

// TestTripScheduleOvernightTrip verifies handling of times > 24:00:00
func TestTripScheduleOvernightTrip(t *testing.T) {
	scorer := &TripScheduleOverlapScorer{}

	// Source trip: 23:00 to 25:00 (overnight, 2 hours)
	sourceFeed := createTestFeedWithTripSchedule("T1", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "23:00:00", "23:00:00"},
		{"B", "25:00:00", "25:00:00"}, // 1 AM next day
	})

	// Target trip: 24:00 to 26:00 (overnight, 2 hours), 1 hour overlap
	targetFeed := createTestFeedWithTripSchedule("T2", []struct {
		stopID    string
		arrival   string
		departure string
	}{
		{"A", "24:00:00", "24:00:00"}, // Midnight
		{"B", "26:00:00", "26:00:00"}, // 2 AM next day
	})

	sourceTrip := sourceFeed.Trips["T1"]
	targetTrip := targetFeed.Trips["T2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceTrip, targetTrip)

	// overlap = 1 hour, formula: (1/2 + 1/2) / 2 = 0.5
	expected := 0.5
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f for overnight trip overlap, got %f", expected, score)
	}
}

// TestParseGTFSTime verifies GTFS time parsing
func TestParseGTFSTime(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"08:00:00", 8 * 3600},
		{"12:30:45", 12*3600 + 30*60 + 45},
		{"24:00:00", 24 * 3600},
		{"25:30:00", 25*3600 + 30*60},
		{"00:00:00", 0},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := parseGTFSTime(tt.input)
		if result != tt.expected {
			t.Errorf("parseGTFSTime(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
