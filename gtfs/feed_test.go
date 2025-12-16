package gtfs

import (
	"testing"
)

func TestNewFeed(t *testing.T) {
	feed := NewFeed()

	if feed == nil {
		t.Fatal("NewFeed() returned nil")
	}

	// Verify all maps are initialized (not nil)
	if feed.Agencies == nil {
		t.Error("Agencies map is nil")
	}
	if feed.Stops == nil {
		t.Error("Stops map is nil")
	}
	if feed.Routes == nil {
		t.Error("Routes map is nil")
	}
	if feed.Trips == nil {
		t.Error("Trips map is nil")
	}
	if feed.Calendars == nil {
		t.Error("Calendars map is nil")
	}
	if feed.CalendarDates == nil {
		t.Error("CalendarDates map is nil")
	}
	if feed.Shapes == nil {
		t.Error("Shapes map is nil")
	}
	if feed.FareAttributes == nil {
		t.Error("FareAttributes map is nil")
	}
	if feed.Areas == nil {
		t.Error("Areas map is nil")
	}

	// Verify slices are initialized (not nil, though empty is fine)
	if feed.StopTimes == nil {
		t.Error("StopTimes slice is nil")
	}
	if feed.Frequencies == nil {
		t.Error("Frequencies slice is nil")
	}
	if feed.Transfers == nil {
		t.Error("Transfers slice is nil")
	}
	if feed.FareRules == nil {
		t.Error("FareRules slice is nil")
	}
	if feed.Pathways == nil {
		t.Error("Pathways slice is nil")
	}

	// FeedInfo can be nil initially
}

func TestFeedAddAgency(t *testing.T) {
	feed := NewFeed()

	agency := &Agency{
		ID:       "agency1",
		Name:     "Test Agency",
		URL:      "https://example.com",
		Timezone: "America/Los_Angeles",
	}

	feed.Agencies[agency.ID] = agency

	// Verify agency was added
	if len(feed.Agencies) != 1 {
		t.Errorf("expected 1 agency, got %d", len(feed.Agencies))
	}

	retrieved := feed.Agencies["agency1"]
	if retrieved == nil {
		t.Fatal("agency not found")
	}
	if retrieved.Name != "Test Agency" {
		t.Errorf("expected name 'Test Agency', got '%s'", retrieved.Name)
	}
}

func TestFeedAddStop(t *testing.T) {
	feed := NewFeed()

	stop := &Stop{
		ID:   "stop1",
		Name: "Main St Station",
		Lat:  47.6062,
		Lon:  -122.3321,
	}

	feed.Stops[stop.ID] = stop

	// Verify stop was added
	if len(feed.Stops) != 1 {
		t.Errorf("expected 1 stop, got %d", len(feed.Stops))
	}

	retrieved := feed.Stops["stop1"]
	if retrieved == nil {
		t.Fatal("stop not found")
	}
	if retrieved.Name != "Main St Station" {
		t.Errorf("expected name 'Main St Station', got '%s'", retrieved.Name)
	}
}

func TestFeedAddRoute(t *testing.T) {
	feed := NewFeed()

	route := &Route{
		ID:        "route1",
		AgencyID:  "agency1",
		ShortName: "1",
		LongName:  "First Avenue Line",
		Type:      3, // Bus
	}

	feed.Routes[route.ID] = route

	// Verify route was added
	if len(feed.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(feed.Routes))
	}

	retrieved := feed.Routes["route1"]
	if retrieved == nil {
		t.Fatal("route not found")
	}
	if retrieved.ShortName != "1" {
		t.Errorf("expected short name '1', got '%s'", retrieved.ShortName)
	}
}

func TestFeedAddTrip(t *testing.T) {
	feed := NewFeed()

	trip := &Trip{
		ID:        "trip1",
		RouteID:   "route1",
		ServiceID: "weekday",
		Headsign:  "Downtown",
	}

	feed.Trips[trip.ID] = trip

	// Verify trip was added
	if len(feed.Trips) != 1 {
		t.Errorf("expected 1 trip, got %d", len(feed.Trips))
	}

	retrieved := feed.Trips["trip1"]
	if retrieved == nil {
		t.Fatal("trip not found")
	}
	if retrieved.Headsign != "Downtown" {
		t.Errorf("expected headsign 'Downtown', got '%s'", retrieved.Headsign)
	}
}

func TestFeedAddStopTime(t *testing.T) {
	feed := NewFeed()

	stopTime := &StopTime{
		TripID:        "trip1",
		StopID:        "stop1",
		StopSequence:  1,
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:01:00",
	}

	feed.StopTimes = append(feed.StopTimes, stopTime)

	// Verify stop time was added
	if len(feed.StopTimes) != 1 {
		t.Errorf("expected 1 stop time, got %d", len(feed.StopTimes))
	}

	retrieved := feed.StopTimes[0]
	if retrieved.ArrivalTime != "08:00:00" {
		t.Errorf("expected arrival '08:00:00', got '%s'", retrieved.ArrivalTime)
	}
}

func TestFeedAddCalendar(t *testing.T) {
	feed := NewFeed()

	calendar := &Calendar{
		ServiceID: "weekday",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    false,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	feed.Calendars[calendar.ServiceID] = calendar

	// Verify calendar was added
	if len(feed.Calendars) != 1 {
		t.Errorf("expected 1 calendar, got %d", len(feed.Calendars))
	}

	retrieved := feed.Calendars["weekday"]
	if retrieved == nil {
		t.Fatal("calendar not found")
	}
	if !retrieved.Monday {
		t.Error("expected Monday to be true")
	}
	if retrieved.Saturday {
		t.Error("expected Saturday to be false")
	}
}

func TestFeedAddCalendarDate(t *testing.T) {
	feed := NewFeed()

	calendarDate := &CalendarDate{
		ServiceID:     "weekday",
		Date:          "20240704",
		ExceptionType: 2, // Service removed
	}

	feed.CalendarDates["weekday"] = append(feed.CalendarDates["weekday"], calendarDate)

	// Verify calendar date was added
	dates := feed.CalendarDates["weekday"]
	if len(dates) != 1 {
		t.Errorf("expected 1 calendar date, got %d", len(dates))
	}

	retrieved := dates[0]
	if retrieved.Date != "20240704" {
		t.Errorf("expected date '20240704', got '%s'", retrieved.Date)
	}
	if retrieved.ExceptionType != 2 {
		t.Errorf("expected exception type 2, got %d", retrieved.ExceptionType)
	}
}

func TestFeedAddShapePoint(t *testing.T) {
	feed := NewFeed()

	shapePoint := &ShapePoint{
		ShapeID:      "shape1",
		Lat:          47.6062,
		Lon:          -122.3321,
		Sequence:     1,
		DistTraveled: 0.0,
	}

	feed.Shapes["shape1"] = append(feed.Shapes["shape1"], shapePoint)

	// Verify shape point was added
	points := feed.Shapes["shape1"]
	if len(points) != 1 {
		t.Errorf("expected 1 shape point, got %d", len(points))
	}

	retrieved := points[0]
	if retrieved.Sequence != 1 {
		t.Errorf("expected sequence 1, got %d", retrieved.Sequence)
	}
}

func TestFeedAddFrequency(t *testing.T) {
	feed := NewFeed()

	frequency := &Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "22:00:00",
		HeadwaySecs: 600, // 10 minutes
	}

	feed.Frequencies = append(feed.Frequencies, frequency)

	// Verify frequency was added
	if len(feed.Frequencies) != 1 {
		t.Errorf("expected 1 frequency, got %d", len(feed.Frequencies))
	}

	retrieved := feed.Frequencies[0]
	if retrieved.HeadwaySecs != 600 {
		t.Errorf("expected headway 600, got %d", retrieved.HeadwaySecs)
	}
}

func TestFeedAddTransfer(t *testing.T) {
	feed := NewFeed()

	transfer := &Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		TransferType:    0, // Recommended transfer
		MinTransferTime: 120,
	}

	feed.Transfers = append(feed.Transfers, transfer)

	// Verify transfer was added
	if len(feed.Transfers) != 1 {
		t.Errorf("expected 1 transfer, got %d", len(feed.Transfers))
	}

	retrieved := feed.Transfers[0]
	if retrieved.FromStopID != "stop1" {
		t.Errorf("expected from stop 'stop1', got '%s'", retrieved.FromStopID)
	}
}

func TestFeedAddFareAttribute(t *testing.T) {
	feed := NewFeed()

	fareAttr := &FareAttribute{
		FareID:        "fare1",
		Price:         2.50,
		CurrencyType:  "USD",
		PaymentMethod: 0,
		Transfers:     0,
	}

	feed.FareAttributes[fareAttr.FareID] = fareAttr

	// Verify fare attribute was added
	if len(feed.FareAttributes) != 1 {
		t.Errorf("expected 1 fare attribute, got %d", len(feed.FareAttributes))
	}

	retrieved := feed.FareAttributes["fare1"]
	if retrieved == nil {
		t.Fatal("fare attribute not found")
	}
	if retrieved.Price != 2.50 {
		t.Errorf("expected price 2.50, got %f", retrieved.Price)
	}
}

func TestFeedAddFareRule(t *testing.T) {
	feed := NewFeed()

	fareRule := &FareRule{
		FareID:  "fare1",
		RouteID: "route1",
	}

	feed.FareRules = append(feed.FareRules, fareRule)

	// Verify fare rule was added
	if len(feed.FareRules) != 1 {
		t.Errorf("expected 1 fare rule, got %d", len(feed.FareRules))
	}

	retrieved := feed.FareRules[0]
	if retrieved.FareID != "fare1" {
		t.Errorf("expected fare ID 'fare1', got '%s'", retrieved.FareID)
	}
}

func TestFeedAddArea(t *testing.T) {
	feed := NewFeed()

	area := &Area{
		ID:   "area1",
		Name: "Downtown Zone",
	}

	feed.Areas[area.ID] = area

	// Verify area was added
	if len(feed.Areas) != 1 {
		t.Errorf("expected 1 area, got %d", len(feed.Areas))
	}

	retrieved := feed.Areas["area1"]
	if retrieved == nil {
		t.Fatal("area not found")
	}
	if retrieved.Name != "Downtown Zone" {
		t.Errorf("expected name 'Downtown Zone', got '%s'", retrieved.Name)
	}
}

func TestFeedAddPathway(t *testing.T) {
	feed := NewFeed()

	pathway := &Pathway{
		ID:              "pathway1",
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		PathwayMode:     1, // Walkway
		IsBidirectional: 1,
	}

	feed.Pathways = append(feed.Pathways, pathway)

	// Verify pathway was added
	if len(feed.Pathways) != 1 {
		t.Errorf("expected 1 pathway, got %d", len(feed.Pathways))
	}

	retrieved := feed.Pathways[0]
	if retrieved.ID != "pathway1" {
		t.Errorf("expected ID 'pathway1', got '%s'", retrieved.ID)
	}
}

func TestFeedSetFeedInfo(t *testing.T) {
	feed := NewFeed()

	feedInfo := &FeedInfo{
		PublisherName: "Transit Authority",
		PublisherURL:  "https://transit.example.com",
		Lang:          "en",
		Version:       "1.0",
	}

	feed.FeedInfo = feedInfo

	// Verify feed info was set
	if feed.FeedInfo == nil {
		t.Fatal("FeedInfo is nil")
	}
	if feed.FeedInfo.PublisherName != "Transit Authority" {
		t.Errorf("expected publisher name 'Transit Authority', got '%s'", feed.FeedInfo.PublisherName)
	}
}
