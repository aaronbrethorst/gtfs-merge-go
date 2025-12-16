package scoring

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// TestScorerInterface verifies that the Scorer interface is correctly defined
func TestScorerInterface(t *testing.T) {
	// Create a mock scorer that implements the interface
	var _ Scorer[*gtfs.Stop] = &mockStopScorer{}
	var _ Scorer[*gtfs.Route] = &mockRouteScorer{}
	var _ Scorer[*gtfs.Trip] = &mockTripScorer{}
}

// mockStopScorer is a test implementation of Scorer for stops
type mockStopScorer struct {
	score float64
}

func (m *mockStopScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Stop) float64 {
	return m.score
}

// mockRouteScorer is a test implementation of Scorer for routes
type mockRouteScorer struct {
	score float64
}

func (m *mockRouteScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Route) float64 {
	return m.score
}

// mockTripScorer is a test implementation of Scorer for trips
type mockTripScorer struct {
	score float64
}

func (m *mockTripScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Trip) float64 {
	return m.score
}

// TestPropertyMatcherAllMatch verifies score is 1.0 when all properties match
func TestPropertyMatcherAllMatch(t *testing.T) {
	matcher := &PropertyMatcher[*gtfs.Agency]{
		Properties: []func(*gtfs.Agency) string{
			func(a *gtfs.Agency) string { return a.Name },
			func(a *gtfs.Agency) string { return a.URL },
		},
	}

	source := &gtfs.Agency{Name: "Test Agency", URL: "http://test.com"}
	target := &gtfs.Agency{Name: "Test Agency", URL: "http://test.com"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := matcher.Score(ctx, source, target)

	if score != 1.0 {
		t.Errorf("Expected score 1.0 when all properties match, got %f", score)
	}
}

// TestPropertyMatcherNoneMatch verifies score is 0.0 when no properties match
func TestPropertyMatcherNoneMatch(t *testing.T) {
	matcher := &PropertyMatcher[*gtfs.Agency]{
		Properties: []func(*gtfs.Agency) string{
			func(a *gtfs.Agency) string { return a.Name },
			func(a *gtfs.Agency) string { return a.URL },
		},
	}

	source := &gtfs.Agency{Name: "Agency A", URL: "http://a.com"}
	target := &gtfs.Agency{Name: "Agency B", URL: "http://b.com"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := matcher.Score(ctx, source, target)

	if score != 0.0 {
		t.Errorf("Expected score 0.0 when no properties match, got %f", score)
	}
}

// TestPropertyMatcherPartialMatch verifies proportional score for partial matches
func TestPropertyMatcherPartialMatch(t *testing.T) {
	matcher := &PropertyMatcher[*gtfs.Agency]{
		Properties: []func(*gtfs.Agency) string{
			func(a *gtfs.Agency) string { return a.Name },
			func(a *gtfs.Agency) string { return a.URL },
			func(a *gtfs.Agency) string { return a.Timezone },
			func(a *gtfs.Agency) string { return a.Lang },
		},
	}

	// 2 out of 4 properties match
	source := &gtfs.Agency{Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York", Lang: "en"}
	target := &gtfs.Agency{Name: "Test Agency", URL: "http://other.com", Timezone: "America/New_York", Lang: "es"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := matcher.Score(ctx, source, target)

	expected := 0.5 // 2/4
	if score != expected {
		t.Errorf("Expected score %f for partial match, got %f", expected, score)
	}
}

// TestPropertyMatcherEmptyProperties verifies behavior with no properties
func TestPropertyMatcherEmptyProperties(t *testing.T) {
	matcher := &PropertyMatcher[*gtfs.Agency]{
		Properties: []func(*gtfs.Agency) string{},
	}

	source := &gtfs.Agency{Name: "Test"}
	target := &gtfs.Agency{Name: "Test"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := matcher.Score(ctx, source, target)

	// With no properties, everything is considered a match
	if score != 1.0 {
		t.Errorf("Expected score 1.0 with empty properties, got %f", score)
	}
}

// TestAndScorerAllAboveThreshold verifies combined scoring with multiplication
func TestAndScorerAllAboveThreshold(t *testing.T) {
	scorer := &AndScorer[*gtfs.Stop]{
		Scorers: []Scorer[*gtfs.Stop]{
			&mockStopScorer{score: 0.8},
			&mockStopScorer{score: 0.9},
		},
		Threshold: 0.5,
	}

	source := &gtfs.Stop{Name: "Stop A"}
	target := &gtfs.Stop{Name: "Stop B"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	// 0.8 * 0.9 = 0.72
	expected := 0.72
	if score < expected-0.001 || score > expected+0.001 {
		t.Errorf("Expected score ~%f (0.8 * 0.9), got %f", expected, score)
	}
}

// TestAndScorerBelowThreshold verifies returns 0 if any scorer returns 0
func TestAndScorerBelowThreshold(t *testing.T) {
	scorer := &AndScorer[*gtfs.Stop]{
		Scorers: []Scorer[*gtfs.Stop]{
			&mockStopScorer{score: 0.8},
			&mockStopScorer{score: 0.0}, // Zero score
			&mockStopScorer{score: 0.9},
		},
		Threshold: 0.5,
	}

	source := &gtfs.Stop{Name: "Stop A"}
	target := &gtfs.Stop{Name: "Stop B"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	// A single 0.0 should result in 0.0 (early exit)
	if score != 0.0 {
		t.Errorf("Expected score 0.0 when any scorer returns 0, got %f", score)
	}
}

// TestAndScorerEarlyExit verifies early exit optimization on zero score
func TestAndScorerEarlyExit(t *testing.T) {
	callCount := 0
	firstScorer := &countingScorer{score: 0.0, callCount: &callCount}
	secondScorer := &countingScorer{score: 1.0, callCount: &callCount}

	scorer := &AndScorer[*gtfs.Stop]{
		Scorers: []Scorer[*gtfs.Stop]{
			firstScorer,
			secondScorer,
		},
		Threshold: 0.5,
	}

	source := &gtfs.Stop{Name: "Stop A"}
	target := &gtfs.Stop{Name: "Stop B"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	scorer.Score(ctx, source, target)

	// Only the first scorer should be called (early exit on 0)
	if callCount != 1 {
		t.Errorf("Expected early exit after first zero score (1 call), got %d calls", callCount)
	}
}

// countingScorer tracks how many times it's called for testing early exit
type countingScorer struct {
	score     float64
	callCount *int
}

func (c *countingScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Stop) float64 {
	*c.callCount++
	return c.score
}

// TestAndScorerEmpty verifies behavior with no scorers
func TestAndScorerEmpty(t *testing.T) {
	scorer := &AndScorer[*gtfs.Stop]{
		Scorers:   []Scorer[*gtfs.Stop]{},
		Threshold: 0.5,
	}

	source := &gtfs.Stop{Name: "Stop A"}
	target := &gtfs.Stop{Name: "Stop B"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	// With no scorers, should return 1.0 (neutral element for multiplication)
	if score != 1.0 {
		t.Errorf("Expected score 1.0 with empty scorers, got %f", score)
	}
}

// TestAndScorerSingleScorer verifies single scorer passthrough
func TestAndScorerSingleScorer(t *testing.T) {
	scorer := &AndScorer[*gtfs.Stop]{
		Scorers: []Scorer[*gtfs.Stop]{
			&mockStopScorer{score: 0.75},
		},
		Threshold: 0.5,
	}

	source := &gtfs.Stop{Name: "Stop A"}
	target := &gtfs.Stop{Name: "Stop B"}

	ctx := strategy.NewMergeContext(gtfs.NewFeed(), gtfs.NewFeed(), "")
	score := scorer.Score(ctx, source, target)

	if score != 0.75 {
		t.Errorf("Expected score 0.75 with single scorer, got %f", score)
	}
}
