# Java GTFS Merge CLI

This directory contains the Java [onebusaway-gtfs-merge-cli](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge-cli) tool, used to validate that the Go implementation produces output equivalent to the original Java tool.

## Setup

### Download the JAR

```bash
./download.sh
```

This downloads the JAR from Maven Central and creates a symlink `onebusaway-gtfs-merge-cli.jar`.

### Requirements

- Java 21 or later
- curl (for downloading)

## Usage

```bash
# Basic merge
java -jar onebusaway-gtfs-merge-cli.jar input1.zip input2.zip output.zip

# With duplicate detection mode
java -jar onebusaway-gtfs-merge-cli.jar \
  --duplicateDetection=none \
  input1.zip input2.zip output.zip

# Available duplicate detection modes:
#   none     - Never detect duplicates, always use prefixing
#   identity - Detect duplicates by ID match
#   fuzzy    - Detect duplicates by property similarity
```

## Files

- `download.sh` - Downloads the JAR from Maven Central
- `onebusaway-gtfs-merge-cli-X.X.X.jar` - The actual JAR (gitignored)
- `onebusaway-gtfs-merge-cli.jar` - Symlink to the versioned JAR (gitignored)

## Version

Current version: 11.2.0

To update, modify the `VERSION` variable in `download.sh` and re-run.

## CI

The GitHub Actions CI workflow automatically downloads and caches this JAR for running comparison tests.
