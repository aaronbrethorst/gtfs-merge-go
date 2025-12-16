package scoring

import (
	"time"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// ServiceDateOverlapScorer scores service calendars by date overlap.
// Uses the IntervalOverlapScore formula for calculating overlap.
type ServiceDateOverlapScorer struct{}

// Score returns a similarity score based on date overlap between two services.
func (s *ServiceDateOverlapScorer) Score(ctx *strategy.MergeContext, sourceID, targetID gtfs.ServiceID) float64 {
	sourceCalendar, sourceOk := ctx.Source.Calendars[sourceID]
	targetCalendar, targetOk := ctx.Target.Calendars[targetID]

	if !sourceOk || !targetOk {
		return 0.0
	}

	// Parse dates
	sourceStart := parseGTFSDate(sourceCalendar.StartDate)
	sourceEnd := parseGTFSDate(sourceCalendar.EndDate)
	targetStart := parseGTFSDate(targetCalendar.StartDate)
	targetEnd := parseGTFSDate(targetCalendar.EndDate)

	if sourceStart == 0 || sourceEnd == 0 || targetStart == 0 || targetEnd == 0 {
		return 0.0
	}

	// Adjust end dates to be inclusive (add one day worth of seconds)
	// This makes "20240101" to "20240131" represent the full month including Jan 31
	sourceEnd += 86400 // 24 * 60 * 60
	targetEnd += 86400

	return IntervalOverlapScore(
		float64(sourceStart), float64(sourceEnd),
		float64(targetStart), float64(targetEnd),
	)
}

// parseGTFSDate parses a GTFS date string (YYYYMMDD) to Unix timestamp.
// Returns 0 for invalid or empty dates.
func parseGTFSDate(dateStr string) int64 {
	if dateStr == "" || len(dateStr) != 8 {
		return 0
	}

	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return 0
	}

	return t.Unix()
}
