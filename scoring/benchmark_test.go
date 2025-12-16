package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// BenchmarkStopDistanceScorer benchmarks the geographic distance scoring
func BenchmarkStopDistanceScorer(b *testing.B) {
	scorer := &StopDistanceScorer{}

	source := &gtfs.Stop{
		ID:   "source",
		Name: "Source Stop",
		Lat:  47.6062,
		Lon:  -122.3321,
	}

	target := &gtfs.Stop{
		ID:   "target",
		Name: "Target Stop",
		Lat:  47.6072,
		Lon:  -122.3331,
	}

	ctx := &strategy.MergeContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(ctx, source, target)
	}
}

// BenchmarkHaversineDistance benchmarks the Haversine distance calculation
func BenchmarkHaversineDistance(b *testing.B) {
	// Seattle coordinates
	lat1, lon1 := 47.6062, -122.3321
	// Nearby location
	lat2, lon2 := 47.6072, -122.3331

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = haversineDistance(lat1, lon1, lat2, lon2)
	}
}

// BenchmarkPropertyMatcher benchmarks property matching
func BenchmarkPropertyMatcher(b *testing.B) {
	type testEntity struct {
		Name     string
		Value    string
		Category string
	}

	matcher := &PropertyMatcher[testEntity]{
		Properties: []func(testEntity) string{
			func(e testEntity) string { return e.Name },
			func(e testEntity) string { return e.Value },
			func(e testEntity) string { return e.Category },
		},
	}

	source := testEntity{Name: "Test", Value: "123", Category: "A"}
	target := testEntity{Name: "Test", Value: "123", Category: "A"}

	ctx := &strategy.MergeContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matcher.Score(ctx, source, target)
	}
}

// BenchmarkAndScorer benchmarks the AndScorer combining multiple scorers
func BenchmarkAndScorer(b *testing.B) {
	type testEntity struct {
		Name  string
		Value int
	}

	andScorer := &AndScorer[testEntity]{
		Scorers: []Scorer[testEntity]{
			&PropertyMatcher[testEntity]{
				Properties: []func(testEntity) string{
					func(e testEntity) string { return e.Name },
				},
			},
			&PropertyMatcher[testEntity]{
				Properties: []func(testEntity) string{
					func(e testEntity) string { return e.Name },
				},
			},
		},
		Threshold: 0.5,
	}

	source := testEntity{Name: "Test", Value: 1}
	target := testEntity{Name: "Test", Value: 2}

	ctx := &strategy.MergeContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = andScorer.Score(ctx, source, target)
	}
}

// BenchmarkElementOverlapScore benchmarks the set overlap calculation
func BenchmarkElementOverlapScore(b *testing.B) {
	setA := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	setB := []string{"c", "d", "e", "f", "g", "k", "l", "m", "n", "o"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ElementOverlapScore(setA, setB)
	}
}

// BenchmarkIntervalOverlapScore benchmarks the interval overlap calculation
func BenchmarkIntervalOverlapScore(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IntervalOverlapScore(10.0, 50.0, 30.0, 70.0)
	}
}

// BenchmarkServiceDateOverlapScorer benchmarks calendar date overlap scoring
func BenchmarkServiceDateOverlapScorer(b *testing.B) {
	scorer := &ServiceDateOverlapScorer{}

	// Create feeds with calendars
	source := gtfs.NewFeed()
	source.Calendars["SVC1"] = &gtfs.Calendar{
		ServiceID: "SVC1",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		StartDate: "20240101",
		EndDate:   "20240630",
	}

	target := gtfs.NewFeed()
	target.Calendars["SVC2"] = &gtfs.Calendar{
		ServiceID: "SVC2",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		StartDate: "20240301",
		EndDate:   "20240930",
	}

	ctx := &strategy.MergeContext{
		Source: source,
		Target: target,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(ctx, gtfs.ServiceID("SVC1"), gtfs.ServiceID("SVC2"))
	}
}
