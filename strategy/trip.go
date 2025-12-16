package strategy

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// TripMergeStrategy handles merging of trips between feeds
type TripMergeStrategy struct {
	BaseStrategy
	// FuzzyThreshold is the minimum score for a fuzzy match (default 0.5)
	FuzzyThreshold float64
}

// NewTripMergeStrategy creates a new TripMergeStrategy
func NewTripMergeStrategy() *TripMergeStrategy {
	return &TripMergeStrategy{
		BaseStrategy:   NewBaseStrategy("trip"),
		FuzzyThreshold: 0.5,
	}
}

// Merge performs the merge operation for trips
func (s *TripMergeStrategy) Merge(ctx *MergeContext) error {
	for _, trip := range ctx.Source.Trips {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Trips[trip.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.TripIDMapping[trip.ID] = existing.ID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate trip detected with ID %q (keeping existing)", trip.ID)
				case LogError:
					return fmt.Errorf("duplicate trip detected with ID %q", trip.ID)
				}

				// Skip adding this trip - use the existing one
				continue
			}
		}

		// Check for fuzzy duplicates
		if s.DuplicateDetection == DetectionFuzzy {
			if matchID := s.findFuzzyMatch(ctx, trip); matchID != "" {
				// Fuzzy duplicate detected - map source ID to existing target ID
				ctx.TripIDMapping[trip.ID] = matchID

				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Fuzzy duplicate trip detected: %q matches %q (keeping existing)", trip.ID, matchID)
				case LogError:
					return fmt.Errorf("fuzzy duplicate trip detected: %q matches %q", trip.ID, matchID)
				}

				// Skip adding this trip - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := trip.ID
		if _, exists := ctx.Target.Trips[trip.ID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.TripID(ctx.Prefix + string(trip.ID))
		}
		ctx.TripIDMapping[trip.ID] = newID

		// Map references
		routeID := trip.RouteID
		if mappedRoute, ok := ctx.RouteIDMapping[routeID]; ok {
			routeID = mappedRoute
		}

		serviceID := trip.ServiceID
		if mappedService, ok := ctx.ServiceIDMapping[serviceID]; ok {
			serviceID = mappedService
		}

		shapeID := trip.ShapeID
		if shapeID != "" {
			if mappedShape, ok := ctx.ShapeIDMapping[shapeID]; ok {
				shapeID = mappedShape
			}
		}

		newTrip := &gtfs.Trip{
			ID:                   newID,
			RouteID:              routeID,
			ServiceID:            serviceID,
			Headsign:             trip.Headsign,
			ShortName:            trip.ShortName,
			DirectionID:          trip.DirectionID,
			BlockID:              trip.BlockID,
			ShapeID:              shapeID,
			WheelchairAccessible: trip.WheelchairAccessible,
			BikesAllowed:         trip.BikesAllowed,
		}
		ctx.Target.Trips[newID] = newTrip
	}

	return nil
}

// findFuzzyMatch searches for a fuzzy duplicate in the target trips.
// Returns the ID of the matching trip if found, or empty string if no match.
// Uses route, service_id, shared stops, and schedule overlap (multiplicative scoring).
// Additionally validates that stop times match exactly.
func (s *TripMergeStrategy) findFuzzyMatch(ctx *MergeContext, source *gtfs.Trip) gtfs.TripID {
	var bestMatch gtfs.TripID
	var bestScore float64

	for _, target := range ctx.Target.Trips {
		// Calculate combined score: route * serviceId * stopsInCommon * scheduleOverlap
		routeScore := tripRouteScore(ctx, source, target)
		serviceScore := tripServiceScore(ctx, source, target)
		stopsScore := tripStopsInCommonScore(ctx, source.ID, target.ID)
		scheduleScore := tripScheduleOverlapScore(ctx, source.ID, target.ID)

		// Multiplicative scoring - any 0 fails the match
		score := routeScore * serviceScore * stopsScore * scheduleScore

		if score >= s.FuzzyThreshold && score > bestScore {
			// Additional validation: check stop times match exactly
			if validateTripStopTimes(ctx, source.ID, target.ID) {
				bestScore = score
				bestMatch = target.ID
			}
		}
	}

	return bestMatch
}

// tripRouteScore returns 1.0 if routes match (considering mappings), 0.0 otherwise.
func tripRouteScore(ctx *MergeContext, source, target *gtfs.Trip) float64 {
	// Get mapped route ID for source
	mappedSourceRoute := source.RouteID
	if mapped, ok := ctx.RouteIDMapping[source.RouteID]; ok {
		mappedSourceRoute = mapped
	}

	if mappedSourceRoute == target.RouteID {
		return 1.0
	}
	return 0.0
}

// tripServiceScore returns 1.0 if service IDs match (considering mappings), 0.0 otherwise.
func tripServiceScore(ctx *MergeContext, source, target *gtfs.Trip) float64 {
	// Get mapped service ID for source
	mappedSourceService := source.ServiceID
	if mapped, ok := ctx.ServiceIDMapping[source.ServiceID]; ok {
		mappedSourceService = mapped
	}

	if mappedSourceService == target.ServiceID {
		return 1.0
	}
	return 0.0
}

// tripStopsInCommonScore returns the element overlap score for stops in two trips.
func tripStopsInCommonScore(ctx *MergeContext, sourceTripID, targetTripID gtfs.TripID) float64 {
	sourceStops := getStopsForTrip(ctx.Source, sourceTripID)
	targetStops := getStopsForTrip(ctx.Target, targetTripID)

	return elementOverlapScore(sourceStops, targetStops)
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

// tripScheduleOverlapScore returns the schedule overlap score for two trips.
func tripScheduleOverlapScore(ctx *MergeContext, sourceTripID, targetTripID gtfs.TripID) float64 {
	sourceStart, sourceEnd := getTripTimeWindow(ctx.Source, sourceTripID)
	targetStart, targetEnd := getTripTimeWindow(ctx.Target, targetTripID)

	return intervalOverlapScore(
		float64(sourceStart), float64(sourceEnd),
		float64(targetStart), float64(targetEnd),
	)
}

// getTripTimeWindow returns the start and end times for a trip in seconds.
// Start is the first stop's departure time, end is the last stop's arrival time.
// Returns (0, 0) if trip has no stop times.
func getTripTimeWindow(feed *gtfs.Feed, tripID gtfs.TripID) (start, end int) {
	var stopTimes []*gtfs.StopTime
	for _, st := range feed.StopTimes {
		if st.TripID == tripID {
			stopTimes = append(stopTimes, st)
		}
	}

	if len(stopTimes) == 0 {
		return 0, 0
	}

	// Find first and last stop by sequence
	var firstStop, lastStop *gtfs.StopTime
	for _, st := range stopTimes {
		if firstStop == nil || st.StopSequence < firstStop.StopSequence {
			firstStop = st
		}
		if lastStop == nil || st.StopSequence > lastStop.StopSequence {
			lastStop = st
		}
	}

	// Use departure time for start, arrival time for end
	start = parseGTFSTime(firstStop.DepartureTime)
	end = parseGTFSTime(lastStop.ArrivalTime)

	return start, end
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS) to seconds since midnight.
// GTFS times can exceed 24:00:00 for overnight trips.
// Returns 0 for invalid or empty times.
func parseGTFSTime(timeStr string) int {
	if timeStr == "" {
		return 0
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}

	return hours*3600 + minutes*60 + seconds
}

// intervalOverlapScore calculates the overlap score between two intervals.
// Formula: (overlap / interval_a_length + overlap / interval_b_length) / 2
// Returns 0.0 if either interval has zero length or there's no overlap.
func intervalOverlapScore(start1, end1, start2, end2 float64) float64 {
	len1 := end1 - start1
	len2 := end2 - start2

	if len1 <= 0 || len2 <= 0 {
		return 0.0
	}

	// Calculate overlap
	overlapStart := max(start1, start2)
	overlapEnd := min(end1, end2)
	overlap := overlapEnd - overlapStart

	if overlap <= 0 {
		return 0.0
	}

	// Apply formula: (overlap/len1 + overlap/len2) / 2
	scoreA := overlap / len1
	scoreB := overlap / len2
	return (scoreA + scoreB) / 2.0
}

// validateTripStopTimes checks if two trips have matching stop times.
// Returns false if:
// - Stop count differs
// - Any stop at the same sequence position differs
// - Any arrival or departure time differs
func validateTripStopTimes(ctx *MergeContext, sourceTripID, targetTripID gtfs.TripID) bool {
	sourceStopTimes := getStopTimesForTrip(ctx.Source, sourceTripID)
	targetStopTimes := getStopTimesForTrip(ctx.Target, targetTripID)

	// Check stop count matches
	if len(sourceStopTimes) != len(targetStopTimes) {
		return false
	}

	// Sort both by stop sequence
	sort.Slice(sourceStopTimes, func(i, j int) bool {
		return sourceStopTimes[i].StopSequence < sourceStopTimes[j].StopSequence
	})
	sort.Slice(targetStopTimes, func(i, j int) bool {
		return targetStopTimes[i].StopSequence < targetStopTimes[j].StopSequence
	})

	// Check each stop time matches
	for i := range sourceStopTimes {
		src := sourceStopTimes[i]
		tgt := targetStopTimes[i]

		// Get mapped stop ID for source
		mappedSourceStop := src.StopID
		if mapped, ok := ctx.StopIDMapping[src.StopID]; ok {
			mappedSourceStop = mapped
		}

		// Check stop matches
		if mappedSourceStop != tgt.StopID {
			return false
		}

		// Check times match exactly
		if src.ArrivalTime != tgt.ArrivalTime {
			return false
		}
		if src.DepartureTime != tgt.DepartureTime {
			return false
		}
	}

	return true
}

// getStopTimesForTrip returns all stop times for a trip.
func getStopTimesForTrip(feed *gtfs.Feed, tripID gtfs.TripID) []*gtfs.StopTime {
	var stopTimes []*gtfs.StopTime
	for _, st := range feed.StopTimes {
		if st.TripID == tripID {
			stopTimes = append(stopTimes, st)
		}
	}
	return stopTimes
}
