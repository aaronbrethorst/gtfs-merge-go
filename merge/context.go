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

// GetPrefixForIndex returns the prefix for a feed at the given index.
// First feed (index 0) gets no prefix.
// Feeds 1-25 get a_, b_, ..., z_.
// Feeds 26+ get 00_, 01_, ..., 99_, etc.
func GetPrefixForIndex(index int) string {
	if index == 0 {
		return ""
	}
	if index <= 26 {
		// Use letters a-z for feeds 1-26
		return string(rune('a'+index-1)) + "_"
	}
	// Use numeric prefixes for feeds 27+
	return fmt.Sprintf("%02d_", index-27)
}
