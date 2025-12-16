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
    // RenameContext - use context prefix (a_, b_, c_, etc.)
    RenameContext RenamingStrategy = iota

    // RenameAgency - use agency-based naming
    RenameAgency
)
```

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

// MergeFiles merges multiple GTFS files into one output file
func (m *Merger) MergeFiles(inputPaths []string, outputPath string) error

// MergeFeeds merges multiple Feed objects into a single Feed
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

// AndScorer combines multiple scorers (all must score above threshold)
type AndScorer[T any] struct {
    Scorers   []Scorer[T]
    Threshold float64
}

func (a *AndScorer[T]) Score(ctx *strategy.MergeContext, source, target T) float64
```

### Specialized Scorers

```go
package scoring

// StopDistanceScorer scores stops by geographic proximity
type StopDistanceScorer struct {
    MaxDistanceMeters float64  // Default: 500m
}

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

// TripScheduleOverlapScorer scores trips by schedule similarity
type TripScheduleOverlapScorer struct {
    MaxTimeDiffSeconds int  // Default: 300 (5 minutes)
}

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
- **Detection**: Matches on `name`, geographic distance (within 500m default)
- **On duplicate**: Updates stop_times, transfers, and pathways
- **ID collision**: Renames stop ID in all references

### Route Merge
- **Detection**: Matches on `agency_id`, `short_name`, `long_name`, and shared stops
- **On duplicate**: Updates trips and fare_rules
- **ID collision**: Renames route ID in all references

### Trip Merge
- **Detection**: Matches on `route_id`, `service_id`, shared stops, schedule overlap
- **On duplicate**: Updates stop_times and frequencies
- **ID collision**: Renames trip ID in all references
- **Validation**: Rejects merge if stop times differ substantially

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
3. **Stops** - References: parent_station (self-referential)
4. **Service Calendars** - No dependencies (calendar.txt + calendar_dates.txt)
5. **Routes** - References: agency_id
6. **Shapes** - No dependencies
7. **Trips** - References: route_id, service_id, shape_id
8. **Stop Times** - References: trip_id, stop_id (handled with trips)
9. **Frequencies** - References: trip_id (handled with trips)
10. **Transfers** - References: from_stop_id, to_stop_id
11. **Fare Attributes** - References: agency_id
12. **Fare Rules** - References: fare_id, route_id
13. **Feed Info** - No dependencies
14. **Pathways** - References: from_stop_id, to_stop_id

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
