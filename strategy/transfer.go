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

		// Check for duplicates (same from_stop_id, to_stop_id, transfer_type, min_transfer_time)
		isDuplicate := false
		if s.DuplicateDetection == DetectionIdentity {
			for _, existing := range ctx.Target.Transfers {
				if existing.FromStopID == fromStopID &&
					existing.ToStopID == toStopID &&
					existing.TransferType == transfer.TransferType &&
					existing.MinTransferTime == transfer.MinTransferTime {
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
		}

		newTransfer := &gtfs.Transfer{
			FromStopID:      fromStopID,
			ToStopID:        toStopID,
			TransferType:    transfer.TransferType,
			MinTransferTime: transfer.MinTransferTime,
		}
		ctx.Target.Transfers = append(ctx.Target.Transfers, newTransfer)
	}

	return nil
}
