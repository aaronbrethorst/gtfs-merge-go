package strategy

import (
	"fmt"
	"log"

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
	for _, agency := range ctx.Source.Agencies {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Agencies[agency.ID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.AgencyIDMapping[agency.ID] = existing.ID

				// Handle logging based on configuration
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate agency detected with ID %q (keeping existing)", agency.ID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate agency detected with ID %q", agency.ID)
				}

				// Skip adding this agency - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.AgencyID(ctx.Prefix + string(agency.ID))
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
	}

	return nil
}
