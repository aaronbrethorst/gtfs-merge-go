package merge

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// ============================================================================
// Milestone 12.1: Merger Options Tests
// ============================================================================

func TestWithDebug(t *testing.T) {
	// Test that WithDebug option sets the debug flag
	tests := []struct {
		name     string
		debug    bool
		expected bool
	}{
		{"debug enabled", true, true},
		{"debug disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(WithDebug(tt.debug))
			if m.debug != tt.expected {
				t.Errorf("WithDebug(%v) set debug to %v, want %v", tt.debug, m.debug, tt.expected)
			}
		})
	}
}

func TestWithDefaultDetection(t *testing.T) {
	// Test that WithDefaultDetection sets detection mode for all strategies
	t.Run("detection none", func(t *testing.T) {
		m := New(WithDefaultDetection(strategy.DetectionNone))

		feedA := gtfs.NewFeed()
		feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

		feedB := gtfs.NewFeed()
		feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

		merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			t.Fatalf("merge failed: %v", err)
		}

		// With DetectionNone, both agencies should be kept (one prefixed)
		if len(merged.Agencies) != 2 {
			t.Errorf("DetectionNone: expected 2 agencies, got %d", len(merged.Agencies))
		}
	})

	t.Run("detection identity", func(t *testing.T) {
		m := New(WithDefaultDetection(strategy.DetectionIdentity))

		feedA := gtfs.NewFeed()
		feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

		feedB := gtfs.NewFeed()
		feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

		merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			t.Fatalf("merge failed: %v", err)
		}

		// With DetectionIdentity, duplicates should be merged
		if len(merged.Agencies) != 1 {
			t.Errorf("DetectionIdentity: expected 1 agency (duplicate merged), got %d", len(merged.Agencies))
		}
	})

	t.Run("detection fuzzy uses stop matching", func(t *testing.T) {
		// Fuzzy detection is most relevant for stops (geo-proximity matching)
		// For agencies, fuzzy falls back to identity-like behavior
		m := New(WithDefaultDetection(strategy.DetectionFuzzy))

		feedA := gtfs.NewFeed()
		feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Main Station", Lat: 47.6, Lon: -122.3}

		feedB := gtfs.NewFeed()
		// Same location and similar name should trigger fuzzy match
		feedB.Stops["stop2"] = &gtfs.Stop{ID: "stop2", Name: "Main Station", Lat: 47.6, Lon: -122.3}

		merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
		if err != nil {
			t.Fatalf("merge failed: %v", err)
		}

		// With DetectionFuzzy, similar stops should be merged
		if len(merged.Stops) != 1 {
			t.Errorf("DetectionFuzzy: expected 1 stop (fuzzy matched), got %d", len(merged.Stops))
		}
	})
}

func TestWithDefaultLogging(t *testing.T) {
	// Test that WithDefaultLogging sets logging mode for all strategies
	tests := []struct {
		name    string
		logging strategy.DuplicateLogging
	}{
		{"logging none", strategy.LogNone},
		{"logging warning", strategy.LogWarning},
		{"logging error", strategy.LogError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create merger with the logging option
			m := New(WithDefaultLogging(tt.logging))

			// Verify the option was applied by checking if merge works
			// (LogError would cause errors if duplicates were detected,
			// but with DetectionNone no duplicates are detected)
			feedA := gtfs.NewFeed()
			feedA.Agencies["a1"] = &gtfs.Agency{ID: "a1", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

			_, err := m.MergeFeeds([]*gtfs.Feed{feedA})
			if err != nil {
				t.Fatalf("merge failed: %v", err)
			}
			// No error expected - option was successfully applied
		})
	}
}

func TestWithDefaultRenaming(t *testing.T) {
	// Test that WithDefaultRenaming sets renaming strategy for all strategies
	tests := []struct {
		name     string
		renaming strategy.RenamingStrategy
	}{
		{"renaming context", strategy.RenameContext},
		{"renaming agency", strategy.RenameAgency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create merger with the renaming option
			m := New(WithDefaultRenaming(tt.renaming))

			// Verify the option was applied by merging feeds
			feedA := gtfs.NewFeed()
			feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

			feedB := gtfs.NewFeed()
			feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

			merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
			if err != nil {
				t.Fatalf("merge failed: %v", err)
			}

			// With DetectionNone (default), both should be kept with different IDs
			if len(merged.Agencies) != 2 {
				t.Errorf("expected 2 agencies after merge, got %d", len(merged.Agencies))
			}

			// Verify the renaming strategy was applied (context prefixing)
			if tt.renaming == strategy.RenameContext {
				// With reverse processing: feedB processed first (no prefix) → "shared"
				// feedA processed second → "b-shared"
				if _, ok := merged.Agencies["shared"]; !ok {
					t.Error("expected agency with ID 'shared'")
				}
				if _, ok := merged.Agencies["b-shared"]; !ok {
					t.Error("expected agency with prefixed ID 'b-shared'")
				}
			}
		})
	}
}

func TestMultipleOptions(t *testing.T) {
	// Test that multiple options can be combined
	m := New(
		WithDebug(true),
		WithDefaultDetection(strategy.DetectionIdentity),
		WithDefaultLogging(strategy.LogWarning),
		WithDefaultRenaming(strategy.RenameContext),
	)

	// Verify debug was set
	if !m.debug {
		t.Error("expected debug to be true")
	}

	// Verify detection by merging overlapping feeds
	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop A", Lat: 47.6, Lon: -122.3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop B", Lat: 40.7, Lon: -74.0}

	merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// With DetectionIdentity, duplicates should be merged
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency with identity detection, got %d", len(merged.Agencies))
	}
	if len(merged.Stops) != 1 {
		t.Errorf("expected 1 stop with identity detection, got %d", len(merged.Stops))
	}
}

// ============================================================================
// Milestone 12.2: Per-Strategy Configuration Tests
// ============================================================================

func TestSetAgencyStrategy(t *testing.T) {
	m := New()

	// Create a custom agency strategy with different detection
	customStrategy := strategy.NewAgencyMergeStrategy()
	customStrategy.SetDuplicateDetection(strategy.DetectionIdentity)

	m.SetAgencyStrategy(customStrategy)

	// Verify the custom strategy is used
	s := m.GetStrategyForFile("agency.txt")
	if s == nil {
		t.Fatal("GetStrategyForFile returned nil for agency.txt")
	}

	// Verify by merging - agency strategy should use identity detection
	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}

	merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// With identity detection on agencies, duplicates should be merged
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency with identity detection, got %d", len(merged.Agencies))
	}
}

func TestSetStopStrategy(t *testing.T) {
	m := New()

	// Create a custom stop strategy with different detection
	customStrategy := strategy.NewStopMergeStrategy()
	customStrategy.SetDuplicateDetection(strategy.DetectionIdentity)

	m.SetStopStrategy(customStrategy)

	// Verify the custom strategy is used
	s := m.GetStrategyForFile("stops.txt")
	if s == nil {
		t.Fatal("GetStrategyForFile returned nil for stops.txt")
	}

	// Verify by merging - stop strategy should use identity detection
	feedA := gtfs.NewFeed()
	feedA.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop A", Lat: 47.6, Lon: -122.3}

	feedB := gtfs.NewFeed()
	feedB.Stops["stop1"] = &gtfs.Stop{ID: "stop1", Name: "Stop B", Lat: 40.7, Lon: -74.0}

	merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// With identity detection on stops, duplicates should be merged
	if len(merged.Stops) != 1 {
		t.Errorf("expected 1 stop with identity detection, got %d", len(merged.Stops))
	}
}

func TestCustomStrategyUsed(t *testing.T) {
	// Test that a custom strategy's behavior is actually used during merge
	m := New()

	// Set only the route strategy to use identity detection
	routeStrategy := strategy.NewRouteMergeStrategy()
	routeStrategy.SetDuplicateDetection(strategy.DetectionIdentity)
	m.SetRouteStrategy(routeStrategy)

	// Agency strategy should still be default (DetectionNone)
	// Route strategy should be identity detection

	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Routes["shared"] = &gtfs.Route{ID: "shared", AgencyID: "shared", ShortName: "R1", Type: 3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Routes["shared"] = &gtfs.Route{ID: "shared", AgencyID: "shared", ShortName: "R2", Type: 3}

	merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Agencies: DetectionNone - both kept (2 agencies)
	if len(merged.Agencies) != 2 {
		t.Errorf("expected 2 agencies (DetectionNone), got %d", len(merged.Agencies))
	}

	// Routes: DetectionIdentity - merged (1 route)
	if len(merged.Routes) != 1 {
		t.Errorf("expected 1 route (DetectionIdentity), got %d", len(merged.Routes))
	}
}

func TestAllStrategySetters(t *testing.T) {
	// Test that all strategy setters work
	m := New()

	// Test each setter
	m.SetAgencyStrategy(strategy.NewAgencyMergeStrategy())
	m.SetStopStrategy(strategy.NewStopMergeStrategy())
	m.SetRouteStrategy(strategy.NewRouteMergeStrategy())
	m.SetTripStrategy(strategy.NewTripMergeStrategy())
	m.SetCalendarStrategy(strategy.NewCalendarMergeStrategy())
	m.SetShapeStrategy(strategy.NewShapeMergeStrategy())
	m.SetFrequencyStrategy(strategy.NewFrequencyMergeStrategy())
	m.SetTransferStrategy(strategy.NewTransferMergeStrategy())
	m.SetFareAttributeStrategy(strategy.NewFareAttributeMergeStrategy())
	m.SetFareRuleStrategy(strategy.NewFareRuleMergeStrategy())
	m.SetFeedInfoStrategy(strategy.NewFeedInfoMergeStrategy())
	m.SetAreaStrategy(strategy.NewAreaMergeStrategy())

	// Verify all are non-nil via GetStrategyForFile
	files := []string{
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"calendar.txt",
		"shapes.txt",
		"frequencies.txt",
		"transfers.txt",
		"fare_attributes.txt",
		"fare_rules.txt",
		"feed_info.txt",
		"areas.txt",
	}

	for _, f := range files {
		s := m.GetStrategyForFile(f)
		if s == nil {
			t.Errorf("GetStrategyForFile(%q) returned nil after setting", f)
		}
	}

	// Verify merger still works
	feedA := gtfs.NewFeed()
	feedA.Agencies["a1"] = &gtfs.Agency{ID: "a1", Name: "Test", URL: "http://test.com", Timezone: "UTC"}

	_, err := m.MergeFeeds([]*gtfs.Feed{feedA})
	if err != nil {
		t.Fatalf("merge failed after setting all strategies: %v", err)
	}
}

func TestMixedStrategyConfiguration(t *testing.T) {
	// Test mixing global options with per-strategy configuration
	m := New(WithDefaultDetection(strategy.DetectionNone))

	// Override just the agency strategy to use identity detection
	agencyStrategy := strategy.NewAgencyMergeStrategy()
	agencyStrategy.SetDuplicateDetection(strategy.DetectionIdentity)
	m.SetAgencyStrategy(agencyStrategy)

	feedA := gtfs.NewFeed()
	feedA.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency A", URL: "http://a.com", Timezone: "UTC"}
	feedA.Stops["shared"] = &gtfs.Stop{ID: "shared", Name: "Stop A", Lat: 47.6, Lon: -122.3}

	feedB := gtfs.NewFeed()
	feedB.Agencies["shared"] = &gtfs.Agency{ID: "shared", Name: "Agency B", URL: "http://b.com", Timezone: "UTC"}
	feedB.Stops["shared"] = &gtfs.Stop{ID: "shared", Name: "Stop B", Lat: 40.7, Lon: -74.0}

	merged, err := m.MergeFeeds([]*gtfs.Feed{feedA, feedB})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Agency: identity detection (overridden) - merged to 1
	if len(merged.Agencies) != 1 {
		t.Errorf("expected 1 agency (identity override), got %d", len(merged.Agencies))
	}

	// Stops: none detection (global default) - both kept as 2
	if len(merged.Stops) != 2 {
		t.Errorf("expected 2 stops (global none), got %d", len(merged.Stops))
	}
}
