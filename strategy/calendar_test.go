package strategy

import (
	"testing"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

func TestCalendarMergeNoDuplicates(t *testing.T) {
	// Given: two feeds with non-overlapping service IDs
	source := gtfs.NewFeed()
	source.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	target := gtfs.NewFeed()
	target.Calendars[gtfs.ServiceID("service2")] = &gtfs.Calendar{
		ServiceID: "service2",
		Saturday:  true,
		Sunday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewCalendarMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)

	// Then: both calendars should be in target
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Calendars) != 2 {
		t.Errorf("Expected 2 calendars, got %d", len(target.Calendars))
	}

	if _, ok := target.Calendars["service1"]; !ok {
		t.Error("Expected service1 to be in target")
	}
	if _, ok := target.Calendars["service2"]; !ok {
		t.Error("Expected service2 to be in target")
	}
}

func TestCalendarMergeIdentityDuplicate(t *testing.T) {
	// Given: both feeds have calendar with service_id "service1"
	source := gtfs.NewFeed()
	source.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		StartDate: "20240601",
		EndDate:   "20241231",
	}

	target := gtfs.NewFeed()
	target.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewCalendarMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged with DetectionIdentity
	err := strategy.Merge(ctx)

	// Then: only one service1 in output (the existing one)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(target.Calendars) != 1 {
		t.Errorf("Expected 1 calendar, got %d", len(target.Calendars))
	}

	cal := target.Calendars["service1"]
	if cal == nil {
		t.Fatal("Expected service1 to be in target")
	}

	// The existing calendar should be kept
	if cal.StartDate != "20240101" {
		t.Errorf("Expected existing calendar to be kept, got start_date %q", cal.StartDate)
	}

	// Check that the ID mapping points to the existing service
	if ctx.ServiceIDMapping["service1"] != "service1" {
		t.Errorf("Expected ServiceIDMapping[service1] = service1, got %q", ctx.ServiceIDMapping["service1"])
	}
}

func TestCalendarMergeUpdatesTripRefs(t *testing.T) {
	// Given: source feed has a calendar
	source := gtfs.NewFeed()
	source.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		StartDate: "20240601",
		EndDate:   "20241231",
	}

	target := gtfs.NewFeed()
	target.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewCalendarMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: calendars are merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: the mapping should point to the existing service
	mappedID := ctx.ServiceIDMapping["service1"]
	if mappedID != "service1" {
		t.Errorf("Expected mapped ID to be service1, got %q", mappedID)
	}
}

func TestCalendarDatesMerged(t *testing.T) {
	// Given: source has calendar dates
	source := gtfs.NewFeed()
	source.CalendarDates[gtfs.ServiceID("service1")] = []*gtfs.CalendarDate{
		{ServiceID: "service1", Date: "20240704", ExceptionType: 1},
	}

	target := gtfs.NewFeed()

	ctx := NewMergeContext(source, target, "")
	ctx.ServiceIDMapping[gtfs.ServiceID("service1")] = gtfs.ServiceID("service1")

	strategy := NewCalendarDateMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: calendar dates should be in target
	dates := target.CalendarDates[gtfs.ServiceID("service1")]
	if len(dates) != 1 {
		t.Errorf("Expected 1 calendar date, got %d", len(dates))
	}

	if dates[0].Date != "20240704" {
		t.Errorf("Expected date 20240704, got %q", dates[0].Date)
	}
}

func TestCalendarDatesWithNewServiceID(t *testing.T) {
	// Given: source has calendar dates for a service not in calendar.txt
	// and target has a collision for that service ID
	source := gtfs.NewFeed()
	source.CalendarDates[gtfs.ServiceID("service_new")] = []*gtfs.CalendarDate{
		{ServiceID: "service_new", Date: "20240704", ExceptionType: 1},
	}

	target := gtfs.NewFeed()
	// Add colliding calendar dates to force prefixing
	target.CalendarDates[gtfs.ServiceID("service_new")] = []*gtfs.CalendarDate{
		{ServiceID: "service_new", Date: "20240101", ExceptionType: 2},
	}

	ctx := NewMergeContext(source, target, "a_")
	// No mapping set for service_new (it's only in calendar_dates)

	strategy := NewCalendarDateMergeStrategy()
	strategy.SetDuplicateDetection(DetectionNone)

	// When: merged with collision (forces prefix)
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: calendar dates should be in target with prefixed service ID
	dates := target.CalendarDates[gtfs.ServiceID("a_service_new")]
	if len(dates) != 1 {
		t.Errorf("Expected 1 calendar date for a_service_new, got %d", len(dates))
	}

	// And the mapping should be updated
	if ctx.ServiceIDMapping["service_new"] != "a_service_new" {
		t.Errorf("Expected ServiceIDMapping[service_new] = a_service_new, got %q", ctx.ServiceIDMapping["service_new"])
	}
}

func TestCalendarMergeErrorOnDuplicate(t *testing.T) {
	// Given: both feeds have calendar with same service_id and error logging enabled
	source := gtfs.NewFeed()
	source.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		StartDate: "20240601",
		EndDate:   "20241231",
	}

	target := gtfs.NewFeed()
	target.Calendars[gtfs.ServiceID("service1")] = &gtfs.Calendar{
		ServiceID: "service1",
		Monday:    true,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	ctx := NewMergeContext(source, target, "")
	strategy := NewCalendarMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)
	strategy.SetDuplicateLogging(LogError)

	// When: merged with LogError
	err := strategy.Merge(ctx)

	// Then: should return an error
	if err == nil {
		t.Fatal("Expected error when duplicate detected with LogError")
	}
}

func TestCalendarDatesDuplicateDetection(t *testing.T) {
	// Given: both feeds have the same calendar date
	source := gtfs.NewFeed()
	source.CalendarDates[gtfs.ServiceID("service1")] = []*gtfs.CalendarDate{
		{ServiceID: "service1", Date: "20240704", ExceptionType: 1},
	}

	target := gtfs.NewFeed()
	target.CalendarDates[gtfs.ServiceID("service1")] = []*gtfs.CalendarDate{
		{ServiceID: "service1", Date: "20240704", ExceptionType: 1},
	}

	ctx := NewMergeContext(source, target, "")
	ctx.ServiceIDMapping[gtfs.ServiceID("service1")] = gtfs.ServiceID("service1")

	strategy := NewCalendarDateMergeStrategy()
	strategy.SetDuplicateDetection(DetectionIdentity)

	// When: merged
	err := strategy.Merge(ctx)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Then: should not have duplicate calendar dates
	dates := target.CalendarDates[gtfs.ServiceID("service1")]
	if len(dates) != 1 {
		t.Errorf("Expected 1 calendar date (duplicate skipped), got %d", len(dates))
	}
}
