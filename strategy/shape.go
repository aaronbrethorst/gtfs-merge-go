package strategy

import (
	"fmt"
	"log"
	"sort"

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
	// Sort shape IDs to ensure deterministic processing order.
	// This is critical because we use a global sequence counter that must
	// assign numbers in the same order as Java to produce matching output.
	shapeIDs := make([]gtfs.ShapeID, 0, len(ctx.Source.Shapes))
	for shapeID := range ctx.Source.Shapes {
		shapeIDs = append(shapeIDs, shapeID)
	}
	sort.Slice(shapeIDs, func(i, j int) bool {
		return string(shapeIDs[i]) < string(shapeIDs[j])
	})

	for _, shapeID := range shapeIDs {
		points := ctx.Source.Shapes[shapeID]
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if _, found := ctx.Target.Shapes[shapeID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.ShapeIDMapping[shapeID] = shapeID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate shape detected with ID %q (keeping existing)", shapeID)
				case LogError:
					return fmt.Errorf("duplicate shape detected with ID %q", shapeID)
				}

				// Skip adding this shape - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := shapeID
		if _, exists := ctx.Target.Shapes[shapeID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.ShapeID(ctx.Prefix + string(shapeID))
		}
		ctx.ShapeIDMapping[shapeID] = newID

		for _, point := range points {
			newPoint := &gtfs.ShapePoint{
				ShapeID:      newID,
				Lat:          point.Lat,
				Lon:          point.Lon,
				Sequence:     ctx.NextShapeSequence(), // Use global counter to match Java
				DistTraveled: point.DistTraveled,
			}
			ctx.Target.Shapes[newID] = append(ctx.Target.Shapes[newID], newPoint)
		}
	}

	return nil
}
