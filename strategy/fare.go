package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// FareAttributeMergeStrategy handles merging of fare attributes between feeds
type FareAttributeMergeStrategy struct {
	BaseStrategy
}

// NewFareAttributeMergeStrategy creates a new FareAttributeMergeStrategy
func NewFareAttributeMergeStrategy() *FareAttributeMergeStrategy {
	return &FareAttributeMergeStrategy{
		BaseStrategy: NewBaseStrategy("fare_attribute"),
	}
}

// Merge performs the merge operation for fare attributes
func (s *FareAttributeMergeStrategy) Merge(ctx *MergeContext) error {
	// Iterate in insertion order to match Java output
	for _, fareID := range ctx.Source.FareAttrOrder {
		fare := ctx.Source.FareAttributes[fareID]
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.FareAttributes[fare.FareID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.FareIDMapping[fare.FareID] = existing.FareID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate fare_attribute detected with fare_id %q (keeping existing)", fare.FareID)
				case LogError:
					return fmt.Errorf("duplicate fare_attribute detected with fare_id %q", fare.FareID)
				}

				// Skip adding this fare - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := fare.FareID
		if _, exists := ctx.Target.FareAttributes[fare.FareID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.FareID(ctx.Prefix + string(fare.FareID))
		}
		ctx.FareIDMapping[fare.FareID] = newID

		// Note: agency_id is NOT remapped - Java doesn't remap this field,
		// it only renames entity IDs (AgencyAndId primary keys)
		newFare := &gtfs.FareAttribute{
			FareID:           newID,
			Price:            fare.Price,
			CurrencyType:     fare.CurrencyType,
			PaymentMethod:    fare.PaymentMethod,
			Transfers:        fare.Transfers,
			AgencyID:         fare.AgencyID, // Keep original, don't remap
			TransferDuration: fare.TransferDuration,
			YouthPrice:       fare.YouthPrice,
			SeniorPrice:      fare.SeniorPrice,
		}
		ctx.Target.FareAttributes[newID] = newFare
		ctx.Target.FareAttrOrder = append(ctx.Target.FareAttrOrder, newID)
	}

	return nil
}

// FareRuleMergeStrategy handles merging of fare rules between feeds
type FareRuleMergeStrategy struct {
	BaseStrategy
}

// NewFareRuleMergeStrategy creates a new FareRuleMergeStrategy
func NewFareRuleMergeStrategy() *FareRuleMergeStrategy {
	return &FareRuleMergeStrategy{
		BaseStrategy: NewBaseStrategy("fare_rule"),
	}
}

// Merge performs the merge operation for fare rules
func (s *FareRuleMergeStrategy) Merge(ctx *MergeContext) error {
	// Build index for O(1) duplicate detection (avoids O(n²) linear scan)
	type fareRuleKey struct {
		fareID        gtfs.FareID
		routeID       gtfs.RouteID
		originID      string
		destinationID string
		containsID    string
	}
	existingKeys := make(map[fareRuleKey]bool)
	if s.DuplicateDetection == DetectionIdentity {
		for _, existing := range ctx.Target.FareRules {
			existingKeys[fareRuleKey{
				existing.FareID,
				existing.RouteID,
				existing.OriginID,
				existing.DestinationID,
				existing.ContainsID,
			}] = true
		}
	}

	for _, rule := range ctx.Source.FareRules {
		// Map references
		fareID := rule.FareID
		if mappedFare, ok := ctx.FareIDMapping[fareID]; ok {
			fareID = mappedFare
		}

		routeID := rule.RouteID
		if routeID != "" {
			if mappedRoute, ok := ctx.RouteIDMapping[routeID]; ok {
				routeID = mappedRoute
			}
		}

		// Check for duplicates using O(1) lookup
		if s.DuplicateDetection == DetectionIdentity {
			key := fareRuleKey{
				fareID,
				routeID,
				rule.OriginID,
				rule.DestinationID,
				rule.ContainsID,
			}
			if existingKeys[key] {
				continue
			}
			// Add to index for subsequent source items
			existingKeys[key] = true
		}

		newRule := &gtfs.FareRule{
			FareID:        fareID,
			RouteID:       routeID,
			OriginID:      rule.OriginID,
			DestinationID: rule.DestinationID,
			ContainsID:    rule.ContainsID,
		}
		ctx.Target.FareRules = append(ctx.Target.FareRules, newRule)
	}

	return nil
}
