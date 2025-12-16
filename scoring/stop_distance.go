package scoring

import (
	"math"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// StopDistanceScorer scores stops by geographic proximity using great-circle distance.
// Uses hardcoded tiered thresholds (not configurable):
//   - < 50m  → 1.0
//   - < 100m → 0.75
//   - < 500m → 0.5
//   - >= 500m → 0.0
type StopDistanceScorer struct{}

// Score returns a similarity score based on geographic distance.
func (s *StopDistanceScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Stop) float64 {
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
