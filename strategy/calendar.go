package strategy

import (
	"fmt"
	"log"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// CalendarMergeStrategy handles merging of calendars between feeds
type CalendarMergeStrategy struct {
	BaseStrategy
}

// NewCalendarMergeStrategy creates a new CalendarMergeStrategy
func NewCalendarMergeStrategy() *CalendarMergeStrategy {
	return &CalendarMergeStrategy{
		BaseStrategy: NewBaseStrategy("calendar"),
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
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate calendar detected with service_id %q (keeping existing)", cal.ServiceID)
				} else if s.DuplicateLogging == LogError {
					return fmt.Errorf("duplicate calendar detected with service_id %q", cal.ServiceID)
				}

				// Skip adding this calendar - use the existing one
				continue
			}
		}

		// No duplicate - add with prefix if needed
		newID := gtfs.ServiceID(ctx.Prefix + string(cal.ServiceID))
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
			newServiceID = gtfs.ServiceID(ctx.Prefix + string(serviceID))
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
				if s.DuplicateLogging == LogWarning {
					log.Printf("WARNING: Duplicate calendar_date detected for service_id %q date %q (keeping existing)", serviceID, date.Date)
				} else if s.DuplicateLogging == LogError {
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
