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

	// makeKey creates a normalized key for a transfer.
	// For symmetric transfers (from_stop_id == to_stop_id), route and trip IDs
	// are normalized to canonical order (smaller first) to ensure consistent deduplication.
	makeKey := func(fromStop, toStop gtfs.StopID, transferType, minTransferTime int,
		fromRoute, toRoute gtfs.RouteID, fromTrip, toTrip gtfs.TripID) transferKey {
		// Normalize symmetric transfers
		if fromStop == toStop {
			// Normalize route IDs: ensure fromRouteID <= toRouteID
			if fromRoute > toRoute {
				fromRoute, toRoute = toRoute, fromRoute
			}
			// Normalize trip IDs: ensure fromTripID <= toTripID
			if fromTrip > toTrip {
				fromTrip, toTrip = toTrip, fromTrip
			}
		}
		return transferKey{
			fromStop, toStop, transferType, minTransferTime,
			fromRoute, toRoute, fromTrip, toTrip,
		}
	}

	// Always build the existing keys index for deduplication
	// Transfers are always deduplicated to avoid duplicates in output
	existingKeys := make(map[transferKey]bool)
	for _, existing := range ctx.Target.Transfers {
		key := makeKey(
			existing.FromStopID, existing.ToStopID,
			existing.TransferType, existing.MinTransferTime,
			existing.FromRouteID, existing.ToRouteID,
			existing.FromTripID, existing.ToTripID,
		)
		existingKeys[key] = true
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

		// Check for duplicates using O(1) lookup (always deduplicate transfers)
		key := makeKey(
			fromStopID, toStopID,
			transfer.TransferType, transfer.MinTransferTime,
			fromRouteID, toRouteID,
			fromTripID, toTripID,
		)
		if existingKeys[key] {
			continue
		}
		// Add to index for subsequent source items
		existingKeys[key] = true

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
