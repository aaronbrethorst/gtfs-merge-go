package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// TripMergeStrategy handles merging of trips between feeds
type TripMergeStrategy struct {
	BaseStrategy
}

// NewTripMergeStrategy creates a new TripMergeStrategy
func NewTripMergeStrategy() *TripMergeStrategy {
	return &TripMergeStrategy{
		BaseStrategy: NewBaseStrategy("trip"),
	}
}

// Merge performs the merge operation for trips
func (s *TripMergeStrategy) Merge(ctx *MergeContext) error {
	for _, trip := range ctx.Source.Trips {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Trips[trip.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.TripIDMapping[trip.ID] = existing.ID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate trip detected with ID %q (keeping existing)", trip.ID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate trip detected with ID %q", trip.ID)
				}

				// Skip adding this trip - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.TripID(ctx.Prefix + string(trip.ID))
		ctx.TripIDMapping[trip.ID] = newID

		// Map references
		routeID := trip.RouteID
		if mappedRoute, ok := ctx.RouteIDMapping[routeID]; ok {
			routeID = mappedRoute
		}

		serviceID := trip.ServiceID
		if mappedService, ok := ctx.ServiceIDMapping[serviceID]; ok {
			serviceID = mappedService
		}

		shapeID := trip.ShapeID
		if shapeID != "" {
			if mappedShape, ok := ctx.ShapeIDMapping[shapeID]; ok {
				shapeID = mappedShape
			}
		}

		newTrip := &gtfs.Trip{
			ID:                   newID,
			RouteID:              routeID,
			ServiceID:            serviceID,
			Headsign:             trip.Headsign,
			ShortName:            trip.ShortName,
			DirectionID:          trip.DirectionID,
			BlockID:              trip.BlockID,
			ShapeID:              shapeID,
			WheelchairAccessible: trip.WheelchairAccessible,
			BikesAllowed:         trip.BikesAllowed,
		}
		ctx.Target.Trips[newID] = newTrip
	}

	return nil
}
