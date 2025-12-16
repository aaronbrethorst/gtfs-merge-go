package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// RouteMergeStrategy handles merging of routes between feeds
type RouteMergeStrategy struct {
	BaseStrategy
}

// NewRouteMergeStrategy creates a new RouteMergeStrategy
func NewRouteMergeStrategy() *RouteMergeStrategy {
	return &RouteMergeStrategy{
		BaseStrategy: NewBaseStrategy("route"),
	}
}

// Merge performs the merge operation for routes
func (s *RouteMergeStrategy) Merge(ctx *MergeContext) error {
	for _, route := range ctx.Source.Routes {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Routes[route.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.RouteIDMapping[route.ID] = existing.ID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate route detected with ID %q (keeping existing)", route.ID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate route detected with ID %q", route.ID)
				}

				// Skip adding this route - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.RouteID(ctx.Prefix + string(route.ID))
		ctx.RouteIDMapping[route.ID] = newID

		// Map agency reference
		agencyID := route.AgencyID
		if agencyID != "" {
			if mappedAgency, ok := ctx.AgencyIDMapping[agencyID]; ok {
				agencyID = mappedAgency
			}
		}

		newRoute := &gtfs.Route{
			ID:                newID,
			AgencyID:          agencyID,
			ShortName:         route.ShortName,
			LongName:          route.LongName,
			Desc:              route.Desc,
			Type:              route.Type,
			URL:               route.URL,
			Color:             route.Color,
			TextColor:         route.TextColor,
			SortOrder:         route.SortOrder,
			ContinuousPickup:  route.ContinuousPickup,
			ContinuousDropOff: route.ContinuousDropOff,
		}
		ctx.Target.Routes[newID] = newRoute
	}

	return nil
}
