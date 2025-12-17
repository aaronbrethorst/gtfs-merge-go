package gtfs

import (
	"io"
	"strings"
	"testing"
)

// Helper function to parse CSV and return a slice of CSVRows
func parseCSVRows(t *testing.T, content string) ([]string, []*CSVRow) {
	t.Helper()
	reader := NewCSVReader(strings.NewReader(content))
	header, err := reader.ReadHeader()
	if err != nil {
		t.Fatalf("failed to read header: %v", err)
	}

	var rows []*CSVRow
	for {
		record, err := reader.ReadRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read record: %v", err)
		}
		rows = append(rows, NewCSVRow(header, record))
	}
	return header, rows
}

// ==================== Agency Tests ====================

func TestParseAgency(t *testing.T) {
	content := `agency_id,agency_name,agency_url,agency_timezone,agency_lang,agency_phone,agency_fare_url,agency_email
agency1,Metro Transit,http://metro.example.com,America/New_York,en,555-1234,http://metro.example.com/fares,info@metro.example.com
agency2,City Bus,http://citybus.example.com,America/Chicago,es,555-5678,http://citybus.example.com/fares,info@citybus.example.com`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	agency := ParseAgency(rows[0])

	if agency.ID != "agency1" {
		t.Errorf("expected ID 'agency1', got '%s'", agency.ID)
	}
	if agency.Name != "Metro Transit" {
		t.Errorf("expected Name 'Metro Transit', got '%s'", agency.Name)
	}
	if agency.URL != "http://metro.example.com" {
		t.Errorf("expected URL 'http://metro.example.com', got '%s'", agency.URL)
	}
	if agency.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got '%s'", agency.Timezone)
	}
	if agency.Lang != "en" {
		t.Errorf("expected Lang 'en', got '%s'", agency.Lang)
	}
	if agency.Phone != "555-1234" {
		t.Errorf("expected Phone '555-1234', got '%s'", agency.Phone)
	}
	if agency.FareURL != "http://metro.example.com/fares" {
		t.Errorf("expected FareURL 'http://metro.example.com/fares', got '%s'", agency.FareURL)
	}
	if agency.Email != "info@metro.example.com" {
		t.Errorf("expected Email 'info@metro.example.com', got '%s'", agency.Email)
	}
}

func TestParseAgencyMinimalFields(t *testing.T) {
	// Only required fields: agency_name, agency_url, agency_timezone
	// agency_id is conditionally required (required if multiple agencies)
	content := `agency_name,agency_url,agency_timezone
Metro Transit,http://metro.example.com,America/New_York`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	agency := ParseAgency(rows[0])

	if agency.ID != "" {
		t.Errorf("expected empty ID, got '%s'", agency.ID)
	}
	if agency.Name != "Metro Transit" {
		t.Errorf("expected Name 'Metro Transit', got '%s'", agency.Name)
	}
	if agency.URL != "http://metro.example.com" {
		t.Errorf("expected URL 'http://metro.example.com', got '%s'", agency.URL)
	}
	if agency.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got '%s'", agency.Timezone)
	}
	// Optional fields should be empty
	if agency.Lang != "" {
		t.Errorf("expected empty Lang, got '%s'", agency.Lang)
	}
	if agency.Phone != "" {
		t.Errorf("expected empty Phone, got '%s'", agency.Phone)
	}
}

// ==================== Stop Tests ====================

func TestParseStops(t *testing.T) {
	content := `stop_id,stop_code,stop_name,stop_desc,stop_lat,stop_lon,zone_id,stop_url,location_type,parent_station,stop_timezone,wheelchair_boarding,level_id,platform_code
stop1,S001,Main Street Station,Main downtown hub,40.7128,-74.0060,zone1,http://stops.example.com/stop1,1,,America/New_York,1,level1,A
stop2,S002,Park Avenue,Near the park,40.7589,-73.9851,zone2,http://stops.example.com/stop2,0,stop1,America/New_York,2,level2,B`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	stop := ParseStop(rows[0])

	if stop.ID != "stop1" {
		t.Errorf("expected ID 'stop1', got '%s'", stop.ID)
	}
	if stop.Code != "S001" {
		t.Errorf("expected Code 'S001', got '%s'", stop.Code)
	}
	if stop.Name != "Main Street Station" {
		t.Errorf("expected Name 'Main Street Station', got '%s'", stop.Name)
	}
	if stop.Desc != "Main downtown hub" {
		t.Errorf("expected Desc 'Main downtown hub', got '%s'", stop.Desc)
	}
	if stop.Lat != 40.7128 {
		t.Errorf("expected Lat 40.7128, got %f", stop.Lat)
	}
	if stop.Lon != -74.0060 {
		t.Errorf("expected Lon -74.0060, got %f", stop.Lon)
	}
	if stop.ZoneID != "zone1" {
		t.Errorf("expected ZoneID 'zone1', got '%s'", stop.ZoneID)
	}
	if stop.URL != "http://stops.example.com/stop1" {
		t.Errorf("expected URL 'http://stops.example.com/stop1', got '%s'", stop.URL)
	}
	if stop.LocationType != 1 {
		t.Errorf("expected LocationType 1, got %d", stop.LocationType)
	}
	if stop.ParentStation != "" {
		t.Errorf("expected empty ParentStation, got '%s'", stop.ParentStation)
	}
	if stop.Timezone != "America/New_York" {
		t.Errorf("expected Timezone 'America/New_York', got '%s'", stop.Timezone)
	}
	if stop.WheelchairBoarding != 1 {
		t.Errorf("expected WheelchairBoarding 1, got %d", stop.WheelchairBoarding)
	}
	if stop.LevelID != "level1" {
		t.Errorf("expected LevelID 'level1', got '%s'", stop.LevelID)
	}
	if stop.PlatformCode != "A" {
		t.Errorf("expected PlatformCode 'A', got '%s'", stop.PlatformCode)
	}
}

func TestParseStopsWithParentStation(t *testing.T) {
	content := `stop_id,stop_name,stop_lat,stop_lon,parent_station
stop1,Main Station,40.7128,-74.0060,
stop2,Platform A,40.7128,-74.0060,stop1`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// First stop has no parent
	stop1 := ParseStop(rows[0])
	if stop1.ParentStation != "" {
		t.Errorf("expected stop1 ParentStation to be empty, got '%s'", stop1.ParentStation)
	}

	// Second stop has parent reference
	stop2 := ParseStop(rows[1])
	if stop2.ParentStation != "stop1" {
		t.Errorf("expected stop2 ParentStation 'stop1', got '%s'", stop2.ParentStation)
	}
}

// ==================== Route Tests ====================

func TestParseRoutes(t *testing.T) {
	content := `route_id,agency_id,route_short_name,route_long_name,route_desc,route_type,route_url,route_color,route_text_color,route_sort_order,continuous_pickup,continuous_drop_off
route1,agency1,R1,Route One,The first route,3,http://routes.example.com/route1,FF0000,FFFFFF,1,0,1`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	route := ParseRoute(rows[0])

	if route.ID != "route1" {
		t.Errorf("expected ID 'route1', got '%s'", route.ID)
	}
	if route.AgencyID != "agency1" {
		t.Errorf("expected AgencyID 'agency1', got '%s'", route.AgencyID)
	}
	if route.ShortName != "R1" {
		t.Errorf("expected ShortName 'R1', got '%s'", route.ShortName)
	}
	if route.LongName != "Route One" {
		t.Errorf("expected LongName 'Route One', got '%s'", route.LongName)
	}
	if route.Desc != "The first route" {
		t.Errorf("expected Desc 'The first route', got '%s'", route.Desc)
	}
	if route.Type != 3 {
		t.Errorf("expected Type 3, got %d", route.Type)
	}
	if route.URL != "http://routes.example.com/route1" {
		t.Errorf("expected URL 'http://routes.example.com/route1', got '%s'", route.URL)
	}
	if route.Color != "FF0000" {
		t.Errorf("expected Color 'FF0000', got '%s'", route.Color)
	}
	if route.TextColor != "FFFFFF" {
		t.Errorf("expected TextColor 'FFFFFF', got '%s'", route.TextColor)
	}
	if route.SortOrder != 1 {
		t.Errorf("expected SortOrder 1, got %d", route.SortOrder)
	}
	if route.ContinuousPickup != 0 {
		t.Errorf("expected ContinuousPickup 0, got %d", route.ContinuousPickup)
	}
	if route.ContinuousDropOff != 1 {
		t.Errorf("expected ContinuousDropOff 1, got %d", route.ContinuousDropOff)
	}
}

// ==================== Trip Tests ====================

func TestParseTrips(t *testing.T) {
	content := `route_id,service_id,trip_id,trip_headsign,trip_short_name,direction_id,block_id,shape_id,wheelchair_accessible,bikes_allowed
route1,service1,trip1,Downtown,Express,0,block1,shape1,1,2`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	trip := ParseTrip(rows[0])

	if trip.ID != "trip1" {
		t.Errorf("expected ID 'trip1', got '%s'", trip.ID)
	}
	if trip.RouteID != "route1" {
		t.Errorf("expected RouteID 'route1', got '%s'", trip.RouteID)
	}
	if trip.ServiceID != "service1" {
		t.Errorf("expected ServiceID 'service1', got '%s'", trip.ServiceID)
	}
	if trip.Headsign != "Downtown" {
		t.Errorf("expected Headsign 'Downtown', got '%s'", trip.Headsign)
	}
	if trip.ShortName != "Express" {
		t.Errorf("expected ShortName 'Express', got '%s'", trip.ShortName)
	}
	if trip.DirectionID == nil || *trip.DirectionID != 0 {
		t.Errorf("expected DirectionID pointer to 0, got %v", trip.DirectionID)
	}
	if trip.BlockID != "block1" {
		t.Errorf("expected BlockID 'block1', got '%s'", trip.BlockID)
	}
	if trip.ShapeID != "shape1" {
		t.Errorf("expected ShapeID 'shape1', got '%s'", trip.ShapeID)
	}
	if trip.WheelchairAccessible != 1 {
		t.Errorf("expected WheelchairAccessible 1, got %d", trip.WheelchairAccessible)
	}
	if trip.BikesAllowed != 2 {
		t.Errorf("expected BikesAllowed 2, got %d", trip.BikesAllowed)
	}
}

// ==================== StopTime Tests ====================

func TestParseStopTimes(t *testing.T) {
	content := `trip_id,arrival_time,departure_time,stop_id,stop_sequence,stop_headsign,pickup_type,drop_off_type,continuous_pickup,continuous_drop_off,shape_dist_traveled,timepoint
trip1,08:00:00,08:02:00,stop1,1,Via Main St,0,0,1,1,0.0,1
trip1,08:15:00,08:16:00,stop2,2,,0,0,0,0,1250.5,0`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	st := ParseStopTime(rows[0])

	if st.TripID != "trip1" {
		t.Errorf("expected TripID 'trip1', got '%s'", st.TripID)
	}
	if st.ArrivalTime != "08:00:00" {
		t.Errorf("expected ArrivalTime '08:00:00', got '%s'", st.ArrivalTime)
	}
	if st.DepartureTime != "08:02:00" {
		t.Errorf("expected DepartureTime '08:02:00', got '%s'", st.DepartureTime)
	}
	if st.StopID != "stop1" {
		t.Errorf("expected StopID 'stop1', got '%s'", st.StopID)
	}
	if st.StopSequence != 1 {
		t.Errorf("expected StopSequence 1, got %d", st.StopSequence)
	}
	if st.StopHeadsign != "Via Main St" {
		t.Errorf("expected StopHeadsign 'Via Main St', got '%s'", st.StopHeadsign)
	}
	if st.PickupType != 0 {
		t.Errorf("expected PickupType 0, got %d", st.PickupType)
	}
	if st.DropOffType != 0 {
		t.Errorf("expected DropOffType 0, got %d", st.DropOffType)
	}
	if st.ContinuousPickup != 1 {
		t.Errorf("expected ContinuousPickup 1, got %d", st.ContinuousPickup)
	}
	if st.ContinuousDropOff != 1 {
		t.Errorf("expected ContinuousDropOff 1, got %d", st.ContinuousDropOff)
	}
	if st.ShapeDistTraveled != 0.0 {
		t.Errorf("expected ShapeDistTraveled 0.0, got %f", st.ShapeDistTraveled)
	}
	if st.Timepoint != 1 {
		t.Errorf("expected Timepoint 1, got %d", st.Timepoint)
	}
}

func TestParseStopTimesTimeFormat(t *testing.T) {
	// GTFS allows times > 24:00:00 for trips extending past midnight
	content := `trip_id,arrival_time,departure_time,stop_id,stop_sequence
trip1,25:30:00,25:35:00,stop1,1`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	st := ParseStopTime(rows[0])

	// Times should be preserved as strings, even when > 24:00:00
	if st.ArrivalTime != "25:30:00" {
		t.Errorf("expected ArrivalTime '25:30:00', got '%s'", st.ArrivalTime)
	}
	if st.DepartureTime != "25:35:00" {
		t.Errorf("expected DepartureTime '25:35:00', got '%s'", st.DepartureTime)
	}
}

// ==================== Calendar Tests ====================

func TestParseCalendar(t *testing.T) {
	content := `service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date
service1,1,1,1,1,1,0,0,20240101,20241231`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	cal := ParseCalendar(rows[0])

	if cal.ServiceID != "service1" {
		t.Errorf("expected ServiceID 'service1', got '%s'", cal.ServiceID)
	}
	if !cal.Monday {
		t.Error("expected Monday to be true")
	}
	if !cal.Tuesday {
		t.Error("expected Tuesday to be true")
	}
	if !cal.Wednesday {
		t.Error("expected Wednesday to be true")
	}
	if !cal.Thursday {
		t.Error("expected Thursday to be true")
	}
	if !cal.Friday {
		t.Error("expected Friday to be true")
	}
	if cal.Saturday {
		t.Error("expected Saturday to be false")
	}
	if cal.Sunday {
		t.Error("expected Sunday to be false")
	}
	if cal.StartDate != "20240101" {
		t.Errorf("expected StartDate '20240101', got '%s'", cal.StartDate)
	}
	if cal.EndDate != "20241231" {
		t.Errorf("expected EndDate '20241231', got '%s'", cal.EndDate)
	}
}

// ==================== CalendarDate Tests ====================

func TestParseCalendarDates(t *testing.T) {
	content := `service_id,date,exception_type
service1,20240704,1
service1,20241225,2`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Test first row (service added)
	cd1 := ParseCalendarDate(rows[0])
	if cd1.ServiceID != "service1" {
		t.Errorf("expected ServiceID 'service1', got '%s'", cd1.ServiceID)
	}
	if cd1.Date != "20240704" {
		t.Errorf("expected Date '20240704', got '%s'", cd1.Date)
	}
	if cd1.ExceptionType != 1 {
		t.Errorf("expected ExceptionType 1 (added), got %d", cd1.ExceptionType)
	}

	// Test second row (service removed)
	cd2 := ParseCalendarDate(rows[1])
	if cd2.ExceptionType != 2 {
		t.Errorf("expected ExceptionType 2 (removed), got %d", cd2.ExceptionType)
	}
}

// ==================== Shape Tests ====================

func TestParseShapes(t *testing.T) {
	content := `shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled
shape1,40.7128,-74.0060,1,0.0
shape1,40.7145,-74.0040,2,250.5
shape1,40.7160,-74.0020,3,500.0`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	sp := ParseShapePoint(rows[0])

	if sp.ShapeID != "shape1" {
		t.Errorf("expected ShapeID 'shape1', got '%s'", sp.ShapeID)
	}
	if sp.Lat != 40.7128 {
		t.Errorf("expected Lat 40.7128, got %f", sp.Lat)
	}
	if sp.Lon != -74.0060 {
		t.Errorf("expected Lon -74.0060, got %f", sp.Lon)
	}
	if sp.Sequence != 1 {
		t.Errorf("expected Sequence 1, got %d", sp.Sequence)
	}
	if sp.DistTraveled != 0.0 {
		t.Errorf("expected DistTraveled 0.0, got %f", sp.DistTraveled)
	}

	// Test second point has distance
	sp2 := ParseShapePoint(rows[1])
	if sp2.DistTraveled != 250.5 {
		t.Errorf("expected DistTraveled 250.5, got %f", sp2.DistTraveled)
	}
}

// ==================== Frequency Tests ====================

func TestParseFrequencies(t *testing.T) {
	content := `trip_id,start_time,end_time,headway_secs,exact_times
trip1,06:00:00,09:00:00,600,0
trip1,09:00:00,16:00:00,900,1`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	freq := ParseFrequency(rows[0])

	if freq.TripID != "trip1" {
		t.Errorf("expected TripID 'trip1', got '%s'", freq.TripID)
	}
	if freq.StartTime != "06:00:00" {
		t.Errorf("expected StartTime '06:00:00', got '%s'", freq.StartTime)
	}
	if freq.EndTime != "09:00:00" {
		t.Errorf("expected EndTime '09:00:00', got '%s'", freq.EndTime)
	}
	if freq.HeadwaySecs != 600 {
		t.Errorf("expected HeadwaySecs 600, got %d", freq.HeadwaySecs)
	}
	if freq.ExactTimes != 0 {
		t.Errorf("expected ExactTimes 0, got %d", freq.ExactTimes)
	}

	// Test second frequency has exact_times = 1
	freq2 := ParseFrequency(rows[1])
	if freq2.ExactTimes != 1 {
		t.Errorf("expected ExactTimes 1, got %d", freq2.ExactTimes)
	}
}

// ==================== Transfer Tests ====================

func TestParseTransfers(t *testing.T) {
	content := `from_stop_id,to_stop_id,transfer_type,min_transfer_time
stop1,stop2,0,180
stop2,stop3,2,`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	tr := ParseTransfer(rows[0])

	if tr.FromStopID != "stop1" {
		t.Errorf("expected FromStopID 'stop1', got '%s'", tr.FromStopID)
	}
	if tr.ToStopID != "stop2" {
		t.Errorf("expected ToStopID 'stop2', got '%s'", tr.ToStopID)
	}
	if tr.TransferType != 0 {
		t.Errorf("expected TransferType 0, got %d", tr.TransferType)
	}
	if tr.MinTransferTime != 180 {
		t.Errorf("expected MinTransferTime 180, got %d", tr.MinTransferTime)
	}

	// Test second transfer with no min_transfer_time
	tr2 := ParseTransfer(rows[1])
	if tr2.MinTransferTime != 0 {
		t.Errorf("expected MinTransferTime 0 (empty), got %d", tr2.MinTransferTime)
	}
}

// ==================== FareAttribute Tests ====================

func TestParseFareAttributes(t *testing.T) {
	content := `fare_id,price,currency_type,payment_method,transfers,agency_id,transfer_duration
fare1,2.50,USD,0,2,agency1,7200`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	fa := ParseFareAttribute(rows[0])

	if fa.FareID != "fare1" {
		t.Errorf("expected FareID 'fare1', got '%s'", fa.FareID)
	}
	if fa.Price != 2.50 {
		t.Errorf("expected Price 2.50, got %f", fa.Price)
	}
	if fa.CurrencyType != "USD" {
		t.Errorf("expected CurrencyType 'USD', got '%s'", fa.CurrencyType)
	}
	if fa.PaymentMethod != 0 {
		t.Errorf("expected PaymentMethod 0, got %d", fa.PaymentMethod)
	}
	if fa.Transfers != 2 {
		t.Errorf("expected Transfers 2, got %d", fa.Transfers)
	}
	if fa.AgencyID != "agency1" {
		t.Errorf("expected AgencyID 'agency1', got '%s'", fa.AgencyID)
	}
	if fa.TransferDuration != 7200 {
		t.Errorf("expected TransferDuration 7200, got %d", fa.TransferDuration)
	}
}

// ==================== FareRule Tests ====================

func TestParseFareRules(t *testing.T) {
	content := `fare_id,route_id,origin_id,destination_id,contains_id
fare1,route1,zone1,zone2,zone3`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	fr := ParseFareRule(rows[0])

	if fr.FareID != "fare1" {
		t.Errorf("expected FareID 'fare1', got '%s'", fr.FareID)
	}
	if fr.RouteID != "route1" {
		t.Errorf("expected RouteID 'route1', got '%s'", fr.RouteID)
	}
	if fr.OriginID != "zone1" {
		t.Errorf("expected OriginID 'zone1', got '%s'", fr.OriginID)
	}
	if fr.DestinationID != "zone2" {
		t.Errorf("expected DestinationID 'zone2', got '%s'", fr.DestinationID)
	}
	if fr.ContainsID != "zone3" {
		t.Errorf("expected ContainsID 'zone3', got '%s'", fr.ContainsID)
	}
}

// ==================== FeedInfo Tests ====================

func TestParseFeedInfo(t *testing.T) {
	content := `feed_publisher_name,feed_publisher_url,feed_lang,default_lang,feed_start_date,feed_end_date,feed_version,feed_contact_email,feed_contact_url
Metro Transit,http://metro.example.com,en,en,20240101,20241231,1.0.0,gtfs@metro.example.com,http://metro.example.com/contact`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	fi := ParseFeedInfo(rows[0])

	if fi.PublisherName != "Metro Transit" {
		t.Errorf("expected PublisherName 'Metro Transit', got '%s'", fi.PublisherName)
	}
	if fi.PublisherURL != "http://metro.example.com" {
		t.Errorf("expected PublisherURL 'http://metro.example.com', got '%s'", fi.PublisherURL)
	}
	if fi.Lang != "en" {
		t.Errorf("expected Lang 'en', got '%s'", fi.Lang)
	}
	if fi.DefaultLang != "en" {
		t.Errorf("expected DefaultLang 'en', got '%s'", fi.DefaultLang)
	}
	if fi.StartDate != "20240101" {
		t.Errorf("expected StartDate '20240101', got '%s'", fi.StartDate)
	}
	if fi.EndDate != "20241231" {
		t.Errorf("expected EndDate '20241231', got '%s'", fi.EndDate)
	}
	if fi.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got '%s'", fi.Version)
	}
	if fi.ContactEmail != "gtfs@metro.example.com" {
		t.Errorf("expected ContactEmail 'gtfs@metro.example.com', got '%s'", fi.ContactEmail)
	}
	if fi.ContactURL != "http://metro.example.com/contact" {
		t.Errorf("expected ContactURL 'http://metro.example.com/contact', got '%s'", fi.ContactURL)
	}
}

// ==================== Area Tests ====================

func TestParseAreas(t *testing.T) {
	content := `area_id,area_name
area1,Downtown Zone
area2,Suburban Zone`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	area := ParseArea(rows[0])

	if area.ID != "area1" {
		t.Errorf("expected ID 'area1', got '%s'", area.ID)
	}
	if area.Name != "Downtown Zone" {
		t.Errorf("expected Name 'Downtown Zone', got '%s'", area.Name)
	}
}

// ==================== Pathway Tests ====================

func TestParsePathways(t *testing.T) {
	content := `pathway_id,from_stop_id,to_stop_id,pathway_mode,is_bidirectional,length,traversal_time,stair_count,max_slope,min_width,signposted_as,reversed_signposted_as
pathway1,stop1,stop2,2,1,150.5,120,24,0.05,2.0,To Platform A,To Exit`

	_, rows := parseCSVRows(t, content)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	pw := ParsePathway(rows[0])

	if pw.ID != "pathway1" {
		t.Errorf("expected ID 'pathway1', got '%s'", pw.ID)
	}
	if pw.FromStopID != "stop1" {
		t.Errorf("expected FromStopID 'stop1', got '%s'", pw.FromStopID)
	}
	if pw.ToStopID != "stop2" {
		t.Errorf("expected ToStopID 'stop2', got '%s'", pw.ToStopID)
	}
	if pw.PathwayMode != 2 {
		t.Errorf("expected PathwayMode 2, got %d", pw.PathwayMode)
	}
	if pw.IsBidirectional != 1 {
		t.Errorf("expected IsBidirectional 1, got %d", pw.IsBidirectional)
	}
	if pw.Length != 150.5 {
		t.Errorf("expected Length 150.5, got %f", pw.Length)
	}
	if pw.TraversalTime != 120 {
		t.Errorf("expected TraversalTime 120, got %d", pw.TraversalTime)
	}
	if pw.StairCount != 24 {
		t.Errorf("expected StairCount 24, got %d", pw.StairCount)
	}
	if pw.MaxSlope != 0.05 {
		t.Errorf("expected MaxSlope 0.05, got %f", pw.MaxSlope)
	}
	if pw.MinWidth != 2.0 {
		t.Errorf("expected MinWidth 2.0, got %f", pw.MinWidth)
	}
	if pw.SignpostedAs != "To Platform A" {
		t.Errorf("expected SignpostedAs 'To Platform A', got '%s'", pw.SignpostedAs)
	}
	if pw.ReversedSignpostedAs != "To Exit" {
		t.Errorf("expected ReversedSignpostedAs 'To Exit', got '%s'", pw.ReversedSignpostedAs)
	}
}
