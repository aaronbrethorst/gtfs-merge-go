package strategy

import (
	"strings"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// AutoDetectConfig holds thresholds for auto-detection of duplicate detection strategy
type AutoDetectConfig struct {
	// MinElementsInCommonScoreForAutoDetect is the ID overlap score needed to consider IDENTITY mode
	MinElementsInCommonScoreForAutoDetect float64

	// MinElementsDuplicateScoreForAutoDetect is the entity match score for strategy selection
	MinElementsDuplicateScoreForAutoDetect float64
}

// DefaultAutoDetectConfig returns the default configuration for auto-detection
func DefaultAutoDetectConfig() AutoDetectConfig {
	return AutoDetectConfig{
		MinElementsInCommonScoreForAutoDetect:  0.5,
		MinElementsDuplicateScoreForAutoDetect: 0.5,
	}
}

// AutoDetectDuplicateDetection automatically chooses the best duplicate detection strategy
// based on feed analysis using default thresholds.
func AutoDetectDuplicateDetection(source, target *gtfs.Feed) DuplicateDetection {
	return AutoDetectDuplicateDetectionWithConfig(source, target, DefaultAutoDetectConfig())
}

// AutoDetectDuplicateDetectionWithConfig automatically chooses the best duplicate detection strategy
// based on feed analysis using the provided configuration.
func AutoDetectDuplicateDetectionWithConfig(source, target *gtfs.Feed, config AutoDetectConfig) DuplicateDetection {
	// Handle empty feeds
	if isEmpty(source) || isEmpty(target) {
		return DetectionNone
	}

	// Calculate ID overlap scores for different entity types
	idOverlapScore := calculateIDOverlapScore(source, target)

	// If ID overlap is significant, use Identity detection
	if idOverlapScore >= config.MinElementsInCommonScoreForAutoDetect {
		return DetectionIdentity
	}

	// No significant ID overlap, check for fuzzy similarity
	fuzzyScore := calculateFuzzySimilarityScore(source, target)

	// If fuzzy similarity is significant, use Fuzzy detection
	if fuzzyScore >= config.MinElementsDuplicateScoreForAutoDetect {
		return DetectionFuzzy
	}

	// No significant overlap or similarity
	return DetectionNone
}

// isEmpty returns true if the feed has no meaningful entities
func isEmpty(feed *gtfs.Feed) bool {
	return len(feed.Agencies) == 0 &&
		len(feed.Stops) == 0 &&
		len(feed.Routes) == 0 &&
		len(feed.Trips) == 0 &&
		len(feed.Calendars) == 0
}

// calculateIDOverlapScore calculates the overall ID overlap score across all entity types.
// Returns the maximum overlap score found across any entity type.
func calculateIDOverlapScore(source, target *gtfs.Feed) float64 {
	var maxScore float64

	// Check agency IDs
	if score := elementOverlapScore(getAgencyIDs(source), getAgencyIDs(target)); score > maxScore {
		maxScore = score
	}

	// Check stop IDs
	if score := elementOverlapScore(getStopIDs(source), getStopIDs(target)); score > maxScore {
		maxScore = score
	}

	// Check route IDs
	if score := elementOverlapScore(getRouteIDs(source), getRouteIDs(target)); score > maxScore {
		maxScore = score
	}

	// Check trip IDs
	if score := elementOverlapScore(getTripIDs(source), getTripIDs(target)); score > maxScore {
		maxScore = score
	}

	// Check service/calendar IDs
	if score := elementOverlapScore(getServiceIDs(source), getServiceIDs(target)); score > maxScore {
		maxScore = score
	}

	return maxScore
}

// calculateFuzzySimilarityScore calculates how similar entities are based on properties.
// Returns a score between 0.0 (completely different) and 1.0 (very similar).
func calculateFuzzySimilarityScore(source, target *gtfs.Feed) float64 {
	var totalScore float64
	var entityTypesChecked int

	// Check agency similarity (by name and URL)
	if len(source.Agencies) > 0 && len(target.Agencies) > 0 {
		score := agencyFuzzySimilarity(source, target)
		if score > 0 {
			totalScore += score
			entityTypesChecked++
		}
	}

	// Check stop similarity (by name and location)
	if len(source.Stops) > 0 && len(target.Stops) > 0 {
		score := stopFuzzySimilarity(source, target)
		if score > 0 {
			totalScore += score
			entityTypesChecked++
		}
	}

	// Check route similarity (by names)
	if len(source.Routes) > 0 && len(target.Routes) > 0 {
		score := routeFuzzySimilarity(source, target)
		if score > 0 {
			totalScore += score
			entityTypesChecked++
		}
	}

	if entityTypesChecked == 0 {
		return 0
	}

	return totalScore / float64(entityTypesChecked)
}

// agencyFuzzySimilarity calculates how similar agencies are between feeds
func agencyFuzzySimilarity(source, target *gtfs.Feed) float64 {
	var matchCount int

	for _, srcAgency := range source.Agencies {
		for _, tgtAgency := range target.Agencies {
			// Check name and URL match (case-insensitive)
			if strings.EqualFold(srcAgency.Name, tgtAgency.Name) ||
				(srcAgency.URL != "" && strings.EqualFold(srcAgency.URL, tgtAgency.URL)) {
				matchCount++
				break // Only count each source agency once
			}
		}
	}

	if len(source.Agencies) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(source.Agencies))
}

// stopFuzzySimilarity calculates how similar stops are between feeds
// based on name matching and geographic proximity
func stopFuzzySimilarity(source, target *gtfs.Feed) float64 {
	var matchCount int

	for _, srcStop := range source.Stops {
		for _, tgtStop := range target.Stops {
			// Check name match (case-insensitive)
			nameMatch := strings.EqualFold(srcStop.Name, tgtStop.Name)

			// Check proximity (within 500m)
			distance := haversineDistance(srcStop.Lat, srcStop.Lon, tgtStop.Lat, tgtStop.Lon)
			proximityMatch := distance < 500

			// Consider a fuzzy match if names match AND locations are close
			if nameMatch && proximityMatch {
				matchCount++
				break // Only count each source stop once
			}
		}
	}

	if len(source.Stops) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(source.Stops))
}

// routeFuzzySimilarity calculates how similar routes are between feeds
// based on short_name and long_name matching
func routeFuzzySimilarity(source, target *gtfs.Feed) float64 {
	var matchCount int

	for _, srcRoute := range source.Routes {
		for _, tgtRoute := range target.Routes {
			// Check short_name or long_name match (case-insensitive)
			shortNameMatch := srcRoute.ShortName != "" && strings.EqualFold(srcRoute.ShortName, tgtRoute.ShortName)
			longNameMatch := srcRoute.LongName != "" && strings.EqualFold(srcRoute.LongName, tgtRoute.LongName)

			if shortNameMatch || longNameMatch {
				matchCount++
				break // Only count each source route once
			}
		}
	}

	if len(source.Routes) == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(source.Routes))
}

// Helper functions for getting ID slices

func getAgencyIDs(feed *gtfs.Feed) []string {
	ids := make([]string, 0, len(feed.Agencies))
	for id := range feed.Agencies {
		ids = append(ids, string(id))
	}
	return ids
}

func getStopIDs(feed *gtfs.Feed) []string {
	ids := make([]string, 0, len(feed.Stops))
	for id := range feed.Stops {
		ids = append(ids, string(id))
	}
	return ids
}

func getRouteIDs(feed *gtfs.Feed) []string {
	ids := make([]string, 0, len(feed.Routes))
	for id := range feed.Routes {
		ids = append(ids, string(id))
	}
	return ids
}

func getTripIDs(feed *gtfs.Feed) []string {
	ids := make([]string, 0, len(feed.Trips))
	for id := range feed.Trips {
		ids = append(ids, string(id))
	}
	return ids
}

func getServiceIDs(feed *gtfs.Feed) []string {
	ids := make([]string, 0, len(feed.Calendars))
	for id := range feed.Calendars {
		ids = append(ids, string(id))
	}
	return ids
}

// Note: elementOverlapScore[T comparable] is defined in route.go
// Note: haversineDistance is defined in stop.go
