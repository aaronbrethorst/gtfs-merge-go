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
	// Build index for O(1) duplicate/collision detection (avoids O(n²) linear scan)
	existingIDs := make(map[string]bool)
	for _, existing := range ctx.Target.Pathways {
		existingIDs[existing.ID] = true
	}

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

		// Check for duplicates/collisions using O(1) lookup
		hasCollision := existingIDs[pathway.ID]
		if hasCollision && s.DuplicateDetection == DetectionIdentity {
			continue // Skip duplicate
		}

		// Only apply prefix if there's a collision
		newID := pathway.ID
		if hasCollision {
			newID = ctx.Prefix + pathway.ID
		}

		// Add to index for subsequent source items
		existingIDs[newID] = true

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
