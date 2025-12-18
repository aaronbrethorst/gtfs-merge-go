package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestAgencyMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping agency IDs
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency2",
		Name:     "Metro Authority",
		URL:      "http://metro.example.com",
		Timezone: "America/Los_Angeles",
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both agencies should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Agencies) != 2 {
		t.Errorf("Expected 2 agencies, got %d", len(target.Agencies))
	}

	if _, ok := target.Agencies["agency1"]; !ok {
		t.Error("Expected agency1 to be in target")
	}
	if _, ok := target.Agencies["agency2"]; !ok {
		t.Error("Expected agency2 to be in target")
	}
}

func TestAgencyMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have agency with ID "agency1"
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Different Transit",
		URL:      "http://different.example.com",
		Timezone: "America/Denver",
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one agency1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Agencies) != 1 {
		t.Errorf("Expected 1 agency, got %d", len(target.Agencies))
	}

	agency := target.Agencies["agency1"]
	if agency == nil {
		t.Fatal("Expected agency1 to be in target")
	}

	// The existing agency should be kept (first one added wins)
	if agency.Name != "Transit Authority" {
		t.Errorf("Expected existing agency to be kept, got name %q", agency.Name)
	}

	// Check that the ID mapping points to the existing agency
	if ctx.AgencyIDMapping["agency1"] != "agency1" {
		t.Errorf("Expected AgencyIDMapping[agency1] = agency1, got %q", ctx.AgencyIDMapping["agency1"])
	}
}

func TestAgencyMergeUpdatesRouteRefs(t *testing.T) {
	// Given: source feed has an agency and a route referencing it
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Source Transit",
		URL:      "http://source.example.com",
		Timezone: "America/Denver",
	})
	source.AddRoute(&gtfs.Route{
		ID:       "route1",
		AgencyID: "agency1",
		LongName: "Test Route",
		Type:     3,
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Target Transit",
		URL:      "http://target.example.com",
		Timezone: "America/New_York",
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: agencies are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing agency
	mappedID := ctx.AgencyIDMapping["agency1"]
	if mappedID != "agency1" {
		t.Errorf("Expected mapped ID to be agency1, got %q", mappedID)
	}

	// When routes are merged (using the mapping), they should reference the existing agency
	// This is verified at a higher level in merger_test.go
}

func TestAgencyMergeLogsWarning(t *testing.T) {
	// Given: both feeds have agency with same ID and warning logging enabled
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Different Transit",
		URL:      "http://different.example.com",
		Timezone: "America/Denver",
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogWarning)

	// When: merged with LogWarning
	err := strategy.Merge(ctx)

	// Then: should not error, duplicate should be logged (we can't easily capture log output in test)
	if err != nil {
		t.Fatalf("Merge should not fail with LogWarning, got: %v", err)
	}

	// Verify the duplicate was handled correctly
	if len(target.Agencies) != 1 {
		t.Errorf("Expected 1 agency, got %d", len(target.Agencies))
	}
}

func TestAgencyMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have agency with same ID and error logging enabled
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Different Transit",
		URL:      "http://different.example.com",
		Timezone: "America/Denver",
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	ctx := NewMergeContext(source, target, "")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogError)

	// When: merged with LogError
	err := strategy.Merge(ctx)

	// Then: should return an error
	if err == nil {
		t.Fatal("Expected error when duplicate detected with LogError")
	}
}

func TestAgencyMergeWithPrefix(t *testing.T) {
	// Given: source feed has an agency that collides with target
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	target := gtfs.NewFeed()
	// Add colliding agency to force prefixing
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Different Transit",
		URL:      "http://different.example.com",
		Timezone: "America/Denver",
	})

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: agency should have prefixed ID (2 total)
	if len(target.Agencies) != 2 {
		t.Errorf("Expected 2 agencies, got %d", len(target.Agencies))
	}

	if _, ok := target.Agencies["a_agency1"]; !ok {
		t.Error("Expected a_agency1 to be in target")
	}

	if ctx.AgencyIDMapping["agency1"] != "a_agency1" {
		t.Errorf("Expected mapping agency1 -> a_agency1, got %q", ctx.AgencyIDMapping["agency1"])
	}
}

func TestAgencyMergeDetectionNone(t *testing.T) {
	// Given: both feeds have agency with same ID but DetectionNone is used
	source := gtfs.NewFeed()
	source.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Different Transit",
		URL:      "http://different.example.com",
		Timezone: "America/Denver",
	})

	target := gtfs.NewFeed()
	target.AddAgency(&gtfs.Agency{
		ID:       "agency1",
		Name:     "Transit Authority",
		URL:      "http://transit.example.com",
		Timezone: "America/New_York",
	})

	ctx := NewMergeContext(source, target, "a_")
	strategy := NewAgencyMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with DetectionNone and a prefix
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: both agencies should exist (source with prefix)
	if len(target.Agencies) != 2 {
		t.Errorf("Expected 2 agencies, got %d", len(target.Agencies))
	}

	if _, ok := target.Agencies["agency1"]; !ok {
		t.Error("Expected agency1 to be in target")
	}
	if _, ok := target.Agencies["a_agency1"]; !ok {
		t.Error("Expected a_agency1 to be in target")
	}
}
