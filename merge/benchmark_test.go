package merge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// BenchmarkMergeTwoFeeds benchmarks merging two simple feeds
func BenchmarkMergeTwoFeeds(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_b"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New()
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeTwoFeedsWithIdentity benchmarks merging with identity detection
func BenchmarkMergeTwoFeedsWithIdentity(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	// Using overlap feed which has IDs that overlap with simple_a
	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "overlap"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New(WithDefaultDetection(strategy.DetectionIdentity))
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeTwoFeedsWithFuzzy benchmarks merging with fuzzy detection
func BenchmarkMergeTwoFeedsWithFuzzy(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "fuzzy_similar"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New(WithDefaultDetection(strategy.DetectionFuzzy))
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeThreeFeeds benchmarks merging three feeds
func BenchmarkMergeThreeFeeds(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_b"))
	if err != nil {
		b.Fatal(err)
	}

	feedC, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "minimal"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New()
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB, feedC})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeFiles benchmarks the complete file-based merge workflow
func BenchmarkMergeFiles(b *testing.B) {
	inputA := filepath.Join("..", "testdata", "simple_a")
	inputB := filepath.Join("..", "testdata", "simple_b")

	tmpDir, err := os.MkdirTemp("", "bench_merge")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "merged.zip")
		merger := New()
		if err := merger.MergeFiles([]string{inputA, inputB}, outputPath); err != nil {
			b.Fatal(err)
		}
		_ = os.Remove(outputPath)
	}
}

// BenchmarkLargeFeedMerge benchmarks merging feeds with larger IDs
func BenchmarkLargeFeedMerge(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "large_ids_feed"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New()
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeWithAllOptionalFiles benchmarks merging feeds with all optional files
func BenchmarkMergeWithAllOptionalFiles(b *testing.B) {
	feedA, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "all_optional_feed"))
	if err != nil {
		b.Fatal(err)
	}

	feedB, err := gtfs.ReadFromPath(filepath.Join("..", "testdata", "simple_a"))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger := New()
		_, err := merger.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			b.Fatal(err)
		}
	}
}
