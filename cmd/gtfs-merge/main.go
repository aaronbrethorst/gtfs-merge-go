// Package main provides the command-line interface for gtfs-merge.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aaronbrethorst/gtfs-merge-go/merge"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// Version is the CLI version (set during build)
var Version = "dev"

// config holds parsed command-line configuration
type config struct {
	inputs             []string
	output             string
	debug              bool
	duplicateDetection string
	logging            string
	files              map[string]fileConfig
	showHelp           bool
	showVersion        bool
}

// fileConfig holds per-file configuration
type fileConfig struct {
	detection string
}

// parseArgs parses command-line arguments into a config
func parseArgs(args []string) (*config, error) {
	cfg := &config{
		files: make(map[string]fileConfig),
	}

	var positional []string
	var currentFile string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// Flag argument
			switch {
			case arg == "--help" || arg == "-h":
				cfg.showHelp = true
			case arg == "--version" || arg == "-v":
				cfg.showVersion = true
			case arg == "--debug":
				cfg.debug = true
			case strings.HasPrefix(arg, "--duplicateDetection="):
				mode := strings.TrimPrefix(arg, "--duplicateDetection=")
				// Validate mode
				if mode != "none" && mode != "identity" && mode != "fuzzy" {
					return nil, fmt.Errorf("invalid duplicate detection mode: %q (must be none, identity, or fuzzy)", mode)
				}
				// If we have a current file, apply to it only (per-file config)
				// Otherwise set as global default
				if currentFile != "" {
					fc := cfg.files[currentFile]
					fc.detection = mode
					cfg.files[currentFile] = fc
				} else {
					cfg.duplicateDetection = mode
				}
			case strings.HasPrefix(arg, "--logging="):
				cfg.logging = strings.TrimPrefix(arg, "--logging=")
			case strings.HasPrefix(arg, "--file="):
				currentFile = strings.TrimPrefix(arg, "--file=")
				cfg.files[currentFile] = fileConfig{}
			default:
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// Positional argument
			positional = append(positional, arg)
			// Reset current file when we hit positional args
			currentFile = ""
		}
	}

	// Early exit for help/version
	if cfg.showHelp || cfg.showVersion {
		return cfg, nil
	}

	// Validate positional arguments
	if len(positional) < 3 {
		return nil, fmt.Errorf("at least 3 arguments required: <input1> <input2> [...] <output>")
	}

	// Last positional is output, rest are inputs
	cfg.inputs = positional[:len(positional)-1]
	cfg.output = positional[len(positional)-1]

	return cfg, nil
}

// runMerge executes the merge operation based on config
func runMerge(cfg *config) error {
	// Build merger options
	var opts []merge.Option

	if cfg.debug {
		opts = append(opts, merge.WithDebug(true))
	}

	if cfg.duplicateDetection != "" {
		detection, err := strategy.ParseDuplicateDetection(cfg.duplicateDetection)
		if err != nil {
			return fmt.Errorf("invalid duplicate detection: %w", err)
		}
		opts = append(opts, merge.WithDefaultDetection(detection))
	}

	if cfg.logging != "" {
		var logging strategy.DuplicateLogging
		switch cfg.logging {
		case "none":
			logging = strategy.LogNone
		case "warning":
			logging = strategy.LogWarning
		case "error":
			logging = strategy.LogError
		default:
			return fmt.Errorf("invalid logging mode: %q (must be none, warning, or error)", cfg.logging)
		}
		opts = append(opts, merge.WithDefaultLogging(logging))
	}

	// Create merger
	m := merge.New(opts...)

	// Apply per-file configurations
	for filename, fc := range cfg.files {
		if fc.detection != "" {
			s := m.GetStrategyForFile(filename)
			if s != nil {
				detection, err := strategy.ParseDuplicateDetection(fc.detection)
				if err != nil {
					return fmt.Errorf("invalid detection for %s: %w", filename, err)
				}
				s.SetDuplicateDetection(detection)
			}
		}
	}

	// Execute merge
	return m.MergeFiles(cfg.inputs, cfg.output)
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println(`gtfs-merge - Merge multiple GTFS feeds into one

Usage:
  gtfs-merge [options] <input1> <input2> [...] <output>

Arguments:
  input1, input2, ...  Input GTFS feeds (zip files or directories)
  output               Output GTFS zip file

Options:
  --help, -h           Show this help message
  --version, -v        Show version information
  --debug              Enable debug output
  --duplicateDetection=MODE
                       Duplicate detection mode: none, identity, fuzzy
                       (default: none)
  --logging=MODE       Logging mode for duplicates: none, warning, error
                       (default: none)
  --file=FILENAME      Apply following options to specific GTFS file
                       (e.g., --file=stops.txt --duplicateDetection=fuzzy)

Examples:
  gtfs-merge feed1.zip feed2.zip merged.zip
  gtfs-merge --duplicateDetection=identity feed1.zip feed2.zip merged.zip
  gtfs-merge --file=stops.txt --duplicateDetection=fuzzy feed1.zip feed2.zip merged.zip`)
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("gtfs-merge version %s\n", Version)
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Use --help for usage information")
		os.Exit(1)
	}

	if cfg.showHelp {
		printUsage()
		os.Exit(0)
	}

	if cfg.showVersion {
		printVersion()
		os.Exit(0)
	}

	if err := runMerge(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully merged %d feeds into %s\n", len(cfg.inputs), cfg.output)
}
