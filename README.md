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

**Check if Java is installed:**

```bash
java -version
```

If Java is not installed:
- **macOS**: `brew install openjdk@17` or download from [Adoptium](https://adoptium.net/)
- **Ubuntu/Debian**: `sudo apt install openjdk-17-jdk`
- **Windows**: Download from [Adoptium](https://adoptium.net/)

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

**Note:** If Java is not installed, the integration tests will be **skipped** (not failed). The CI pipeline runs these tests automatically with Java 17.

The integration tests use a `//go:build java` tag, so they are excluded by default when running `go test ./...`. This allows development without Java installed.

## Project Status

See [spec.md](spec.md) for the full project specification and milestone tracking.