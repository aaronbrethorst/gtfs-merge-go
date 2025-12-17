package strategy

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// StopTimeMergeStrategy handles merging of stop times between feeds
type StopTimeMergeStrategy struct {
	BaseStrategy
}

// NewStopTimeMergeStrategy creates a new StopTimeMergeStrategy
func NewStopTimeMergeStrategy() *StopTimeMergeStrategy {
	return &StopTimeMergeStrategy{
		BaseStrategy: NewBaseStrategy("stop_time"),
	}
}

// Merge performs the merge operation for stop times
func (s *StopTimeMergeStrategy) Merge(ctx *MergeContext) error {
	// Build index for O(1) duplicate detection (avoids O(n²) linear scan)
	type stopTimeKey struct {
		tripID       gtfs.TripID
		stopSequence int
	}
	existingKeys := make(map[stopTimeKey]bool)
	if s.DuplicateDetection == DetectionIdentity {
		for _, existing := range ctx.Target.StopTimes {
			existingKeys[stopTimeKey{existing.TripID, existing.StopSequence}] = true
		}
	}

	for _, st := range ctx.Source.StopTimes {
		// Map references
		tripID := st.TripID
		if mappedTrip, ok := ctx.TripIDMapping[tripID]; ok {
			tripID = mappedTrip
		}

		stopID := st.StopID
		if mappedStop, ok := ctx.StopIDMapping[stopID]; ok {
			stopID = mappedStop
		}

		// Check for duplicates (same trip_id, stop_sequence) using O(1) lookup
		if s.DuplicateDetection == DetectionIdentity {
			key := stopTimeKey{tripID, st.StopSequence}
			if existingKeys[key] {
				continue
			}
			// Add to index for subsequent source items
			existingKeys[key] = true
		}

		newST := &gtfs.StopTime{
			TripID:            tripID,
			ArrivalTime:       st.ArrivalTime,
			DepartureTime:     st.DepartureTime,
			StopID:            stopID,
			StopSequence:      st.StopSequence,
			StopHeadsign:      st.StopHeadsign,
			PickupType:        st.PickupType,
			DropOffType:       st.DropOffType,
			ContinuousPickup:  st.ContinuousPickup,
			ContinuousDropOff: st.ContinuousDropOff,
			ShapeDistTraveled: st.ShapeDistTraveled,
			Timepoint:         st.Timepoint,
		}
		ctx.Target.StopTimes = append(ctx.Target.StopTimes, newST)
	}

	return nil
}
