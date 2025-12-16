package strategy

import (
	"fmt"
	"log"
	"math"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// StopMergeStrategy handles merging of stops between feeds
type StopMergeStrategy struct {
	BaseStrategy
	// FuzzyThreshold is the minimum score for a fuzzy match (default 0.5)
	FuzzyThreshold float64
	// Concurrent controls concurrent processing for fuzzy matching
	Concurrent ConcurrentConfig
}

// NewStopMergeStrategy creates a new StopMergeStrategy
func NewStopMergeStrategy() *StopMergeStrategy {
	return &StopMergeStrategy{
		BaseStrategy:   NewBaseStrategy("stop"),
		FuzzyThreshold: 0.5,
		Concurrent:     DefaultConcurrentConfig(),
	}
}

// SetConcurrent enables or disables concurrent fuzzy matching
func (s *StopMergeStrategy) SetConcurrent(enabled bool) {
	s.Concurrent.Enabled = enabled
}

// SetConcurrentWorkers sets the number of worker goroutines for concurrent processing
func (s *StopMergeStrategy) SetConcurrentWorkers(n int) {
	if n > 0 {
		s.Concurrent.NumWorkers = n
	}
}

// Merge performs the merge operation for stops
func (s *StopMergeStrategy) Merge(ctx *MergeContext) error {
	for _, stop := range ctx.Source.Stops {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Stops[stop.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.StopIDMapping[stop.ID] = existing.ID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate stop detected with ID %q (keeping existing)", stop.ID)
				case LogError:
					return fmt.Errorf("duplicate stop detected with ID %q", stop.ID)
				}

				// Skip adding this stop - use the existing one
				continue
			}
		}

		// Check for fuzzy duplicates
		if s.DuplicateDetection == DetectionFuzzy {
			if matchID := s.findFuzzyMatch(ctx, stop); matchID != "" {
				// Fuzzy duplicate detected - map source ID to existing target ID
				ctx.StopIDMapping[stop.ID] = matchID

				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Fuzzy duplicate stop detected: %q matches %q (keeping existing)", stop.ID, matchID)
				case LogError:
					return fmt.Errorf("fuzzy duplicate stop detected: %q matches %q", stop.ID, matchID)
				}

				// Skip adding this stop - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := stop.ID
		if _, exists := ctx.Target.Stops[stop.ID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.StopID(ctx.Prefix + string(stop.ID))
		}
		ctx.StopIDMapping[stop.ID] = newID

		// Handle parent_station reference
		parentStation := stop.ParentStation
		if parentStation != "" {
			if mappedParent, ok := ctx.StopIDMapping[parentStation]; ok {
				parentStation = mappedParent
			} else if _, exists := ctx.Target.Stops[parentStation]; exists {
				// Parent exists in target with collision - would have been prefixed
				parentStation = gtfs.StopID(ctx.Prefix + string(parentStation))
			}
			// Otherwise keep as-is (no collision)
		}

		newStop := &gtfs.Stop{
			ID:                 newID,
			Code:               stop.Code,
			Name:               stop.Name,
			Desc:               stop.Desc,
			Lat:                stop.Lat,
			Lon:                stop.Lon,
			ZoneID:             stop.ZoneID,
			URL:                stop.URL,
			LocationType:       stop.LocationType,
			ParentStation:      parentStation,
			Timezone:           stop.Timezone,
			WheelchairBoarding: stop.WheelchairBoarding,
			LevelID:            stop.LevelID,
			PlatformCode:       stop.PlatformCode,
		}
		ctx.Target.Stops[newID] = newStop
	}

	return nil
}

// findFuzzyMatch searches for a fuzzy duplicate in the target stops.
// Returns the ID of the matching stop if found, or empty string if no match.
// Uses name matching combined with geographic distance (multiplicative scoring).
// Supports concurrent processing when enabled.
func (s *StopMergeStrategy) findFuzzyMatch(ctx *MergeContext, source *gtfs.Stop) gtfs.StopID {
	// Convert map to slice for concurrent processing
	targets := make([]*gtfs.Stop, 0, len(ctx.Target.Stops))
	for _, stop := range ctx.Target.Stops {
		targets = append(targets, stop)
	}

	// Use concurrent matching if enabled and enough items
	if s.Concurrent.Enabled && len(targets) >= s.Concurrent.MinItemsForConcurrency {
		return findBestMatchConcurrent(
			targets,
			func(stop *gtfs.Stop) gtfs.StopID { return stop.ID },
			func(target *gtfs.Stop) float64 {
				nameScore := stopNameScore(source, target)
				distScore := stopDistanceScore(source, target)
				return nameScore * distScore
			},
			s.FuzzyThreshold,
			s.Concurrent,
		)
	}

	// Sequential matching (default)
	var bestMatch gtfs.StopID
	var bestScore float64

	for _, target := range targets {
		// Calculate combined score: name match * distance score
		nameScore := stopNameScore(source, target)
		distScore := stopDistanceScore(source, target)
		score := nameScore * distScore

		if score >= s.FuzzyThreshold && score > bestScore {
			bestScore = score
			bestMatch = target.ID
		}
	}

	return bestMatch
}

// stopNameScore returns 1.0 if names match, 0.0 otherwise.
func stopNameScore(source, target *gtfs.Stop) float64 {
	if source.Name == target.Name {
		return 1.0
	}
	return 0.0
}

// stopDistanceScore returns a score based on geographic distance.
// Uses tiered thresholds: <50m → 1.0, <100m → 0.75, <500m → 0.5, >=500m → 0.0
func stopDistanceScore(source, target *gtfs.Stop) float64 {
	distanceKm := haversineDistance(source.Lat, source.Lon, target.Lat, target.Lon)
	distanceM := distanceKm * 1000

	switch {
	case distanceM < 50:
		return 1.0
	case distanceM < 100:
		return 0.75
	case distanceM < 500:
		return 0.5
	default:
		return 0.0
	}
}

// earthRadiusKm is the mean radius of the Earth in kilometers
const earthRadiusKm = 6371.0

// haversineDistance calculates the great-circle distance between two points
// on the Earth's surface using the Haversine formula.
// Returns distance in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
