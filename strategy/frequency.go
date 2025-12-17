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
	// Build index for O(1) duplicate detection (avoids O(n²) linear scan)
	type frequencyKey struct {
		tripID      gtfs.TripID
		startTime   string
		endTime     string
		headwaySecs int
	}
	existingKeys := make(map[frequencyKey]bool)
	if s.DuplicateDetection == DetectionIdentity {
		for _, existing := range ctx.Target.Frequencies {
			existingKeys[frequencyKey{existing.TripID, existing.StartTime, existing.EndTime, existing.HeadwaySecs}] = true
		}
	}

	for _, freq := range ctx.Source.Frequencies {
		// Map trip reference
		tripID := freq.TripID
		if mappedTrip, ok := ctx.TripIDMapping[tripID]; ok {
			tripID = mappedTrip
		}

		// Check for duplicates using O(1) lookup
		if s.DuplicateDetection == DetectionIdentity {
			key := frequencyKey{tripID, freq.StartTime, freq.EndTime, freq.HeadwaySecs}
			if existingKeys[key] {
				continue
			}
			// Add to index for subsequent source items
			existingKeys[key] = true
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
