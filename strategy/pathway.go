package strategy

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// PathwayMergeStrategy handles merging of pathways between feeds
type PathwayMergeStrategy struct {
	BaseStrategy
}

// NewPathwayMergeStrategy creates a new PathwayMergeStrategy
func NewPathwayMergeStrategy() *PathwayMergeStrategy {
	return &PathwayMergeStrategy{
		BaseStrategy: NewBaseStrategy("pathway"),
	}
}

// Merge performs the merge operation for pathways
func (s *PathwayMergeStrategy) Merge(ctx *MergeContext) error {
	for _, pathway := range ctx.Source.Pathways {
		// Map stop references
		fromStopID := pathway.FromStopID
		if mappedStop, ok := ctx.StopIDMapping[fromStopID]; ok {
			fromStopID = mappedStop
		}

		toStopID := pathway.ToStopID
		if mappedStop, ok := ctx.StopIDMapping[toStopID]; ok {
			toStopID = mappedStop
		}

		newID := ctx.Prefix + pathway.ID

		// Check for duplicates based on ID
		isDuplicate := false
		if s.DuplicateDetection == DetectionIdentity {
			for _, existing := range ctx.Target.Pathways {
				if existing.ID == newID || existing.ID == pathway.ID {
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
		}

		newPathway := &gtfs.Pathway{
			ID:                   newID,
			FromStopID:           fromStopID,
			ToStopID:             toStopID,
			PathwayMode:          pathway.PathwayMode,
			IsBidirectional:      pathway.IsBidirectional,
			Length:               pathway.Length,
			TraversalTime:        pathway.TraversalTime,
			StairCount:           pathway.StairCount,
			MaxSlope:             pathway.MaxSlope,
			MinWidth:             pathway.MinWidth,
			SignpostedAs:         pathway.SignpostedAs,
			ReversedSignpostedAs: pathway.ReversedSignpostedAs,
		}
		ctx.Target.Pathways = append(ctx.Target.Pathways, newPathway)
	}

	return nil
}
