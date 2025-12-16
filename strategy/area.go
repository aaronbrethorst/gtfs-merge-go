package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// AreaMergeStrategy handles merging of areas between feeds
type AreaMergeStrategy struct {
	BaseStrategy
}

// NewAreaMergeStrategy creates a new AreaMergeStrategy
func NewAreaMergeStrategy() *AreaMergeStrategy {
	return &AreaMergeStrategy{
		BaseStrategy: NewBaseStrategy("area"),
	}
}

// Merge performs the merge operation for areas
func (s *AreaMergeStrategy) Merge(ctx *MergeContext) error {
	for _, area := range ctx.Source.Areas {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Areas[area.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.AreaIDMapping[area.ID] = existing.ID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate area detected with ID %q (keeping existing)", area.ID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate area detected with ID %q", area.ID)
				}

				// Skip adding this area - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := area.ID
		if _, exists := ctx.Target.Areas[area.ID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.AreaID(ctx.Prefix + string(area.ID))
		}
		ctx.AreaIDMapping[area.ID] = newID

		newArea := &gtfs.Area{
			ID:   newID,
			Name: area.Name,
		}
		ctx.Target.Areas[newID] = newArea
	}

	return nil
}
