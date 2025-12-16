package strategy

import (
	"fmt"
	"log"
	"time"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// CalendarMergeStrategy handles merging of calendars between feeds
type CalendarMergeStrategy struct {
	BaseStrategy
	// FuzzyThreshold is the minimum score for a fuzzy match (default 0.5)
	FuzzyThreshold float64
}

// NewCalendarMergeStrategy creates a new CalendarMergeStrategy
func NewCalendarMergeStrategy() *CalendarMergeStrategy {
	return &CalendarMergeStrategy{
		BaseStrategy:   NewBaseStrategy("calendar"),
		FuzzyThreshold: 0.5,
	}
}

// Merge performs the merge operation for calendars
func (s *CalendarMergeStrategy) Merge(ctx *MergeContext) error {
	// First, merge calendar.txt entries
	for _, cal := range ctx.Source.Calendars {
		// Check for duplicates based on detection mode
		if s.DuplicateDetection == DetectionIdentity {
			if existing, found := ctx.Target.Calendars[cal.ServiceID]; found {
				// Duplicate detected - map source ID to existing target ID
				ctx.ServiceIDMapping[cal.ServiceID] = existing.ServiceID

				// Handle logging based on configuration
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate calendar detected with service_id %q (keeping existing)", cal.ServiceID)
				case LogError:
					return fmt.Errorf("duplicate calendar detected with service_id %q", cal.ServiceID)
				}

				// Skip adding this calendar - use the existing one
				continue
			}
		}

		// Check for fuzzy duplicates
		if s.DuplicateDetection == DetectionFuzzy {
			if matchID := s.findFuzzyMatch(ctx, cal); matchID != "" {
				// Fuzzy duplicate detected - map source ID to existing target ID
				ctx.ServiceIDMapping[cal.ServiceID] = matchID

				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Fuzzy duplicate calendar detected: %q matches %q (keeping existing)", cal.ServiceID, matchID)
				case LogError:
					return fmt.Errorf("fuzzy duplicate calendar detected: %q matches %q", cal.ServiceID, matchID)
				}

				// Skip adding this calendar - use the existing one
				continue
			}
		}

		// Determine new ID - only apply prefix if there's a collision
		newID := cal.ServiceID
		if _, exists := ctx.Target.Calendars[cal.ServiceID]; exists {
			// Collision detected - apply prefix
			newID = gtfs.ServiceID(ctx.Prefix + string(cal.ServiceID))
		}
		ctx.ServiceIDMapping[cal.ServiceID] = newID

		newCal := &gtfs.Calendar{
			ServiceID: newID,
			Monday:    cal.Monday,
			Tuesday:   cal.Tuesday,
			Wednesday: cal.Wednesday,
			Thursday:  cal.Thursday,
			Friday:    cal.Friday,
			Saturday:  cal.Saturday,
			Sunday:    cal.Sunday,
			StartDate: cal.StartDate,
			EndDate:   cal.EndDate,
		}
		ctx.Target.Calendars[newID] = newCal
	}

	return nil
}

// findFuzzyMatch searches for a fuzzy duplicate in the target calendars.
// Returns the ID of the matching calendar if found, or empty string if no match.
// Uses date overlap scoring.
func (s *CalendarMergeStrategy) findFuzzyMatch(ctx *MergeContext, source *gtfs.Calendar) gtfs.ServiceID {
	var bestMatch gtfs.ServiceID
	var bestScore float64

	for _, target := range ctx.Target.Calendars {
		score := calendarDateOverlapScore(source, target)

		if score >= s.FuzzyThreshold && score > bestScore {
			bestScore = score
			bestMatch = target.ServiceID
		}
	}

	return bestMatch
}

// calendarDateOverlapScore returns the date overlap score between two calendars.
func calendarDateOverlapScore(source, target *gtfs.Calendar) float64 {
	sourceStart := parseGTFSDate(source.StartDate)
	sourceEnd := parseGTFSDate(source.EndDate)
	targetStart := parseGTFSDate(target.StartDate)
	targetEnd := parseGTFSDate(target.EndDate)

	if sourceStart == 0 || sourceEnd == 0 || targetStart == 0 || targetEnd == 0 {
		return 0.0
	}

	// Adjust end dates to be inclusive (add one day worth of seconds)
	// This makes "20240101" to "20240131" represent the full month including Jan 31
	sourceEnd += 86400 // 24 * 60 * 60
	targetEnd += 86400

	return calendarIntervalOverlapScore(
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

// calendarIntervalOverlapScore calculates the overlap score between two intervals.
// Formula: (overlap / interval_a_length + overlap / interval_b_length) / 2
// Returns 0.0 if either interval has zero length or there's no overlap.
func calendarIntervalOverlapScore(start1, end1, start2, end2 float64) float64 {
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

// CalendarDateMergeStrategy handles merging of calendar dates between feeds
type CalendarDateMergeStrategy struct {
	BaseStrategy
}

// NewCalendarDateMergeStrategy creates a new CalendarDateMergeStrategy
func NewCalendarDateMergeStrategy() *CalendarDateMergeStrategy {
	return &CalendarDateMergeStrategy{
		BaseStrategy: NewBaseStrategy("calendar_dates"),
	}
}

// Merge performs the merge operation for calendar dates
func (s *CalendarDateMergeStrategy) Merge(ctx *MergeContext) error {
	for serviceID, dates := range ctx.Source.CalendarDates {
		newServiceID := ctx.ServiceIDMapping[serviceID]
		if newServiceID == "" {
			// Service may only be defined in calendar_dates, not calendar
			// Only apply prefix if there's a collision
			newServiceID = serviceID
			if _, exists := ctx.Target.CalendarDates[serviceID]; exists {
				newServiceID = gtfs.ServiceID(ctx.Prefix + string(serviceID))
			}
			ctx.ServiceIDMapping[serviceID] = newServiceID
		}

		for _, date := range dates {
			// Check for exact duplicate calendar date (same service_id, date, exception_type)
			isDuplicate := false
			if s.DuplicateDetection == DetectionIdentity {
				for _, existingDate := range ctx.Target.CalendarDates[newServiceID] {
					if existingDate.Date == date.Date && existingDate.ExceptionType == date.ExceptionType {
						isDuplicate = true
						break
					}
				}
			}

			if isDuplicate {
				switch s.DuplicateLogging {
				case LogWarning:
					log.Printf("WARNING: Duplicate calendar_date detected for service_id %q date %q (keeping existing)", serviceID, date.Date)
				case LogError:
					return fmt.Errorf("duplicate calendar_date detected for service_id %q date %q", serviceID, date.Date)
				}
				continue
			}

			newDate := &gtfs.CalendarDate{
				ServiceID:     newServiceID,
				Date:          date.Date,
				ExceptionType: date.ExceptionType,
			}
			ctx.Target.CalendarDates[newServiceID] = append(ctx.Target.CalendarDates[newServiceID], newDate)
		}
	}

	return nil
}
