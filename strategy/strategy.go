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
}

// NewMergeContext creates a new merge context
func NewMergeContext(source, target *gtfs.Feed, prefix string) *MergeContext {
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
