package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// TestAutoDetectIdentityWhenIDsOverlap verifies that when feeds share many IDs,
// auto-detection returns DetectionIdentity
func TestAutoDetectIdentityWhenIDsOverlap(t *testing.T) {
	// Create source feed with specific IDs
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Agency One"}
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One", Lat: 47.6, Lon: -122.3}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop Two", Lat: 47.61, Lon: -122.31}
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1"}

	// Create target feed with same IDs (high overlap)
	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency1")] = &gtfs.Agency{ID: "agency1", Name: "Agency One Target"}
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One Target", Lat: 47.6, Lon: -122.3}
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop Two Target", Lat: 47.61, Lon: -122.31}
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", AgencyID: "agency1", ShortName: "R1"}

	// Auto-detect should choose Identity mode due to high ID overlap
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for high ID overlap, got %v", detection)
	}
}

// TestAutoDetectIdentityPartialOverlap verifies that partial ID overlap (above threshold)
// still returns DetectionIdentity
func TestAutoDetectIdentityPartialOverlap(t *testing.T) {
	// Create source feed
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One"}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop Two"}
	source.Stops[gtfs.StopID("stop3")] = &gtfs.Stop{ID: "stop3", Name: "Stop Three"}
	source.Stops[gtfs.StopID("stop4")] = &gtfs.Stop{ID: "stop4", Name: "Stop Four"}

	// Create target feed with 50% overlap (2 of 4 stops match)
	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One"}
	target.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop Two"}
	target.Stops[gtfs.StopID("stopA")] = &gtfs.Stop{ID: "stopA", Name: "Stop A"}
	target.Stops[gtfs.StopID("stopB")] = &gtfs.Stop{ID: "stopB", Name: "Stop B"}

	// With 50% overlap, should still use Identity mode (at threshold)
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for partial ID overlap >= threshold, got %v", detection)
	}
}

// TestAutoDetectFuzzyWhenNoIDOverlap verifies that when feeds share no IDs
// but have similar entities, auto-detection returns DetectionFuzzy
func TestAutoDetectFuzzyWhenNoIDOverlap(t *testing.T) {
	// Create source feed
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("src_agency")] = &gtfs.Agency{ID: "src_agency", Name: "Transit Agency"}
	source.Stops[gtfs.StopID("src_stop1")] = &gtfs.Stop{ID: "src_stop1", Name: "Main Street Station", Lat: 47.6, Lon: -122.3}
	source.Stops[gtfs.StopID("src_stop2")] = &gtfs.Stop{ID: "src_stop2", Name: "Downtown Stop", Lat: 47.61, Lon: -122.31}

	// Create target feed with different IDs but similar entities (same names, nearby locations)
	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("tgt_agency")] = &gtfs.Agency{ID: "tgt_agency", Name: "Transit Agency"}
	target.Stops[gtfs.StopID("tgt_stop1")] = &gtfs.Stop{ID: "tgt_stop1", Name: "Main Street Station", Lat: 47.6001, Lon: -122.3001} // Within 50m
	target.Stops[gtfs.StopID("tgt_stop2")] = &gtfs.Stop{ID: "tgt_stop2", Name: "Downtown Stop", Lat: 47.6101, Lon: -122.3101}       // Within 50m

	// Auto-detect should choose Fuzzy mode (no ID overlap but similar entities)
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionFuzzy {
		t.Errorf("Expected DetectionFuzzy for no ID overlap with similar entities, got %v", detection)
	}
}

// TestAutoDetectFuzzyWithSimilarAgencies verifies that agencies with same name
// but different IDs trigger Fuzzy detection
func TestAutoDetectFuzzyWithSimilarAgencies(t *testing.T) {
	// Create source feed
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("agency_a")] = &gtfs.Agency{
		ID:   "agency_a",
		Name: "Metro Transit",
		URL:  "https://metro.example.com",
	}

	// Create target feed with same agency name/url but different ID
	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("agency_b")] = &gtfs.Agency{
		ID:   "agency_b",
		Name: "Metro Transit",
		URL:  "https://metro.example.com",
	}

	// Should detect fuzzy potential
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionFuzzy {
		t.Errorf("Expected DetectionFuzzy for similar agencies with different IDs, got %v", detection)
	}
}

// TestAutoDetectNoneWhenNoSimilarity verifies that when feeds are completely different,
// auto-detection returns DetectionNone
func TestAutoDetectNoneWhenNoSimilarity(t *testing.T) {
	// Create source feed
	source := gtfs.NewFeed()
	source.Agencies[gtfs.AgencyID("seattle_metro")] = &gtfs.Agency{
		ID:   "seattle_metro",
		Name: "Seattle Metro",
		URL:  "https://seattle.gov/metro",
	}
	source.Stops[gtfs.StopID("sea_stop1")] = &gtfs.Stop{
		ID:   "sea_stop1",
		Name: "Pike Place Market",
		Lat:  47.6097,
		Lon:  -122.3425,
	}

	// Create completely different target feed (different city, different entities)
	target := gtfs.NewFeed()
	target.Agencies[gtfs.AgencyID("nyc_mta")] = &gtfs.Agency{
		ID:   "nyc_mta",
		Name: "NYC MTA",
		URL:  "https://mta.info",
	}
	target.Stops[gtfs.StopID("nyc_stop1")] = &gtfs.Stop{
		ID:   "nyc_stop1",
		Name: "Times Square",
		Lat:  40.7580,
		Lon:  -73.9855, // Very far from Seattle
	}

	// Auto-detect should choose None mode (no overlap, no similarity)
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionNone {
		t.Errorf("Expected DetectionNone for completely different feeds, got %v", detection)
	}
}

// TestAutoDetectNoneForEmptyFeeds verifies that empty feeds return DetectionNone
func TestAutoDetectNoneForEmptyFeeds(t *testing.T) {
	source := gtfs.NewFeed()
	target := gtfs.NewFeed()

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionNone {
		t.Errorf("Expected DetectionNone for empty feeds, got %v", detection)
	}
}

// TestAutoDetectNoneWhenSourceEmpty verifies that empty source returns DetectionNone
func TestAutoDetectNoneWhenSourceEmpty(t *testing.T) {
	source := gtfs.NewFeed()

	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop"}

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionNone {
		t.Errorf("Expected DetectionNone for empty source feed, got %v", detection)
	}
}

// TestAutoDetectNoneWhenTargetEmpty verifies that empty target returns DetectionNone
func TestAutoDetectNoneWhenTargetEmpty(t *testing.T) {
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop"}

	target := gtfs.NewFeed()

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionNone {
		t.Errorf("Expected DetectionNone for empty target feed, got %v", detection)
	}
}

// TestAutoDetectWithRouteOverlap verifies route ID overlap triggers Identity detection
func TestAutoDetectWithRouteOverlap(t *testing.T) {
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Routes[gtfs.RouteID("route2")] = &gtfs.Route{ID: "route2", ShortName: "2"}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Routes[gtfs.RouteID("route2")] = &gtfs.Route{ID: "route2", ShortName: "2"}

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for route ID overlap, got %v", detection)
	}
}

// TestAutoDetectWithTripOverlap verifies trip ID overlap triggers Identity detection
func TestAutoDetectWithTripOverlap(t *testing.T) {
	source := gtfs.NewFeed()
	source.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{ID: "trip1"}
	source.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{ID: "trip2"}

	target := gtfs.NewFeed()
	target.Trips[gtfs.TripID("trip1")] = &gtfs.Trip{ID: "trip1"}
	target.Trips[gtfs.TripID("trip2")] = &gtfs.Trip{ID: "trip2"}

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for trip ID overlap, got %v", detection)
	}
}

// TestAutoDetectBelowThreshold verifies that low overlap returns Fuzzy/None based on similarity
func TestAutoDetectBelowThreshold(t *testing.T) {
	// Create source feed with many stops
	source := gtfs.NewFeed()
	for i := 0; i < 10; i++ {
		id := gtfs.StopID(string(rune('a' + i)))
		source.Stops[id] = &gtfs.Stop{ID: id, Name: "Stop " + string(id)}
	}

	// Create target feed with only 1 overlapping ID (10% overlap, below 50% threshold)
	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("a")] = &gtfs.Stop{ID: "a", Name: "Stop a"}
	for i := 0; i < 9; i++ {
		id := gtfs.StopID("x" + string(rune('0'+i)))
		target.Stops[id] = &gtfs.Stop{ID: id, Name: "Stop x" + string(rune('0'+i))}
	}

	// With low ID overlap and no fuzzy similarity, should return None
	detection := AutoDetectDuplicateDetection(source, target)

	// Below threshold should NOT return Identity
	if detection == DetectionIdentity {
		t.Errorf("Expected non-Identity detection for low ID overlap, got %v", detection)
	}
}

// TestAutoDetectWithCalendarOverlap verifies calendar/service ID overlap triggers Identity detection
func TestAutoDetectWithCalendarOverlap(t *testing.T) {
	source := gtfs.NewFeed()
	source.Calendars[gtfs.ServiceID("weekday")] = &gtfs.Calendar{ServiceID: "weekday"}
	source.Calendars[gtfs.ServiceID("weekend")] = &gtfs.Calendar{ServiceID: "weekend"}

	target := gtfs.NewFeed()
	target.Calendars[gtfs.ServiceID("weekday")] = &gtfs.Calendar{ServiceID: "weekday"}
	target.Calendars[gtfs.ServiceID("weekend")] = &gtfs.Calendar{ServiceID: "weekend"}

	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for calendar ID overlap, got %v", detection)
	}
}

// TestAutoDetectThresholds tests the configured thresholds for auto-detection
func TestAutoDetectThresholds(t *testing.T) {
	config := DefaultAutoDetectConfig()

	// Default thresholds should be 0.5
	if config.MinElementsInCommonScoreForAutoDetect != 0.5 {
		t.Errorf("Expected MinElementsInCommonScoreForAutoDetect to be 0.5, got %v", config.MinElementsInCommonScoreForAutoDetect)
	}
	if config.MinElementsDuplicateScoreForAutoDetect != 0.5 {
		t.Errorf("Expected MinElementsDuplicateScoreForAutoDetect to be 0.5, got %v", config.MinElementsDuplicateScoreForAutoDetect)
	}
}

// TestAutoDetectWithConfiguredThresholds tests custom threshold configuration
func TestAutoDetectWithConfiguredThresholds(t *testing.T) {
	source := gtfs.NewFeed()
	source.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One"}
	source.Stops[gtfs.StopID("stop2")] = &gtfs.Stop{ID: "stop2", Name: "Stop Two"}
	source.Stops[gtfs.StopID("stop3")] = &gtfs.Stop{ID: "stop3", Name: "Stop Three"}

	// 1 of 3 overlap (33%)
	target := gtfs.NewFeed()
	target.Stops[gtfs.StopID("stop1")] = &gtfs.Stop{ID: "stop1", Name: "Stop One"}
	target.Stops[gtfs.StopID("stopA")] = &gtfs.Stop{ID: "stopA", Name: "Stop A"}
	target.Stops[gtfs.StopID("stopB")] = &gtfs.Stop{ID: "stopB", Name: "Stop B"}

	// With default threshold (0.5), this should not return Identity
	detection := AutoDetectDuplicateDetection(source, target)
	if detection == DetectionIdentity {
		t.Errorf("Expected non-Identity for 33%% overlap with default threshold, got %v", detection)
	}

	// With lowered threshold (0.3), this should return Identity
	config := AutoDetectConfig{
		MinElementsInCommonScoreForAutoDetect:  0.3,
		MinElementsDuplicateScoreForAutoDetect: 0.3,
	}
	detectionWithConfig := AutoDetectDuplicateDetectionWithConfig(source, target, config)
	if detectionWithConfig != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity for 33%% overlap with lowered threshold, got %v", detectionWithConfig)
	}
}

// TestAutoDetectRouteOverlapNoStopOverlap verifies that routes with same IDs
// but stops with different IDs still triggers Identity detection
func TestAutoDetectRouteOverlapNoStopOverlap(t *testing.T) {
	source := gtfs.NewFeed()
	source.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	source.Stops[gtfs.StopID("src_stop1")] = &gtfs.Stop{ID: "src_stop1", Name: "Source Stop"}

	target := gtfs.NewFeed()
	target.Routes[gtfs.RouteID("route1")] = &gtfs.Route{ID: "route1", ShortName: "1"}
	target.Stops[gtfs.StopID("tgt_stop1")] = &gtfs.Stop{ID: "tgt_stop1", Name: "Target Stop"}

	// Route IDs overlap, so should return Identity even though stop IDs don't
	detection := AutoDetectDuplicateDetection(source, target)

	if detection != DetectionIdentity {
		t.Errorf("Expected DetectionIdentity when any entity type has significant ID overlap, got %v", detection)
	}
}
