//go:build java

package compare

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/merge"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

func skipIfNoJava(t *testing.T) string {
	// Check if Java is actually installed and working (not just a stub)
	cmd := exec.Command("java", "-version")
	if err := cmd.Run(); err != nil {
		t.Skip("Java not installed or not working - skipping integration test")
	}

	// Check if JAR file exists
	jarPath := GetDefaultJARPath()
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		t.Skipf("JAR file not found at %s - run testdata/java/download.sh first", jarPath)
	}
	return jarPath
}

func TestCompare_IdenticalFeeds(t *testing.T) {
	// Given: same feed processed by both tools
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// When: merged by both tools
	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Then: compare outputs (both should be valid GTFS)
	diffs, err := CompareGTFS(javaOutput, goOutput)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	// Log differences but don't fail - this test establishes baseline
	if len(diffs) > 0 {
		t.Logf("Found %d file(s) with differences between Java and Go outputs:", len(diffs))
		for _, diff := range diffs {
			t.Logf("  %s: %d differences", diff.File, len(diff.Differences))
			for _, d := range diff.Differences[:min(5, len(diff.Differences))] {
				t.Logf("    %s at %s", d.Type, d.Location)
			}
		}
	}
}

func TestCompare_SimpleMergeNoOverlap(t *testing.T) {
	// Given: simple_a + simple_b (no overlapping IDs)
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// When: merged by Java and Go tools
	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Then: normalized outputs should be comparable
	diffs, err := CompareGTFS(javaOutput, goOutput)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	// For simple merge without overlap, outputs should be very similar
	logDifferences(t, diffs)
}

func TestCompare_SimpleMergeWithPrefixing(t *testing.T) {
	// Given: simple_a + overlap (some overlapping IDs)
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// When: merged with detection=none (forces prefixing)
	err := javaMerger.MergeQuiet([]string{inputA, inputOverlap}, javaOutput, WithDuplicateDetection("none"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputOverlap}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Then: both tools should apply prefixing
	diffs, err := CompareGTFS(javaOutput, goOutput)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	logDifferences(t, diffs)
}

func TestCompare_MinimalFeed(t *testing.T) {
	// Given: minimal test feed merged with simple_b
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputMinimal := "../testdata/minimal"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// When: merged
	err := javaMerger.MergeQuiet([]string{inputMinimal, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputMinimal, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Then: outputs should match
	diffs, err := CompareGTFS(javaOutput, goOutput)
	if err != nil {
		t.Fatalf("Comparison failed: %v", err)
	}

	logDifferences(t, diffs)
}

func TestCompare_EntityCounts(t *testing.T) {
	// This test verifies that both tools produce the same entity counts
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read both feeds and compare entity counts
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	// Compare counts
	compareCounts(t, "Agencies", len(javaFeed.Agencies), len(goFeed.Agencies))
	compareCounts(t, "Stops", len(javaFeed.Stops), len(goFeed.Stops))
	compareCounts(t, "Routes", len(javaFeed.Routes), len(goFeed.Routes))
	compareCounts(t, "Trips", len(javaFeed.Trips), len(goFeed.Trips))
	compareCounts(t, "StopTimes", len(javaFeed.StopTimes), len(goFeed.StopTimes))
	compareCounts(t, "Calendars", len(javaFeed.Calendars), len(goFeed.Calendars))
}

func compareCounts(t *testing.T, entity string, java, golang int) {
	if java != golang {
		t.Errorf("%s count mismatch: Java=%d, Go=%d", entity, java, golang)
	} else {
		t.Logf("%s count: %d (matched)", entity, java)
	}
}

func logDifferences(t *testing.T, diffs []DiffResult) {
	if len(diffs) == 0 {
		t.Log("No differences found - outputs match!")
		return
	}

	t.Logf("Found %d file(s) with differences:", len(diffs))
	for _, diff := range diffs {
		t.Logf("  %s: %d differences", diff.File, len(diff.Differences))
		for i, d := range diff.Differences {
			if i >= 10 {
				t.Logf("    ... and %d more", len(diff.Differences)-10)
				break
			}
			switch d.Type {
			case RowMissing:
				t.Logf("    Missing row at %s: %s", d.Location, d.Expected)
			case RowExtra:
				t.Logf("    Extra row at %s: %s", d.Location, d.Actual)
			case RowDifferent:
				t.Logf("    Different at %s:\n      Expected: %s\n      Actual:   %s", d.Location, d.Expected, d.Actual)
			}
		}
	}
}

// ============================================================================
// Validation Integration Tests
// ============================================================================

func TestValidation_GoMergedFeedPassesValidation(t *testing.T) {
	// Verify that Go's merge output produces a valid GTFS feed
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// Merge with Go
	err := goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read and validate
	feed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Go merged feed has %d validation errors:", len(errs))
		for _, e := range errs {
			t.Errorf("  - %v", e)
		}
	} else {
		t.Log("Go merged feed passes validation")
	}
}

func TestValidation_JavaMergedFeedPassesValidation(t *testing.T) {
	// Verify that Java's merge output passes our validation
	// This ensures our validation isn't too strict
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")

	// Merge with Java
	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	// Read and validate
	feed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Java merged feed has %d validation errors:", len(errs))
		for _, e := range errs {
			t.Errorf("  - %v", e)
		}
	} else {
		t.Log("Java merged feed passes validation")
	}
}

func TestValidation_BothMergedFeedsValidate(t *testing.T) {
	// Compare validation results for both Go and Java merged outputs
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// Merge with both tools
	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput)
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read both feeds
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	// Validate both
	javaErrs := javaFeed.Validate()
	goErrs := goFeed.Validate()

	t.Logf("Java merged feed validation errors: %d", len(javaErrs))
	t.Logf("Go merged feed validation errors: %d", len(goErrs))

	// Both should pass validation
	if len(javaErrs) > 0 {
		t.Errorf("Java merged feed has validation errors:")
		for _, e := range javaErrs {
			t.Errorf("  - %v", e)
		}
	}

	if len(goErrs) > 0 {
		t.Errorf("Go merged feed has validation errors:")
		for _, e := range goErrs {
			t.Errorf("  - %v", e)
		}
	}

	if len(javaErrs) == 0 && len(goErrs) == 0 {
		t.Log("Both merged feeds pass validation!")
	}
}

func TestValidation_MergeWithOverlapPassesValidation(t *testing.T) {
	// Verify that merging feeds with overlapping IDs still produces valid output
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_merged.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// Merge with both tools (forces prefixing due to overlapping IDs)
	err := javaMerger.MergeQuiet([]string{inputA, inputOverlap}, javaOutput, WithDuplicateDetection("none"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputOverlap}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read and validate both
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	javaErrs := javaFeed.Validate()
	goErrs := goFeed.Validate()

	t.Logf("Java merged (with overlap) validation errors: %d", len(javaErrs))
	t.Logf("Go merged (with overlap) validation errors: %d", len(goErrs))

	if len(javaErrs) > 0 {
		t.Errorf("Java merged feed (with overlap) has validation errors:")
		for _, e := range javaErrs {
			t.Errorf("  - %v", e)
		}
	}

	if len(goErrs) > 0 {
		t.Errorf("Go merged feed (with overlap) has validation errors:")
		for _, e := range goErrs {
			t.Errorf("  - %v", e)
		}
	}

	if len(javaErrs) == 0 && len(goErrs) == 0 {
		t.Log("Both merged feeds (with overlap) pass validation!")
	}
}

// ============================================================================
// Detection Mode Integration Tests
// ============================================================================

func TestDetectionModes_JavaIdentityVsNone(t *testing.T) {
	// Compare Java's behavior with identity vs none detection modes
	// Identity mode should merge entities with matching IDs
	// None mode should prefix and keep all entities
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	identityOutput := filepath.Join(tmpDir, "java_identity.zip")
	noneOutput := filepath.Join(tmpDir, "java_none.zip")

	// Merge with identity detection
	err := javaMerger.MergeQuiet([]string{inputA, inputOverlap}, identityOutput, WithDuplicateDetection("identity"))
	if err != nil {
		t.Fatalf("Java identity merge failed: %v", err)
	}

	// Merge with none detection
	err = javaMerger.MergeQuiet([]string{inputA, inputOverlap}, noneOutput, WithDuplicateDetection("none"))
	if err != nil {
		t.Fatalf("Java none merge failed: %v", err)
	}

	// Read both feeds
	identityFeed, err := gtfs.ReadFromPath(identityOutput)
	if err != nil {
		t.Fatalf("Failed to read identity output: %v", err)
	}

	noneFeed, err := gtfs.ReadFromPath(noneOutput)
	if err != nil {
		t.Fatalf("Failed to read none output: %v", err)
	}

	// With overlapping IDs:
	// - Identity mode should detect duplicates and merge them (fewer entities)
	// - None mode should keep all entities (more entities, with prefixes)
	t.Logf("Java identity detection - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(identityFeed.Agencies), len(identityFeed.Stops), len(identityFeed.Routes), len(identityFeed.Trips))
	t.Logf("Java none detection - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(noneFeed.Agencies), len(noneFeed.Stops), len(noneFeed.Routes), len(noneFeed.Trips))

	// Verify both produce valid feeds
	identityErrs := identityFeed.Validate()
	noneErrs := noneFeed.Validate()

	if len(identityErrs) > 0 {
		t.Errorf("Java identity merged feed has validation errors: %v", identityErrs)
	}
	if len(noneErrs) > 0 {
		t.Errorf("Java none merged feed has validation errors: %v", noneErrs)
	}

	// None mode should generally have >= entity counts compared to identity
	// (identity merges duplicates, none keeps all with prefixes)
	if len(noneFeed.Agencies) < len(identityFeed.Agencies) {
		t.Log("Note: none mode has fewer agencies than identity - may be expected based on Java's algorithm")
	}
}

func TestDetectionModes_GoMatchesJavaNone(t *testing.T) {
	// Go currently implements DetectionNone behavior
	// This test verifies Go's output matches Java with detection=none
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_none.zip")
	goOutput := filepath.Join(tmpDir, "go_merged.zip")

	// Java with explicit none detection
	err := javaMerger.MergeQuiet([]string{inputA, inputB}, javaOutput, WithDuplicateDetection("none"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	// Go (currently always uses none detection behavior)
	err = goMerger.MergeFiles([]string{inputA, inputB}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Compare entity counts
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	t.Logf("Java (none detection) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d, StopTimes: %d",
		len(javaFeed.Agencies), len(javaFeed.Stops), len(javaFeed.Routes), len(javaFeed.Trips), len(javaFeed.StopTimes))
	t.Logf("Go (none detection) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d, StopTimes: %d",
		len(goFeed.Agencies), len(goFeed.Stops), len(goFeed.Routes), len(goFeed.Trips), len(goFeed.StopTimes))

	// For non-overlapping feeds, counts should match exactly
	if len(javaFeed.Agencies) != len(goFeed.Agencies) {
		t.Errorf("Agency count mismatch: Java=%d, Go=%d", len(javaFeed.Agencies), len(goFeed.Agencies))
	}
	if len(javaFeed.Stops) != len(goFeed.Stops) {
		t.Errorf("Stop count mismatch: Java=%d, Go=%d", len(javaFeed.Stops), len(goFeed.Stops))
	}
	if len(javaFeed.Routes) != len(goFeed.Routes) {
		t.Errorf("Route count mismatch: Java=%d, Go=%d", len(javaFeed.Routes), len(goFeed.Routes))
	}
	if len(javaFeed.Trips) != len(goFeed.Trips) {
		t.Errorf("Trip count mismatch: Java=%d, Go=%d", len(javaFeed.Trips), len(goFeed.Trips))
	}
	if len(javaFeed.StopTimes) != len(goFeed.StopTimes) {
		t.Errorf("StopTime count mismatch: Java=%d, Go=%d", len(javaFeed.StopTimes), len(goFeed.StopTimes))
	}
}

func TestDetectionModes_ThreeFeedMerge(t *testing.T) {
	// Test merging three feeds with different detection modes
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New()

	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"
	inputMinimal := "../testdata/minimal"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_three.zip")
	goOutput := filepath.Join(tmpDir, "go_three.zip")

	// Java merge of three feeds
	err := javaMerger.MergeQuiet([]string{inputA, inputB, inputMinimal}, javaOutput, WithDuplicateDetection("none"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	// Go merge of three feeds
	err = goMerger.MergeFiles([]string{inputA, inputB, inputMinimal}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read both feeds
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	t.Logf("Three-feed merge (Java) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(javaFeed.Agencies), len(javaFeed.Stops), len(javaFeed.Routes), len(javaFeed.Trips))
	t.Logf("Three-feed merge (Go) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(goFeed.Agencies), len(goFeed.Stops), len(goFeed.Routes), len(goFeed.Trips))

	// Both should produce valid feeds
	javaErrs := javaFeed.Validate()
	goErrs := goFeed.Validate()

	if len(javaErrs) > 0 {
		t.Errorf("Java three-feed merge has validation errors: %v", javaErrs)
	}
	if len(goErrs) > 0 {
		t.Errorf("Go three-feed merge has validation errors: %v", goErrs)
	}

	// Entity counts should match for non-overlapping feeds with none detection
	if len(javaFeed.Agencies) != len(goFeed.Agencies) {
		t.Logf("Note: Agency count differs - Java=%d, Go=%d", len(javaFeed.Agencies), len(goFeed.Agencies))
	}
	if len(javaFeed.Stops) != len(goFeed.Stops) {
		t.Logf("Note: Stop count differs - Java=%d, Go=%d", len(javaFeed.Stops), len(goFeed.Stops))
	}
	if len(javaFeed.Routes) != len(goFeed.Routes) {
		t.Logf("Note: Route count differs - Java=%d, Go=%d", len(javaFeed.Routes), len(goFeed.Routes))
	}
	if len(javaFeed.Trips) != len(goFeed.Trips) {
		t.Logf("Note: Trip count differs - Java=%d, Go=%d", len(javaFeed.Trips), len(goFeed.Trips))
	}
}

func TestDetectionModes_OverlapWithIdentity(t *testing.T) {
	// Test Go's identity detection matches Java's identity detection
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New(merge.WithDefaultDetection(strategy.DetectionIdentity))

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_identity.zip")
	goOutput := filepath.Join(tmpDir, "go_identity.zip")

	// Java with identity detection
	err := javaMerger.MergeQuiet([]string{inputA, inputOverlap}, javaOutput, WithDuplicateDetection("identity"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	// Go with identity detection
	err = goMerger.MergeFiles([]string{inputA, inputOverlap}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read feeds
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	t.Logf("Overlap merge (Java identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(javaFeed.Agencies), len(javaFeed.Stops), len(javaFeed.Routes), len(javaFeed.Trips))
	t.Logf("Overlap merge (Go identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(goFeed.Agencies), len(goFeed.Stops), len(goFeed.Routes), len(goFeed.Trips))

	// With identity detection, both should have similar entity counts
	// (duplicates are merged, not prefixed)
	compareCounts(t, "Agencies", len(javaFeed.Agencies), len(goFeed.Agencies))
	compareCounts(t, "Stops", len(javaFeed.Stops), len(goFeed.Stops))
	compareCounts(t, "Routes", len(javaFeed.Routes), len(goFeed.Routes))
	compareCounts(t, "Trips", len(javaFeed.Trips), len(goFeed.Trips))

	// Both should still produce valid output
	javaErrs := javaFeed.Validate()
	goErrs := goFeed.Validate()

	if len(javaErrs) > 0 {
		t.Errorf("Java merged feed has validation errors: %v", javaErrs)
	}
	if len(goErrs) > 0 {
		t.Errorf("Go merged feed has validation errors: %v", goErrs)
	}
}

// ============================================================================
// Identity Detection Integration Tests (Milestone 8)
// ============================================================================

func TestIdentityDetection_GoMatchesJavaIdentity(t *testing.T) {
	// Test that Go's identity detection produces similar results to Java's
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New(merge.WithDefaultDetection(strategy.DetectionIdentity))

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_identity.zip")
	goOutput := filepath.Join(tmpDir, "go_identity.zip")

	// Both with identity detection
	err := javaMerger.MergeQuiet([]string{inputA, inputOverlap}, javaOutput, WithDuplicateDetection("identity"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles([]string{inputA, inputOverlap}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Compare entity counts
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	t.Logf("Java (identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d, StopTimes: %d",
		len(javaFeed.Agencies), len(javaFeed.Stops), len(javaFeed.Routes), len(javaFeed.Trips), len(javaFeed.StopTimes))
	t.Logf("Go (identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d, StopTimes: %d",
		len(goFeed.Agencies), len(goFeed.Stops), len(goFeed.Routes), len(goFeed.Trips), len(goFeed.StopTimes))

	// Entity counts should match (or be very close)
	compareCounts(t, "Agencies", len(javaFeed.Agencies), len(goFeed.Agencies))
	compareCounts(t, "Stops", len(javaFeed.Stops), len(goFeed.Stops))
	compareCounts(t, "Routes", len(javaFeed.Routes), len(goFeed.Routes))
	compareCounts(t, "Trips", len(javaFeed.Trips), len(goFeed.Trips))
}

func TestIdentityDetection_PreservesReferentialIntegrity(t *testing.T) {
	// Verify Go's identity detection maintains valid foreign key references
	goMerger := merge.New(merge.WithDefaultDetection(strategy.DetectionIdentity))

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	goOutput := filepath.Join(tmpDir, "go_identity.zip")

	err := goMerger.MergeFiles([]string{inputA, inputOverlap}, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read and validate
	feed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Go identity-merged feed has %d validation errors:", len(errs))
		for _, e := range errs {
			t.Errorf("  - %v", e)
		}
	} else {
		t.Log("Go identity-merged feed passes validation!")
	}
}

func TestIdentityDetection_ThreeFeedMerge(t *testing.T) {
	// Test identity detection with three feeds
	jarPath := skipIfNoJava(t)

	javaMerger := NewJavaMerger(jarPath)
	goMerger := merge.New(merge.WithDefaultDetection(strategy.DetectionIdentity))

	inputs := []string{
		"../testdata/simple_a",
		"../testdata/simple_b",
		"../testdata/minimal",
	}

	tmpDir := t.TempDir()
	javaOutput := filepath.Join(tmpDir, "java_three_identity.zip")
	goOutput := filepath.Join(tmpDir, "go_three_identity.zip")

	err := javaMerger.MergeQuiet(inputs, javaOutput, WithDuplicateDetection("identity"))
	if err != nil {
		t.Fatalf("Java merge failed: %v", err)
	}

	err = goMerger.MergeFiles(inputs, goOutput)
	if err != nil {
		t.Fatalf("Go merge failed: %v", err)
	}

	// Read feeds
	javaFeed, err := gtfs.ReadFromPath(javaOutput)
	if err != nil {
		t.Fatalf("Failed to read Java output: %v", err)
	}

	goFeed, err := gtfs.ReadFromPath(goOutput)
	if err != nil {
		t.Fatalf("Failed to read Go output: %v", err)
	}

	t.Logf("Three-feed merge (Java identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(javaFeed.Agencies), len(javaFeed.Stops), len(javaFeed.Routes), len(javaFeed.Trips))
	t.Logf("Three-feed merge (Go identity) - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(goFeed.Agencies), len(goFeed.Stops), len(goFeed.Routes), len(goFeed.Trips))

	// Validate both
	javaErrs := javaFeed.Validate()
	goErrs := goFeed.Validate()

	if len(javaErrs) > 0 {
		t.Errorf("Java three-feed identity merge has validation errors: %v", javaErrs)
	}
	if len(goErrs) > 0 {
		t.Errorf("Go three-feed identity merge has validation errors: %v", goErrs)
	}
}

func TestIdentityDetection_CompareNoneVsIdentity(t *testing.T) {
	// Compare Go's output with none vs identity detection
	goMergerNone := merge.New() // DetectionNone is default
	goMergerIdentity := merge.New(merge.WithDefaultDetection(strategy.DetectionIdentity))

	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	tmpDir := t.TempDir()
	noneOutput := filepath.Join(tmpDir, "go_none.zip")
	identityOutput := filepath.Join(tmpDir, "go_identity.zip")

	err := goMergerNone.MergeFiles([]string{inputA, inputOverlap}, noneOutput)
	if err != nil {
		t.Fatalf("Go none merge failed: %v", err)
	}

	err = goMergerIdentity.MergeFiles([]string{inputA, inputOverlap}, identityOutput)
	if err != nil {
		t.Fatalf("Go identity merge failed: %v", err)
	}

	// Read feeds
	noneFeed, err := gtfs.ReadFromPath(noneOutput)
	if err != nil {
		t.Fatalf("Failed to read none output: %v", err)
	}

	identityFeed, err := gtfs.ReadFromPath(identityOutput)
	if err != nil {
		t.Fatalf("Failed to read identity output: %v", err)
	}

	t.Logf("Go none detection - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(noneFeed.Agencies), len(noneFeed.Stops), len(noneFeed.Routes), len(noneFeed.Trips))
	t.Logf("Go identity detection - Agencies: %d, Stops: %d, Routes: %d, Trips: %d",
		len(identityFeed.Agencies), len(identityFeed.Stops), len(identityFeed.Routes), len(identityFeed.Trips))

	// With overlapping feeds:
	// - none detection should keep all entities (more entities, with prefixes)
	// - identity detection should merge duplicates (fewer entities)
	if len(noneFeed.Agencies) <= len(identityFeed.Agencies) {
		t.Log("Note: none detection should have more entities than identity detection for overlapping feeds")
	}

	// Both should produce valid feeds
	noneErrs := noneFeed.Validate()
	identityErrs := identityFeed.Validate()

	if len(noneErrs) > 0 {
		t.Errorf("Go none merged feed has validation errors: %v", noneErrs)
	}
	if len(identityErrs) > 0 {
		t.Errorf("Go identity merged feed has validation errors: %v", identityErrs)
	}
}
