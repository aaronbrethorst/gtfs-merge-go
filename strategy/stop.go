package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// StopMergeStrategy handles merging of stops between feeds
type StopMergeStrategy struct {
	BaseStrategy
}

// NewStopMergeStrategy creates a new StopMergeStrategy
func NewStopMergeStrategy() *StopMergeStrategy {
	return &StopMergeStrategy{
		BaseStrategy: NewBaseStrategy("stop"),
	}
}

// Merge performs the merge operation for stops
func (s *StopMergeStrategy) Merge(ctx *MergeContext) error {
	for _, stop := range ctx.Source.Stops {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Stops[stop.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.StopIDMapping[stop.ID] = existing.ID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate stop detected with ID %q (keeping existing)", stop.ID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate stop detected with ID %q", stop.ID)
				}

				// Skip adding this stop - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.StopID(ctx.Prefix + string(stop.ID))
		ctx.StopIDMapping[stop.ID] = newID

		// Handle parent_station reference
		parentStation := stop.ParentStation
		if parentStation != "" {
			if mappedParent, ok := ctx.StopIDMapping[parentStation]; ok {
				parentStation = mappedParent
			} else {
				parentStation = gtfs.StopID(ctx.Prefix + string(parentStation))
			}
		}

		newStop := &gtfs.Stop{
			ID:                 newID,
			Code:               stop.Code,
			Name:               stop.Name,
			Desc:               stop.Desc,
			Lat:                stop.Lat,
			Lon:                stop.Lon,
			ZoneID:             stop.ZoneID,
			URL:                stop.URL,
			LocationType:       stop.LocationType,
			ParentStation:      parentStation,
			Timezone:           stop.Timezone,
			WheelchairBoarding: stop.WheelchairBoarding,
			LevelID:            stop.LevelID,
			PlatformCode:       stop.PlatformCode,
		}
		ctx.Target.Stops[newID] = newStop
	}

	return nil
}
