package strategy

import (
	"path/filepath"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// BenchmarkFuzzyScoring benchmarks the fuzzy duplicate detection for stops
func BenchmarkFuzzyScoring(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "fuzzy_similar"))
	if err != nil {
		b.Fatal(err)
	}

	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewMergeContext(feedB, feedA, "")
		if err := strategy.Merge(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFuzzyScoringStops benchmarks stop fuzzy matching specifically
func BenchmarkFuzzyScoringStops(b *testing.B) {
	// Create feeds with many stops for testing
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	// Add 100 stops to target
	for i := 0; i < 100; i++ {
		target.Stops[gtfs.StopID(string(rune('A'+i%26))+string(rune('0'+i/26)))] = &gtfs.Stop{
			ID:   gtfs.StopID(string(rune('A'+i%26)) + string(rune('0'+i/26))),
			Name: "Stop " + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			Lat:  47.6062 + float64(i)*0.001,
			Lon:  -122.3321 + float64(i)*0.001,
		}
	}

	// Add 50 stops to source, some similar
	for i := 0; i < 50; i++ {
		source.Stops[gtfs.StopID("src_"+string(rune('0'+i/10))+string(rune('0'+i%10)))] = &gtfs.Stop{
			ID:   gtfs.StopID("src_" + string(rune('0'+i/10)) + string(rune('0'+i%10))),
			Name: "Stop " + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			Lat:  47.6062 + float64(i)*0.001,
			Lon:  -122.3321 + float64(i)*0.001,
		}
	}

	// Add required entities for a valid feed
	source.Agencies["agency"] = &gtfs.Agency{ID: "agency", Name: "Test", URL: "http://test.com", Timezone: "America/Los_Angeles"}
	source.Routes["route"] = &gtfs.Route{ID: "route", AgencyID: "agency", ShortName: "R1", Type: 3}
	source.Trips["trip"] = &gtfs.Trip{ID: "trip", RouteID: "route", ServiceID: "svc"}
	source.StopTimes = append(source.StopTimes, &gtfs.StopTime{TripID: "trip", StopID: "src_00", StopSequence: 1})
	source.Calendars["svc"] = &gtfs.Calendar{ServiceID: "svc", StartDate: "20240101", EndDate: "20241231"}

	target.Agencies["agency"] = &gtfs.Agency{ID: "agency", Name: "Test", URL: "http://test.com", Timezone: "America/Los_Angeles"}
	target.Routes["route"] = &gtfs.Route{ID: "route", AgencyID: "agency", ShortName: "R1", Type: 3}
	target.Trips["trip"] = &gtfs.Trip{ID: "trip", RouteID: "route", ServiceID: "svc"}
	target.StopTimes = append(target.StopTimes, &gtfs.StopTime{TripID: "trip", StopID: "A0", StopSequence: 1})
	target.Calendars["svc"] = &gtfs.Calendar{ServiceID: "svc", StartDate: "20240101", EndDate: "20241231"}

	strategy := NewStopMergeStrategy()
	strategy.SetDuplicateDetection(DetectionFuzzy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewMergeContext(source, target, "")
		if err := strategy.Merge(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFuzzyScoringRoutes benchmarks route fuzzy matching
func BenchmarkFuzzyScoringRoutes(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "fuzzy_similar"))
	if err != nil {
		b.Fatal(err)
	}

	// First merge stops so we have mappings
	stopStrategy := NewStopMergeStrategy()
	stopStrategy.SetDuplicateDetection(DetectionFuzzy)

	ctx := NewMergeContext(feedB, feedA, "")
	if err := stopStrategy.Merge(ctx); err != nil {
		b.Fatal(err)
	}

	routeStrategy := NewRouteMergeStrategy()
	routeStrategy.SetDuplicateDetection(DetectionFuzzy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewMergeContext(feedB, feedA, "")
		if err := routeStrategy.Merge(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFuzzyScoringTrips benchmarks trip fuzzy matching
func BenchmarkFuzzyScoringTrips(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "fuzzy_similar"))
	if err != nil {
		b.Fatal(err)
	}

	// First merge dependencies
	stopStrategy := NewStopMergeStrategy()
	stopStrategy.SetDuplicateDetection(DetectionFuzzy)
	ctx := NewMergeContext(feedB, feedA, "")
	if err := stopStrategy.Merge(ctx); err != nil {
		b.Fatal(err)
	}

	calendarStrategy := NewCalendarMergeStrategy()
	calendarStrategy.SetDuplicateDetection(DetectionFuzzy)
	if err := calendarStrategy.Merge(ctx); err != nil {
		b.Fatal(err)
	}

	routeStrategy := NewRouteMergeStrategy()
	routeStrategy.SetDuplicateDetection(DetectionFuzzy)
	if err := routeStrategy.Merge(ctx); err != nil {
		b.Fatal(err)
	}

	tripStrategy := NewTripMergeStrategy()
	tripStrategy.SetDuplicateDetection(DetectionFuzzy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewMergeContext(feedB, feedA, "")
		if err := tripStrategy.Merge(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIdentityDetection benchmarks identity duplicate detection
func BenchmarkIdentityDetection(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "overlap"))
	if err != nil {
		b.Fatal(err)
	}

	stopStrategy := NewStopMergeStrategy()
	stopStrategy.SetDuplicateDetection(DetectionIdentity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewMergeContext(feedB, feedA, "")
		if err := stopStrategy.Merge(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAutoDetect benchmarks the auto-detection of duplicate strategy
func BenchmarkAutoDetect(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "overlap"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AutoDetectDuplicateDetection(feedA, feedB)
	}
}

// BenchmarkAutoDetectWithConfig benchmarks the auto-detection with custom config
func BenchmarkAutoDetectWithConfig(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "overlap"))
	if err != nil {
		b.Fatal(err)
	}

	config := DefaultAutoDetectConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AutoDetectDuplicateDetectionWithConfig(feedA, feedB, config)
	}
}
