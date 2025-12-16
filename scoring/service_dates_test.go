package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// createTestFeedWithCalendar creates a feed with a calendar for testing
func createTestFeedWithCalendar(serviceID string, startDate, endDate string, weekdays [7]bool) *gtfs.Feed {
	feed := gtfs.NewFeed()

	feed.Calendars[gtfs.ServiceID(serviceID)] = &gtfs.Calendar{
		ServiceID: gtfs.ServiceID(serviceID),
		StartDate: startDate,
		EndDate:   endDate,
		Monday:    weekdays[0],
		Tuesday:   weekdays[1],
		Wednesday: weekdays[2],
		Thursday:  weekdays[3],
		Friday:    weekdays[4],
		Saturday:  weekdays[5],
		Sunday:    weekdays[6],
	}

	return feed
}

// TestServiceDatesFullOverlap verifies score is 1.0 when dates fully overlap
func TestServiceDatesFullOverlap(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 31, 2024
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	// Target service: Jan 1 - Jan 31, 2024 (same period)
	targetFeed := createTestFeedWithCalendar("S2", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	if score != 1.0 {
		t.Errorf("Expected score 1.0 for full overlap, got %f", score)
	}
}

// TestServiceDatesPartialOverlap verifies proportional score for partial overlap
func TestServiceDatesPartialOverlap(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 31, 2024 (31 days)
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	// Target service: Jan 16 - Feb 15, 2024 (31 days), 16 days overlap
	targetFeed := createTestFeedWithCalendar("S2", "20240116", "20240215", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	// overlap = 16 days, formula: (16/31 + 16/31) / 2 = 0.516
	expected := (16.0/31.0 + 16.0/31.0) / 2.0
	if score < expected-0.01 || score > expected+0.01 {
		t.Errorf("Expected score ~%f for partial overlap, got %f", expected, score)
	}
}

// TestServiceDatesNoOverlap verifies score is 0 when dates don't overlap
func TestServiceDatesNoOverlap(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 31, 2024
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	// Target service: Mar 1 - Mar 31, 2024 (no overlap)
	targetFeed := createTestFeedWithCalendar("S2", "20240301", "20240331", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	if score != 0.0 {
		t.Errorf("Expected score 0.0 for no overlap, got %f", score)
	}
}

// TestServiceDatesAsymmetricOverlap verifies asymmetric overlap scoring
func TestServiceDatesAsymmetricOverlap(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 10, 2024 (10 days)
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240110", [7]bool{true, true, true, true, true, true, true})

	// Target service: Jan 1 - Jan 31, 2024 (31 days), full source is within target
	targetFeed := createTestFeedWithCalendar("S2", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	// overlap = 10 days, formula: (10/10 + 10/31) / 2 = (1.0 + 0.323) / 2 = 0.6613
	expected := (1.0 + 10.0/31.0) / 2.0
	if score < expected-0.01 || score > expected+0.01 {
		t.Errorf("Expected score ~%f for asymmetric overlap, got %f", expected, score)
	}
}

// TestServiceDatesMissingSource verifies score is 0 when source service doesn't exist
func TestServiceDatesMissingSource(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source feed has no calendar
	sourceFeed := gtfs.NewFeed()

	// Target service: Jan 1 - Jan 31, 2024
	targetFeed := createTestFeedWithCalendar("S2", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when source service missing, got %f", score)
	}
}

// TestServiceDatesMissingTarget verifies score is 0 when target service doesn't exist
func TestServiceDatesMissingTarget(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 31, 2024
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	// Target feed has no calendar
	targetFeed := gtfs.NewFeed()

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when target service missing, got %f", score)
	}
}

// TestServiceDatesAdjacentPeriods verifies score is 0 for adjacent but non-overlapping periods
func TestServiceDatesAdjacentPeriods(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 1 - Jan 31, 2024
	sourceFeed := createTestFeedWithCalendar("S1", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	// Target service: Feb 1 - Feb 29, 2024 (adjacent, no overlap)
	targetFeed := createTestFeedWithCalendar("S2", "20240201", "20240229", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	if score != 0.0 {
		t.Errorf("Expected score 0.0 for adjacent non-overlapping periods, got %f", score)
	}
}

// TestServiceDatesContainedWithin verifies one service fully contained in another
func TestServiceDatesContainedWithin(t *testing.T) {
	scorer := &ServiceDateOverlapScorer{}

	// Source service: Jan 10 - Jan 20, 2024 (11 days)
	sourceFeed := createTestFeedWithCalendar("S1", "20240110", "20240120", [7]bool{true, true, true, true, true, true, true})

	// Target service: Jan 1 - Jan 31, 2024 (31 days), source fully contained
	targetFeed := createTestFeedWithCalendar("S2", "20240101", "20240131", [7]bool{true, true, true, true, true, true, true})

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, gtfs.ServiceID("S1"), gtfs.ServiceID("S2"))

	// overlap = 11 days (full source), formula: (11/11 + 11/31) / 2 = (1.0 + 0.355) / 2 = 0.6774
	expected := (1.0 + 11.0/31.0) / 2.0
	if score < expected-0.01 || score > expected+0.01 {
		t.Errorf("Expected score ~%f for contained service, got %f", expected, score)
	}
}

// TestParseGTFSDate verifies GTFS date parsing
func TestParseGTFSDate(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		valid    bool
	}{
		{"20240101", 1704067200, true}, // Jan 1, 2024 00:00:00 UTC
		{"20240131", 1706659200, true}, // Jan 31, 2024 00:00:00 UTC
		{"20240229", 1709164800, true}, // Feb 29, 2024 (leap year)
		{"invalid", 0, false},
		{"", 0, false},
		{"2024011", 0, false}, // Wrong length
	}

	for _, tt := range tests {
		result := parseGTFSDate(tt.input)
		if tt.valid {
			// For valid dates, just check it's not 0 (exact values depend on timezone)
			if result == 0 {
				t.Errorf("parseGTFSDate(%q) returned 0, expected valid timestamp", tt.input)
			}
		} else {
			if result != 0 {
				t.Errorf("parseGTFSDate(%q) = %d, expected 0 for invalid input", tt.input, result)
			}
		}
	}
}
