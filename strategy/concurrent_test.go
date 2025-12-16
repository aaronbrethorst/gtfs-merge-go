package strategy

import (
	"runtime"
	"sync"
	"testing"
)

func TestDefaultConcurrentConfig(t *testing.T) {
	config := DefaultConcurrentConfig()

	if config.Enabled {
		t.Error("Expected Enabled to be false by default")
	}
	if config.NumWorkers != runtime.NumCPU() {
		t.Errorf("Expected NumWorkers to be %d, got %d", runtime.NumCPU(), config.NumWorkers)
	}
	if config.MinItemsForConcurrency != 100 {
		t.Errorf("Expected MinItemsForConcurrency to be 100, got %d", config.MinItemsForConcurrency)
	}
}

func TestConcurrentScorerSetters(t *testing.T) {
	cs := NewConcurrentScorer()

	// Test SetConcurrent
	cs.SetConcurrent(true)
	if !cs.Config.Enabled {
		t.Error("Expected Enabled to be true after SetConcurrent(true)")
	}

	// Test SetNumWorkers
	cs.SetNumWorkers(4)
	if cs.Config.NumWorkers != 4 {
		t.Errorf("Expected NumWorkers to be 4, got %d", cs.Config.NumWorkers)
	}

	// Test SetNumWorkers with invalid value
	cs.SetNumWorkers(-1)
	if cs.Config.NumWorkers != 4 {
		t.Error("Expected NumWorkers to remain 4 after invalid SetNumWorkers(-1)")
	}

	// Test SetMinItemsForConcurrency
	cs.SetMinItemsForConcurrency(50)
	if cs.Config.MinItemsForConcurrency != 50 {
		t.Errorf("Expected MinItemsForConcurrency to be 50, got %d", cs.Config.MinItemsForConcurrency)
	}
}

type testCandidate struct {
	id    string
	value int
}

func TestFindBestMatchSequential(t *testing.T) {
	candidates := []testCandidate{
		{id: "a", value: 50},
		{id: "b", value: 80},
		{id: "c", value: 60},
		{id: "d", value: 90},
		{id: "e", value: 70},
	}

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 100.0 }

	// Test finding best match above threshold
	result := findBestMatchSequential(candidates, getID, score, 0.5)
	if result != "d" {
		t.Errorf("Expected best match 'd' (score 0.9), got '%s'", result)
	}

	// Test with higher threshold
	result = findBestMatchSequential(candidates, getID, score, 0.85)
	if result != "d" {
		t.Errorf("Expected best match 'd' (score 0.9), got '%s'", result)
	}

	// Test with threshold too high
	result = findBestMatchSequential(candidates, getID, score, 0.95)
	if result != "" {
		t.Errorf("Expected no match with threshold 0.95, got '%s'", result)
	}

	// Test with empty candidates
	result = findBestMatchSequential([]testCandidate{}, getID, score, 0.5)
	if result != "" {
		t.Errorf("Expected empty result for empty candidates, got '%s'", result)
	}
}

func TestFindBestMatchConcurrent_Disabled(t *testing.T) {
	candidates := make([]testCandidate, 200)
	for i := 0; i < 200; i++ {
		candidates[i] = testCandidate{id: string(rune('a' + i%26)), value: i}
	}

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 200.0 }

	// With concurrent disabled, should use sequential
	config := DefaultConcurrentConfig()
	config.Enabled = false

	result := findBestMatchConcurrent(candidates, getID, score, 0.5, config)
	// The last candidate has value 199, so score = 0.995
	// ID will be candidates[199].id
	if result == "" {
		t.Error("Expected a match, got empty string")
	}
}

func TestFindBestMatchConcurrent_Enabled(t *testing.T) {
	candidates := make([]testCandidate, 200)
	for i := 0; i < 200; i++ {
		candidates[i] = testCandidate{id: string(rune('0'+i/26)) + string(rune('a'+i%26)), value: i}
	}
	// Set specific best candidate
	candidates[150].value = 300 // This will have highest score

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 400.0 }

	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             4,
		MinItemsForConcurrency: 50,
	}

	result := findBestMatchConcurrent(candidates, getID, score, 0.5, config)
	expectedID := candidates[150].id
	if result != expectedID {
		t.Errorf("Expected best match '%s', got '%s'", expectedID, result)
	}
}

func TestFindBestMatchConcurrent_BelowMinItems(t *testing.T) {
	// With only 10 items and MinItemsForConcurrency=100, should use sequential
	candidates := make([]testCandidate, 10)
	for i := 0; i < 10; i++ {
		candidates[i] = testCandidate{id: string(rune('a' + i)), value: (i + 1) * 10}
	}

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 100.0 }

	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             4,
		MinItemsForConcurrency: 100,
	}

	result := findBestMatchConcurrent(candidates, getID, score, 0.5, config)
	// Last candidate has score 1.0
	if result != "j" {
		t.Errorf("Expected best match 'j', got '%s'", result)
	}
}

func TestConcurrentScoringCorrectness(t *testing.T) {
	// Verify that concurrent and sequential produce the same results
	candidates := make([]testCandidate, 500)
	for i := 0; i < 500; i++ {
		candidates[i] = testCandidate{
			id:    string(rune('0'+i/100)) + string(rune('0'+i/10%10)) + string(rune('a'+i%10)),
			value: i,
		}
	}
	// Insert a specific best candidate
	candidates[333].value = 1000

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 1100.0 }
	threshold := 0.5

	// Sequential result
	sequentialResult := findBestMatchSequential(candidates, getID, score, threshold)

	// Concurrent result
	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             4,
		MinItemsForConcurrency: 10,
	}
	concurrentResult := findBestMatchConcurrent(candidates, getID, score, threshold, config)

	if sequentialResult != concurrentResult {
		t.Errorf("Sequential result '%s' does not match concurrent result '%s'", sequentialResult, concurrentResult)
	}
}

func TestConcurrentScoringPerformance(t *testing.T) {
	// This test ensures concurrent processing doesn't have correctness issues under load
	candidates := make([]testCandidate, 1000)
	for i := 0; i < 1000; i++ {
		candidates[i] = testCandidate{
			id:    string(rune('0'+i/100)) + string(rune('0'+i/10%10)) + string(rune('a'+i%10)),
			value: i % 100, // All scores are low
		}
	}
	// Insert one high-scoring candidate
	candidates[500].value = 500

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 600.0 }
	threshold := 0.7

	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             runtime.NumCPU(),
		MinItemsForConcurrency: 10,
	}

	// Run multiple times to check for race conditions
	var wg sync.WaitGroup
	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = findBestMatchConcurrent(candidates, getID, score, threshold, config)
		}(i)
	}
	wg.Wait()

	// All results should be the same
	expected := candidates[500].id
	for i, result := range results {
		if result != expected {
			t.Errorf("Run %d: expected '%s', got '%s'", i, expected, result)
		}
	}
}

func TestFindBestMatchConcurrent_EmptyCandidates(t *testing.T) {
	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return 1.0 }

	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             4,
		MinItemsForConcurrency: 10,
	}

	result := findBestMatchConcurrent([]testCandidate{}, getID, score, 0.5, config)
	if result != "" {
		t.Errorf("Expected empty result for empty candidates, got '%s'", result)
	}
}

func TestFindBestMatchConcurrent_NoBestMatch(t *testing.T) {
	candidates := []testCandidate{
		{id: "a", value: 10},
		{id: "b", value: 20},
		{id: "c", value: 30},
	}

	getID := func(c testCandidate) string { return c.id }
	score := func(c testCandidate) float64 { return float64(c.value) / 100.0 } // All below 0.5

	config := ConcurrentConfig{
		Enabled:                true,
		NumWorkers:             2,
		MinItemsForConcurrency: 1,
	}

	result := findBestMatchConcurrent(candidates, getID, score, 0.5, config)
	if result != "" {
		t.Errorf("Expected no match (all below threshold), got '%s'", result)
	}
}
