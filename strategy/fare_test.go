package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestFareAttributeMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping fare IDs
	source := gtfs.NewFeed()
	source.FareAttributes[gtfs.FareID("fare1")] = &gtfs.FareAttribute{
		FareID:       "fare1",
		Price:        2.50,
		CurrencyType: "USD",
	}

	target := gtfs.NewFeed()
	target.FareAttributes[gtfs.FareID("fare2")] = &gtfs.FareAttribute{
		FareID:       "fare2",
		Price:        3.00,
		CurrencyType: "USD",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFareAttributeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both fares should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FareAttributes) != 2 {
		t.Errorf("Expected 2 fares, got %d", len(target.FareAttributes))
	}
}

func TestFareAttributeMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have fare with ID "fare1"
	source := gtfs.NewFeed()
	source.FareAttributes[gtfs.FareID("fare1")] = &gtfs.FareAttribute{
		FareID:       "fare1",
		Price:        3.00,
		CurrencyType: "USD",
	}

	target := gtfs.NewFeed()
	target.FareAttributes[gtfs.FareID("fare1")] = &gtfs.FareAttribute{
		FareID:       "fare1",
		Price:        2.50,
		CurrencyType: "USD",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewFareAttributeMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one fare1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FareAttributes) != 1 {
		t.Errorf("Expected 1 fare, got %d", len(target.FareAttributes))
	}

	fare := target.FareAttributes["fare1"]
	if fare.Price != 2.50 {
		t.Errorf("Expected existing fare to be kept, got price %f", fare.Price)
	}
}

func TestFareRuleMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping fare rules
	source := gtfs.NewFeed()
	source.FareRules = append(source.FareRules, &gtfs.FareRule{
		FareID:  "fare1",
		RouteID: "route1",
	})

	target := gtfs.NewFeed()
	target.FareRules = append(target.FareRules, &gtfs.FareRule{
		FareID:  "fare2",
		RouteID: "route2",
	})

	ctx := NewMergeContext(source, target, "")
	ctx.FareIDMapping[gtfs.FareID("fare1")] = gtfs.FareID("fare1")
	ctx.RouteIDMapping[gtfs.RouteID("route1")] = gtfs.RouteID("route1")

	strategy := NewFareRuleMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both rules should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FareRules) != 2 {
		t.Errorf("Expected 2 fare rules, got %d", len(target.FareRules))
	}
}

func TestFareRuleMergeIdentical(t *testing.T) {
	// Given: both feeds have identical fare rule
	source := gtfs.NewFeed()
	source.FareRules = append(source.FareRules, &gtfs.FareRule{
		FareID:  "fare1",
		RouteID: "route1",
	})

	target := gtfs.NewFeed()
	target.FareRules = append(target.FareRules, &gtfs.FareRule{
		FareID:  "fare1",
		RouteID: "route1",
	})

	ctx := NewMergeContext(source, target, "")
	ctx.FareIDMapping[gtfs.FareID("fare1")] = gtfs.FareID("fare1")
	ctx.RouteIDMapping[gtfs.RouteID("route1")] = gtfs.RouteID("route1")

	strategy := NewFareRuleMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one fare rule in output
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.FareRules) != 1 {
		t.Errorf("Expected 1 fare rule (duplicate skipped), got %d", len(target.FareRules))
	}
}
