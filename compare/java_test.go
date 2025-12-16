//go:build java

package compare

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// skipIfNoJavaEnv skips the test if Java is not installed or JAR is missing
func skipIfNoJavaEnv(t *testing.T) string {
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

func TestJavaToolExists(t *testing.T) {
	// Given: download script has been run
	// When: checking for JAR file
	jarPath := GetDefaultJARPath()

	// Then: file exists at expected path
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		t.Skipf("JAR file not found at %s - run testdata/java/download.sh first", jarPath)
	}
}

func TestJavaToolMerge(t *testing.T) {
	// Given: two input feeds (simple_a, simple_b)
	jarPath := skipIfNoJavaEnv(t)

	merger := NewJavaMerger(jarPath)
	inputA := "../testdata/simple_a"
	inputB := "../testdata/simple_b"

	// When: running java -jar merge-cli input1 input2 output
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	err := merger.Merge([]string{inputA, inputB}, output)

	// Then: output file is created
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", output)
	}
}

func TestJavaToolMergeWithDetection(t *testing.T) {
	// Given: two input feeds with overlapping IDs
	jarPath := skipIfNoJavaEnv(t)

	merger := NewJavaMerger(jarPath)
	inputA := "../testdata/simple_a"
	inputOverlap := "../testdata/overlap"

	// When: running with --duplicateDetection=none
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	err := merger.Merge([]string{inputA, inputOverlap}, output, WithDuplicateDetection("none"))

	// Then: output contains prefixed entities (merge succeeds)
	if err != nil {
		t.Fatalf("Merge with detection=none failed: %v", err)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", output)
	}
}

func TestJavaToolMergeMultipleFeeds(t *testing.T) {
	// Given: three input feeds
	jarPath := skipIfNoJavaEnv(t)

	merger := NewJavaMerger(jarPath)
	inputs := []string{
		"../testdata/minimal",
		"../testdata/simple_a",
		"../testdata/simple_b",
	}

	// When: merging three feeds
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	err := merger.Merge(inputs, output)

	// Then: merge succeeds
	if err != nil {
		t.Fatalf("Merge of three feeds failed: %v", err)
	}
}
