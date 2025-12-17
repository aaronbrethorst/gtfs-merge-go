// Package merge provides GTFS feed merge functionality.
package merge

import (
	"fmt"

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
func NewMergeContext(source, target *gtfs.Feed) *MergeContext {
	return &MergeContext{
		Source:           source,
		Target:           target,
		Prefix:           "",
		EntityByRawID:    make(map[string]interface{}),
		AgencyIDMapping:  make(map[gtfs.AgencyID]gtfs.AgencyID),
		StopIDMapping:    make(map[gtfs.StopID]gtfs.StopID),
		RouteIDMapping:   make(map[gtfs.RouteID]gtfs.RouteID),
		TripIDMapping:    make(map[gtfs.TripID]gtfs.TripID),
		ServiceIDMapping: make(map[gtfs.ServiceID]gtfs.ServiceID),
		ShapeIDMapping:   make(map[gtfs.ShapeID]gtfs.ShapeID),
		FareIDMapping:    make(map[gtfs.FareID]gtfs.FareID),
		AreaIDMapping:    make(map[gtfs.AreaID]gtfs.AreaID),
	}
}

// GetPrefixForIndex returns the prefix for a feed at the given ORIGINAL array index.
// This matches Java behavior where:
// - index 0 → "a-"
// - index 1 → "b-"
// - index 2 → "c-"
// - etc.
// The prefix is only applied when there's an ID collision during merge.
// Note: Uses hyphen (-) delimiter to match Java implementation.
func GetPrefixForIndex(index int) string {
	if index < 26 {
		// Use letters a-z for feeds 0-25
		return string(rune('a'+index)) + "-"
	}
	// Use numeric prefixes for feeds 26+
	return fmt.Sprintf("%02d-", index-26)
}
