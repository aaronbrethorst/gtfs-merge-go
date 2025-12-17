//go:build java

package compare

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/merge"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// realWorldScenario defines a test scenario for real-world GTFS feeds
type realWorldScenario struct {
	name      string
	feeds     []string
	detection string
}

// realWorldScenarios defines all test scenarios to run
// 12 total: 4 feed combinations × 3 detection modes
var realWorldScenarios = []realWorldScenario{
	// 2-feed combinations: Pierce + Intercity
	{"pierce_intercity_none", []string{"pierce_transit", "intercity_transit"}, "none"},
	{"pierce_intercity_identity", []string{"pierce_transit", "intercity_transit"}, "identity"},
	{"pierce_intercity_fuzzy", []string{"pierce_transit", "intercity_transit"}, "fuzzy"},

	// 2-feed combinations: Community + Everett
	{"community_everett_none", []string{"community_transit", "everett_transit"}, "none"},
	{"community_everett_identity", []string{"community_transit", "everett_transit"}, "identity"},
	{"community_everett_fuzzy", []string{"community_transit", "everett_transit"}, "fuzzy"},

	// 3-feed merge
	{"three_feed_none", []string{"pierce_transit", "intercity_transit", "community_transit"}, "none"},
	{"three_feed_identity", []string{"pierce_transit", "intercity_transit", "community_transit"}, "identity"},
	{"three_feed_fuzzy", []string{"pierce_transit", "intercity_transit", "community_transit"}, "fuzzy"},

	// 4-feed merge (all feeds)
	{"four_feed_none", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "none"},
	{"four_feed_identity", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "identity"},
	{"four_feed_fuzzy", []string{"pierce_transit", "intercity_transit", "community_transit", "everett_transit"}, "fuzzy"},
}

// TestRealWorld_JavaGoComparison runs all real-world GTFS merge scenarios
// comparing Java CLI output against Go CLI output
func TestRealWorld_JavaGoComparison(t *testing.T) {
	jarPath := skipIfNoJava(t)
	feedPaths := skipIfNoRealWorldFeeds(t)

	for _, scenario := range realWorldScenarios {
		scenario := scenario // capture range variable for parallel execution
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			// Resolve feed paths
			inputs := make([]string, len(scenario.feeds))
			for i, name := range scenario.feeds {
				inputs[i] = feedPaths[name]
			}

			// Create temp directory for outputs
			tmpDir := t.TempDir()
			javaOutput := filepath.Join(tmpDir, "java_merged.zip")
			goOutput := filepath.Join(tmpDir, "go_merged.zip")

			// Run Java merger with increased memory for real feeds
			javaMerger := NewJavaMerger(jarPath)
			javaMerger.MaxMemory = "1g"
			err := javaMerger.MergeQuiet(inputs, javaOutput, WithDuplicateDetection(scenario.detection))
			if err != nil {
				t.Fatalf("Java merge failed: %v", err)
			}

			// Run Go merger
			goMerger := createGoMergerWithDetection(scenario.detection)
			err = goMerger.MergeFiles(inputs, goOutput)
			if err != nil {
				t.Fatalf("Go merge failed: %v", err)
			}

			// Compare outputs - must be 100% identical after normalization
			diffs, err := CompareGTFS(javaOutput, goOutput)
			if err != nil {
				t.Fatalf("Comparison failed: %v", err)
			}

			if len(diffs) > 0 {
				diffOutput := FormatDiffStyleOutput(diffs)
				t.Fatalf("Java and Go outputs differ:\n%s", diffOutput)
			}
		})
	}
}

// skipIfNoRealWorldFeeds checks if real-world feeds are available
// Returns a map of feed name to file path, or skips the test if feeds are missing
func skipIfNoRealWorldFeeds(t *testing.T) map[string]string {
	t.Helper()

	feedDir := getRealWorldFeedDir()

	feeds := map[string]string{
		"pierce_transit":    filepath.Join(feedDir, "pierce_transit.zip"),
		"intercity_transit": filepath.Join(feedDir, "intercity_transit.zip"),
		"community_transit": filepath.Join(feedDir, "community_transit.zip"),
		"everett_transit":   filepath.Join(feedDir, "everett_transit.zip"),
	}

	for name, path := range feeds {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("Real-world feed %s not found at %s - run testdata/realworld/download.sh first", name, path)
		}
	}

	return feeds
}

// getRealWorldFeedDir returns the absolute path to testdata/realworld
func getRealWorldFeedDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "testdata/realworld"
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "testdata", "realworld")
}

// createGoMergerWithDetection creates a Go merger with the specified detection mode.
// This matches Java CLI behavior where --file=stops.txt --duplicateDetection={mode}
// applies the detection mode only to stops.txt, not all entity types.
func createGoMergerWithDetection(detection string) *merge.Merger {
	var mode strategy.DuplicateDetection
	switch detection {
	case "none":
		mode = strategy.DetectionNone
	case "identity":
		mode = strategy.DetectionIdentity
	case "fuzzy":
		mode = strategy.DetectionFuzzy
	default:
		mode = strategy.DetectionNone
	}

	// Match Java CLI behavior: only apply to stops.txt
	// Java uses: --file=stops.txt --duplicateDetection={mode}
	// Other entity types use their default (NONE) mode
	return merge.New(merge.WithFileDetection("stops.txt", mode))
}
