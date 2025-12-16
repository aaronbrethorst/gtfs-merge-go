//go:build java

package compare

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/merge"
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
