package merge

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestNewMergeContext(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target)

	if ctx.Source != source {
		t.Error("expected source feed to be set")
	}
	if ctx.Target != target {
		t.Error("expected target feed to be set")
	}
	if ctx.Prefix != "" {
		t.Errorf("expected empty prefix for first feed, got %q", ctx.Prefix)
	}
}

func TestMergeContextWithPrefix(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target)
	ctx.Prefix = "a-"

	if ctx.Prefix != "a-" {
		t.Errorf("expected prefix 'a-', got %q", ctx.Prefix)
	}
}

func TestMergeContextEntityTracking(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target)

	// Track an entity by its raw ID
	agency := &gtfs.Agency{ID: "agency1", Name: "Test Agency"}
	ctx.EntityByRawID["agency:agency1"] = agency

	// Verify we can retrieve it
	retrieved, ok := ctx.EntityByRawID["agency:agency1"]
	if !ok {
		t.Error("expected to find tracked entity")
	}
	if retrieved != agency {
		t.Error("expected to get same agency back")
	}
}

func TestMergeContextIDMappings(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target)

	// Test agency ID mapping
	ctx.AgencyIDMapping["old_agency"] = "new_agency"
	if ctx.AgencyIDMapping["old_agency"] != "new_agency" {
		t.Error("agency ID mapping not working")
	}

	// Test stop ID mapping
	ctx.StopIDMapping["old_stop"] = "new_stop"
	if ctx.StopIDMapping["old_stop"] != "new_stop" {
		t.Error("stop ID mapping not working")
	}

	// Test route ID mapping
	ctx.RouteIDMapping["old_route"] = "new_route"
	if ctx.RouteIDMapping["old_route"] != "new_route" {
		t.Error("route ID mapping not working")
	}

	// Test trip ID mapping
	ctx.TripIDMapping["old_trip"] = "new_trip"
	if ctx.TripIDMapping["old_trip"] != "new_trip" {
		t.Error("trip ID mapping not working")
	}

	// Test service ID mapping
	ctx.ServiceIDMapping["old_service"] = "new_service"
	if ctx.ServiceIDMapping["old_service"] != "new_service" {
		t.Error("service ID mapping not working")
	}

	// Test shape ID mapping
	ctx.ShapeIDMapping["old_shape"] = "new_shape"
	if ctx.ShapeIDMapping["old_shape"] != "new_shape" {
		t.Error("shape ID mapping not working")
	}

	// Test fare ID mapping
	ctx.FareIDMapping["old_fare"] = "new_fare"
	if ctx.FareIDMapping["old_fare"] != "new_fare" {
		t.Error("fare ID mapping not working")
	}

	// Test area ID mapping
	ctx.AreaIDMapping["old_area"] = "new_area"
	if ctx.AreaIDMapping["old_area"] != "new_area" {
		t.Error("area ID mapping not working")
	}
}

func TestGetPrefixForIndex(t *testing.T) {
	// Tests the Java-compatible prefix assignment based on ORIGINAL array index.
	// Java uses: index 0 → "a-", index 1 → "b-", etc.
	// The prefix is only applied when there's an ID collision during merge.
	tests := []struct {
		index    int
		expected string
	}{
		{0, "a-"},    // Feed at index 0
		{1, "b-"},    // Feed at index 1
		{2, "c-"},    // Feed at index 2
		{25, "z-"},   // Feed at index 25 (last letter)
		{26, "00-"},  // Feed at index 26 (first numeric)
		{27, "01-"},  // Feed at index 27
		{125, "99-"}, // Feed at index 125
	}

	for _, tt := range tests {
		result := GetPrefixForIndex(tt.index)
		if result != tt.expected {
			t.Errorf("GetPrefixForIndex(%d) = %q, want %q", tt.index, result, tt.expected)
		}
	}
}
