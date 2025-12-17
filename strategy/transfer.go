package strategy

import (
	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// TransferMergeStrategy handles merging of transfers between feeds
type TransferMergeStrategy struct {
	BaseStrategy
}

// NewTransferMergeStrategy creates a new TransferMergeStrategy
func NewTransferMergeStrategy() *TransferMergeStrategy {
	return &TransferMergeStrategy{
		BaseStrategy: NewBaseStrategy("transfer"),
	}
}

// Merge performs the merge operation for transfers
func (s *TransferMergeStrategy) Merge(ctx *MergeContext) error {
	// Build index for O(1) duplicate detection (avoids O(n²) linear scan)
	type transferKey struct {
		fromStopID      gtfs.StopID
		toStopID        gtfs.StopID
		transferType    int
		minTransferTime int
		fromRouteID     gtfs.RouteID
		toRouteID       gtfs.RouteID
		fromTripID      gtfs.TripID
		toTripID        gtfs.TripID
	}
	existingKeys := make(map[transferKey]bool)
	if s.DuplicateDetection == DetectionIdentity {
		for _, existing := range ctx.Target.Transfers {
			existingKeys[transferKey{
				existing.FromStopID,
				existing.ToStopID,
				existing.TransferType,
				existing.MinTransferTime,
				existing.FromRouteID,
				existing.ToRouteID,
				existing.FromTripID,
				existing.ToTripID,
			}] = true
		}
	}

	for _, transfer := range ctx.Source.Transfers {
		// Map stop references
		fromStopID := transfer.FromStopID
		if mappedStop, ok := ctx.StopIDMapping[fromStopID]; ok {
			fromStopID = mappedStop
		}

		toStopID := transfer.ToStopID
		if mappedStop, ok := ctx.StopIDMapping[toStopID]; ok {
			toStopID = mappedStop
		}

		// Map route references
		fromRouteID := transfer.FromRouteID
		if fromRouteID != "" {
			if mappedRoute, ok := ctx.RouteIDMapping[fromRouteID]; ok {
				fromRouteID = mappedRoute
			}
		}

		toRouteID := transfer.ToRouteID
		if toRouteID != "" {
			if mappedRoute, ok := ctx.RouteIDMapping[toRouteID]; ok {
				toRouteID = mappedRoute
			}
		}

		// Map trip references
		fromTripID := transfer.FromTripID
		if fromTripID != "" {
			if mappedTrip, ok := ctx.TripIDMapping[fromTripID]; ok {
				fromTripID = mappedTrip
			}
		}

		toTripID := transfer.ToTripID
		if toTripID != "" {
			if mappedTrip, ok := ctx.TripIDMapping[toTripID]; ok {
				toTripID = mappedTrip
			}
		}

		// Check for duplicates using O(1) lookup
		if s.DuplicateDetection == DetectionIdentity {
			key := transferKey{
				fromStopID,
				toStopID,
				transfer.TransferType,
				transfer.MinTransferTime,
				fromRouteID,
				toRouteID,
				fromTripID,
				toTripID,
			}
			if existingKeys[key] {
				continue
			}
			// Add to index for subsequent source items
			existingKeys[key] = true
		}

		newTransfer := &gtfs.Transfer{
			FromStopID:      fromStopID,
			ToStopID:        toStopID,
			TransferType:    transfer.TransferType,
			MinTransferTime: transfer.MinTransferTime,
			FromRouteID:     fromRouteID,
			ToRouteID:       toRouteID,
			FromTripID:      fromTripID,
			ToTripID:        toTripID,
		}
		ctx.Target.Transfers = append(ctx.Target.Transfers, newTransfer)
	}

	return nil
}
