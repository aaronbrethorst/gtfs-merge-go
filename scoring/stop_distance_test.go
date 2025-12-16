package scoring

import (
	"math"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// TestStopDistanceSameLocation verifies score is 1.0 for same coordinates
func TestStopDistanceSameLocation(t *testing.T) {
	scorer := &StopDistanceScorer{}

	source := &gtfs.Stop{
		ID:   "stop1",
		Name: "Test Stop",
		Lat:  47.6062,
		Lon:  -122.3321,
	}
	target := &gtfs.Stop{
		ID:   "stop2",
		Name: "Same Location",
		Lat:  47.6062,
		Lon:  -122.3321,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 for same location, got %f", score)
	}
}

// TestStopDistanceWithin50m verifies score is 1.0 for stops within 50m
func TestStopDistanceWithin50m(t *testing.T) {
	scorer := &StopDistanceScorer{}

	// Two points approximately 30m apart (less than 50m)
	// Using Seattle area coordinates
	source := &gtfs.Stop{
		ID:   "stop1",
		Name: "Stop A",
		Lat:  47.6062,
		Lon:  -122.3321,
	}
	target := &gtfs.Stop{
		ID:   "stop2",
		Name: "Stop B",
		Lat:  47.6064, // ~22m north
		Lon:  -122.3321,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 for stops within 50m, got %f", score)
	}
}

// TestStopDistanceWithin100m verifies score is 0.75 for stops between 50-100m
func TestStopDistanceWithin100m(t *testing.T) {
	scorer := &StopDistanceScorer{}

	// Two points approximately 75m apart (between 50m and 100m)
	source := &gtfs.Stop{
		ID:   "stop1",
		Name: "Stop A",
		Lat:  47.6062,
		Lon:  -122.3321,
	}
	target := &gtfs.Stop{
		ID:   "stop2",
		Name: "Stop B",
		Lat:  47.6069, // ~77m north
		Lon:  -122.3321,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 0.75 {
		t.Errorf("Expected score 0.75 for stops between 50-100m, got %f", score)
	}
}

// TestStopDistanceWithin500m verifies score is 0.5 for stops between 100-500m
func TestStopDistanceWithin500m(t *testing.T) {
	scorer := &StopDistanceScorer{}

	// Two points approximately 300m apart (between 100m and 500m)
	source := &gtfs.Stop{
		ID:   "stop1",
		Name: "Stop A",
		Lat:  47.6062,
		Lon:  -122.3321,
	}
	target := &gtfs.Stop{
		ID:   "stop2",
		Name: "Stop B",
		Lat:  47.6089, // ~300m north
		Lon:  -122.3321,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 0.5 {
		t.Errorf("Expected score 0.5 for stops between 100-500m, got %f", score)
	}
}

// TestStopDistanceBeyondThreshold verifies score is 0 beyond 500m
func TestStopDistanceBeyondThreshold(t *testing.T) {
	scorer := &StopDistanceScorer{}

	// Two points approximately 1km apart (beyond 500m)
	source := &gtfs.Stop{
		ID:   "stop1",
		Name: "Stop A",
		Lat:  47.6062,
		Lon:  -122.3321,
	}
	target := &gtfs.Stop{
		ID:   "stop2",
		Name: "Stop B",
		Lat:  47.6162, // ~1.1km north
		Lon:  -122.3321,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 for stops beyond 500m, got %f", score)
	}
}

// TestStopDistanceHaversine verifies correct distance calculation using haversine formula
func TestStopDistanceHaversine(t *testing.T) {
	// Test known distance: New York to Los Angeles is approximately 3944 km
	lat1, lon1 := 40.7128, -74.0060  // New York
	lat2, lon2 := 34.0522, -118.2437 // Los Angeles

	distance := haversineDistance(lat1, lon1, lat2, lon2)

	// Should be approximately 3944 km, allow 50km margin
	expectedMin := 3900.0
	expectedMax := 4000.0
	if distance < expectedMin || distance > expectedMax {
		t.Errorf("Haversine distance between NY and LA should be ~3944km, got %f km", distance)
	}
}

// TestStopDistanceHaversineSmall verifies haversine for small distances
func TestStopDistanceHaversineSmall(t *testing.T) {
	// Test a small known distance in Seattle
	// From Space Needle to Pike Place Market is approximately 1.3 km
	lat1, lon1 := 47.6205, -122.3493 // Space Needle
	lat2, lon2 := 47.6097, -122.3422 // Pike Place Market

	distance := haversineDistance(lat1, lon1, lat2, lon2)

	// Should be approximately 1.3 km, allow 0.2km margin
	expectedMin := 1.1
	expectedMax := 1.5
	if distance < expectedMin || distance > expectedMax {
		t.Errorf("Haversine distance should be ~1.3km, got %f km", distance)
	}
}

// TestStopDistanceZeroCoordinates verifies handling of zero coordinates
func TestStopDistanceZeroCoordinates(t *testing.T) {
	scorer := &StopDistanceScorer{}

	source := &gtfs.Stop{
		ID:  "stop1",
		Lat: 0.0,
		Lon: 0.0,
	}
	target := &gtfs.Stop{
		ID:  "stop2",
		Lat: 0.0,
		Lon: 0.0,
	}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	// Same location (even at 0,0) should be 1.0
	if score != 1.0 {
		t.Errorf("Expected score 1.0 for same zero coordinates, got %f", score)
	}
}

// TestHaversineDistanceMeters verifies the meters conversion helper
func TestHaversineDistanceMeters(t *testing.T) {
	// Two points approximately 100m apart
	lat1, lon1 := 47.6062, -122.3321
	lat2, lon2 := 47.6071, -122.3321 // ~100m north

	distanceKm := haversineDistance(lat1, lon1, lat2, lon2)
	distanceM := distanceKm * 1000

	// Should be approximately 100m, allow 10m margin
	if math.Abs(distanceM-100) > 10 {
		t.Errorf("Expected ~100m, got %f meters", distanceM)
	}
}
