// Package strategy provides entity-specific merge strategies for GTFS feeds.
// It defines the EntityMergeStrategy interface and implementations for all GTFS
// entity types, supporting identity-based and fuzzy duplicate detection with
// configurable logging and renaming options.
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

	// ID mappings from source IDs to target IDs
	AgencyIDMapping  map[gtfs.AgencyID]gtfs.AgencyID
	StopIDMapping    map[gtfs.StopID]gtfs.StopID
	RouteIDMapping   map[gtfs.RouteID]gtfs.RouteID
	TripIDMapping    map[gtfs.TripID]gtfs.TripID
	ServiceIDMapping map[gtfs.ServiceID]gtfs.ServiceID
	ShapeIDMapping   map[gtfs.ShapeID]gtfs.ShapeID
	FareIDMapping    map[gtfs.FareID]gtfs.FareID
	AreaIDMapping    map[gtfs.AreaID]gtfs.AreaID

	// JustAddedStops tracks stop IDs added in the current feed.
	// Used to prevent within-feed fuzzy matching (matches Java behavior).
	JustAddedStops map[gtfs.StopID]struct{}

	// JustAddedRoutes tracks route IDs added in the current feed.
	// Used to prevent within-feed fuzzy matching (matches Java behavior).
	JustAddedRoutes map[gtfs.RouteID]struct{}

	// ShapeSequenceCounter is a counter for shape point sequences within this context.
	// Deprecated: Use sharedShapeCounter for multi-feed merges to match Java behavior.
	ShapeSequenceCounter int

	// sharedShapeCounter points to a counter that persists across all feeds
	// in a single merge operation. Used for shape point sequence numbering
	// to match Java's behavior of globally incrementing sequences.
	sharedShapeCounter *int
}

// NextShapeSequence returns the next shape point sequence number.
// This mimics Java's behavior where all shape points get globally incrementing
// sequence numbers rather than preserving original sequences.
func (ctx *MergeContext) NextShapeSequence() int {
	if ctx.sharedShapeCounter != nil {
		*ctx.sharedShapeCounter++
		return *ctx.sharedShapeCounter
	}
	// Fallback for single-context usage
	ctx.ShapeSequenceCounter++
	return ctx.ShapeSequenceCounter
}

// SetSharedShapeCounter sets a shared counter for shape sequences that persists
// across multiple merge contexts. This is used to match Java's behavior where
// shape point sequences increment globally across all feeds in a merge.
func (ctx *MergeContext) SetSharedShapeCounter(counter *int) {
	ctx.sharedShapeCounter = counter
}

// NewMergeContext creates a new merge context.
// If the source feed has empty order slices but non-empty maps, SyncOrderSlices
// is called to populate them. This supports test code that uses direct map assignments.
func NewMergeContext(source, target *gtfs.Feed, prefix string) *MergeContext {
	// Sync order slices if they're empty but maps have data.
	// This supports test code that populates maps directly.
	if len(source.AgencyOrder) == 0 && len(source.Agencies) > 0 ||
		len(source.StopOrder) == 0 && len(source.Stops) > 0 ||
		len(source.RouteOrder) == 0 && len(source.Routes) > 0 ||
		len(source.TripOrder) == 0 && len(source.Trips) > 0 ||
		len(source.CalendarOrder) == 0 && len(source.Calendars) > 0 ||
		len(source.CalendarDateOrder) == 0 && len(source.CalendarDates) > 0 ||
		len(source.FareAttrOrder) == 0 && len(source.FareAttributes) > 0 ||
		len(source.FeedInfoOrder) == 0 && len(source.FeedInfos) > 0 ||
		len(source.AreaOrder) == 0 && len(source.Areas) > 0 {
		source.SyncOrderSlices()
	}

	return &MergeContext{
		Source:            source,
		Target:            target,
		Prefix:            prefix,
		EntityByRawID:     make(map[string]interface{}),
		ResolvedDetection: DetectionNone,
		AgencyIDMapping:   make(map[gtfs.AgencyID]gtfs.AgencyID),
		StopIDMapping:     make(map[gtfs.StopID]gtfs.StopID),
		RouteIDMapping:    make(map[gtfs.RouteID]gtfs.RouteID),
		TripIDMapping:     make(map[gtfs.TripID]gtfs.TripID),
		ServiceIDMapping:  make(map[gtfs.ServiceID]gtfs.ServiceID),
		ShapeIDMapping:    make(map[gtfs.ShapeID]gtfs.ShapeID),
		FareIDMapping:     make(map[gtfs.FareID]gtfs.FareID),
		AreaIDMapping:     make(map[gtfs.AreaID]gtfs.AreaID),
		JustAddedStops:    make(map[gtfs.StopID]struct{}),
		JustAddedRoutes:   make(map[gtfs.RouteID]struct{}),
	}
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

// BaseStrategy provides default implementations for EntityMergeStrategy methods.
// Embed this in concrete strategy implementations.
type BaseStrategy struct {
	name               string
	DuplicateDetection DuplicateDetection
	DuplicateLogging   DuplicateLogging
	RenamingStrategy   RenamingStrategy
}

// NewBaseStrategy creates a new BaseStrategy with the given name
func NewBaseStrategy(name string) BaseStrategy {
	return BaseStrategy{
		name:               name,
		DuplicateDetection: DetectionNone,
		DuplicateLogging:   LogNone,
		RenamingStrategy:   RenameContext,
	}
}

// Name returns the strategy name
func (b *BaseStrategy) Name() string {
	return b.name
}

// SetDuplicateDetection configures duplicate detection
func (b *BaseStrategy) SetDuplicateDetection(d DuplicateDetection) {
	b.DuplicateDetection = d
}

// SetDuplicateLogging configures duplicate logging behavior
func (b *BaseStrategy) SetDuplicateLogging(l DuplicateLogging) {
	b.DuplicateLogging = l
}

// SetRenamingStrategy configures ID renaming behavior
func (b *BaseStrategy) SetRenamingStrategy(r RenamingStrategy) {
	b.RenamingStrategy = r
}
