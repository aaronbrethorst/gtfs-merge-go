# gtfs-merge-go

A Go port of [onebusaway-gtfs-merge](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge) for merging multiple GTFS feeds into a single unified feed.

## Features

- Merge multiple GTFS feeds into a single output
- Three duplicate detection modes:
  - **None**: Add all entities (with prefix on collision)
  - **Identity**: Match duplicates by ID
  - **Fuzzy**: Match duplicates by properties (name, location, etc.)
- Concurrent fuzzy matching for improved performance
- Maintains referential integrity across all GTFS entity types
- CLI tool and Go library

## Installation

### CLI Tool

```bash
go install github.com/aaronbrethorst/gtfs-merge-go/cmd/gtfs-merge@latest
```

### Library

```bash
go get github.com/aaronbrethorst/gtfs-merge-go
```

## CLI Usage

```bash
# Basic merge
gtfs-merge feed1.zip feed2.zip merged.zip

# Merge with identity duplicate detection
gtfs-merge --duplicateDetection=identity feed1.zip feed2.zip merged.zip

# Merge with fuzzy duplicate detection
gtfs-merge --duplicateDetection=fuzzy feed1.zip feed2.zip merged.zip

# Show help
gtfs-merge --help
```

## Library Usage

### Basic Merge

```go
import (
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
)

// Merge two GTFS feeds
merger := merge.New()
err := merger.MergeFiles([]string{"feed1.zip", "feed2.zip"}, "merged.zip")
if err != nil {
    log.Fatal(err)
}
```

### Merge with Duplicate Detection

```go
import (
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// Use identity-based duplicate detection (match by ID)
merger := merge.New(
    merge.WithDefaultDetection(strategy.DetectionIdentity),
)
err := merger.MergeFiles([]string{"feed1.zip", "feed2.zip"}, "merged.zip")

// Use fuzzy duplicate detection (match by properties)
fuzzyMerger := merge.New(
    merge.WithDefaultDetection(strategy.DetectionFuzzy),
)
err = fuzzyMerger.MergeFiles([]string{"feed1.zip", "feed2.zip"}, "merged.zip")
```

### Working with Feed Objects Directly

```go
import (
    "github.com/aaronbrethorst/gtfs-merge-go/gtfs"
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
)

// Read feeds
feed1, err := gtfs.ReadFromPath("feed1.zip")
if err != nil {
    log.Fatal(err)
}

feed2, err := gtfs.ReadFromPath("feed2.zip")
if err != nil {
    log.Fatal(err)
}

// Merge feeds in memory
merger := merge.New()
merged, err := merger.MergeFeeds([]*gtfs.Feed{feed1, feed2})
if err != nil {
    log.Fatal(err)
}

// Write result
err = gtfs.WriteToPath(merged, "merged.zip")
```

### Enable Concurrent Fuzzy Matching

For large feeds, you can enable concurrent fuzzy matching:

```go
import (
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

merger := merge.New(
    merge.WithDefaultDetection(strategy.DetectionFuzzy),
)

// Enable concurrent processing for stops
stopStrategy := strategy.NewStopMergeStrategy()
stopStrategy.SetDuplicateDetection(strategy.DetectionFuzzy)
stopStrategy.SetConcurrent(true)
stopStrategy.SetConcurrentWorkers(8)
merger.SetStopStrategy(stopStrategy)

err := merger.MergeFiles([]string{"feed1.zip", "feed2.zip"}, "merged.zip")
```

## Architecture

The codebase follows a modular structure:

- **`gtfs/`** - GTFS data model and I/O
- **`merge/`** - Core merge orchestration
- **`strategy/`** - Entity-specific merge strategies with duplicate detection
- **`scoring/`** - Duplicate similarity scoring for fuzzy matching
- **`compare/`** - Java-Go comparison testing framework
- **`cmd/gtfs-merge/`** - CLI application

### Entity Processing Order

Entities are merged in dependency order to maintain referential integrity:
1. Agencies, Areas
2. Stops (handles self-referential parent_station)
3. Service Calendars
4. Routes (references agency)
5. Shapes
6. Trips (references route, service, shape)
7. Stop Times, Frequencies (reference trip, stop)
8. Transfers, Pathways (reference stops)
9. Fare Attributes, Fare Rules
10. Feed Info

## Development

### Running Tests

```bash
# Run all unit tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests for a specific package
go test ./gtfs/...
go test ./merge/...

# Run benchmarks
go test -bench=. ./...
```

### Running Integration Tests

The integration tests compare the Go implementation's output against the original Java [onebusaway-gtfs-merge-cli](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge-cli) tool to ensure compatibility.

**Requirements:**
- Java 21 or later

**Setup:**

```bash
# Download the Java GTFS merge CLI tool
./testdata/java/download.sh
```

**Run integration tests:**

```bash
go test -v -tags=java ./compare/...
```

**Note:** If Java is not installed, the integration tests will be **skipped** (not failed). The CI pipeline runs these tests automatically with Java 21.

## Project Status

**All milestones are complete.** This project is fully functional with duplicate detection (none, identity, fuzzy), Java comparison testing, and feed validation.

See [docs/archived/spec.md](docs/archived/spec.md) for the original project specification and milestone tracking.

## License

MIT License - see [LICENSE](LICENSE) for details.
