package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// RouteMergeStrategy handles merging of routes between feeds
type RouteMergeStrategy struct {
	BaseStrategy
	// FuzzyThreshold is the minimum score for a fuzzy match (default 0.5)
	FuzzyThreshold float64
}

// NewRouteMergeStrategy creates a new RouteMergeStrategy
func NewRouteMergeStrategy() *RouteMergeStrategy {
	return &RouteMergeStrategy{
		BaseStrategy:   NewBaseStrategy("route"),
		FuzzyThreshold: 0.5,
	}
}

// Merge performs the merge operation for routes
func (s *RouteMergeStrategy) Merge(ctx *MergeContext) error {
	for _, route := range ctx.Source.Routes {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Routes[route.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.RouteIDMapping[route.ID] = existing.ID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate route detected with ID %q (keeping existing)", route.ID)
				case LogError:
					return fmt.Errorf("duplicate route detected with ID %q", route.ID)
				}

				// Skip adding this route - use the existing one
				continue
			}
		}

		// Check for fuzzy duplicates
		if s.DuplicateDetection == DetectionFuzzy {
			if matchID := s.findFuzzyMatch(ctx, route); matchID != "" {
				// Fuzzy duplicate detected - map source ID to existing target ID
				ctx.RouteIDMapping[route.ID] = matchID

				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Fuzzy duplicate route detected: %q matches %q (keeping existing)", route.ID, matchID)
				case LogError:
					return fmt.Errorf("fuzzy duplicate route detected: %q matches %q", route.ID, matchID)
				}

				// Skip adding this route - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := route.ID
		if _, exists := ctx.Target.Routes[route.ID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.RouteID(ctx.Prefix + string(route.ID))
		}
		ctx.RouteIDMapping[route.ID] = newID

		// Map agency reference
		agencyID := route.AgencyID
		if agencyID != "" {
			if mappedAgency, ok := ctx.AgencyIDMapping[agencyID]; ok {
				agencyID = mappedAgency
			}
		}

		newRoute := &gtfs.Route{
			ID:                newID,
			AgencyID:          agencyID,
			ShortName:         route.ShortName,
			LongName:          route.LongName,
			Desc:              route.Desc,
			Type:              route.Type,
			URL:               route.URL,
			Color:             route.Color,
			TextColor:         route.TextColor,
			SortOrder:         route.SortOrder,
			ContinuousPickup:  route.ContinuousPickup,
			ContinuousDropOff: route.ContinuousDropOff,
		}
		ctx.Target.Routes[newID] = newRoute
	}

	return nil
}

// findFuzzyMatch searches for a fuzzy duplicate in the target routes.
// Returns the ID of the matching route if found, or empty string if no match.
// Uses agency, short_name, long_name matching combined with shared stops (multiplicative scoring).
func (s *RouteMergeStrategy) findFuzzyMatch(ctx *MergeContext, source *gtfs.Route) gtfs.RouteID {
	var bestMatch gtfs.RouteID
	var bestScore float64

	for _, target := range ctx.Target.Routes {
		// Calculate combined score: agency * shortName * longName * stopsInCommon
		agencyScore := routeAgencyScore(ctx, source, target)
		shortNameScore := routePropertyScore(source.ShortName, target.ShortName)
		longNameScore := routePropertyScore(source.LongName, target.LongName)
		stopsScore := routeStopsInCommonScore(ctx, source.ID, target.ID)

		// Multiplicative scoring - any 0 fails the match
		score := agencyScore * shortNameScore * longNameScore * stopsScore

		if score >= s.FuzzyThreshold && score > bestScore {
			bestScore = score
			bestMatch = target.ID
		}
	}

	return bestMatch
}

// routeAgencyScore returns 1.0 if agencies match (considering mappings), 0.0 otherwise.
// Also returns 1.0 if either agency is empty (not comparable).
func routeAgencyScore(ctx *MergeContext, source, target *gtfs.Route) float64 {
	if source.AgencyID == "" || target.AgencyID == "" {
		return 1.0 // Can't compare - neutral score
	}

	// Get mapped agency ID for source
	mappedSourceAgency := source.AgencyID
	if mapped, ok := ctx.AgencyIDMapping[source.AgencyID]; ok {
		mappedSourceAgency = mapped
	}

	if mappedSourceAgency == target.AgencyID {
		return 1.0
	}
	return 0.0
}

// routePropertyScore returns 1.0 if strings match, 0.0 otherwise.
// Empty strings are treated as a match (neutral).
func routePropertyScore(source, target string) float64 {
	if source == "" || target == "" {
		return 1.0 // Can't compare - neutral score
	}
	if source == target {
		return 1.0
	}
	return 0.0
}

// routeStopsInCommonScore returns the element overlap score for stops served by two routes.
func routeStopsInCommonScore(ctx *MergeContext, sourceRouteID, targetRouteID gtfs.RouteID) float64 {
	sourceStops := getStopsForRoute(ctx.Source, sourceRouteID)
	targetStops := getStopsForRoute(ctx.Target, targetRouteID)

	return elementOverlapScore(sourceStops, targetStops)
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

// elementOverlapScore calculates the overlap score between two sets.
// Formula: (common_count / a.size + common_count / b.size) / 2
// Returns 0.0 if either collection is empty.
func elementOverlapScore[T comparable](a, b []T) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Build a set from b for O(1) lookups
	bSet := make(map[T]struct{}, len(b))
	for _, item := range b {
		bSet[item] = struct{}{}
	}

	// Count common elements
	common := 0
	for _, item := range a {
		if _, ok := bSet[item]; ok {
			common++
		}
	}

	// Apply formula: (common/a + common/b) / 2
	scoreA := float64(common) / float64(len(a))
	scoreB := float64(common) / float64(len(b))
	return (scoreA + scoreB) / 2.0
}
