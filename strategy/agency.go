package strategy

import (
	"fmt"
	"log"
	"sort"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// AgencyMergeStrategy handles merging of agencies between feeds
type AgencyMergeStrategy struct {
	BaseStrategy
}

// NewAgencyMergeStrategy creates a new AgencyMergeStrategy
func NewAgencyMergeStrategy() *AgencyMergeStrategy {
	return &AgencyMergeStrategy{
		BaseStrategy: NewBaseStrategy("agency"),
	}
}

// Merge performs the merge operation for agencies
func (s *AgencyMergeStrategy) Merge(ctx *MergeContext) error {
	// Sort source agency IDs to match Java output order
	// Java processes each feed's agencies in sorted order within that feed
	sortedAgencyIDs := make([]gtfs.AgencyID, len(ctx.Source.AgencyOrder))
	copy(sortedAgencyIDs, ctx.Source.AgencyOrder)
	sort.Slice(sortedAgencyIDs, func(i, j int) bool {
		return sortedAgencyIDs[i] < sortedAgencyIDs[j]
	})

	for _, agencyID := range sortedAgencyIDs {
		agency := ctx.Source.Agencies[agencyID]
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Agencies[agency.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.AgencyIDMapping[agency.ID] = existing.ID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate agency detected with ID %q (keeping existing)", agency.ID)
				case LogError:
					return fmt.Errorf("duplicate agency detected with ID %q", agency.ID)
				}

				// Skip adding this agency - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := agency.ID
		if _, exists := ctx.Target.Agencies[agency.ID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.AgencyID(ctx.Prefix + string(agency.ID))
		}
		ctx.AgencyIDMapping[agency.ID] = newID

		newAgency := &gtfs.Agency{
			ID:       newID,
			Name:     agency.Name,
			URL:      agency.URL,
			Timezone: agency.Timezone,
			Lang:     agency.Lang,
			Phone:    agency.Phone,
			FareURL:  agency.FareURL,
			Email:    agency.Email,
		}
		ctx.Target.Agencies[newID] = newAgency
		ctx.Target.AgencyOrder = append(ctx.Target.AgencyOrder, newID)
	}

	return nil
}
