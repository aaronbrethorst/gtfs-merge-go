package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// ShapeMergeStrategy handles merging of shapes between feeds
type ShapeMergeStrategy struct {
	BaseStrategy
}

// NewShapeMergeStrategy creates a new ShapeMergeStrategy
func NewShapeMergeStrategy() *ShapeMergeStrategy {
	return &ShapeMergeStrategy{
		BaseStrategy: NewBaseStrategy("shape"),
	}
}

// Merge performs the merge operation for shapes
func (s *ShapeMergeStrategy) Merge(ctx *MergeContext) error {
	for shapeID, points := range ctx.Source.Shapes {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if _, found := ctx.Target.Shapes[shapeID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.ShapeIDMapping[shapeID] = shapeID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate shape detected with ID %q (keeping existing)", shapeID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate shape detected with ID %q", shapeID)
				}

				// Skip adding this shape - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.ShapeID(ctx.Prefix + string(shapeID))
		ctx.ShapeIDMapping[shapeID] = newID

		for _, point := range points {
			newPoint := &gtfs.ShapePoint{
				ShapeID:      newID,
				Lat:          point.Lat,
				Lon:          point.Lon,
				Sequence:     point.Sequence,
				DistTraveled: point.DistTraveled,
			}
			ctx.Target.Shapes[newID] = append(ctx.Target.Shapes[newID], newPoint)
		}
	}

	return nil
}
