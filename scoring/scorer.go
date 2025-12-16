// Package scoring provides duplicate similarity scoring for GTFS merge operations.
package scoring

import (
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// Scorer calculates similarity between entities.
// Score returns a similarity score between 0.0 and 1.0.
// 0.0 = completely different, 1.0 = identical.
type Scorer[T any] interface {
	Score(ctx *strategy.MergeContext, source, target T) float64
}

// PropertyMatcher scores based on matching property values.
// For each property function, if source and target return the same value,
// it counts as a match. The final score is matches/total.
type PropertyMatcher[T any] struct {
	Properties []func(T) string
}

// Score returns the proportion of properties that match.
// Returns 1.0 if all properties match, 0.0 if none match.
// Returns 1.0 if there are no properties defined.
func (p *PropertyMatcher[T]) Score(ctx *strategy.MergeContext, source, target T) float64 {
	if len(p.Properties) == 0 {
		return 1.0
	}

	matches := 0
	for _, prop := range p.Properties {
		if prop(source) == prop(target) {
			matches++
		}
	}

	return float64(matches) / float64(len(p.Properties))
}

// AndScorer combines multiple scorers using MULTIPLICATION.
// Final score = scorer1 * scorer2 * ... * scorerN
// A single 0.0 score fails the entire match (early exit optimization).
// All entity merge strategies use AndScorer to combine their scoring rules.
type AndScorer[T any] struct {
	Scorers   []Scorer[T]
	Threshold float64 // Minimum score to be considered a match
}

// Score returns the product of all individual scorer scores.
// Returns 0.0 immediately if any scorer returns 0.0 (early exit).
// Returns 1.0 if there are no scorers defined (neutral element for multiplication).
func (a *AndScorer[T]) Score(ctx *strategy.MergeContext, source, target T) float64 {
	if len(a.Scorers) == 0 {
		return 1.0
	}

	result := 1.0
	for _, scorer := range a.Scorers {
		score := scorer.Score(ctx, source, target)
		if score == 0.0 {
			return 0.0 // Early exit on zero
		}
		result *= score
	}

	return result
}

// ElementOverlapScore calculates the overlap score between two sets.
// Formula: (common_count / a.size + common_count / b.size) / 2
// Returns 0.0 if either collection is empty.
// Returns 1.0 for identical sets.
func ElementOverlapScore[T comparable](a, b []T) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Build a set from b for O(1) lookups
	bSet := make(map[T]struct{}, len(b))
	for _, item := range b {
		bSet[item] = struct{}{}
	}

	// Count common elements
	common := 0
	for _, item := range a {
		if _, ok := bSet[item]; ok {
			common++
		}
	}

	// Apply formula: (common/a + common/b) / 2
	scoreA := float64(common) / float64(len(a))
	scoreB := float64(common) / float64(len(b))
	return (scoreA + scoreB) / 2.0
}

// IntervalOverlapScore calculates the overlap score between two intervals.
// Formula: (overlap / interval_a_length + overlap / interval_b_length) / 2
// Returns 0.0 if either interval has zero length or there's no overlap.
func IntervalOverlapScore(start1, end1, start2, end2 float64) float64 {
	len1 := end1 - start1
	len2 := end2 - start2

	if len1 <= 0 || len2 <= 0 {
		return 0.0
	}

	// Calculate overlap
	overlapStart := max(start1, start2)
	overlapEnd := min(end1, end2)
	overlap := overlapEnd - overlapStart

	if overlap <= 0 {
		return 0.0
	}

	// Apply formula: (overlap/len1 + overlap/len2) / 2
	scoreA := overlap / len1
	scoreB := overlap / len2
	return (scoreA + scoreB) / 2.0
}
