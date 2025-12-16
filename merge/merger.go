// Package merge provides the core GTFS feed merge orchestration.
// It handles merging multiple GTFS feeds into a single unified feed,
// with configurable duplicate detection, entity renaming, and referential integrity.
package merge

import (
	"errors"
	"fmt"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// ErrNoInputFeeds indicates no input feeds were provided
var ErrNoInputFeeds = errors.New("at least one input feed is required")

// Merger orchestrates the merging of multiple GTFS feeds
type Merger struct {
	// Strategy configurations
	agencyStrategy       strategy.EntityMergeStrategy
	areaStrategy         strategy.EntityMergeStrategy
	stopStrategy         strategy.EntityMergeStrategy
	calendarStrategy     strategy.EntityMergeStrategy
	calendarDateStrategy strategy.EntityMergeStrategy
	routeStrategy        strategy.EntityMergeStrategy
	shapeStrategy        strategy.EntityMergeStrategy
	tripStrategy         strategy.EntityMergeStrategy
	stopTimeStrategy     strategy.EntityMergeStrategy
	frequencyStrategy    strategy.EntityMergeStrategy
	transferStrategy     strategy.EntityMergeStrategy
	pathwayStrategy      strategy.EntityMergeStrategy
	fareAttrStrategy     strategy.EntityMergeStrategy
	fareRuleStrategy     strategy.EntityMergeStrategy
	feedInfoStrategy     strategy.EntityMergeStrategy

	// Options
	debug bool
}

// New creates a new Merger with default strategies
func New(opts ...Option) *Merger {
	m := &Merger{
		agencyStrategy:       strategy.NewAgencyMergeStrategy(),
		areaStrategy:         strategy.NewAreaMergeStrategy(),
		stopStrategy:         strategy.NewStopMergeStrategy(),
		calendarStrategy:     strategy.NewCalendarMergeStrategy(),
		calendarDateStrategy: strategy.NewCalendarDateMergeStrategy(),
		routeStrategy:        strategy.NewRouteMergeStrategy(),
		shapeStrategy:        strategy.NewShapeMergeStrategy(),
		tripStrategy:         strategy.NewTripMergeStrategy(),
		stopTimeStrategy:     strategy.NewStopTimeMergeStrategy(),
		frequencyStrategy:    strategy.NewFrequencyMergeStrategy(),
		transferStrategy:     strategy.NewTransferMergeStrategy(),
		pathwayStrategy:      strategy.NewPathwayMergeStrategy(),
		fareAttrStrategy:     strategy.NewFareAttributeMergeStrategy(),
		fareRuleStrategy:     strategy.NewFareRuleMergeStrategy(),
		feedInfoStrategy:     strategy.NewFeedInfoMergeStrategy(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// MergeFiles merges multiple GTFS files into one output file.
// IMPORTANT: Input feeds are processed in REVERSE order (newest/last first).
// This ensures entities from newer feeds are added first and older duplicates
// are potentially dropped.
func (m *Merger) MergeFiles(inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 {
		return ErrNoInputFeeds
	}

	// Read all feeds
	feeds := make([]*gtfs.Feed, 0, len(inputPaths))
	for _, path := range inputPaths {
		feed, err := gtfs.ReadFromPath(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		feeds = append(feeds, feed)
	}

	// Merge feeds
	merged, err := m.MergeFeeds(feeds)
	if err != nil {
		return err
	}

	// Write output
	return gtfs.WriteToPath(merged, outputPath)
}

// MergeFeeds merges multiple Feed objects into a single Feed.
// IMPORTANT: Feeds are processed in REVERSE order (last element first).
func (m *Merger) MergeFeeds(feeds []*gtfs.Feed) (*gtfs.Feed, error) {
	if len(feeds) == 0 {
		return nil, ErrNoInputFeeds
	}

	// Start with an empty target feed
	target := gtfs.NewFeed()

	// Process feeds in reverse order (last feed first, which gets no prefix)
	for i := len(feeds) - 1; i >= 0; i-- {
		feedIndex := len(feeds) - 1 - i // 0 for last feed, 1 for second-to-last, etc.
		prefix := GetPrefixForIndex(feedIndex)

		ctx := strategy.NewMergeContext(feeds[i], target, prefix)

		if err := m.mergeFeed(ctx); err != nil {
			return nil, fmt.Errorf("merging feed %d: %w", i, err)
		}
	}

	return target, nil
}

// mergeFeed merges a single source feed into the target
func (m *Merger) mergeFeed(ctx *strategy.MergeContext) error {
	// Merge entities in dependency order:
	// 1. Agencies (no dependencies)
	if err := m.agencyStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging agencies: %w", err)
	}

	// 2. Areas (no dependencies)
	if err := m.areaStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging areas: %w", err)
	}

	// 3. Stops (references: parent_station)
	if err := m.stopStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging stops: %w", err)
	}

	// 4. Service Calendars (no dependencies)
	if err := m.calendarStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging calendars: %w", err)
	}
	if err := m.calendarDateStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging calendar_dates: %w", err)
	}

	// 5. Routes (references: agency_id)
	if err := m.routeStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging routes: %w", err)
	}

	// 6. Shapes (no dependencies)
	if err := m.shapeStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging shapes: %w", err)
	}

	// 7. Trips (references: route_id, service_id, shape_id)
	if err := m.tripStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging trips: %w", err)
	}

	// 8. Stop Times (references: trip_id, stop_id)
	if err := m.stopTimeStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging stop_times: %w", err)
	}

	// 9. Frequencies (references: trip_id)
	if err := m.frequencyStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging frequencies: %w", err)
	}

	// 10. Transfers (references: from_stop_id, to_stop_id)
	if err := m.transferStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging transfers: %w", err)
	}

	// 11. Pathways (references: from_stop_id, to_stop_id)
	if err := m.pathwayStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging pathways: %w", err)
	}

	// 12. Fare Attributes (references: agency_id)
	if err := m.fareAttrStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging fare_attributes: %w", err)
	}

	// 13. Fare Rules (references: fare_id, route_id)
	if err := m.fareRuleStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging fare_rules: %w", err)
	}

	// 14. Feed Info (no dependencies)
	if err := m.feedInfoStrategy.Merge(ctx); err != nil {
		return fmt.Errorf("merging feed_info: %w", err)
	}

	return nil
}

// Strategy setters for customization

// SetAgencyStrategy sets the agency merge strategy
func (m *Merger) SetAgencyStrategy(s strategy.EntityMergeStrategy) {
	m.agencyStrategy = s
}

// SetStopStrategy sets the stop merge strategy
func (m *Merger) SetStopStrategy(s strategy.EntityMergeStrategy) {
	m.stopStrategy = s
}

// SetRouteStrategy sets the route merge strategy
func (m *Merger) SetRouteStrategy(s strategy.EntityMergeStrategy) {
	m.routeStrategy = s
}

// SetTripStrategy sets the trip merge strategy
func (m *Merger) SetTripStrategy(s strategy.EntityMergeStrategy) {
	m.tripStrategy = s
}

// SetCalendarStrategy sets the calendar merge strategy
func (m *Merger) SetCalendarStrategy(s strategy.EntityMergeStrategy) {
	m.calendarStrategy = s
}

// SetShapeStrategy sets the shape merge strategy
func (m *Merger) SetShapeStrategy(s strategy.EntityMergeStrategy) {
	m.shapeStrategy = s
}

// SetFrequencyStrategy sets the frequency merge strategy
func (m *Merger) SetFrequencyStrategy(s strategy.EntityMergeStrategy) {
	m.frequencyStrategy = s
}

// SetTransferStrategy sets the transfer merge strategy
func (m *Merger) SetTransferStrategy(s strategy.EntityMergeStrategy) {
	m.transferStrategy = s
}

// SetFareAttributeStrategy sets the fare attribute merge strategy
func (m *Merger) SetFareAttributeStrategy(s strategy.EntityMergeStrategy) {
	m.fareAttrStrategy = s
}

// SetFareRuleStrategy sets the fare rule merge strategy
func (m *Merger) SetFareRuleStrategy(s strategy.EntityMergeStrategy) {
	m.fareRuleStrategy = s
}

// SetFeedInfoStrategy sets the feed info merge strategy
func (m *Merger) SetFeedInfoStrategy(s strategy.EntityMergeStrategy) {
	m.feedInfoStrategy = s
}

// SetAreaStrategy sets the area merge strategy
func (m *Merger) SetAreaStrategy(s strategy.EntityMergeStrategy) {
	m.areaStrategy = s
}

// GetStrategyForFile returns the strategy for a specific GTFS file
func (m *Merger) GetStrategyForFile(filename string) strategy.EntityMergeStrategy {
	switch filename {
	case "agency.txt":
		return m.agencyStrategy
	case "areas.txt":
		return m.areaStrategy
	case "stops.txt":
		return m.stopStrategy
	case "calendar.txt":
		return m.calendarStrategy
	case "calendar_dates.txt":
		return m.calendarDateStrategy
	case "routes.txt":
		return m.routeStrategy
	case "shapes.txt":
		return m.shapeStrategy
	case "trips.txt":
		return m.tripStrategy
	case "stop_times.txt":
		return m.stopTimeStrategy
	case "frequencies.txt":
		return m.frequencyStrategy
	case "transfers.txt":
		return m.transferStrategy
	case "pathways.txt":
		return m.pathwayStrategy
	case "fare_attributes.txt":
		return m.fareAttrStrategy
	case "fare_rules.txt":
		return m.fareRuleStrategy
	case "feed_info.txt":
		return m.feedInfoStrategy
	default:
		return nil
	}
}

// SetDuplicateDetectionForAll sets duplicate detection for all strategies
func (m *Merger) SetDuplicateDetectionForAll(d strategy.DuplicateDetection) {
	m.agencyStrategy.SetDuplicateDetection(d)
	m.areaStrategy.SetDuplicateDetection(d)
	m.stopStrategy.SetDuplicateDetection(d)
	m.calendarStrategy.SetDuplicateDetection(d)
	m.calendarDateStrategy.SetDuplicateDetection(d)
	m.routeStrategy.SetDuplicateDetection(d)
	m.shapeStrategy.SetDuplicateDetection(d)
	m.tripStrategy.SetDuplicateDetection(d)
	m.stopTimeStrategy.SetDuplicateDetection(d)
	m.frequencyStrategy.SetDuplicateDetection(d)
	m.transferStrategy.SetDuplicateDetection(d)
	m.pathwayStrategy.SetDuplicateDetection(d)
	m.fareAttrStrategy.SetDuplicateDetection(d)
	m.fareRuleStrategy.SetDuplicateDetection(d)
	m.feedInfoStrategy.SetDuplicateDetection(d)
}
