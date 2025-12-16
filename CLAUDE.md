# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go port of [onebusaway-gtfs-merge](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge) - a library for merging multiple static GTFS (General Transit Feed Specification) feeds into a single unified feed. The project handles duplicate detection, entity renaming, and referential integrity across all GTFS entity types.

## Build Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./gtfs/...
go test ./merge/...
go test ./compare/...

# Run Java comparison tests (requires Java 21+)
./testdata/java/download.sh  # Download JAR first
go test -v -tags=java ./compare/...

# Run a single test
go test -run TestMergeTwoSimpleFeeds ./merge/...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector (matches CI)
go test -v -race ./...

# Run benchmarks
go test -bench=. ./...

# Build and run CLI
go build -o gtfs-merge ./cmd/gtfs-merge
./gtfs-merge feed1.zip feed2.zip merged.zip
./gtfs-merge --duplicateDetection=identity feed1.zip feed2.zip merged.zip
./gtfs-merge --help
```

## Architecture

The codebase follows a modular structure with clear separation of concerns:

### Package Structure (Current)

- **`gtfs/`** - GTFS data model and I/O
  - `model.go` - Entity structs (Agency, Stop, Route, Trip, StopTime, Calendar, etc.) with typed IDs
  - `feed.go` - Feed container holding all entities with maps for O(1) lookup
  - `csv.go` - CSVReader/CSVRow for parsing with UTF-8 BOM handling and type conversions
  - `csv_writer.go` - CSVWriter for output
  - `parse.go` - Entity-specific parse functions (ParseAgency, ParseStop, etc.)
  - `reader.go` - ReadFromPath() auto-detects zip vs directory, ReadFromZip() for io.ReaderAt
  - `writer.go` - WriteToPath() and WriteToZip() for complete feeds

- **`merge/`** - Core merge orchestration
  - `merger.go` - Merger with MergeFiles() and MergeFeeds(), processes feeds in reverse order
  - `context.go` - MergeContext tracks source/target feeds, ID mappings for all entity types, and prefix
  - `options.go` - Functional options pattern (WithDebug)

- **`compare/`** - Java-Go comparison testing framework
  - `java.go` - JavaMerger wrapper for invoking onebusaway-gtfs-merge-cli v11.2.0
  - `normalize.go` - CSV normalization for comparison (column order, row sorting, float precision)
  - `compare.go` - CompareGTFS() compares two GTFS outputs with detailed diff reporting
  - Tests use `//go:build java` tag (skipped without Java 21+, run in CI)

- **`strategy/`** - Entity-specific merge strategies with duplicate detection
  - `strategy.go` - EntityMergeStrategy interface, MergeContext, BaseStrategy
  - `enums.go` - DuplicateDetection, DuplicateLogging, RenamingStrategy enums
  - `autodetect.go` - AutoDetectDuplicateDetection() for automatic mode selection
  - Entity strategies: agency.go, stop.go, route.go, trip.go, calendar.go, etc.

- **`scoring/`** - Duplicate similarity scoring for fuzzy matching
  - `scorer.go` - Scorer interface, PropertyMatcher, AndScorer
  - Specialized scorers: stop_distance.go, route_stops.go, trip_stops.go, etc.

- **`cmd/gtfs-merge/`** - CLI application
  - `main.go` - Argument parsing, merge execution, help/version output

### Entity Processing Order

Entities are merged in dependency order to maintain referential integrity:
1. Agencies, Areas
2. Stops (self-referential parent_station)
3. Service Calendars
4. Routes (references agency)
5. Shapes
6. Trips (references route, service, shape)
7. Stop Times, Frequencies (reference trip, stop)
8. Transfers, Pathways (reference stops)
9. Fare Attributes, Fare Rules
10. Feed Info

### Key Design Patterns

- **Reverse Processing Order**: Feeds are processed in reverse order (last feed first). The last feed gets no prefix, earlier feeds get prefixes (a_, b_, c_, etc.). This ensures newer data takes precedence.
- **ID Prefixing**: When IDs collide, earlier feeds get prefixes applied to all entities and references. Prefixes: "" (last feed), "a_" (second-to-last), "b_", ..., "z_", then "00_", "01_", etc.
- **Duplicate Detection**: Three modes - `DetectionNone` (always add), `DetectionIdentity` (same ID), `DetectionFuzzy` (property similarity)
- **Functional Options**: `merge.New(WithDebug(true))`

## Development Approach

This project follows **Test-Driven Development (TDD)**:
1. Write tests that define expected behavior
2. Run tests and observe them fail (red)
3. Write minimal code to make tests pass (green)
4. Refactor while keeping tests passing

Test fixtures should go in `testdata/` with minimal valid GTFS feeds for testing various scenarios.

## Project Status

**All milestones are complete.** The project has a fully functional GTFS merge CLI with duplicate detection (none, identity, fuzzy), Java comparison testing, and feed validation.

The original project specification and milestone tracking have been archived to [docs/archived/spec.md](docs/archived/spec.md).

### QA Process

At every significant change, follow this process:

1. **Code Review** - Review changes for correctness and style
2. **Run CI Checks Locally**:
   ```bash
   gofmt -l .              # Check formatting
   go vet ./...            # Static analysis
   golangci-lint run       # Linter
   go test -v -race ./...  # Tests with race detector
   ```
3. **Verify Functionality** - Confirm deliverables work as expected
4. **Commit** - Clear message describing the change

### Key Files

- `docs/archived/spec.md` - Original project specification and milestone tracking (completed)
- `.github/workflows/ci.yml` - CI configuration (must pass before merge)
