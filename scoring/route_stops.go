package scoring

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// RouteStopsInCommonScorer scores routes by shared stops.
// Uses the ElementOverlapScore formula for calculating overlap.
type RouteStopsInCommonScorer struct{}

// Score returns a similarity score based on shared stops between routes.
// Collects all stops served by each route (across all trips) and calculates overlap.
func (r *RouteStopsInCommonScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Route) float64 {
	sourceStops := getStopsForRoute(ctx.Source, source.ID)
	targetStops := getStopsForRoute(ctx.Target, target.ID)

	return ElementOverlapScore(sourceStops, targetStops)
}

// getStopsForRoute returns all unique stop IDs served by a route's trips.
func getStopsForRoute(feed *gtfs.Feed, routeID gtfs.RouteID) []gtfs.StopID {
	// Find all trips for this route
	tripIDs := make(map[gtfs.TripID]struct{})
	for tripID, trip := range feed.Trips {
		if trip.RouteID == routeID {
			tripIDs[tripID] = struct{}{}
		}
	}

	// Collect all unique stops from those trips
	stopSet := make(map[gtfs.StopID]struct{})
	for _, st := range feed.StopTimes {
		if _, ok := tripIDs[st.TripID]; ok {
			stopSet[st.StopID] = struct{}{}
		}
	}

	// Convert to slice
	stops := make([]gtfs.StopID, 0, len(stopSet))
	for stopID := range stopSet {
		stops = append(stops, stopID)
	}

	return stops
}
