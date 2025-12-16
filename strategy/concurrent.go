package strategy

import (
	"runtime"
	"sync"
)

// ConcurrentConfig holds configuration for concurrent operations
type ConcurrentConfig struct {
	// Enabled controls whether concurrent processing is used
	Enabled bool
	// NumWorkers is the number of worker goroutines (default: runtime.NumCPU())
	NumWorkers int
	// MinItemsForConcurrency is the minimum number of items to trigger concurrent processing
	MinItemsForConcurrency int
}

// DefaultConcurrentConfig returns the default concurrent configuration
func DefaultConcurrentConfig() ConcurrentConfig {
	return ConcurrentConfig{
		Enabled:                false,
		NumWorkers:             runtime.NumCPU(),
		MinItemsForConcurrency: 100,
	}
}

// scoredResult holds the result of scoring an entity
type scoredResult[ID comparable] struct {
	ID    ID
	Score float64
}

// findBestMatchConcurrent finds the best scoring match from a collection using concurrent processing.
// It takes a slice of candidates and a scoring function, and returns the ID of the best match
// (above threshold) or the zero value if no match is found.
func findBestMatchConcurrent[T any, ID comparable](
	candidates []T,
	getID func(T) ID,
	score func(T) float64,
	threshold float64,
	config ConcurrentConfig,
) ID {
	var zeroID ID

	if len(candidates) == 0 {
		return zeroID
	}

	// Fall back to sequential if concurrent is disabled or not enough items
	if !config.Enabled || len(candidates) < config.MinItemsForConcurrency {
		return findBestMatchSequential(candidates, getID, score, threshold)
	}

	numWorkers := config.NumWorkers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	// Limit workers to number of candidates
	if numWorkers > len(candidates) {
		numWorkers = len(candidates)
	}

	// Channel for work items
	jobs := make(chan T, len(candidates))
	results := make(chan scoredResult[ID], len(candidates))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				s := score(candidate)
				if s >= threshold {
					results <- scoredResult[ID]{
						ID:    getID(candidate),
						Score: s,
					}
				}
			}
		}()
	}

	// Send work
	go func() {
		for _, candidate := range candidates {
			jobs <- candidate
		}
		close(jobs)
	}()

	// Wait for workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect best result
	var bestID ID
	var bestScore float64
	for result := range results {
		if result.Score > bestScore {
			bestScore = result.Score
			bestID = result.ID
		}
	}

	return bestID
}

// findBestMatchSequential finds the best scoring match sequentially
func findBestMatchSequential[T any, ID comparable](
	candidates []T,
	getID func(T) ID,
	score func(T) float64,
	threshold float64,
) ID {
	var zeroID ID
	var bestID ID
	var bestScore float64

	for _, candidate := range candidates {
		s := score(candidate)
		if s >= threshold && s > bestScore {
			bestScore = s
			bestID = getID(candidate)
		}
	}

	if bestScore >= threshold {
		return bestID
	}
	return zeroID
}

// ConcurrentScorer provides concurrent scoring capabilities to strategies
type ConcurrentScorer struct {
	Config ConcurrentConfig
}

// NewConcurrentScorer creates a new ConcurrentScorer with default configuration
func NewConcurrentScorer() *ConcurrentScorer {
	return &ConcurrentScorer{
		Config: DefaultConcurrentConfig(),
	}
}

// SetConcurrent enables or disables concurrent processing
func (cs *ConcurrentScorer) SetConcurrent(enabled bool) {
	cs.Config.Enabled = enabled
}

// SetNumWorkers sets the number of worker goroutines
func (cs *ConcurrentScorer) SetNumWorkers(n int) {
	if n > 0 {
		cs.Config.NumWorkers = n
	}
}

// SetMinItemsForConcurrency sets the minimum number of items for concurrent processing
func (cs *ConcurrentScorer) SetMinItemsForConcurrency(n int) {
	if n > 0 {
		cs.Config.MinItemsForConcurrency = n
	}
}
