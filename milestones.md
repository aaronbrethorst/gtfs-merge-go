
### Milestone 1: Project Setup and GTFS Model

**Goal**: Establish project structure and define all GTFS entity types.

#### 1.1 Initialize Go Module
```bash
go mod init github.com/aaronbrethorst/gtfs-merge-go
```

#### 1.1.1 Set Up GitHub Actions CI

Create `.github/workflows/ci.yml` to run continuous integration on every push and pull request:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run tests
        run: go test -v -race ./...

  fmt:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Check formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Go files are not formatted:"
            gofmt -d .
            exit 1
          fi

  vet:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run go vet
        run: go vet ./...
```

All CI checks must pass before merging pull requests.

#### 1.1.2 Quality Assurance Process

At every milestone completion and at reasonable internal checkpoints during longer milestones, apply the following QA process:

**1. Code Review**
- Use the `code-review-expert` subagent to review all new or modified code
- Address any issues identified before proceeding

**2. Run CI Checks Locally**
```bash
# Check formatting (should produce no output if properly formatted)
gofmt -l .

# Run static analysis
go vet ./...

# Run linter (matches CI golangci-lint)
golangci-lint run

# Run tests with race detector
go test -v -race ./...
```

**3. Verify Functionality**
- Confirm the milestone deliverables work as expected
- Test any new features or changes manually if appropriate

**4. Commit Changes**
- Create a commit with a clear, descriptive message
- Reference the milestone number in commit messages when completing milestones

**5. Update Milestone Tracking**
- Mark the milestone as complete in the tracking section below
- Add any feedback, notes, or lessons learned

This process ensures consistent quality and catches issues locally before they reach CI.

#### 1.2 Define GTFS Entity Types (TDD)

**Tests to write first** (`gtfs/model_test.go`):
```go
func TestAgencyFields(t *testing.T)           // Verify Agency struct has all GTFS fields
func TestStopFields(t *testing.T)             // Verify Stop struct has all GTFS fields
func TestRouteFields(t *testing.T)            // Verify Route struct has all GTFS fields
func TestTripFields(t *testing.T)             // Verify Trip struct has all GTFS fields
func TestStopTimeFields(t *testing.T)         // Verify StopTime struct has all GTFS fields
func TestCalendarFields(t *testing.T)         // Verify Calendar struct has all GTFS fields
func TestCalendarDateFields(t *testing.T)     // Verify CalendarDate struct has all GTFS fields
func TestShapePointFields(t *testing.T)       // Verify ShapePoint struct has all GTFS fields
func TestFrequencyFields(t *testing.T)        // Verify Frequency struct has all GTFS fields
func TestTransferFields(t *testing.T)         // Verify Transfer struct has all GTFS fields
func TestFareAttributeFields(t *testing.T)    // Verify FareAttribute struct has all GTFS fields
func TestFareRuleFields(t *testing.T)         // Verify FareRule struct has all GTFS fields
func TestFeedInfoFields(t *testing.T)         // Verify FeedInfo struct has all GTFS fields
func TestAreaFields(t *testing.T)             // Verify Area struct has all GTFS fields
func TestPathwayFields(t *testing.T)          // Verify Pathway struct has all GTFS fields
```

**Implementation**: Create `gtfs/model.go` with all entity structs.

#### 1.3 Define Feed Container (TDD)

**Tests to write first** (`gtfs/feed_test.go`):
```go
func TestNewFeed(t *testing.T)                // NewFeed returns initialized empty feed
func TestFeedAddAgency(t *testing.T)          // Can add agency to feed
func TestFeedAddStop(t *testing.T)            // Can add stop to feed
func TestFeedAddRoute(t *testing.T)           // Can add route to feed
func TestFeedAddTrip(t *testing.T)            // Can add trip to feed
func TestFeedAddStopTime(t *testing.T)        // Can add stop time to feed
func TestFeedAddCalendar(t *testing.T)        // Can add calendar to feed
func TestFeedAddCalendarDate(t *testing.T)    // Can add calendar date to feed
func TestFeedAddShapePoint(t *testing.T)      // Can add shape point to feed
```

**Implementation**: Create `gtfs/feed.go` with Feed struct and helper methods.

---

### Milestone 2: CSV Parsing Foundation

**Goal**: Parse individual GTFS CSV files into entity structs.

#### 2.1 CSV Reader Utility (TDD)

**Tests to write first** (`gtfs/csv_test.go`):
```go
func TestParseCSVHeader(t *testing.T)         // Correctly parses CSV header row
func TestParseCSVWithAllFields(t *testing.T)  // Parses row with all columns present
func TestParseCSVWithMissingOptionalFields(t *testing.T)  // Handles missing optional columns
func TestParseCSVWithExtraFields(t *testing.T)            // Ignores unknown columns
func TestParseCSVEmptyFile(t *testing.T)      // Handles empty file (header only)
func TestParseCSVQuotedFields(t *testing.T)   // Handles quoted fields with commas
func TestParseCSVUTF8BOM(t *testing.T)        // Handles UTF-8 BOM at start of file
func TestParseCSVTrailingNewline(t *testing.T) // Handles trailing newlines
```

**Implementation**: Create `gtfs/csv.go` with generic CSV parsing utilities.

#### 2.2 Entity-Specific Parsers (TDD)

**Tests to write first** (`gtfs/parse_test.go`):
```go
func TestParseAgency(t *testing.T)            // Parse agency.txt content
func TestParseAgencyMinimalFields(t *testing.T)  // Parse with only required fields
func TestParseStops(t *testing.T)             // Parse stops.txt content
func TestParseStopsWithParentStation(t *testing.T)  // Parse stops with parent references
func TestParseRoutes(t *testing.T)            // Parse routes.txt content
func TestParseTrips(t *testing.T)             // Parse trips.txt content
func TestParseStopTimes(t *testing.T)         // Parse stop_times.txt content
func TestParseStopTimesTimeFormat(t *testing.T)   // Handle times > 24:00:00
func TestParseCalendar(t *testing.T)          // Parse calendar.txt content
func TestParseCalendarDates(t *testing.T)     // Parse calendar_dates.txt content
func TestParseShapes(t *testing.T)            // Parse shapes.txt content
func TestParseFrequencies(t *testing.T)       // Parse frequencies.txt content
func TestParseTransfers(t *testing.T)         // Parse transfers.txt content
func TestParseFareAttributes(t *testing.T)    // Parse fare_attributes.txt content
func TestParseFareRules(t *testing.T)         // Parse fare_rules.txt content
func TestParseFeedInfo(t *testing.T)          // Parse feed_info.txt content
```

**Implementation**: Create parsing functions for each entity type.

---

### Milestone 3: GTFS Reader (Zip/Directory)

**Goal**: Read complete GTFS feeds from zip files or directories.

#### 3.1 Create Test Fixtures

Create minimal valid GTFS test feeds in `testdata/`:
- `testdata/minimal/` - Bare minimum valid feed (1 agency, 1 stop, 1 route, 1 trip, 1 stop_time, 1 calendar)
- `testdata/simple_a/` - Simple feed A (2 agencies, 5 stops, 2 routes, 4 trips)
- `testdata/simple_b/` - Simple feed B (1 agency, 3 stops, 1 route, 2 trips) - no overlap with A
- `testdata/overlap/` - Feed with IDs that overlap with simple_a

#### 3.2 Directory Reader (TDD)

**Tests to write first** (`gtfs/reader_test.go`):
```go
func TestReadFromDirectory(t *testing.T)      // Read feed from directory
func TestReadFromDirectoryMissingRequired(t *testing.T)  // Error on missing required file
func TestReadFromDirectoryOptionalFiles(t *testing.T)    // Handle missing optional files
func TestReadFromDirectoryInvalidCSV(t *testing.T)       // Error on malformed CSV
```

**Implementation**: Create `gtfs/reader.go` with `ReadFromDirectory()`.

#### 3.3 Zip Reader (TDD)

**Tests to write first** (`gtfs/reader_test.go`):
```go
func TestReadFromZipPath(t *testing.T)        // Read feed from zip file path
func TestReadFromZipReader(t *testing.T)      // Read feed from io.ReaderAt
func TestReadFromPathAutoDetect(t *testing.T) // Auto-detect zip vs directory
func TestReadFromZipNestedDirectory(t *testing.T)  // Handle zip with nested folder
```

**Implementation**: Add zip reading support to `gtfs/reader.go`.

---

### Milestone 4: GTFS Writer

**Goal**: Write Feed objects to GTFS zip files.

#### 4.1 CSV Writer Utility (TDD)

**Tests to write first** (`gtfs/write_test.go`):
```go
func TestWriteCSVHeader(t *testing.T)         // Write correct header row
func TestWriteCSVEmptySlice(t *testing.T)     // Handle empty entity slice
func TestWriteCSVEscapeCommas(t *testing.T)   // Escape commas in values
func TestWriteCSVEscapeQuotes(t *testing.T)   // Escape quotes in values
```

**Implementation**: Create `gtfs/csv_writer.go`.

#### 4.2 Feed Writer (TDD)

**Tests to write first** (`gtfs/writer_test.go`):
```go
func TestWriteToPath(t *testing.T)            // Write feed to zip file
func TestWriteAndReadRoundTrip(t *testing.T)  // Write then read produces same data
func TestWriteEmptyFeed(t *testing.T)         // Handle feed with only required files
func TestWriteAllOptionalFiles(t *testing.T)  // Write all optional files when present
func TestWriteSkipsEmptyOptionalFiles(t *testing.T)  // Don't write empty optional files
```

**Implementation**: Create `gtfs/writer.go`.

---

### Milestone 5: Simple Feed Merge (No Duplicate Detection)

**Goal**: Merge two simple, non-overlapping feeds into one valid output feed.

This is the critical early milestone that proves the core merge capability works.

#### 5.1 Merge Context (TDD)

**Tests to write first** (`merge/context_test.go`):
```go
func TestNewMergeContext(t *testing.T)        // Create context with source/target
func TestMergeContextPrefix(t *testing.T)     // Context has correct prefix
func TestMergeContextEntityTracking(t *testing.T)  // Track entities by raw ID
```

**Implementation**: Create `merge/context.go`.

#### 5.2 Basic Merger - Concatenation Only (TDD)

**Tests to write first** (`merge/merger_test.go`):
```go
// Core merge tests - these are the critical "it works" tests
func TestMergeTwoSimpleFeeds(t *testing.T) {
    // Given: two simple feeds with no overlapping IDs
    // When: merged
    // Then: output contains all entities from both feeds
}

func TestMergeTwoFeedsAgencies(t *testing.T) {
    // Given: feed A has agencies [A1, A2], feed B has agency [B1]
    // When: merged
    // Then: output has agencies [A1, A2, B1]
}

func TestMergeTwoFeedsStops(t *testing.T) {
    // Given: feed A has stops [S1, S2], feed B has stops [S3, S4]
    // When: merged
    // Then: output has stops [S1, S2, S3, S4]
}

func TestMergeTwoFeedsRoutes(t *testing.T) {
    // Given: feed A has routes [R1], feed B has routes [R2]
    // When: merged with agency references updated
    // Then: output has routes [R1, R2] with correct agency refs
}

func TestMergeTwoFeedsTrips(t *testing.T) {
    // Given: feed A has trips [T1, T2], feed B has trips [T3]
    // When: merged with route/service references updated
    // Then: output has trips [T1, T2, T3] with correct refs
}

func TestMergeTwoFeedsStopTimes(t *testing.T) {
    // Given: each feed has stop_times for their trips
    // When: merged
    // Then: output has all stop_times with correct trip/stop refs
}

func TestMergeTwoFeedsCalendars(t *testing.T) {
    // Given: feed A has service [SVC1], feed B has service [SVC2]
    // When: merged
    // Then: output has services [SVC1, SVC2]
}

func TestMergeTwoFeedsPreservesReferentialIntegrity(t *testing.T) {
    // Given: two complete feeds
    // When: merged
    // Then: all foreign key references are valid in output
}

func TestMergeProducesValidGTFS(t *testing.T) {
    // Given: two valid feeds
    // When: merged
    // Then: output passes GTFS validation
}

func TestMergeFilesEndToEnd(t *testing.T) {
    // Given: two zip files
    // When: MergeFiles() called
    // Then: output zip is valid and contains merged data
}
```

**Implementation**: Create `merge/merger.go` with basic concatenation logic.

#### 5.3 ID Prefixing for Non-Overlapping Merge (TDD)

**Tests to write first** (`merge/merger_test.go`):
```go
func TestMergeAppliesPrefixToSecondFeed(t *testing.T) {
    // Given: feeds with same IDs
    // When: merged with DetectionNone
    // Then: second feed entities get prefix
}

func TestMergePrefixUpdatesAllReferences(t *testing.T) {
    // Given: feed B entity IDs get prefixed
    // When: merged
    // Then: all references to those IDs are also prefixed
}

func TestMergePrefixSequence(t *testing.T) {
    // Given: three feeds
    // When: merged
    // Then: prefixes are a_, b_, c_ (or similar)
}
```

**Implementation**: Add prefix logic to merger.

---

### Milestone 5.5: Java-Go Comparison Testing Framework

**Goal**: Establish a comparison testing framework to validate the Go implementation produces output equivalent to the original Java onebusaway-gtfs-merge tool.

#### 5.5.1 Java Tool Integration (TDD)

**Tests to write first** (`compare/java_test.go`):
```go
//go:build java

func TestJavaToolExists(t *testing.T)           // JAR file exists at expected path
func TestJavaToolMerge(t *testing.T)            // Basic merge produces valid output
func TestJavaToolMergeWithDetection(t *testing.T) // Detection modes work
func TestJavaToolMergeMultipleFeeds(t *testing.T) // Three-feed merge works
```

**Implementation**:
- Create `testdata/java/download.sh` to download JAR from Maven Central
- Create `compare/java.go` with `JavaMerger` struct and `Merge()` method

#### 5.5.2 CSV Normalization Utilities (TDD)

**Tests to write first** (`compare/normalize_test.go`):
```go
func TestNormalizeRowOrder(t *testing.T)        // Sort rows by primary key
func TestNormalizeColumnOrder(t *testing.T)     // Reorder to canonical GTFS order
func TestNormalizeFloatPrecision(t *testing.T)  // Normalize to 6 decimal places
func TestNormalizeEmptyFields(t *testing.T)     // Treat empty and missing as equivalent
func TestNormalizeWhitespace(t *testing.T)      // Normalize line endings and trim
func TestNormalizeAgencyTxt(t *testing.T)       // File-specific normalization
func TestNormalizeStopsTxt(t *testing.T)
func TestNormalizeStopTimesTxt(t *testing.T)    // Composite key sorting
func TestPrimaryKey(t *testing.T)               // Primary key definitions
func TestGTFSColumnOrder(t *testing.T)          // Canonical column orders
func TestStripUTF8BOM(t *testing.T)             // BOM removal
```

**Implementation**: Create `compare/normalize.go` with `NormalizeCSV()`, `PrimaryKey()`, `GTFSColumnOrder()`.

#### 5.5.3 Comparison Framework (TDD)

**Tests to write first** (`compare/compare_test.go`):
```go
//go:build java

func TestCompare_IdenticalFeeds(t *testing.T)       // Sanity check - same inputs match
func TestCompare_SimpleMergeNoOverlap(t *testing.T) // simple_a + simple_b
func TestCompare_SimpleMergeWithPrefixing(t *testing.T) // With ID collisions
func TestCompare_MinimalFeed(t *testing.T)          // Minimal feed merge
func TestCompare_EntityCounts(t *testing.T)         // Entity count comparison
```

**Implementation**: Create `compare/compare.go` with `CompareGTFS()`, `CompareCSV()`, `DiffResult` struct.

#### 5.5.4 CI Integration

- Add `compare-java` job to `.github/workflows/ci.yml`
- Setup: Go 1.22, Java 17 (Temurin), cache JAR
- Run: `go test -v -tags=java ./compare/...`

**Note**: Tests with `//go:build java` tag are skipped when Java is not available. CI runs them with `-tags=java`.

---

### Milestone 6: Feed Validation

**Goal**: Validate feeds for GTFS compliance and referential integrity.

#### 6.1 Referential Integrity Validation (TDD)

**Tests to write first** (`gtfs/validate_test.go`):
```go
func TestValidateRouteAgencyRef(t *testing.T)      // Route references valid agency
func TestValidateTripRouteRef(t *testing.T)        // Trip references valid route
func TestValidateTripServiceRef(t *testing.T)      // Trip references valid service
func TestValidateStopTimeStopRef(t *testing.T)     // StopTime references valid stop
func TestValidateStopTimeTripRef(t *testing.T)     // StopTime references valid trip
func TestValidateStopParentRef(t *testing.T)       // Stop parent_station is valid
func TestValidateTransferStopRefs(t *testing.T)    // Transfer stop refs are valid
func TestValidateFareRuleRefs(t *testing.T)        // FareRule refs are valid
func TestValidateShapeInTrip(t *testing.T)         // Trip shape_id is valid
func TestValidateFrequencyTripRef(t *testing.T)    // Frequency trip ref is valid
```

**Implementation**: Create `gtfs/validate.go`.

#### 6.2 Required Field Validation (TDD)

**Tests to write first** (`gtfs/validate_test.go`):
```go
func TestValidateAgencyRequired(t *testing.T)      // Agency has required fields
func TestValidateStopRequired(t *testing.T)        // Stop has required fields
func TestValidateRouteRequired(t *testing.T)       // Route has required fields
func TestValidateTripRequired(t *testing.T)        // Trip has required fields
func TestValidateStopTimeRequired(t *testing.T)    // StopTime has required fields
func TestValidateCalendarRequired(t *testing.T)    // Calendar has required fields
```

**Implementation**: Add required field checks to `gtfs/validate.go`.

---

### Milestone 7: Strategy Interface and Base Classes

**Goal**: Define the strategy interface and abstract base implementations.

#### 7.1 Strategy Interface (TDD)

**Tests to write first** (`strategy/strategy_test.go`):
```go
func TestStrategyInterfaceCompliance(t *testing.T)  // All strategies implement interface
func TestMergeContextCreation(t *testing.T)         // Context created correctly
```

**Implementation**: Create `strategy/strategy.go` with interface definition.

#### 7.2 Detection/Logging/Renaming Enums (TDD)

**Tests to write first** (`strategy/enums_test.go`):
```go
func TestDuplicateDetectionValues(t *testing.T)    // Enum has correct values
func TestDuplicateLoggingValues(t *testing.T)      // Enum has correct values
func TestRenamingStrategyValues(t *testing.T)      // Enum has correct values
func TestDuplicateDetectionString(t *testing.T)    // String representation
```

**Implementation**: Create enum types in `strategy/` package.

---

### Milestone 8: Identity-Based Duplicate Detection

**Goal**: Detect duplicates when entities have the same ID.

#### 8.1 Agency Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/agency_test.go`):
```go
func TestAgencyMergeNoDuplicates(t *testing.T)     // No duplicates, all copied
func TestAgencyMergeIdentityDuplicate(t *testing.T) {
    // Given: both feeds have agency with ID "agency1"
    // When: merged with DetectionIdentity
    // Then: only one agency1 in output, routes updated
}
func TestAgencyMergeUpdatesRouteRefs(t *testing.T)  // Merged agency updates route refs
func TestAgencyMergeLogsWarning(t *testing.T)       // Warning logged when configured
func TestAgencyMergeErrorOnDuplicate(t *testing.T)  // Error when configured
```

**Implementation**: Create `strategy/agency.go`.

#### 8.2 Stop Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/stop_test.go`):
```go
func TestStopMergeNoDuplicates(t *testing.T)
func TestStopMergeIdentityDuplicate(t *testing.T)
func TestStopMergeUpdatesStopTimeRefs(t *testing.T)
func TestStopMergeUpdatesTransferRefs(t *testing.T)
func TestStopMergeUpdatesParentStation(t *testing.T)
```

**Implementation**: Create `strategy/stop.go`.

#### 8.3 Route Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/route_test.go`):
```go
func TestRouteMergeNoDuplicates(t *testing.T)
func TestRouteMergeIdentityDuplicate(t *testing.T)
func TestRouteMergeUpdatesTripRefs(t *testing.T)
func TestRouteMergeUpdatesFareRuleRefs(t *testing.T)
```

**Implementation**: Create `strategy/route.go`.

#### 8.4 Trip Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/trip_test.go`):
```go
func TestTripMergeNoDuplicates(t *testing.T)
func TestTripMergeIdentityDuplicate(t *testing.T)
func TestTripMergeUpdatesStopTimeRefs(t *testing.T)
func TestTripMergeUpdatesFrequencyRefs(t *testing.T)
```

**Implementation**: Create `strategy/trip.go`.

#### 8.5 Calendar Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/calendar_test.go`):
```go
func TestCalendarMergeNoDuplicates(t *testing.T)
func TestCalendarMergeIdentityDuplicate(t *testing.T)
func TestCalendarMergeUpdatesTripRefs(t *testing.T)
func TestCalendarMergeMergesDateRanges(t *testing.T)
func TestCalendarDatesMerged(t *testing.T)
```

**Implementation**: Create `strategy/calendar.go`.

#### 8.6 Shape Merge Strategy with Identity Detection (TDD)

**Tests to write first** (`strategy/shape_test.go`):
```go
func TestShapeMergeNoDuplicates(t *testing.T)
func TestShapeMergeIdentityDuplicate(t *testing.T)
func TestShapeMergeUpdatesTripRefs(t *testing.T)
```

**Implementation**: Create `strategy/shape.go`.

#### 8.7 Remaining Entity Strategies (TDD)

**Tests to write first** (`strategy/*_test.go`):
```go
// Frequency
func TestFrequencyMergeNoDuplicates(t *testing.T)
func TestFrequencyMergeIdentical(t *testing.T)

// Transfer
func TestTransferMergeNoDuplicates(t *testing.T)
func TestTransferMergeIdentical(t *testing.T)

// FareAttribute
func TestFareAttributeMergeNoDuplicates(t *testing.T)
func TestFareAttributeMergeIdentityDuplicate(t *testing.T)

// FareRule
func TestFareRuleMergeNoDuplicates(t *testing.T)
func TestFareRuleMergeIdentical(t *testing.T)

// FeedInfo
func TestFeedInfoMerge(t *testing.T)
func TestFeedInfoMergeCombinesVersions(t *testing.T)
func TestFeedInfoMergeExpandsDateRange(t *testing.T)

// Area
func TestAreaMergeNoDuplicates(t *testing.T)
func TestAreaMergeIdentityDuplicate(t *testing.T)

// Pathway
func TestPathwayMergeNoDuplicates(t *testing.T)
```

**Implementation**: Create remaining strategy files.

---

### Milestone 9: Duplicate Scoring Infrastructure

**Goal**: Build scoring framework for fuzzy duplicate detection.

#### 9.1 Scorer Interface (TDD)

**Tests to write first** (`scoring/scorer_test.go`):
```go
func TestScorerInterface(t *testing.T)            // Interface is correctly defined
func TestPropertyMatcherAllMatch(t *testing.T)    // Score 1.0 when all match
func TestPropertyMatcherNoneMatch(t *testing.T)   // Score 0.0 when none match
func TestPropertyMatcherPartialMatch(t *testing.T) // Proportional score
func TestAndScorerAllAboveThreshold(t *testing.T)  // Combined scoring
func TestAndScorerBelowThreshold(t *testing.T)     // Returns 0 if any below
```

**Implementation**: Create `scoring/scorer.go`.

#### 9.2 Stop Distance Scorer (TDD)

**Tests to write first** (`scoring/stop_distance_test.go`):
```go
func TestStopDistanceSameLocation(t *testing.T)    // Score 1.0 for same coords
func TestStopDistanceWithinThreshold(t *testing.T) // Score > 0 within distance
func TestStopDistanceBeyondThreshold(t *testing.T) // Score 0 beyond distance
func TestStopDistanceHaversine(t *testing.T)       // Correct distance calculation
```

**Implementation**: Create `scoring/stop_distance.go`.

#### 9.3 Route Stops In Common Scorer (TDD)

**Tests to write first** (`scoring/route_stops_test.go`):
```go
func TestRouteStopsAllInCommon(t *testing.T)       // Score 1.0 for same stops
func TestRouteStopsNoneInCommon(t *testing.T)      // Score 0 for no overlap
func TestRouteStopsPartialOverlap(t *testing.T)    // Proportional score
```

**Implementation**: Create `scoring/route_stops.go`.

#### 9.4 Trip Scorers (TDD)

**Tests to write first** (`scoring/trip_test.go`):
```go
// Trip Stops In Common
func TestTripStopsAllInCommon(t *testing.T)
func TestTripStopsNoneInCommon(t *testing.T)
func TestTripStopsPartialOverlap(t *testing.T)

// Trip Schedule Overlap
func TestTripScheduleExactMatch(t *testing.T)
func TestTripScheduleWithinTolerance(t *testing.T)
func TestTripScheduleNoOverlap(t *testing.T)
```

**Implementation**: Create `scoring/trip_stops.go` and `scoring/trip_schedule.go`.

#### 9.5 Service Date Overlap Scorer (TDD)

**Tests to write first** (`scoring/service_dates_test.go`):
```go
func TestServiceDatesFullOverlap(t *testing.T)
func TestServiceDatesPartialOverlap(t *testing.T)
func TestServiceDatesNoOverlap(t *testing.T)
```

**Implementation**: Create `scoring/service_dates.go`.

---

### Milestone 10: Fuzzy Duplicate Detection

**Goal**: Detect duplicates based on similar properties, not just ID.

#### 10.1 Fuzzy Detection in Stop Strategy (TDD)

**Tests to write first** (`strategy/stop_test.go`):
```go
func TestStopMergeFuzzyByName(t *testing.T) {
    // Given: stops with different IDs but same name
    // When: merged with DetectionFuzzy
    // Then: detected as duplicates
}
func TestStopMergeFuzzyByDistance(t *testing.T) {
    // Given: stops with different IDs but within 500m
    // When: merged with DetectionFuzzy
    // Then: detected as duplicates if names similar
}
func TestStopMergeFuzzyNoMatch(t *testing.T) {
    // Given: stops with different names and far apart
    // When: merged with DetectionFuzzy
    // Then: not detected as duplicates
}
```

**Implementation**: Add fuzzy detection to `strategy/stop.go`.

#### 10.2 Fuzzy Detection in Route Strategy (TDD)

**Tests to write first** (`strategy/route_test.go`):
```go
func TestRouteMergeFuzzyByName(t *testing.T)
func TestRouteMergeFuzzyByStops(t *testing.T)
func TestRouteMergeFuzzyNoMatch(t *testing.T)
```

**Implementation**: Add fuzzy detection to `strategy/route.go`.

#### 10.3 Fuzzy Detection in Trip Strategy (TDD)

**Tests to write first** (`strategy/trip_test.go`):
```go
func TestTripMergeFuzzyByStops(t *testing.T)
func TestTripMergeFuzzyBySchedule(t *testing.T)
func TestTripMergeFuzzyRejectsOnStopTimeDiff(t *testing.T)
```

**Implementation**: Add fuzzy detection to `strategy/trip.go`.

#### 10.4 Fuzzy Detection in Calendar Strategy (TDD)

**Tests to write first** (`strategy/calendar_test.go`):
```go
func TestCalendarMergeFuzzyByDateOverlap(t *testing.T)
func TestCalendarMergeFuzzyNoOverlap(t *testing.T)
```

**Implementation**: Add fuzzy detection to `strategy/calendar.go`.

---

### Milestone 11: Auto-Detection of Strategy

**Goal**: Automatically choose best duplicate detection strategy based on feed analysis.

#### 11.1 Strategy Auto-Detection (TDD)

**Tests to write first** (`strategy/autodetect_test.go`):
```go
func TestAutoDetectIdentityWhenIDsOverlap(t *testing.T) {
    // Given: feeds share many IDs
    // When: auto-detecting strategy
    // Then: returns DetectionIdentity
}
func TestAutoDetectFuzzyWhenNoIDOverlap(t *testing.T) {
    // Given: feeds share no IDs but have similar entities
    // When: auto-detecting strategy
    // Then: returns DetectionFuzzy
}
func TestAutoDetectNoneWhenNoSimilarity(t *testing.T) {
    // Given: completely different feeds
    // When: auto-detecting strategy
    // Then: returns DetectionNone
}
```

**Implementation**: Add auto-detection logic to strategy base class.

---

### Milestone 12: Merger Configuration Options

**Goal**: Support functional options pattern for configuration.

#### 12.1 Merger Options (TDD)

**Tests to write first** (`merge/options_test.go`):
```go
func TestWithDebug(t *testing.T)
func TestWithDefaultDetection(t *testing.T)
func TestWithDefaultLogging(t *testing.T)
func TestWithDefaultRenaming(t *testing.T)
func TestMultipleOptions(t *testing.T)
```

**Implementation**: Create `merge/options.go`.

#### 12.2 Per-Strategy Configuration (TDD)

**Tests to write first** (`merge/merger_test.go`):
```go
func TestSetAgencyStrategy(t *testing.T)
func TestSetStopStrategy(t *testing.T)
func TestGetStrategyForFile(t *testing.T)
func TestCustomStrategyUsed(t *testing.T)
```

**Implementation**: Add strategy setter methods.

---

### Milestone 13: CLI Application

**Goal**: Build command-line interface for the merge tool.

#### 13.1 Argument Parsing (TDD)

**Tests to write first** (`cmd/gtfs-merge/main_test.go`):
```go
func TestParseArgsMinimum(t *testing.T)         // Two inputs + one output
func TestParseArgsMultipleInputs(t *testing.T)  // Multiple inputs
func TestParseArgsWithOptions(t *testing.T)     // With flags
func TestParseArgsFileOption(t *testing.T)      // --file specific config
func TestParseArgsDuplicateDetection(t *testing.T)
func TestParseArgsHelp(t *testing.T)
func TestParseArgsVersion(t *testing.T)
func TestParseArgsInvalid(t *testing.T)         // Error on invalid args
```

**Implementation**: Create `cmd/gtfs-merge/main.go`.

#### 13.2 CLI End-to-End (TDD)

**Tests to write first** (`cmd/gtfs-merge/main_test.go`):
```go
func TestCLIMergeTwoFeeds(t *testing.T)         // Basic merge works
func TestCLIWithDuplicateDetection(t *testing.T)
func TestCLIWithPerFileConfig(t *testing.T)
func TestCLIDebugOutput(t *testing.T)
func TestCLIErrorOnInvalidInput(t *testing.T)
func TestCLIOutputFileSize(t *testing.T)        // Reports file size
```

**Implementation**: Complete CLI implementation.

---

### Milestone 14: Integration Tests with Real Data

**Goal**: Verify correctness with real-world GTFS feeds.

#### 14.1 Sample Feed Tests

**Tests to write**:
```go
func TestMergeRealWorldFeeds(t *testing.T)      // Test with real feeds
func TestMergeLargeFeed(t *testing.T)           // Performance with large feed
func TestMergeThreeFeeds(t *testing.T)          // Multiple feed merge
func TestMergeFeedWithAllOptionalFiles(t *testing.T)
```

#### 14.2 Edge Case Tests

**Tests to write**:
```go
func TestMergeEmptyOptionalFiles(t *testing.T)
func TestMergeMissingOptionalFiles(t *testing.T)
func TestMergeUnicodeContent(t *testing.T)
func TestMergeLargeIDs(t *testing.T)
func TestMergeSpecialCharactersInIDs(t *testing.T)
```

---

### Milestone 15: Performance and Polish

**Goal**: Optimize performance and add finishing touches.

#### 15.1 Benchmarks

**Tests to write** (`*_bench_test.go`):
```go
func BenchmarkReadFeed(b *testing.B)
func BenchmarkWriteFeed(b *testing.B)
func BenchmarkMergeTwoFeeds(b *testing.B)
func BenchmarkFuzzyScoring(b *testing.B)
func BenchmarkLargeFeedMerge(b *testing.B)
```

#### 15.2 Concurrent Fuzzy Matching (TDD)

**Tests to write**:
```go
func TestConcurrentScoringCorrectness(t *testing.T)
func TestConcurrentScoringPerformance(t *testing.T)
```

**Implementation**: Add goroutine-based parallel scoring.

#### 15.3 Documentation

- Add GoDoc comments to all exported types and functions
- Create README.md with usage examples
- Add CONTRIBUTING.md

---

## Milestone Summary

| # | Milestone | Key Deliverable |
|---|-----------|-----------------|
| 1 | Project Setup | Go module, entity types, Feed struct |
| 2 | CSV Parsing | Parse all GTFS CSV files |
| 3 | GTFS Reader | Read from zip/directory |
| 4 | GTFS Writer | Write to zip file |
| **5** | **Simple Merge** | **Merge two feeds with no duplicates** |
| **5.5** | **Java Comparison** | **Validate Go output matches Java tool** |
| 6 | Validation | Referential integrity checks |
| 7 | Strategy Interface | Base strategy implementation |
| 8 | Identity Detection | Detect duplicates by ID |
| 9 | Scoring Infrastructure | Similarity scoring framework |
| 10 | Fuzzy Detection | Detect duplicates by properties |
| 11 | Auto-Detection | Choose best strategy automatically |
| 12 | Configuration | Functional options, per-file config |
| 13 | CLI | Command-line application |
| 14 | Integration Tests | Real-world feed testing |
| 15 | Performance | Optimization and polish |

**Milestone 5 is the critical proof-of-concept** - after completing it, you have a working merge tool that can combine two disjoint feeds into one valid output.

---
