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

		// Check for duplicates (same trip_id, stop_sequence)
		isDuplicate := false
		if s.DuplicateDetection == DetectionIdentity {
			for _, existing := range ctx.Target.StopTimes {
				if existing.TripID == tripID && existing.StopSequence == st.StopSequence {
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
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
