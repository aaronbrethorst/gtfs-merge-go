package scoring

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// TripStopsInCommonScorer scores trips by shared stops.
// Uses the ElementOverlapScore formula for calculating overlap.
type TripStopsInCommonScorer struct{}

// Score returns a similarity score based on shared stops between trips.
// Collects all stops for each trip and calculates overlap.
func (s *TripStopsInCommonScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Trip) float64 {
	sourceStops := getStopsForTrip(ctx.Source, source.ID)
	targetStops := getStopsForTrip(ctx.Target, target.ID)

	return ElementOverlapScore(sourceStops, targetStops)
}

// getStopsForTrip returns all unique stop IDs for a trip.
func getStopsForTrip(feed *gtfs.Feed, tripID gtfs.TripID) []gtfs.StopID {
	stopSet := make(map[gtfs.StopID]struct{})

	for _, st := range feed.StopTimes {
		if st.TripID == tripID {
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
