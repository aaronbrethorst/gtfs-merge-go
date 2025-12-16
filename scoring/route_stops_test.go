package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// createTestFeedWithRouteStops creates a feed with routes and their stop times for testing
func createTestFeedWithRouteStops(routeID string, tripID string, stopIDs []string) *gtfs.Feed {
	feed := gtfs.NewFeed()

	// Add route
	feed.Routes[gtfs.RouteID(routeID)] = &gtfs.Route{
		ID:        gtfs.RouteID(routeID),
		ShortName: routeID,
	}

	// Add trip
	feed.Trips[gtfs.TripID(tripID)] = &gtfs.Trip{
		ID:      gtfs.TripID(tripID),
		RouteID: gtfs.RouteID(routeID),
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

// TestRouteStopsAllInCommon verifies score is 1.0 when all stops are shared
func TestRouteStopsAllInCommon(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route serves stops A, B, C
	sourceFeed := createTestFeedWithRouteStops("R1", "T1", []string{"A", "B", "C"})

	// Target route serves the same stops A, B, C
	targetFeed := createTestFeedWithRouteStops("R2", "T2", []string{"A", "B", "C"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 when all stops are shared, got %f", score)
	}
}

// TestRouteStopsNoneInCommon verifies score is 0 when no stops are shared
func TestRouteStopsNoneInCommon(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route serves stops A, B, C
	sourceFeed := createTestFeedWithRouteStops("R1", "T1", []string{"A", "B", "C"})

	// Target route serves different stops X, Y, Z
	targetFeed := createTestFeedWithRouteStops("R2", "T2", []string{"X", "Y", "Z"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when no stops are shared, got %f", score)
	}
}

// TestRouteStopsPartialOverlap verifies proportional score for partial overlap
func TestRouteStopsPartialOverlap(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route serves stops A, B, C (3 stops)
	sourceFeed := createTestFeedWithRouteStops("R1", "T1", []string{"A", "B", "C"})

	// Target route serves stops B, C, D (3 stops), 2 in common
	targetFeed := createTestFeedWithRouteStops("R2", "T2", []string{"B", "C", "D"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	// common = 2, formula: (2/3 + 2/3) / 2 = 0.6667
	expected := (2.0/3.0 + 2.0/3.0) / 2.0
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f, got %f", expected, score)
	}
}

// TestRouteStopsAsymmetricOverlap verifies asymmetric overlap scoring
func TestRouteStopsAsymmetricOverlap(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route serves stops A, B (2 stops)
	sourceFeed := createTestFeedWithRouteStops("R1", "T1", []string{"A", "B"})

	// Target route serves stops A, B, C, D (4 stops), 2 in common
	targetFeed := createTestFeedWithRouteStops("R2", "T2", []string{"A", "B", "C", "D"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	// common = 2, formula: (2/2 + 2/4) / 2 = (1.0 + 0.5) / 2 = 0.75
	expected := 0.75
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f, got %f", expected, score)
	}
}

// TestRouteStopsEmptySource verifies score is 0 when source has no stops
func TestRouteStopsEmptySource(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route has no stops
	sourceFeed := gtfs.NewFeed()
	sourceFeed.Routes["R1"] = &gtfs.Route{ID: "R1", ShortName: "R1"}

	// Target route serves stops A, B, C
	targetFeed := createTestFeedWithRouteStops("R2", "T2", []string{"A", "B", "C"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when source has no stops, got %f", score)
	}
}

// TestRouteStopsEmptyTarget verifies score is 0 when target has no stops
func TestRouteStopsEmptyTarget(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Source route serves stops A, B, C
	sourceFeed := createTestFeedWithRouteStops("R1", "T1", []string{"A", "B", "C"})

	// Target route has no stops
	targetFeed := gtfs.NewFeed()
	targetFeed.Routes["R2"] = &gtfs.Route{ID: "R2", ShortName: "R2"}

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when target has no stops, got %f", score)
	}
}

// TestRouteStopsMultipleTrips verifies stops are collected from all trips for a route
func TestRouteStopsMultipleTrips(t *testing.T) {
	scorer := &RouteStopsInCommonScorer{}

	// Create source feed with one route and two trips
	sourceFeed := gtfs.NewFeed()
	sourceFeed.Routes["R1"] = &gtfs.Route{ID: "R1", ShortName: "R1"}
	sourceFeed.Trips["T1"] = &gtfs.Trip{ID: "T1", RouteID: "R1"}
	sourceFeed.Trips["T2"] = &gtfs.Trip{ID: "T2", RouteID: "R1"}
	sourceFeed.Stops["A"] = &gtfs.Stop{ID: "A", Name: "A"}
	sourceFeed.Stops["B"] = &gtfs.Stop{ID: "B", Name: "B"}
	sourceFeed.Stops["C"] = &gtfs.Stop{ID: "C", Name: "C"}
	// Trip 1 serves A, B
	sourceFeed.StopTimes = append(sourceFeed.StopTimes, &gtfs.StopTime{TripID: "T1", StopID: "A", StopSequence: 1})
	sourceFeed.StopTimes = append(sourceFeed.StopTimes, &gtfs.StopTime{TripID: "T1", StopID: "B", StopSequence: 2})
	// Trip 2 serves B, C (so route serves A, B, C total)
	sourceFeed.StopTimes = append(sourceFeed.StopTimes, &gtfs.StopTime{TripID: "T2", StopID: "B", StopSequence: 1})
	sourceFeed.StopTimes = append(sourceFeed.StopTimes, &gtfs.StopTime{TripID: "T2", StopID: "C", StopSequence: 2})

	// Target route serves A, B, C
	targetFeed := createTestFeedWithRouteStops("R2", "T3", []string{"A", "B", "C"})

	sourceRoute := sourceFeed.Routes["R1"]
	targetRoute := targetFeed.Routes["R2"]

	ctx := strategy.NewMergeContext(sourceFeed, targetFeed, "")
	score := scorer.Score(ctx, sourceRoute, targetRoute)

	// Both routes have exactly the same stops {A, B, C}
	if score != 1.0 {
		t.Errorf("Expected score 1.0 when stops from multiple trips match, got %f", score)
	}
}
