// Package compare provides utilities for comparing GTFS outputs between
// the Java onebusaway-gtfs-merge tool and the Go implementation.
package compare

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// JavaMerger invokes the onebusaway-gtfs-merge-cli Java tool
type JavaMerger struct {
	JARPath   string
	JavaBin   string // defaults to "java"
	MaxMemory string // defaults to "512m"
}

// javaConfig holds configuration for a single merge operation
type javaConfig struct {
	duplicateDetection string // none, identity, fuzzy
	logDuplicates      bool
}

// JavaOption configures a Java merge operation
type JavaOption func(*javaConfig)

// WithDuplicateDetection sets the duplicate detection mode
func WithDuplicateDetection(mode string) JavaOption {
	return func(c *javaConfig) {
		c.duplicateDetection = mode
	}
}

// WithLogDuplicates enables logging of duplicate entities
func WithLogDuplicates(enabled bool) JavaOption {
	return func(c *javaConfig) {
		c.logDuplicates = enabled
	}
}

// NewJavaMerger creates a new JavaMerger with the given JAR path
func NewJavaMerger(jarPath string) *JavaMerger {
	return &JavaMerger{
		JARPath:   jarPath,
		JavaBin:   "java",
		MaxMemory: "512m",
	}
}

// GetDefaultJARPath returns the default path to the Java GTFS merge CLI JAR
func GetDefaultJARPath() string {
	// Get the directory of this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "testdata/java/onebusaway-gtfs-merge-cli.jar"
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "testdata", "java", "onebusaway-gtfs-merge-cli.jar")
}

// Merge runs the Java merge tool and writes output to the given path
func (j *JavaMerger) Merge(inputs []string, output string, opts ...JavaOption) error {
	if len(inputs) < 2 {
		return fmt.Errorf("at least two input feeds are required")
	}

	// Check JAR exists
	if _, err := os.Stat(j.JARPath); os.IsNotExist(err) {
		return fmt.Errorf("JAR file not found: %s (run testdata/java/download.sh)", j.JARPath)
	}

	// Apply options
	cfg := &javaConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build command args
	args := []string{
		"-Xmx" + j.MaxMemory,
		"-jar", j.JARPath,
	}

	// Add options
	if cfg.duplicateDetection != "" {
		args = append(args, "--duplicateDetection="+cfg.duplicateDetection)
	}
	if cfg.logDuplicates {
		args = append(args, "--logDroppedDuplicates")
	}

	// Add inputs and output
	args = append(args, inputs...)
	args = append(args, output)

	// Run the command
	cmd := exec.Command(j.JavaBin, args...)
	cmd.Stderr = os.Stderr // Show errors
	cmd.Stdout = os.Stdout // Show output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("java merge failed: %w", err)
	}

	return nil
}

// MergeQuiet runs the merge without printing to stdout/stderr
func (j *JavaMerger) MergeQuiet(inputs []string, output string, opts ...JavaOption) error {
	if len(inputs) < 2 {
		return fmt.Errorf("at least two input feeds are required")
	}

	// Check JAR exists
	if _, err := os.Stat(j.JARPath); os.IsNotExist(err) {
		return fmt.Errorf("JAR file not found: %s (run testdata/java/download.sh)", j.JARPath)
	}

	// Apply options
	cfg := &javaConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build command args
	args := []string{
		"-Xmx" + j.MaxMemory,
		"-jar", j.JARPath,
	}

	// Add options
	if cfg.duplicateDetection != "" {
		args = append(args, "--duplicateDetection="+cfg.duplicateDetection)
	}
	if cfg.logDuplicates {
		args = append(args, "--logDroppedDuplicates")
	}

	// Add inputs and output
	args = append(args, inputs...)
	args = append(args, output)

	// Run the command quietly
	cmd := exec.Command(j.JavaBin, args...)

	output_bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("java merge failed: %w\nOutput: %s", err, string(output_bytes))
	}

	return nil
}
