package strategy

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// FrequencyMergeStrategy handles merging of frequencies between feeds
type FrequencyMergeStrategy struct {
	BaseStrategy
}

// NewFrequencyMergeStrategy creates a new FrequencyMergeStrategy
func NewFrequencyMergeStrategy() *FrequencyMergeStrategy {
	return &FrequencyMergeStrategy{
		BaseStrategy: NewBaseStrategy("frequency"),
	}
}

// Merge performs the merge operation for frequencies
func (s *FrequencyMergeStrategy) Merge(ctx *MergeContext) error {
	for _, freq := range ctx.Source.Frequencies {
		// Map trip reference
		tripID := freq.TripID
		if mappedTrip, ok := ctx.TripIDMapping[tripID]; ok {
			tripID = mappedTrip
		}

		// Check for duplicates (same trip_id, start_time, end_time, headway_secs)
		isDuplicate := false
		if s.DuplicateDetection == DetectionIdentity {
			for _, existing := range ctx.Target.Frequencies {
				if existing.TripID == tripID &&
					existing.StartTime == freq.StartTime &&
					existing.EndTime == freq.EndTime &&
					existing.HeadwaySecs == freq.HeadwaySecs {
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
		}

		newFreq := &gtfs.Frequency{
			TripID:      tripID,
			StartTime:   freq.StartTime,
			EndTime:     freq.EndTime,
			HeadwaySecs: freq.HeadwaySecs,
			ExactTimes:  freq.ExactTimes,
		}
		ctx.Target.Frequencies = append(ctx.Target.Frequencies, newFreq)
	}

	return nil
}
