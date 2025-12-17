
| Milestone | Status | Commit | Notes |
|-----------|--------|--------|-------|
| 1.1 Initialize Go Module | ✅ Complete | `12a9687` | Created `go.mod` with module path `github.com/aaronbrethorst/gtfs-merge-go` |
| 1.1.1 Set Up GitHub Actions CI | ✅ Complete | `2350017` | Added `.github/workflows/ci.yml` with lint, test, fmt, vet jobs |
| 1.1.2 Quality Assurance Process | ✅ Complete | `1852d40` | Defined 5-step QA process, added milestone tracking section |
| 1.2 Define GTFS Entity Types | ✅ Complete | `a7c295a` | All 15 entity structs + 8 ID types, 16 tests passing |
| 1.3 Define Feed Container | ✅ Complete | `9e2a22f` | Feed struct with maps/slices for all entities, NewFeed(), 16 new tests |
| 2.1 CSV Reader Utility | ✅ Complete | `6576ff1` | CSVReader and CSVRow types, 15 tests for CSV parsing |
| 2.2 Entity-Specific Parsers | ✅ Complete | `746684e` | Parse functions for all 15 GTFS entities, 18 new tests |
| 3.1 Create Test Fixtures | ✅ Complete | `c59dcb0` | Created 4 test feeds in testdata/ directory |
| 3.2 Directory Reader | ✅ Complete | `c59dcb0` | ReadFromPath() for directory input, 6 tests |
| 3.3 Zip Reader | ✅ Complete | `c59dcb0` | ReadFromZip() for zip input, handles nested directories, 5 tests |
| 4.1 CSV Writer Utility | ✅ Complete | `f120a3e` | CSVWriter type wrapping standard csv.Writer, 8 tests |
| 4.2 Feed Writer | ✅ Complete | `f120a3e` | WriteToPath() and WriteToZip() for complete feeds, 5 tests |
| 5.1 Merge Context | ✅ Complete | `0383cdb` | MergeContext struct with ID mappings, GetPrefixForIndex(), 5 tests |
| 5.2 Basic Merger | ✅ Complete | `0383cdb` | Merger with MergeFiles() and MergeFeeds(), 12 tests |
| 5.3 ID Prefixing | ✅ Complete | `0383cdb` | Feeds processed in reverse order with a_, b_, c_ prefixes, 4 tests |
| 5.5.1 Java Tool Integration | ✅ Complete | `5e3d1f1` | JavaMerger wrapper, download script, 4 tests. Updated to v11.2.0 (requires Java 21+) in `a96171b` |
| 5.5.2 CSV Normalization | ✅ Complete | `5e3d1f1` | NormalizeCSV(), PrimaryKey(), GTFSColumnOrder(), 11 tests |
| 5.5.3 Comparison Framework | ✅ Complete | `5e3d1f1` | CompareGTFS(), CompareCSV(), DiffResult, 6 tests |
| 5.5.4 CI Integration | ✅ Complete | `5e3d1f1` | Added compare-java job to CI workflow |
| 6.1 Referential Integrity Validation | ✅ Complete | `aca9ebc` | Validate() method with refs for routes, trips, stop_times, transfers, fares, pathways, 15 tests |
| 6.2 Required Field Validation | ✅ Complete | `aca9ebc` | Required field checks for agency, stop, route, trip, stop_time, calendar, 6 tests |
| 7.1 Strategy Interface | ✅ Complete | `a53053c` | EntityMergeStrategy interface, MergeContext, BaseStrategy with default implementations, 11 tests |
| 7.2 Detection/Logging/Renaming Enums | ✅ Complete | `a53053c` | DuplicateDetection, DuplicateLogging, RenamingStrategy enums with String() and Parse methods, 7 tests |
| 8.1-8.7 Identity-Based Duplicate Detection | ✅ Complete | `afd0146` | All entity strategies with identity detection, 45+ strategy tests, 5 Java integration tests |
| 9.1-9.5 Duplicate Scoring Infrastructure | ✅ Complete | - | `scoring/` package with Scorer interface, PropertyMatcher, AndScorer, specialized scorers, 55 tests |
| 10.1-10.4 Fuzzy Duplicate Detection | ✅ Complete | - | Fuzzy detection in Stop, Route, Trip, Calendar strategies; 18 new unit tests, 5 integration tests |
| 11.1 Strategy Auto-Detection | ✅ Complete | `1e6558b` | AutoDetectDuplicateDetection() with configurable thresholds, 15 unit tests, 6 integration tests |
| 12.1-12.2 Merger Configuration Options | ✅ Complete | - | Functional options (WithDebug, WithDefaultDetection, WithDefaultLogging, WithDefaultRenaming), per-strategy setters, GetStrategyForFile(), 10 unit tests, 6 Java integration tests |
| 13.1-13.2 CLI Application | ✅ Complete | `3985f1b` | Command-line interface in `cmd/gtfs-merge/`, parseArgs() with flags, runMerge() integration, 15 unit tests, 7 Java integration tests |
| 14.1-14.2 Integration Tests with Real Data | ✅ Complete | - | Edge case fixtures, sample feed tests, unicode/large ID/special char handling, 20+ Java integration tests |
| 15.1 Benchmarks | ✅ Complete | - | Comprehensive benchmarks for gtfs, merge, scoring, strategy packages |
| 15.2 Concurrent Fuzzy Matching | ✅ Complete | - | Worker pool pattern, ConcurrentConfig, integrated into Stop/Route strategies |
| 15.3 Documentation | ✅ Complete | - | Package comments, README.md with examples, CONTRIBUTING.md |

### Feedback & Notes

#### Milestone 1.1 - Initialize Go Module
- Straightforward initialization
- Go version 1.24.2 used

#### Milestone 1.1.1 - Set Up GitHub Actions CI
- Uses Go 1.22 in CI (standard ubuntu-latest)
- Four parallel jobs: lint, test, fmt, vet
- golangci-lint-action v4 for linting

#### Milestone 1.1.2 - Quality Assurance Process
- Defined 5-step QA process: code review, local CI checks, verification, commit, tracking update
- Added milestone tracking section at end of spec.md (now archived to docs/archived/spec.md)
- Code review suggested optional improvements (failure handling, date column) for future consideration
- This QA process will be applied to all future milestones

#### Milestone 1.2 - Define GTFS Entity Types
- Created `gtfs/model.go` with all 15 GTFS entity structs
- Defined 8 type aliases for IDs (AgencyID, StopID, RouteID, etc.)
- TDD approach: wrote reflection-based tests first, then implemented structs
- All tests verify struct fields match the GTFS specification exactly
- 16 tests total, all passing with race detector

#### Milestone 1.3 - Define Feed Container
- Created `gtfs/feed.go` with Feed struct containing all entity collections
- Maps for entities with unique IDs (agencies, stops, routes, trips, calendars, etc.)
- Slices for entities keyed by relationships (stop_times, frequencies, transfers, etc.)
- NewFeed() initializes all maps and slices to avoid nil pointer issues
- 16 new tests verify all entity types can be added to feed
- Total: 32 tests passing with race detector

#### Milestone 2.1 - CSV Reader Utility
- Created `gtfs/csv.go` with CSVReader and CSVRow types
- CSVReader wraps standard csv.Reader with GTFS-specific handling:
  - UTF-8 BOM stripping from first field
  - Whitespace trimming from header column names
  - Empty line/trailing newline skipping
  - Protection against calling ReadRecord before ReadHeader
- CSVRow provides convenient field access by column name with type conversions:
  - Get(column) - returns string, empty for missing fields
  - GetInt(column) - returns int, 0 for invalid/missing
  - GetFloat(column) - returns float64, 0.0 for invalid/missing
  - GetBool(column) - returns bool, true for "1" or "true"
- Returns io.EOF when no more records (follows Go idioms)
- 15 new tests covering edge cases identified in code review
- Total: 47 tests passing with race detector

#### Milestone 2.2 - Entity-Specific Parsers
- Created `gtfs/parse.go` with 15 Parse functions (one per GTFS entity type)
- Each function takes a CSVRow and returns the corresponding entity pointer
- Functions: ParseAgency, ParseStop, ParseRoute, ParseTrip, ParseStopTime, ParseCalendar, ParseCalendarDate, ParseShapePoint, ParseFrequency, ParseTransfer, ParseFareAttribute, ParseFareRule, ParseFeedInfo, ParseArea, ParsePathway
- All column names follow GTFS specification precisely
- Type conversions handled via CSVRow methods (Get, GetInt, GetFloat, GetBool)
- TDD approach: wrote 18 tests first, then implemented parse functions
- Tests cover: all entity types, minimal fields, parent station references, time formats > 24:00:00, exception types, etc.
- Code review grade: A (Excellent) - no blockers, all field mappings verified correct
- Total: 65 tests passing with race detector

#### Milestone 3 - GTFS Reader (Zip/Directory)
- Created 4 test fixtures in `testdata/`:
  - `minimal/` - bare minimum valid feed (1 agency, 1 stop, 1 route, 1 trip)
  - `simple_a/` - larger feed (2 agencies, 5 stops, 2 routes, 4 trips)
  - `simple_b/` - different feed with no ID overlap (1 agency, 3 stops, 1 route, 2 trips)
  - `overlap/` - feed with IDs that collide with simple_a (for future merge testing)
- Created `gtfs/reader.go` with:
  - `ReadFromPath()` - auto-detects directory vs zip file
  - `ReadFromZip()` - reads from io.ReaderAt for programmatic access
  - Validates required GTFS files and at least one calendar file
  - Handles nested directories in zip files (common in real-world feeds)
  - Uses opener pattern for code reuse between directory/zip reading
- Changed CSV reader to use strict quote handling (LazyQuotes=false) to properly detect malformed CSV
- TDD approach: wrote 11 new reader tests, then implemented
- All linter issues fixed (errcheck for defer Close() calls)
- Total: 77 tests passing with race detector

#### Milestone 4 - GTFS Writer
- Created `gtfs/csv_writer.go` with CSVWriter type:
  - Wraps standard csv.Writer for GTFS output
  - Uses Unix-style line endings (LF) for consistency
  - WriteHeader() and WriteRecord() methods
  - Flush() for writing buffered data
- Created `gtfs/writer.go` with WriteToPath() and WriteToZip() functions:
  - Writes all 15 GTFS file types to zip archive
  - Required files always written (agency, stops, routes, trips, stop_times, calendar)
  - Optional files only written when data is present
  - Helper functions formatInt(), formatFloat(), formatBool() for value conversion
- TDD approach: wrote 13 tests first, then implemented
- Round-trip test verifies write/read produces identical data
- All QA checks pass: gofmt, go vet, golangci-lint, race detector
- Total: 89 tests passing with race detector

#### Milestone 5 - Simple Feed Merge (No Duplicate Detection)
- **This is the critical proof-of-concept milestone** - proves core merge capability works
- Created `merge/` package with:
  - `context.go` - MergeContext struct with source/target feeds, ID mappings for all entity types
  - `merger.go` - Merger struct with MergeFiles() and MergeFeeds() methods
  - `options.go` - Functional options pattern (WithDebug)
- Merger processes feeds in **reverse order** (last feed first, gets no prefix)
  - First feed (index 0, processed last) gets prefix based on position
  - GetPrefixForIndex(): 0="" (no prefix), 1="a_", 2="b_", ..., 26="z_", 27="00_", etc.
- Merges all 15 entity types in dependency order:
  1. Agencies, Areas (no dependencies)
  2. Stops (self-referential parent_station)
  3. Calendars, CalendarDates
  4. Routes (references agency)
  5. Shapes
  6. Trips (references route, service, shape)
  7. StopTimes (references trip, stop)
  8. Frequencies (references trip)
  9. Transfers, Pathways (reference stops)
  10. FareAttributes (references agency)
  11. FareRules (references fare, route)
  12. FeedInfo
- All ID references are correctly updated when prefixes are applied
- TDD approach: wrote 21 tests first across context and merger
- Tests verify:
  - Entity concatenation from multiple feeds
  - Referential integrity maintained after merge
  - Round-trip write/read produces valid GTFS
  - ID prefixing with collisions
  - Three-feed merge with correct prefix sequence
- All QA checks pass: gofmt, go vet, golangci-lint (0 issues), race detector
- Total: 110 tests passing with race detector

#### Milestone 5.5 - Java-Go Comparison Testing Framework
- Created new `compare/` package with Java tool integration and GTFS comparison utilities
- **Java Tool Integration** (`compare/java.go`):
  - `JavaMerger` struct wraps the onebusaway-gtfs-merge-cli JAR
  - Downloads JAR v2.0.0 from Maven Central via `testdata/java/download.sh`
  - Supports `--duplicateDetection` option
  - Tests use `//go:build java` tag to skip when Java not available
- **CSV Normalization** (`compare/normalize.go`):
  - `NormalizeCSV()` normalizes GTFS files for comparison
  - Handles: column reordering, row sorting by primary key, float precision, BOM stripping, whitespace
  - `PrimaryKey()` and `GTFSColumnOrder()` define canonical GTFS structures
- **Comparison Framework** (`compare/compare.go`):
  - `CompareGTFS()` compares two GTFS zip files
  - Normalizes both outputs before comparing
  - Returns `DiffResult` with detailed differences
- **CI Integration**:
  - Added `compare-java` job to CI workflow
  - Sets up Java 17 (Temurin), caches JAR, runs comparison tests
- Tests: 11 normalization tests (always run) + 6 comparison tests (require Java)
- Total test count: 121 tests (without Java tag)

#### Milestone 6 - Feed Validation
- Created `gtfs/validate.go` with `Validate()` method on Feed struct
- Returns `[]error` for all validation issues found
- **Referential Integrity Validation** (6.1):
  - Route -> Agency reference (required when multiple agencies exist)
  - Trip -> Route, Service (calendar or calendar_dates), Shape references
  - StopTime -> Trip, Stop references
  - Stop -> ParentStation self-reference
  - Transfer -> FromStop, ToStop references
  - Frequency -> Trip reference
  - FareAttribute -> Agency reference (optional field)
  - FareRule -> Fare (required), Route (optional) references
  - Pathway -> FromStop, ToStop references
- **Required Field Validation** (6.2):
  - Agency: name, url, timezone required
  - Stop: stop_id required, stop_name required for location_type 0-2
  - Route: route_id required, at least one of short_name or long_name required
  - Trip: trip_id, route_id, service_id required
  - StopTime: trip_id, stop_id required
  - Calendar: service_id, start_date, end_date required
- Defined `ValidationError` type with EntityType, EntityID, Field, Message for clear error reporting
- Also fixed pre-existing linter errors in compare/compare.go (errcheck warnings)
- TDD approach: wrote 21 validation tests first, then implemented
- All QA checks pass: gofmt, go vet, golangci-lint (0 issues), race detector
- **Java Integration Tests for Validation** (4 new tests in `compare/compare_test.go`):
  - `TestValidation_GoMergedFeedPassesValidation` - Go merged output is valid GTFS
  - `TestValidation_JavaMergedFeedPassesValidation` - Java merged output passes our validation
  - `TestValidation_BothMergedFeedsValidate` - Compare validation results for both outputs
  - `TestValidation_MergeWithOverlapPassesValidation` - Overlapping ID merge produces valid output
- Total: 142 tests (without Java tag), 155 tests (with Java tag)

#### Milestone 7 - Strategy Interface and Base Classes
- Created new `strategy/` package with core merge strategy abstractions
- **Enum Types** (`strategy/enums.go`):
  - `DuplicateDetection`: DetectionNone, DetectionIdentity, DetectionFuzzy
  - `DuplicateLogging`: LogNone, LogWarning, LogError
  - `RenamingStrategy`: RenameContext, RenameAgency
  - All enums have `String()` method for debugging/logging
  - `ParseDuplicateDetection()` function for CLI parsing (case-insensitive)
- **Strategy Interface** (`strategy/strategy.go`):
  - `EntityMergeStrategy` interface with Name(), Merge(), SetDuplicateDetection(), SetDuplicateLogging(), SetRenamingStrategy()
  - `MergeContext` struct with Source/Target feeds, Prefix, ID mappings, ResolvedDetection
  - `BaseStrategy` struct for embedding in concrete strategies with default implementations
  - `NewMergeContext()` and `NewBaseStrategy()` constructor functions
- **Java-Go Integration Tests** (4 new tests for detection modes):
  - `TestDetectionModes_JavaIdentityVsNone` - Compare Java's identity vs none detection
  - `TestDetectionModes_GoMatchesJavaNone` - Verify Go matches Java with none detection
  - `TestDetectionModes_ThreeFeedMerge` - Three-feed merge comparison
  - `TestDetectionModes_OverlapWithIdentity` - Document difference until identity detection is implemented
- TDD approach: wrote 18 new tests (11 strategy + 7 enum), then implemented
- All QA checks pass: gofmt, go vet, golangci-lint (0 issues), race detector
- Total: 153 tests (without Java tag), 170 tests (with Java tag)

#### Milestone 8 - Identity-Based Duplicate Detection
- Implemented entity-specific merge strategies with identity-based duplicate detection
- **Strategy Implementations** (`strategy/*.go`):
  - `AgencyMergeStrategy` - Detects duplicate agencies by ID, maps to existing if found
  - `StopMergeStrategy` - Handles stops with parent_station references
  - `RouteMergeStrategy` - Maps agency references correctly
  - `TripMergeStrategy` - Maps route, service, shape references
  - `CalendarMergeStrategy` - Merges calendars and calendar_dates
  - `ShapeMergeStrategy` - Handles shape points as groups
  - `StopTimeMergeStrategy` - Detects duplicates by trip_id + stop_sequence
  - `FrequencyMergeStrategy` - Detects duplicates by trip/start/end/headway
  - `TransferMergeStrategy` - Detects duplicates by from/to stop and type
  - `PathwayMergeStrategy` - Maps stop references, prefixes pathway IDs
  - `FareAttributeMergeStrategy` - Maps agency references
  - `FareRuleMergeStrategy` - Maps fare and route references
  - `FeedInfoMergeStrategy` - Combines versions, expands date ranges
  - `AreaMergeStrategy` - Identity-based area merging
- **Merger Integration** (`merge/merger.go`):
  - Refactored to use strategy instances for all entity types
  - Added `SetDuplicateDetectionForAll()` for bulk configuration
  - Added `GetStrategyForFile()` for file-based strategy lookup
  - Added strategy setter methods for customization
  - `WithDefaultDetection()` option for functional configuration
- **New Tests**:
  - 45 new strategy tests covering identity detection for all entity types
  - 8 new merger integration tests for identity detection
  - 5 new Java integration tests comparing Go identity vs Java identity
- All strategies:
  - Support `DetectionNone` (always add with prefix) and `DetectionIdentity` (merge by ID)
  - Support `LogWarning` and `LogError` for duplicate logging
  - Correctly map ID references to downstream entities
- All QA checks pass: gofmt, go vet, race detector
- Total: 198+ tests (without Java tag)

#### Milestone 9 - Duplicate Scoring Infrastructure
- Created new `scoring/` package with duplicate similarity scoring framework
- **Core Scorer Types** (`scoring/scorer.go`):
  - `Scorer[T]` generic interface for entity scoring
  - `PropertyMatcher[T]` - scores based on matching property values
  - `AndScorer[T]` - combines multiple scorers using multiplication with early exit
  - `ElementOverlapScore()` - calculates overlap score for sets (common/a + common/b) / 2
  - `IntervalOverlapScore()` - calculates overlap score for intervals
- **Specialized Scorers**:
  - `StopDistanceScorer` - scores stops by geographic proximity using Haversine formula
    - < 50m → 1.0, < 100m → 0.75, < 500m → 0.5, >= 500m → 0.0
  - `RouteStopsInCommonScorer` - scores routes by shared stops across all trips
  - `TripStopsInCommonScorer` - scores trips by shared stops
  - `TripScheduleOverlapScorer` - scores trips by time window overlap
  - `ServiceDateOverlapScorer` - scores service calendars by date range overlap
- **Helper Functions**:
  - `haversineDistance()` - great-circle distance calculation
  - `parseGTFSTime()` - parses GTFS time strings (handles > 24:00:00)
  - `parseGTFSDate()` - parses GTFS date strings (YYYYMMDD)
- TDD approach: wrote 55 new tests, then implemented
- All QA checks pass: gofmt, go vet, race detector
- Total: 269 tests (without Java tag)

#### Milestone 10 - Fuzzy Duplicate Detection
- Implemented `DetectionFuzzy` mode for Stop, Route, Trip, and Calendar strategies
- **Stop Strategy** (`strategy/stop.go`):
  - `findFuzzyMatch()` method searches for similar stops by name + distance
  - `stopNameScore()` - returns 1.0 for exact match, 0.0 otherwise
  - `stopDistanceScore()` - uses Haversine formula: < 50m → 1.0, < 100m → 0.75, < 500m → 0.5, else 0.0
  - Score = nameScore * distScore; matches if >= FuzzyThreshold (default 0.8)
- **Route Strategy** (`strategy/route.go`):
  - `findFuzzyMatch()` scores routes by agency + names + shared stops
  - `routeAgencyScore()`, `routePropertyScore()`, `routeStopsInCommonScore()`, `getStopsForRoute()`
  - Uses `elementOverlapScore()` for set overlap calculations
- **Trip Strategy** (`strategy/trip.go`):
  - `findFuzzyMatch()` scores trips by route + service + stops + schedule overlap
  - `validateTripStopTimes()` - validates exact stop time match before accepting fuzzy duplicate
  - Ensures stop count, sequence, and times all match exactly
  - Uses `tripScheduleOverlapScore()` for time window overlap
- **Calendar Strategy** (`strategy/calendar.go`):
  - `findFuzzyMatch()` scores calendars by date range overlap
  - Uses `calendarDateOverlapScore()` with interval overlap calculation
  - `parseGTFSDate()` helper for date parsing
- **Test Data**:
  - Created `testdata/fuzzy_similar/` with entities having different IDs but similar properties to `simple_a`
- **New Tests**:
  - 5 new tests for Stop fuzzy detection
  - 4 new tests for Route fuzzy detection
  - 5 new tests for Trip fuzzy detection (including validation rejection test)
  - 4 new tests for Calendar fuzzy detection
  - 5 new Java integration tests for fuzzy detection validation
- **Key Design Decision**: Implemented scoring logic directly in strategy files to avoid import cycle
  (scoring package imports strategy for MergeContext; strategy could not import scoring)
- All QA checks pass: gofmt, go vet, race detector
- Total: 287+ tests (without Java tag)

#### Milestone 11 - Auto-Detection of Strategy
- Created `strategy/autodetect.go` with auto-detection logic
- **AutoDetectConfig** struct with configurable thresholds:
  - `MinElementsInCommonScoreForAutoDetect` (default 0.5) - ID overlap threshold for Identity mode
  - `MinElementsDuplicateScoreForAutoDetect` (default 0.5) - Similarity threshold for Fuzzy mode
- **AutoDetectDuplicateDetection()** function analyzes source and target feeds:
  - Calculates ID overlap across agencies, stops, routes, trips, calendars
  - If any entity type has overlap >= threshold → DetectionIdentity
  - If no ID overlap but fuzzy similarity (names/locations) → DetectionFuzzy
  - Otherwise → DetectionNone
- **Fuzzy similarity detection** checks:
  - Agencies: matching name or URL
  - Stops: matching name AND within 500m (Haversine distance)
  - Routes: matching short_name or long_name
- Reuses existing `elementOverlapScore[T]` from route.go and `haversineDistance` from stop.go
- **Tests**:
  - 15 unit tests for auto-detection scenarios
  - 6 Java integration tests for auto-detection validation
- All QA checks pass: gofmt, go vet, race detector
- Total: 307+ tests (without Java tag)

#### Milestone 12 - Merger Configuration Options
- Completed implementation of functional options pattern in `merge/options.go`:
  - `WithDebug(bool)` - Enables debug output
  - `WithDefaultDetection(DuplicateDetection)` - Sets detection mode for all strategies
  - `WithDefaultLogging(DuplicateLogging)` - Sets logging mode for all strategies
  - `WithDefaultRenaming(RenamingStrategy)` - Sets renaming strategy for all strategies
- Per-strategy configuration via setters on Merger:
  - `SetAgencyStrategy()`, `SetStopStrategy()`, `SetRouteStrategy()`, etc.
  - `GetStrategyForFile(filename)` - Returns strategy for a GTFS file
  - `SetDuplicateDetectionForAll()` - Bulk configuration
- **Unit Tests** (`merge/options_test.go`):
  - `TestWithDebug` - Verifies debug flag setting
  - `TestWithDefaultDetection` - Tests all detection modes
  - `TestWithDefaultLogging` - Tests all logging modes
  - `TestWithDefaultRenaming` - Tests renaming strategies
  - `TestMultipleOptions` - Combined options test
  - `TestSetAgencyStrategy`, `TestSetStopStrategy` - Per-strategy setters
  - `TestCustomStrategyUsed` - Verifies custom strategy behavior
  - `TestAllStrategySetters` - All 12 strategy setters
  - `TestMixedStrategyConfiguration` - Global vs per-strategy config
- **Java Integration Tests** (`compare/compare_test.go`):
  - `TestConfigOptions_GoDebugModeProducesValidOutput`
  - `TestConfigOptions_CompareGoOptionsWithJava`
  - `TestConfigOptions_PerStrategyConfiguration`
  - `TestConfigOptions_JavaVsGoIdentityDetection`
  - `TestConfigOptions_CombinedOptionsWithJava`
  - `TestConfigOptions_StrategySettersProduceValidOutput`
- All QA checks pass: gofmt, go vet, race detector
- Total: 317+ tests (without Java tag)

#### Milestone 13 - CLI Application
- Created `cmd/gtfs-merge/` package with full CLI implementation
- **Argument Parsing** (`cmd/gtfs-merge/main.go`):
  - `parseArgs()` function with comprehensive flag parsing
  - Support for `--help`, `--version`, `--debug` flags
  - `--duplicateDetection=MODE` (none, identity, fuzzy)
  - `--logging=MODE` (none, warning, error)
  - `--file=FILENAME` for per-file strategy configuration
  - Proper validation and error handling
- **CLI End-to-End** (`cmd/gtfs-merge/main.go`):
  - `runMerge()` integrates with merge package options
  - Applies per-file configurations to strategies
  - Clear usage help with examples
- **Unit Tests** (`cmd/gtfs-merge/main_test.go`):
  - `TestParseArgsMinimum`, `TestParseArgsMultipleInputs`
  - `TestParseArgsWithOptions`, `TestParseArgsFileOption`
  - `TestParseArgsDuplicateDetection`, `TestParseArgsHelp`
  - `TestParseArgsVersion`, `TestParseArgsInvalid`
  - `TestCLIMergeTwoFeeds`, `TestCLIWithDuplicateDetection`
  - `TestCLIWithPerFileConfig`, `TestCLIDebugOutput`
  - `TestCLIErrorOnInvalidInput`, `TestCLIOutputFileSize`
  - `TestCLIMergeThreeFeeds`
- **Java Integration Tests** (`compare/compare_test.go`):
  - `TestCLI_JavaVsGoBasicMerge`
  - `TestCLI_JavaVsGoIdentityDetection`
  - `TestCLI_JavaVsGoThreeFeeds`
  - `TestCLI_GoValidOutputWithAllModes`
  - `TestCLI_JavaVsGoFuzzyDetection`
  - `TestCLI_OutputValidation`
- All QA checks pass: gofmt, go vet, race detector
- Total: 327 tests (without Java tag)

#### Milestone 14 - Integration Tests with Real Data
- Created edge case test fixtures in `testdata/`:
  - `unicode_feed/` - Feed with unicode characters in names (German, French, Japanese)
  - `large_ids_feed/` - Feed with very long ID strings (25+ characters)
  - `special_chars_feed/` - Feed with dots, dashes, underscores in IDs
  - `all_optional_feed/` - Feed with all optional GTFS files (shapes, transfers, frequencies, fares, feed_info)
- **14.1 Sample Feed Tests** (`compare/compare_test.go`):
  - `TestMergeRealWorldFeeds` - Tests 4 feed combinations (simple_a+b, simple+minimal, overlap, fuzzy)
  - `TestMergeLargeFeed` - Tests with large_ids_feed for performance verification
  - `TestMergeThreeFeedsVariations` - Tests 3 different three-feed combinations
  - `TestMergeFeedWithAllOptionalFiles` - Verifies shapes, transfers, frequencies, fares merge correctly
- **14.2 Edge Case Tests** (`compare/compare_test.go`):
  - `TestMergeEmptyOptionalFiles` - Tests merging when optional files are missing
  - `TestMergeMissingOptionalFiles` - Tests one feed with optionals, one without
  - `TestMergeUnicodeContent` - Verifies unicode preservation in stop names
  - `TestMergeLargeIDs` - Tests handling of very large ID strings
  - `TestMergeSpecialCharactersInIDs` - Tests dots, dashes, underscores in IDs
  - `TestMergeAllDetectionModesWithEdgeCases` - Tests all 3 detection modes × 3 edge case feeds (9 combinations)
  - `TestMergeFourFeeds` - Tests prefix assignment with 4 feeds
- All tests compare Go output with Java output where applicable
- All tests verify Go output passes validation
- All QA checks pass: gofmt, go vet, race detector
- Total: 327+ tests (without Java tag), 20+ new Java integration tests

#### Milestone 15 - Performance and Polish
- **15.1 Benchmarks** - Added comprehensive benchmark tests:
  - `gtfs/benchmark_test.go`: BenchmarkReadFeed, BenchmarkReadFeedFromZip, BenchmarkWriteFeed, BenchmarkReadLargeFeed, BenchmarkParsing
  - `merge/benchmark_test.go`: BenchmarkMergeTwoFeeds, BenchmarkMergeTwoFeedsWithIdentity, BenchmarkMergeTwoFeedsWithFuzzy, BenchmarkMergeThreeFeeds, BenchmarkMergeFiles, BenchmarkLargeFeedMerge, BenchmarkMergeWithAllOptionalFiles
  - `scoring/benchmark_test.go`: BenchmarkStopDistanceScorer, BenchmarkHaversineDistance, BenchmarkPropertyMatcher, BenchmarkAndScorer, BenchmarkElementOverlapScore, BenchmarkIntervalOverlapScore, BenchmarkServiceDateOverlapScorer
  - `strategy/benchmark_test.go`: BenchmarkFuzzyScoring, BenchmarkFuzzyScoringStops, BenchmarkFuzzyScoringRoutes, BenchmarkFuzzyScoringTrips, BenchmarkIdentityDetection, BenchmarkAutoDetect, BenchmarkAutoDetectWithConfig
- **15.2 Concurrent Fuzzy Matching**:
  - Created `strategy/concurrent.go` with goroutine-based parallel scoring
  - `ConcurrentConfig` struct with Enabled, NumWorkers, MinItemsForConcurrency settings
  - `findBestMatchConcurrent[T, ID]` generic function using worker pool pattern
  - Integrated into StopMergeStrategy and RouteMergeStrategy
  - `SetConcurrent(bool)` and `SetConcurrentWorkers(int)` methods for configuration
  - 9 unit tests for concurrent correctness and performance
  - 3 integration tests verifying concurrent produces same results as sequential
- **15.3 Documentation**:
  - Added package comments to merge/ and strategy/ packages
  - Updated README.md with comprehensive usage examples:
    - CLI usage examples
    - Library usage examples (basic, duplicate detection, direct feeds, concurrent)
    - Architecture overview and entity processing order
  - Created CONTRIBUTING.md with development guidelines
- All QA checks pass: gofmt, go vet, golangci-lint (0 issues), race detector
- Total: 340+ tests (without Java tag)
