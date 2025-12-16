package gtfs

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkReadFeed benchmarks reading a GTFS feed from a directory
func BenchmarkReadFeed(b *testing.B) {
	feedPath := filepath.Join("..", "testdata", "simple_a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadFromPath(feedPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReadFeedFromZip benchmarks reading a GTFS feed from a zip file
func BenchmarkReadFeedFromZip(b *testing.B) {
	// Create a temporary zip file from simple_a
	feedPath := filepath.Join("..", "testdata", "simple_a")
	feed, err := ReadFromPath(feedPath)
	if err != nil {
		b.Fatal(err)
	}

	tmpFile, err := os.CreateTemp("", "bench_feed_*.zip")
	if err != nil {
		b.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := WriteToPath(feed, tmpPath); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadFromPath(tmpPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWriteFeed benchmarks writing a GTFS feed to a zip file
func BenchmarkWriteFeed(b *testing.B) {
	feedPath := filepath.Join("..", "testdata", "simple_a")
	feed, err := ReadFromPath(feedPath)
	if err != nil {
		b.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "bench_write")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "output.zip")
		if err := WriteToPath(feed, outputPath); err != nil {
			b.Fatal(err)
		}
		_ = os.Remove(outputPath)
	}
}

// BenchmarkReadLargeFeed benchmarks reading a feed with larger IDs
func BenchmarkReadLargeFeed(b *testing.B) {
	feedPath := filepath.Join("..", "testdata", "large_ids_feed")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadFromPath(feedPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParsing benchmarks CSV parsing overhead
func BenchmarkParsing(b *testing.B) {
	feedPath := filepath.Join("..", "testdata", "simple_a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, err := ReadFromPath(feedPath)
		if err != nil {
			b.Fatal(err)
		}
		// Access all entities to ensure parsing is complete
		_ = len(feed.Agencies)
		_ = len(feed.Stops)
		_ = len(feed.Routes)
		_ = len(feed.Trips)
		_ = len(feed.StopTimes)
		_ = len(feed.Calendars)
	}
}
