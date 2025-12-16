# gtfs-merge-go

A Go port of [onebusaway-gtfs-merge](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge) for merging multiple GTFS feeds into a single unified feed.

## Installation

```bash
go get github.com/aaronbrethorst/gtfs-merge-go
```

## Usage

```go
import (
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
)

// Merge multiple GTFS feeds
merger := merge.New()
err := merger.MergeFiles([]string{"feed1.zip", "feed2.zip"}, "merged.zip")
```

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
go test ./compare/...
```

### Running Integration Tests

The integration tests compare the Go implementation's output against the original Java [onebusaway-gtfs-merge-cli](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge-cli) tool to ensure compatibility.

**Requirements:**
- Java 11 or later (Java 17 recommended)

**Setup:**

```bash
# Download the Java GTFS merge CLI tool
./testdata/java/download.sh
```

**Run integration tests:**

```bash
# Run Java comparison tests
go test -v -tags=java ./compare/...
```

The integration tests use a `//go:build java` tag, so they are skipped by default when running `go test ./...`. This allows development without Java installed. The CI pipeline runs these tests automatically with Java 17.

## Project Status

See [spec.md](spec.md) for the full project specification and milestone tracking.