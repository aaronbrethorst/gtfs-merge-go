package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// Milestone 13.1: Argument Parsing Tests (TDD)
// ============================================================================

func TestParseArgsMinimum(t *testing.T) {
	// Minimum valid arguments: two inputs + one output
	args := []string{"feed1.zip", "feed2.zip", "output.zip"}

	cfg, err := parseArgs(args)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}

	if len(cfg.inputs) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(cfg.inputs))
	}
	if cfg.inputs[0] != "feed1.zip" {
		t.Errorf("expected input[0]='feed1.zip', got %q", cfg.inputs[0])
	}
	if cfg.inputs[1] != "feed2.zip" {
		t.Errorf("expected input[1]='feed2.zip', got %q", cfg.inputs[1])
	}
	if cfg.output != "output.zip" {
		t.Errorf("expected output='output.zip', got %q", cfg.output)
	}
}

func TestParseArgsMultipleInputs(t *testing.T) {
	// Multiple inputs
	args := []string{"feed1.zip", "feed2.zip", "feed3.zip", "output.zip"}

	cfg, err := parseArgs(args)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}

	if len(cfg.inputs) != 3 {
		t.Errorf("expected 3 inputs, got %d", len(cfg.inputs))
	}
	if cfg.output != "output.zip" {
		t.Errorf("expected output='output.zip', got %q", cfg.output)
	}
}

func TestParseArgsWithOptions(t *testing.T) {
	// With flags
	tests := []struct {
		name            string
		args            []string
		expectDebug     bool
		expectDetection string
		expectLogging   string
	}{
		{
			name:            "debug flag",
			args:            []string{"--debug", "feed1.zip", "feed2.zip", "output.zip"},
			expectDebug:     true,
			expectDetection: "",
			expectLogging:   "",
		},
		{
			name:            "detection flag",
			args:            []string{"--duplicateDetection=identity", "feed1.zip", "feed2.zip", "output.zip"},
			expectDebug:     false,
			expectDetection: "identity",
			expectLogging:   "",
		},
		{
			name:            "logging flag",
			args:            []string{"--logging=warning", "feed1.zip", "feed2.zip", "output.zip"},
			expectDebug:     false,
			expectDetection: "",
			expectLogging:   "warning",
		},
		{
			name:            "multiple flags",
			args:            []string{"--debug", "--duplicateDetection=fuzzy", "--logging=error", "feed1.zip", "feed2.zip", "output.zip"},
			expectDebug:     true,
			expectDetection: "fuzzy",
			expectLogging:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}

			if cfg.debug != tt.expectDebug {
				t.Errorf("debug: expected %v, got %v", tt.expectDebug, cfg.debug)
			}
			if cfg.duplicateDetection != tt.expectDetection {
				t.Errorf("duplicateDetection: expected %q, got %q", tt.expectDetection, cfg.duplicateDetection)
			}
			if cfg.logging != tt.expectLogging {
				t.Errorf("logging: expected %q, got %q", tt.expectLogging, cfg.logging)
			}
		})
	}
}

func TestParseArgsFileOption(t *testing.T) {
	// --file option for per-file configuration
	tests := []struct {
		name         string
		args         []string
		expectFiles  map[string]string // file -> detection mode
		expectInputs int
	}{
		{
			name:        "single file option",
			args:        []string{"--file=stops.txt", "feed1.zip", "feed2.zip", "output.zip"},
			expectFiles: map[string]string{"stops.txt": ""},
		},
		{
			name:        "file with detection",
			args:        []string{"--file=stops.txt", "--duplicateDetection=fuzzy", "feed1.zip", "feed2.zip", "output.zip"},
			expectFiles: map[string]string{"stops.txt": "fuzzy"},
		},
		{
			name:        "multiple file options",
			args:        []string{"--file=agency.txt", "--file=routes.txt", "feed1.zip", "feed2.zip", "output.zip"},
			expectFiles: map[string]string{"agency.txt": "", "routes.txt": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}

			if len(cfg.files) != len(tt.expectFiles) {
				t.Errorf("expected %d file configs, got %d", len(tt.expectFiles), len(cfg.files))
			}
			for file := range tt.expectFiles {
				if _, ok := cfg.files[file]; !ok {
					t.Errorf("expected file config for %q, not found", file)
				}
			}
		})
	}
}

func TestParseArgsDuplicateDetection(t *testing.T) {
	// Test all duplicate detection modes
	modes := []string{"none", "identity", "fuzzy"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			args := []string{"--duplicateDetection=" + mode, "feed1.zip", "feed2.zip", "output.zip"}

			cfg, err := parseArgs(args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}

			if cfg.duplicateDetection != mode {
				t.Errorf("expected duplicateDetection=%q, got %q", mode, cfg.duplicateDetection)
			}
		})
	}
}

func TestParseArgsHelp(t *testing.T) {
	// --help flag
	tests := []struct {
		name string
		args []string
	}{
		{"long help", []string{"--help"}},
		{"short help", []string{"-h"}},
		{"help with other args", []string{"--help", "feed1.zip", "feed2.zip", "output.zip"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}

			if !cfg.showHelp {
				t.Error("expected showHelp=true")
			}
		})
	}
}

func TestParseArgsVersion(t *testing.T) {
	// --version flag
	tests := []struct {
		name string
		args []string
	}{
		{"long version", []string{"--version"}},
		{"short version", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}

			if !cfg.showVersion {
				t.Error("expected showVersion=true")
			}
		})
	}
}

func TestParseArgsInvalid(t *testing.T) {
	// Error on invalid args
	tests := []struct {
		name      string
		args      []string
		expectErr string
	}{
		{
			name:      "no args",
			args:      []string{},
			expectErr: "at least 3 arguments required",
		},
		{
			name:      "one arg",
			args:      []string{"feed1.zip"},
			expectErr: "at least 3 arguments required",
		},
		{
			name:      "two args",
			args:      []string{"feed1.zip", "feed2.zip"},
			expectErr: "at least 3 arguments required",
		},
		{
			name:      "unknown flag",
			args:      []string{"--unknown", "feed1.zip", "feed2.zip", "output.zip"},
			expectErr: "unknown flag",
		},
		{
			name:      "invalid detection mode",
			args:      []string{"--duplicateDetection=invalid", "feed1.zip", "feed2.zip", "output.zip"},
			expectErr: "invalid duplicate detection mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseArgs(tt.args)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectErr)
			}
			if !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectErr, err.Error())
			}
		})
	}
}

// ============================================================================
// Milestone 13.2: CLI End-to-End Tests (TDD)
// ============================================================================

func TestCLIMergeTwoFeeds(t *testing.T) {
	// Basic merge works
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	// Use test fixtures
	inputA := "../../testdata/simple_a"
	inputB := "../../testdata/simple_b"

	// Verify test fixtures exist
	if _, err := os.Stat(inputA); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputA)
	}
	if _, err := os.Stat(inputB); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputB)
	}

	cfg := &config{
		inputs: []string{inputA, inputB},
		output: output,
	}

	err := runMerge(cfg)
	if err != nil {
		t.Fatalf("runMerge failed: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created")
	}
}

func TestCLIWithDuplicateDetection(t *testing.T) {
	// Test each detection mode
	modes := []string{"none", "identity", "fuzzy"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			tmpDir := t.TempDir()
			output := filepath.Join(tmpDir, "merged.zip")

			inputA := "../../testdata/simple_a"
			inputOverlap := "../../testdata/overlap"

			// Verify test fixtures exist
			if _, err := os.Stat(inputA); os.IsNotExist(err) {
				t.Skipf("test fixture not found: %s", inputA)
			}
			if _, err := os.Stat(inputOverlap); os.IsNotExist(err) {
				t.Skipf("test fixture not found: %s", inputOverlap)
			}

			cfg := &config{
				inputs:             []string{inputA, inputOverlap},
				output:             output,
				duplicateDetection: mode,
			}

			err := runMerge(cfg)
			if err != nil {
				t.Fatalf("runMerge with %s detection failed: %v", mode, err)
			}

			// Verify output exists
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Errorf("output file was not created with %s detection", mode)
			}
		})
	}
}

func TestCLIWithPerFileConfig(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	inputA := "../../testdata/simple_a"
	inputOverlap := "../../testdata/overlap"

	// Verify test fixtures exist
	if _, err := os.Stat(inputA); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputA)
	}
	if _, err := os.Stat(inputOverlap); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputOverlap)
	}

	cfg := &config{
		inputs:             []string{inputA, inputOverlap},
		output:             output,
		duplicateDetection: "identity",
		files:              map[string]fileConfig{"stops.txt": {detection: "fuzzy"}},
	}

	err := runMerge(cfg)
	if err != nil {
		t.Fatalf("runMerge with per-file config failed: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created with per-file config")
	}
}

func TestCLIDebugOutput(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	inputA := "../../testdata/simple_a"
	inputB := "../../testdata/simple_b"

	// Verify test fixtures exist
	if _, err := os.Stat(inputA); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputA)
	}
	if _, err := os.Stat(inputB); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputB)
	}

	cfg := &config{
		inputs: []string{inputA, inputB},
		output: output,
		debug:  true,
	}

	err := runMerge(cfg)
	if err != nil {
		t.Fatalf("runMerge with debug failed: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created with debug mode")
	}
}

func TestCLIErrorOnInvalidInput(t *testing.T) {
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	cfg := &config{
		inputs: []string{"/nonexistent/feed1.zip", "/nonexistent/feed2.zip"},
		output: output,
	}

	err := runMerge(cfg)
	if err == nil {
		t.Error("expected error for invalid input, got nil")
	}
}

func TestCLIOutputFileSize(t *testing.T) {
	// Verify output file has non-zero size
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	inputA := "../../testdata/simple_a"
	inputB := "../../testdata/simple_b"

	// Verify test fixtures exist
	if _, err := os.Stat(inputA); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputA)
	}
	if _, err := os.Stat(inputB); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputB)
	}

	cfg := &config{
		inputs: []string{inputA, inputB},
		output: output,
	}

	err := runMerge(cfg)
	if err != nil {
		t.Fatalf("runMerge failed: %v", err)
	}

	// Check file size
	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("failed to stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file has zero size")
	}
	t.Logf("Output file size: %d bytes", info.Size())
}

func TestCLIMergeThreeFeeds(t *testing.T) {
	// Test merging three feeds
	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, "merged.zip")

	inputA := "../../testdata/simple_a"
	inputB := "../../testdata/simple_b"
	inputMinimal := "../../testdata/minimal"

	// Verify test fixtures exist
	if _, err := os.Stat(inputA); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputA)
	}
	if _, err := os.Stat(inputB); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputB)
	}
	if _, err := os.Stat(inputMinimal); os.IsNotExist(err) {
		t.Skipf("test fixture not found: %s", inputMinimal)
	}

	cfg := &config{
		inputs: []string{inputA, inputB, inputMinimal},
		output: output,
	}

	err := runMerge(cfg)
	if err != nil {
		t.Fatalf("runMerge with three feeds failed: %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("output file was not created")
	}
}
