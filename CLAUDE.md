# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go port of [onebusaway-gtfs-merge](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge) - a library for merging multiple static GTFS (General Transit Feed Specification) feeds into a single unified feed. The project handles duplicate detection, entity renaming, and referential integrity across all GTFS entity types.

## Build Commands

```bash
# Initialize module (when starting)
go mod init github.com/aaronbrethorst/gtfs-merge-go

# Run all tests
go test ./...

# Run tests for a specific package
go test ./gtfs/...
go test ./merge/...
go test ./strategy/...
go test ./scoring/...

# Run a single test
go test -run TestMergeTwoSimpleFeeds ./merge/...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Build CLI
go build -o gtfs-merge ./cmd/gtfs-merge

# Run CLI
./gtfs-merge feed1.zip feed2.zip merged.zip
```

## Architecture

The codebase follows a modular structure with clear separation of concerns:

### Package Structure

- **`gtfs/`** - GTFS data model and I/O
  - Entity structs (Agency, Stop, Route, Trip, StopTime, Calendar, etc.)
  - Feed container holding all entities with maps for O(1) lookup
  - CSV parsing/writing utilities
  - Zip/directory reader and writer

- **`merge/`** - Core merge orchestration
  - `Merger` - main orchestrator that processes feeds in dependency order
  - `MergeContext` - tracks source/target feeds, ID mappings, and prefix
  - Functional options pattern for configuration

- **`strategy/`** - Entity-specific merge strategies
  - `EntityMergeStrategy` interface with `Merge()`, duplicate detection/logging/renaming configuration
  - One strategy per entity type (AgencyMergeStrategy, StopMergeStrategy, etc.)
  - Three detection modes: None, Identity (same ID), Fuzzy (similar properties)

- **`scoring/`** - Duplicate similarity scoring
  - `Scorer[T]` generic interface returning 0.0-1.0 similarity
  - `PropertyMatcher`, `AndScorer` for combining scorers
  - Specialized scorers: StopDistanceScorer (geographic), RouteStopsInCommonScorer, TripScheduleOverlapScorer

- **`cmd/gtfs-merge/`** - CLI application

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

- **Duplicate Detection**: Three modes - `DetectionNone` (always add), `DetectionIdentity` (same ID), `DetectionFuzzy` (property similarity)
- **ID Prefixing**: When IDs collide, feeds get prefixes (a_, b_, c_) applied to all entities and references
- **Functional Options**: `merge.New(WithDebug(true), WithDefaultDetection(DetectionFuzzy))`

## Development Approach

This project follows **Test-Driven Development (TDD)**:
1. Write tests that define expected behavior
2. Run tests and observe them fail (red)
3. Write minimal code to make tests pass (green)
4. Refactor while keeping tests passing

Test fixtures should go in `testdata/` with minimal valid GTFS feeds for testing various scenarios.
