# GTFS Merge Go - API Specification

A Golang port of [onebusaway-gtfs-merge](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge) and [onebusaway-gtfs-merge-cli](https://github.com/OneBusAway/onebusaway-gtfs-modules/tree/main/onebusaway-gtfs-merge-cli).

## Overview

This library merges two or more static GTFS (General Transit Feed Specification) feeds into a single unified feed. It handles duplicate detection, entity renaming, and referential integrity across all GTFS entity types.

## GTFS Files Supported

### Required Files
- `agency.txt` - Transit agencies
- `stops.txt` - Stop locations
- `routes.txt` - Transit routes
- `trips.txt` - Trips for each route
- `stop_times.txt` - Arrival/departure times at stops
- `calendar.txt` - Service schedules (weekly patterns)
- `calendar_dates.txt` - Service exceptions

### Optional Files
- `fare_attributes.txt` - Fare pricing information
- `fare_rules.txt` - Fare rules for routes
- `shapes.txt` - Geographic path of routes
- `frequencies.txt` - Headway-based schedules
- `transfers.txt` - Transfer rules between stops
- `feed_info.txt` - Feed metadata
- `areas.txt` - Geographic areas
- `pathways.txt` - Station pathways

---

## Package Structure

```
gtfsmerge/
├── gtfs/                    # GTFS data model and I/O
│   ├── model.go            # Entity structs
│   ├── reader.go           # GTFS zip/directory reader
│   └── writer.go           # GTFS zip writer
├── merge/                   # Core merge functionality
│   ├── merger.go           # Main GtfsMerger type
│   ├── context.go          # GtfsMergeContext type
│   └── options.go          # Configuration options
├── strategy/                # Merge strategies
│   ├── strategy.go         # EntityMergeStrategy interface
│   ├── agency.go           # AgencyMergeStrategy
│   ├── stop.go             # StopMergeStrategy
│   ├── route.go            # RouteMergeStrategy
│   ├── trip.go             # TripMergeStrategy
│   ├── calendar.go         # ServiceCalendarMergeStrategy
│   ├── shape.go            # ShapePointMergeStrategy
│   ├── frequency.go        # FrequencyMergeStrategy
│   ├── transfer.go         # TransferMergeStrategy
│   ├── fare.go             # FareAttribute/FareRule strategies
│   ├── feedinfo.go         # FeedInfoMergeStrategy
│   └── area.go             # AreaMergeStrategy
├── scoring/                 # Duplicate scoring
│   ├── scorer.go           # DuplicateScoringStrategy interface
│   ├── stop_distance.go    # Geographic distance scoring
│   ├── route_stops.go      # Route stops in common
│   ├── trip_stops.go       # Trip stops in common
│   └── trip_schedule.go    # Schedule overlap scoring
└── cmd/                     # CLI application
    └── gtfs-merge/
        └── main.go
```

---

## Core Types

### GTFS Model (`gtfs/model.go`)

```go
package gtfs

// AgencyID is a unique identifier for an agency
type AgencyID string

// StopID is a unique identifier for a stop
type StopID string

// RouteID is a unique identifier for a route
type RouteID string

// TripID is a unique identifier for a trip
type TripID string

// ServiceID is a unique identifier for a service calendar
type ServiceID string

// ShapeID is a unique identifier for a shape
type ShapeID string

// FareID is a unique identifier for a fare attribute
type FareID string

// AreaID is a unique identifier for an area
type AreaID string

// Agency represents a transit agency (agency.txt)
type Agency struct {
    ID         AgencyID
    Name       string
    URL        string
    Timezone   string
    Lang       string
    Phone      string
    FareURL    string
    Email      string
}

// Stop represents a stop location (stops.txt)
type Stop struct {
    ID                 StopID
    Code               string
    Name               string
    Desc               string
    Lat                float64
    Lon                float64
    ZoneID             string
    URL                string
    LocationType       int
    ParentStation      StopID
    Timezone           string
    WheelchairBoarding int
    LevelID            string
    PlatformCode       string
}

// Route represents a transit route (routes.txt)
type Route struct {
    ID              RouteID
    AgencyID        AgencyID
    ShortName       string
    LongName        string
    Desc            string
    Type            int
    URL             string
    Color           string
    TextColor       string
    SortOrder       int
    ContinuousPickup   int
    ContinuousDropOff  int
}

// Trip represents a trip (trips.txt)
type Trip struct {
    ID                   TripID
    RouteID              RouteID
    ServiceID            ServiceID
    Headsign             string
    ShortName            string
    DirectionID          int
    BlockID              string
    ShapeID              ShapeID
    WheelchairAccessible int
    BikesAllowed         int
}

// StopTime represents a stop time (stop_times.txt)
type StopTime struct {
    TripID            TripID
    ArrivalTime       string  // HH:MM:SS format, can exceed 24:00:00
    DepartureTime     string
    StopID            StopID
    StopSequence      int
    StopHeadsign      string
    PickupType        int
    DropOffType       int
    ContinuousPickup  int
    ContinuousDropOff int
    ShapeDistTraveled float64
    Timepoint         int
}

// Calendar represents a service calendar (calendar.txt)
type Calendar struct {
    ServiceID ServiceID
    Monday    bool
    Tuesday   bool
    Wednesday bool
    Thursday  bool
    Friday    bool
    Saturday  bool
    Sunday    bool
    StartDate string  // YYYYMMDD format
    EndDate   string
}

// CalendarDate represents a calendar exception (calendar_dates.txt)
type CalendarDate struct {
    ServiceID     ServiceID
    Date          string  // YYYYMMDD format
    ExceptionType int     // 1=added, 2=removed
}

// ShapePoint represents a point in a shape (shapes.txt)
type ShapePoint struct {
    ShapeID      ShapeID
    Lat          float64
    Lon          float64
    Sequence     int
    DistTraveled float64
}

// Frequency represents frequency-based service (frequencies.txt)
type Frequency struct {
    TripID      TripID
    StartTime   string
    EndTime     string
    HeadwaySecs int
    ExactTimes  int
}

// Transfer represents a transfer rule (transfers.txt)
type Transfer struct {
    FromStopID      StopID
    ToStopID        StopID
    TransferType    int
    MinTransferTime int
}

// FareAttribute represents fare pricing (fare_attributes.txt)
type FareAttribute struct {
    FareID           FareID
    Price            float64
    CurrencyType     string
    PaymentMethod    int
    Transfers        int
    AgencyID         AgencyID
    TransferDuration int
}

// FareRule represents fare rules (fare_rules.txt)
type FareRule struct {
    FareID        FareID
    RouteID       RouteID
    OriginID      string
    DestinationID string
    ContainsID    string
}

// FeedInfo represents feed metadata (feed_info.txt)
type FeedInfo struct {
    PublisherName string
    PublisherURL  string
    Lang          string
    DefaultLang   string
    StartDate     string
    EndDate       string
    Version       string
    ContactEmail  string
    ContactURL    string
}

// Area represents a geographic area (areas.txt)
type Area struct {
    ID   AreaID
    Name string
}

// Pathway represents a station pathway (pathways.txt)
type Pathway struct {
    ID                   string
    FromStopID           StopID
    ToStopID             StopID
    PathwayMode          int
    IsBidirectional      int
    Length               float64
    TraversalTime        int
    StairCount           int
    MaxSlope             float64
    MinWidth             float64
    SignpostedAs         string
    ReversedSignpostedAs string
}
```

### Feed Container (`gtfs/feed.go`)

```go
package gtfs

// Feed represents a complete GTFS feed
type Feed struct {
    Agencies       map[AgencyID]*Agency
    Stops          map[StopID]*Stop
    Routes         map[RouteID]*Route
    Trips          map[TripID]*Trip
    StopTimes      []*StopTime  // Keyed by TripID+Sequence
    Calendars      map[ServiceID]*Calendar
    CalendarDates  map[ServiceID][]*CalendarDate
    Shapes         map[ShapeID][]*ShapePoint
    Frequencies    []*Frequency
    Transfers      []*Transfer
    FareAttributes map[FareID]*FareAttribute
    FareRules      []*FareRule
    FeedInfo       *FeedInfo
    Areas          map[AreaID]*Area
    Pathways       []*Pathway
}

// NewFeed creates an empty feed
func NewFeed() *Feed

// Clone creates a deep copy of the feed
func (f *Feed) Clone() *Feed

// Validate checks the feed for referential integrity
func (f *Feed) Validate() []error

// GetStopTimesForTrip returns stop times for a specific trip
func (f *Feed) GetStopTimesForTrip(tripID TripID) []*StopTime

// GetTripsForRoute returns all trips for a specific route
func (f *Feed) GetTripsForRoute(routeID RouteID) []*Trip

// GetTripsForServiceID returns all trips for a specific service
func (f *Feed) GetTripsForServiceID(serviceID ServiceID) []*Trip

// GetServiceDates returns all active dates for a service ID
func (f *Feed) GetServiceDates(serviceID ServiceID) []string
```

### Feed I/O (`gtfs/reader.go`, `gtfs/writer.go`)

```go
package gtfs

import "io"

// ReadFromPath reads a GTFS feed from a file path (zip or directory)
func ReadFromPath(path string) (*Feed, error)

// ReadFromZip reads a GTFS feed from a zip file reader
func ReadFromZip(r io.ReaderAt, size int64) (*Feed, error)

// WriteToPath writes a GTFS feed to a zip file
func WriteToPath(feed *Feed, path string) error

// WriteToZip writes a GTFS feed to a zip writer
func WriteToZip(feed *Feed, w io.Writer) error
```

---

## Merge Configuration

### Duplicate Detection Strategy (`strategy/detection.go`)

```go
package strategy

// DuplicateDetection specifies how duplicates are detected
type DuplicateDetection int

const (
    // DetectionNone - entities are never considered duplicates
    DetectionNone DuplicateDetection = iota

    // DetectionIdentity - entities with same ID are duplicates
    DetectionIdentity

    // DetectionFuzzy - entities with similar properties are duplicates
    DetectionFuzzy
)
```

### Duplicate Logging Strategy (`strategy/logging.go`)

```go
package strategy

// DuplicateLogging specifies how to handle detected duplicates
type DuplicateLogging int

const (
    // LogNone - no logging when duplicates are detected
    LogNone DuplicateLogging = iota

    // LogWarning - log a warning when duplicates are detected
    LogWarning

    // LogError - return an error when duplicates are detected
    LogError
)
```

### Renaming Strategy (`strategy/renaming.go`)

```go
package strategy

// RenamingStrategy specifies how duplicate IDs are renamed
type RenamingStrategy int

const (
    // RenameContext - use context prefix (a-, b-, c-, etc. for up to 26 feeds, or 00-, 01-, etc. for more)
    RenameContext RenamingStrategy = iota

    // RenameAgency - use agency-based naming
    RenameAgency
)
```

### Auto-Detection Thresholds

By default, no duplicate detection strategy is specified, triggering **auto-detection**.
The system analyzes source and target feeds to pick the best strategy (IDENTITY, FUZZY, or NONE).

The following thresholds control auto-detection behavior (all default to 0.5):

| Threshold | Default | Purpose |
|-----------|---------|---------|
| `MinElementsInCommonScoreForAutoDetect` | 0.5 | ID overlap score needed to consider IDENTITY mode |
| `MinElementsDuplicateScoreForAutoDetect` | 0.5 | Entity match score for strategy selection |
| `MinElementDuplicateScoreForFuzzyMatch` | 0.5 | Candidate filtering for fuzzy matching |

---

## Merge Strategy Interface (`strategy/strategy.go`)

```go
package strategy

import (
    "github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// MergeContext provides context during merge operations
type MergeContext struct {
    // Source is the feed being merged into the target
    Source *gtfs.Feed

    // Target is the output feed being built
    Target *gtfs.Feed

    // Prefix is the current feed's prefix (e.g., "a_", "b_")
    Prefix string

    // EntityByRawID tracks entities by their original IDs
    EntityByRawID map[string]interface{}

    // ResolvedDetection is the auto-detected or configured strategy
    ResolvedDetection DuplicateDetection
}

// EntityMergeStrategy defines the interface for entity-specific merge logic
type EntityMergeStrategy interface {
    // Name returns a human-readable name for this strategy
    Name() string

    // Merge performs the merge operation for this entity type
    Merge(ctx *MergeContext) error

    // SetDuplicateDetection configures duplicate detection
    SetDuplicateDetection(d DuplicateDetection)

    // SetDuplicateLogging configures duplicate logging behavior
    SetDuplicateLogging(l DuplicateLogging)

    // SetRenamingStrategy configures ID renaming behavior
    SetRenamingStrategy(r RenamingStrategy)
}
```

---

## Main Merger API (`merge/merger.go`)

```go
package merge

import (
    "github.com/aaronbrethorst/gtfs-merge-go/gtfs"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// Merger orchestrates the merging of multiple GTFS feeds
type Merger struct {
    // Strategy configurations
    agencyStrategy   strategy.EntityMergeStrategy
    stopStrategy     strategy.EntityMergeStrategy
    routeStrategy    strategy.EntityMergeStrategy
    tripStrategy     strategy.EntityMergeStrategy
    calendarStrategy strategy.EntityMergeStrategy
    shapeStrategy    strategy.EntityMergeStrategy
    frequencyStrategy strategy.EntityMergeStrategy
    transferStrategy strategy.EntityMergeStrategy
    fareAttrStrategy strategy.EntityMergeStrategy
    fareRuleStrategy strategy.EntityMergeStrategy
    feedInfoStrategy strategy.EntityMergeStrategy
    areaStrategy     strategy.EntityMergeStrategy

    // Options
    debug bool
}

// New creates a new Merger with default strategies
func New(opts ...Option) *Merger

// MergeFiles merges multiple GTFS files into one output file.
// IMPORTANT: Input feeds are processed in REVERSE order (newest/last first).
// This ensures entities from newer feeds are added first and older duplicates
// are potentially dropped.
func (m *Merger) MergeFiles(inputPaths []string, outputPath string) error

// MergeFeeds merges multiple Feed objects into a single Feed.
// IMPORTANT: Feeds are processed in REVERSE order (last element first).
func (m *Merger) MergeFeeds(feeds []*gtfs.Feed) (*gtfs.Feed, error)

// Strategy setters for customization
func (m *Merger) SetAgencyStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetStopStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetRouteStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetTripStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetCalendarStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetShapeStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetFrequencyStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetTransferStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetFareAttributeStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetFareRuleStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetFeedInfoStrategy(s strategy.EntityMergeStrategy)
func (m *Merger) SetAreaStrategy(s strategy.EntityMergeStrategy)

// GetStrategyForFile returns the strategy for a specific GTFS file
func (m *Merger) GetStrategyForFile(filename string) strategy.EntityMergeStrategy
```

### Merger Options (`merge/options.go`)

```go
package merge

// Option configures a Merger
type Option func(*Merger)

// WithDebug enables debug output
func WithDebug(debug bool) Option

// WithDefaultDetection sets default duplicate detection for all strategies
func WithDefaultDetection(d strategy.DuplicateDetection) Option

// WithDefaultLogging sets default duplicate logging for all strategies
func WithDefaultLogging(l strategy.DuplicateLogging) Option

// WithDefaultRenaming sets default renaming strategy for all strategies
func WithDefaultRenaming(r strategy.RenamingStrategy) Option
```

---

## Duplicate Scoring (`scoring/scorer.go`)

```go
package scoring

import (
    "github.com/aaronbrethorst/gtfs-merge-go/gtfs"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// Scorer calculates similarity between entities
type Scorer[T any] interface {
    // Score returns a similarity score between 0.0 and 1.0
    // 0.0 = completely different, 1.0 = identical
    Score(ctx *strategy.MergeContext, source, target T) float64
}

// PropertyMatcher scores based on matching property values
type PropertyMatcher[T any] struct {
    Properties []func(T) string
}

func (p *PropertyMatcher[T]) Score(ctx *strategy.MergeContext, source, target T) float64

// AndScorer combines multiple scorers using MULTIPLICATION.
// Final score = scorer1 * scorer2 * ... * scorerN
// A single 0.0 score fails the entire match (early exit optimization).
// All entity merge strategies use AndScorer to combine their scoring rules.
type AndScorer[T any] struct {
    Scorers   []Scorer[T]
    Threshold float64
}

func (a *AndScorer[T]) Score(ctx *strategy.MergeContext, source, target T) float64
```

### Scoring Formulas

The following formulas are used for overlap calculations:

**Element Overlap** (for sets of IDs, stops, dates):
```
score = (common_count / a.size + common_count / b.size) / 2
```
Returns 0.0 if either collection is empty. Returns 1.0 for identical sets.

**Interval Overlap** (for time windows):
```
overlap = max(0, min(end1, end2) - max(start1, start2))
score = (overlap / interval_a_length + overlap / interval_b_length) / 2
```

### Specialized Scorers

```go
package scoring

// StopDistanceScorer scores stops by geographic proximity using great-circle distance.
// Uses hardcoded tiered thresholds (not configurable):
//   - < 50m  → 1.0
//   - < 100m → 0.75
//   - < 500m → 0.5
//   - >= 500m → 0.0
type StopDistanceScorer struct{}

func (s *StopDistanceScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Stop) float64

// RouteStopsInCommonScorer scores routes by shared stops
type RouteStopsInCommonScorer struct {
    MinCommonStops int
}

func (r *RouteStopsInCommonScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Route) float64

// TripStopsInCommonScorer scores trips by shared stops
type TripStopsInCommonScorer struct {
    MinCommonStops int
}

func (t *TripStopsInCommonScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Trip) float64

// TripScheduleOverlapScorer scores trips by schedule similarity.
// Computes overlap of time windows [first_stop_departure, last_stop_arrival].
// Uses interval overlap formula: (overlap/interval_a + overlap/interval_b) / 2
type TripScheduleOverlapScorer struct{}

func (t *TripScheduleOverlapScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Trip) float64

// ServiceDateOverlapScorer scores service calendars by date overlap
type ServiceDateOverlapScorer struct{}

func (s *ServiceDateOverlapScorer) Score(ctx *strategy.MergeContext, source, target gtfs.ServiceID) float64
```

---

## CLI Application

### Usage

```
gtfs-merge [OPTIONS] INPUT1 INPUT2 [INPUT...] OUTPUT

Arguments:
  INPUT...    Input GTFS files (zip) or directories (minimum 2)
  OUTPUT      Output GTFS file path (zip)

Options:
  --file FILENAME              Specify which GTFS file to configure
  --duplicate-detection MODE   Set duplicate detection: none, identity, fuzzy
  --log-duplicates            Enable warning logging for dropped duplicates
  --error-on-duplicates       Raise error when duplicates are dropped
  --debug                     Show detailed configuration before merging
  -h, --help                  Show help message
  -v, --version               Show version

Examples:
  # Basic merge of two feeds
  gtfs-merge feed1.zip feed2.zip merged.zip

  # Merge with identity-based duplicate detection
  gtfs-merge --duplicate-detection identity feed1.zip feed2.zip merged.zip

  # Configure specific file handling
  gtfs-merge --file stops.txt --duplicate-detection fuzzy \
             --file routes.txt --duplicate-detection identity \
             feed1.zip feed2.zip merged.zip

  # Merge with duplicate warnings
  gtfs-merge --log-duplicates feed1.zip feed2.zip merged.zip

  # Fail on duplicates
  gtfs-merge --error-on-duplicates feed1.zip feed2.zip merged.zip
```

### CLI Implementation (`cmd/gtfs-merge/main.go`)

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/aaronbrethorst/gtfs-merge-go/merge"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

func main() {
    // Parse flags and arguments
    // ...

    merger := merge.New(
        merge.WithDebug(debug),
    )

    // Apply per-file configurations
    // ...

    if err := merger.MergeFiles(inputs, output); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Merged %d feeds into %s\n", len(inputs), output)
}
```

---

## Entity Merge Behavior

### Agency Merge
- **Detection**: Matches on `name` and `url`
- **On duplicate**: Updates all routes to reference the merged agency
- **ID collision**: Renames agency ID across all referencing entities

### Stop Merge
- **Detection**: Matches on `name` AND geographic distance (combined multiplicatively)
  - PropertyMatch("name") * StopDistanceScorer
- **On duplicate**: Updates stop_times, transfers, and pathways
- **ID collision**: Renames stop ID in all references

### Route Merge
- **Detection**: Matches on `agency`, `short_name`, `long_name`, AND shared stops (all combined multiplicatively)
  - PropertyMatch("agency") * PropertyMatch("shortName") * PropertyMatch("longName") * RouteStopsInCommonScorer
- **On duplicate**: Updates trips and fare_rules
- **ID collision**: Renames route ID in all references

### Trip Merge
- **Detection**: Matches on `route`, `service_id`, shared stops, schedule overlap (all combined multiplicatively)
  - PropertyMatch("route") * PropertyMatch("serviceId") * TripStopsInCommonScorer * TripScheduleOverlapScorer
- **On duplicate**: Updates stop_times and frequencies
- **ID collision**: Renames trip ID in all references
- **Validation**: Rejects merge if:
  - Stop count differs between trips
  - Any stop at the same sequence position differs
  - Any arrival or departure time differs (exact integer comparison, not fuzzy)

### Service Calendar Merge
- **Detection**: Matches on date overlap
- **On duplicate**: Merges date ranges
- **ID collision**: Renames service ID in calendars, calendar_dates, and trips

### Shape Merge
- **Detection**: Identity-based only (TODO: geographic similarity)
- **On duplicate**: Uses existing shape
- **ID collision**: Renames shape ID in trips and shape_points

### Frequency Merge
- **Detection**: Matches on `trip_id`, `start_time`, `end_time`, `headway_secs`
- **On duplicate**: Drops duplicate frequency

### Transfer Merge
- **Detection**: Matches on `from_stop_id`, `to_stop_id`, `transfer_type`, `min_transfer_time`
- **On duplicate**: Drops duplicate transfer

### Fare Attribute Merge
- **Detection**: Matches on fare properties
- **On duplicate**: Updates fare_rules
- **ID collision**: Renames fare ID

### Fare Rule Merge
- **Detection**: Matches on all fields
- **On duplicate**: Drops duplicate rule

### Feed Info Merge
- **Detection**: Matches on `publisher_name`, `publisher_url`, `lang`
- **On duplicate**: Merges versions, takes earliest start/latest end dates

### Area Merge
- **Detection**: Matches on `name`
- **On duplicate**: Uses existing area
- **ID collision**: Renames area ID

---

## Processing Order

Entities are merged in dependency order to maintain referential integrity:

1. **Agencies** - No dependencies
2. **Areas** - No dependencies
3. **Stops** - References: parent_station (self-referential). Also updates Transfers and Pathways when stops merge.
4. **Service Calendars** - No dependencies (calendar.txt + calendar_dates.txt)
5. **Routes** - References: agency_id
6. **Trips** - References: route_id, service_id, shape_id. Also handles StopTimes.
7. **Shapes** - No dependencies (shape_id collection)
8. **Frequencies** - References: trip_id
9. **Transfers** - References: from_stop_id, to_stop_id
10. **Fare Attributes** - References: agency_id
11. **Fare Rules** - References: fare_id, route_id
12. **Feed Info** - No dependencies

Note: Pathways are updated within StopMergeStrategy when stops are merged, not as a separate processing step.

---

## Error Handling

```go
package merge

import "errors"

var (
    // ErrNoInputFeeds indicates no input feeds were provided
    ErrNoInputFeeds = errors.New("at least one input feed is required")

    // ErrInvalidFeed indicates a feed failed validation
    ErrInvalidFeed = errors.New("invalid GTFS feed")

    // ErrDuplicateDetected indicates a duplicate was found when error mode is enabled
    ErrDuplicateDetected = errors.New("duplicate entity detected")

    // ErrMergeConflict indicates entities cannot be safely merged
    ErrMergeConflict = errors.New("merge conflict detected")
)

// MergeError provides detailed error information
type MergeError struct {
    EntityType string
    EntityID   string
    Message    string
    Cause      error
}

func (e *MergeError) Error() string
func (e *MergeError) Unwrap() error
```

---

## Example Usage

### Basic Merge

```go
package main

import (
    "log"
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
)

func main() {
    merger := merge.New()

    err := merger.MergeFiles(
        []string{"feed1.zip", "feed2.zip"},
        "merged.zip",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Custom Configuration

```go
package main

import (
    "log"
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
    "github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

func main() {
    merger := merge.New(
        merge.WithDebug(true),
        merge.WithDefaultDetection(strategy.DetectionFuzzy),
        merge.WithDefaultLogging(strategy.LogWarning),
    )

    // Configure specific strategies
    stopStrategy := strategy.NewStopMergeStrategy()
    stopStrategy.SetDuplicateDetection(strategy.DetectionFuzzy)
    merger.SetStopStrategy(stopStrategy)

    err := merger.MergeFiles(
        []string{"feed1.zip", "feed2.zip", "feed3.zip"},
        "merged.zip",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Working with Feed Objects

```go
package main

import (
    "log"
    "github.com/aaronbrethorst/gtfs-merge-go/gtfs"
    "github.com/aaronbrethorst/gtfs-merge-go/merge"
)

func main() {
    // Read feeds
    feed1, err := gtfs.ReadFromPath("feed1.zip")
    if err != nil {
        log.Fatal(err)
    }

    feed2, err := gtfs.ReadFromPath("feed2.zip")
    if err != nil {
        log.Fatal(err)
    }

    // Merge
    merger := merge.New()
    merged, err := merger.MergeFeeds([]*gtfs.Feed{feed1, feed2})
    if err != nil {
        log.Fatal(err)
    }

    // Validate
    if errs := merged.Validate(); len(errs) > 0 {
        for _, e := range errs {
            log.Printf("Validation warning: %v", e)
        }
    }

    // Write
    err = gtfs.WriteToPath(merged, "merged.zip")
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## Implementation Notes

### Differences from Java Version

1. **No external dependencies** - Pure Go implementation
2. **Concurrent processing** - Uses goroutines for parallel scoring
3. **Generics** - Uses Go generics for type-safe strategies
4. **Functional options** - Uses functional options pattern for configuration
5. **Error handling** - Uses Go-style error returns instead of exceptions

### Performance Considerations

1. **Memory**: Large feeds are loaded entirely into memory
2. **Concurrency**: Fuzzy matching uses worker pools for parallel scoring
3. **I/O**: Uses buffered readers/writers for zip operations

### Testing Strategy

1. Unit tests for each merge strategy
2. Integration tests with sample GTFS feeds
3. Fuzzing tests for CSV parsing
4. Benchmark tests for large feeds

---

## Implementation Milestones

All milestones follow **Test-Driven Development (TDD)**:
1. Write tests that define expected behavior
2. Run tests and observe them fail (red)
3. Write minimal code to make tests pass (green)
4. Refactor while keeping tests passing (refactor)

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

## Milestone Tracking

This section tracks completed milestones with feedback and notes.

### Completed Milestones

| Milestone | Status | Commit | Notes |
|-----------|--------|--------|-------|
| 1.1 Initialize Go Module | ✅ Complete | `12a9687` | Created `go.mod` with module path `github.com/aaronbrethorst/gtfs-merge-go` |
| 1.1.1 Set Up GitHub Actions CI | ✅ Complete | `2350017` | Added `.github/workflows/ci.yml` with lint, test, fmt, vet jobs |
| 1.1.2 Quality Assurance Process | ✅ Complete | `1852d40` | Defined 5-step QA process, added milestone tracking section |
| 1.2 Define GTFS Entity Types | ✅ Complete | `a7c295a` | All 15 entity structs + 8 ID types, 16 tests passing |

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
- Added milestone tracking section at end of spec.md
- Code review suggested optional improvements (failure handling, date column) for future consideration
- This QA process will be applied to all future milestones

#### Milestone 1.2 - Define GTFS Entity Types
- Created `gtfs/model.go` with all 15 GTFS entity structs
- Defined 8 type aliases for IDs (AgencyID, StopID, RouteID, etc.)
- TDD approach: wrote reflection-based tests first, then implemented structs
- All tests verify struct fields match the GTFS specification exactly
- 16 tests total, all passing with race detector
