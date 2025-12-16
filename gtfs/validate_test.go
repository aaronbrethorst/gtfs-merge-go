package gtfs

import (
	"strings"
	"testing"
)

// ============================================================================
// Referential Integrity Validation Tests (6.1)
// ============================================================================

func TestValidateRouteAgencyRef(t *testing.T) {
	// Route references valid agency
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid route-agency reference, got: %v", errs)
	}

	// Route references invalid agency
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "nonexistent", ShortName: "R1", Type: 3}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid agency reference")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "route") && strings.Contains(err.Error(), "agency") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about route referencing invalid agency, got: %v", errs)
	}
}

func TestValidateRouteAgencyRefOptional(t *testing.T) {
	// Route with empty agency_id is valid when there's only one agency
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "", ShortName: "R1", Type: 3}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for route with empty agency_id when single agency, got: %v", errs)
	}

	// Route with empty agency_id is invalid when there are multiple agencies
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency 1", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Agencies["agency2"] = &Agency{ID: "agency2", Name: "Test Agency 2", URL: "http://test2.com", Timezone: "America/Los_Angeles"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "", ShortName: "R1", Type: 3}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for route with empty agency_id when multiple agencies exist")
	}
}

func TestValidateTripRouteRef(t *testing.T) {
	// Trip references valid route
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid trip-route reference, got: %v", errs)
	}

	// Trip references invalid route
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "nonexistent", ServiceID: "service1"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid route reference")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "trip") && strings.Contains(err.Error(), "route") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about trip referencing invalid route, got: %v", errs)
	}
}

func TestValidateTripServiceRef(t *testing.T) {
	// Trip references valid service
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid trip-service reference, got: %v", errs)
	}

	// Trip references invalid service (not in calendar or calendar_dates)
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "nonexistent"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid service reference")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "trip") && strings.Contains(err.Error(), "service") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about trip referencing invalid service, got: %v", errs)
	}

	// Trip references service that exists only in calendar_dates (valid)
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed3.CalendarDates["service2"] = []*CalendarDate{{ServiceID: "service2", Date: "20240101", ExceptionType: 1}}
	feed3.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service2"}

	errs = feed3.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for service in calendar_dates only, got: %v", errs)
	}
}

func TestValidateStopTimeStopRef(t *testing.T) {
	// StopTime references valid stop
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed.StopTimes = append(feed.StopTimes, &StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid stoptime-stop reference, got: %v", errs)
	}

	// StopTime references invalid stop
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed2.StopTimes = append(feed2.StopTimes, &StopTime{TripID: "trip1", StopID: "nonexistent", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid stop reference in stop_time")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "stop_time") && strings.Contains(err.Error(), "stop") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about stop_time referencing invalid stop, got: %v", errs)
	}
}

func TestValidateStopTimeTripRef(t *testing.T) {
	// StopTime references valid trip
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed.StopTimes = append(feed.StopTimes, &StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid stoptime-trip reference, got: %v", errs)
	}

	// StopTime references invalid trip
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed2.StopTimes = append(feed2.StopTimes, &StopTime{TripID: "nonexistent", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid trip reference in stop_time")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "stop_time") && strings.Contains(err.Error(), "trip") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about stop_time referencing invalid trip, got: %v", errs)
	}
}

func TestValidateStopParentRef(t *testing.T) {
	// Stop with valid parent_station reference
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["station1"] = &Stop{ID: "station1", Name: "Main Station", Lat: 40.0, Lon: -74.0, LocationType: 1}
	feed.Stops["platform1"] = &Stop{ID: "platform1", Name: "Platform 1", Lat: 40.0, Lon: -74.0, LocationType: 0, ParentStation: "station1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid parent_station reference, got: %v", errs)
	}

	// Stop with invalid parent_station reference
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["platform1"] = &Stop{ID: "platform1", Name: "Platform 1", Lat: 40.0, Lon: -74.0, LocationType: 0, ParentStation: "nonexistent"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid parent_station reference")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "stop") && strings.Contains(err.Error(), "parent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about stop referencing invalid parent_station, got: %v", errs)
	}

	// Stop with empty parent_station (valid)
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0, ParentStation: ""}

	errs = feed3.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for stop with empty parent_station, got: %v", errs)
	}
}

func TestValidateTransferStopRefs(t *testing.T) {
	// Transfer with valid stop references
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Stops["stop2"] = &Stop{ID: "stop2", Name: "Stop 2", Lat: 40.1, Lon: -74.1}
	feed.Transfers = append(feed.Transfers, &Transfer{FromStopID: "stop1", ToStopID: "stop2", TransferType: 0})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid transfer stop references, got: %v", errs)
	}

	// Transfer with invalid from_stop_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed2.Transfers = append(feed2.Transfers, &Transfer{FromStopID: "nonexistent", ToStopID: "stop1", TransferType: 0})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid from_stop_id in transfer")
	}

	// Transfer with invalid to_stop_id
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed3.Transfers = append(feed3.Transfers, &Transfer{FromStopID: "stop1", ToStopID: "nonexistent", TransferType: 0})

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid to_stop_id in transfer")
	}
}

func TestValidateFareRuleRefs(t *testing.T) {
	// FareRule with valid references
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0}
	feed.FareRules = append(feed.FareRules, &FareRule{FareID: "fare1", RouteID: "route1"})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid fare_rule references, got: %v", errs)
	}

	// FareRule with invalid fare_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0}
	feed2.FareRules = append(feed2.FareRules, &FareRule{FareID: "nonexistent", RouteID: "route1"})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid fare_id in fare_rule")
	}

	// FareRule with invalid route_id (when specified)
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed3.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0}
	feed3.FareRules = append(feed3.FareRules, &FareRule{FareID: "fare1", RouteID: "nonexistent"})

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid route_id in fare_rule")
	}

	// FareRule with empty route_id (valid - applies to all routes)
	feed4 := NewFeed()
	feed4.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed4.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0}
	feed4.FareRules = append(feed4.FareRules, &FareRule{FareID: "fare1", RouteID: ""})

	errs = feed4.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for fare_rule with empty route_id, got: %v", errs)
	}
}

func TestValidateShapeInTrip(t *testing.T) {
	// Trip with valid shape_id reference
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Shapes["shape1"] = []*ShapePoint{{ShapeID: "shape1", Lat: 40.0, Lon: -74.0, Sequence: 1}}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1", ShapeID: "shape1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid trip-shape reference, got: %v", errs)
	}

	// Trip with invalid shape_id reference
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Shapes["shape1"] = []*ShapePoint{{ShapeID: "shape1", Lat: 40.0, Lon: -74.0, Sequence: 1}}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1", ShapeID: "nonexistent"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid shape_id in trip")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "trip") && strings.Contains(err.Error(), "shape") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about trip referencing invalid shape, got: %v", errs)
	}

	// Trip with empty shape_id (valid)
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed3.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed3.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1", ShapeID: ""}

	errs = feed3.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for trip with empty shape_id, got: %v", errs)
	}
}

func TestValidateFrequencyTripRef(t *testing.T) {
	// Frequency with valid trip_id reference
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed.Frequencies = append(feed.Frequencies, &Frequency{TripID: "trip1", StartTime: "06:00:00", EndTime: "22:00:00", HeadwaySecs: 600})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid frequency-trip reference, got: %v", errs)
	}

	// Frequency with invalid trip_id reference
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed2.Frequencies = append(feed2.Frequencies, &Frequency{TripID: "nonexistent", StartTime: "06:00:00", EndTime: "22:00:00", HeadwaySecs: 600})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid trip_id in frequency")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "frequency") && strings.Contains(err.Error(), "trip") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about frequency referencing invalid trip, got: %v", errs)
	}
}

func TestValidatePathwayStopRefs(t *testing.T) {
	// Pathway with valid stop references
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Stops["stop2"] = &Stop{ID: "stop2", Name: "Stop 2", Lat: 40.1, Lon: -74.1}
	feed.Pathways = append(feed.Pathways, &Pathway{ID: "pathway1", FromStopID: "stop1", ToStopID: "stop2", PathwayMode: 1, IsBidirectional: 1})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid pathway stop references, got: %v", errs)
	}

	// Pathway with invalid from_stop_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed2.Pathways = append(feed2.Pathways, &Pathway{ID: "pathway1", FromStopID: "nonexistent", ToStopID: "stop1", PathwayMode: 1, IsBidirectional: 1})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid from_stop_id in pathway")
	}

	// Pathway with invalid to_stop_id
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed3.Pathways = append(feed3.Pathways, &Pathway{ID: "pathway1", FromStopID: "stop1", ToStopID: "nonexistent", PathwayMode: 1, IsBidirectional: 1})

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid to_stop_id in pathway")
	}
}

func TestValidateFareAttributeAgencyRef(t *testing.T) {
	// FareAttribute with valid agency_id reference
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0, AgencyID: "agency1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid fare_attribute-agency reference, got: %v", errs)
	}

	// FareAttribute with invalid agency_id reference
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0, AgencyID: "nonexistent"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for invalid agency_id in fare_attribute")
	}

	// FareAttribute with empty agency_id (valid)
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.FareAttributes["fare1"] = &FareAttribute{FareID: "fare1", Price: 2.50, CurrencyType: "USD", PaymentMethod: 0, Transfers: 0, AgencyID: ""}

	errs = feed3.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for fare_attribute with empty agency_id, got: %v", errs)
	}
}

// ============================================================================
// Required Field Validation Tests (6.2)
// ============================================================================

func TestValidateAgencyRequired(t *testing.T) {
	// Agency with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for agency with all required fields, got: %v", errs)
	}

	// Agency missing name
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "", URL: "http://test.com", Timezone: "America/New_York"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for agency missing name")
	}

	// Agency missing url
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "", Timezone: "America/New_York"}

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for agency missing url")
	}

	// Agency missing timezone
	feed4 := NewFeed()
	feed4.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: ""}

	errs = feed4.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for agency missing timezone")
	}
}

func TestValidateStopRequired(t *testing.T) {
	// Stop with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for stop with all required fields, got: %v", errs)
	}

	// Stop missing name (for location_type 0, 1, 2 - required)
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "", Lat: 40.0, Lon: -74.0, LocationType: 0}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for stop missing name (location_type=0)")
	}

	// Stop missing stop_id
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Stops[""] = &Stop{ID: "", Name: "Stop 1", Lat: 40.0, Lon: -74.0}

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for stop missing stop_id")
	}
}

func TestValidateRouteRequired(t *testing.T) {
	// Route with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for route with all required fields, got: %v", errs)
	}

	// Route missing route_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes[""] = &Route{ID: "", AgencyID: "agency1", ShortName: "R1", Type: 3}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for route missing route_id")
	}

	// Route must have either short_name or long_name
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "", LongName: "", Type: 3}

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for route missing both short_name and long_name")
	}

	// Route with only long_name is valid
	feed4 := NewFeed()
	feed4.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed4.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "", LongName: "Long Name", Type: 3}

	errs = feed4.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for route with only long_name, got: %v", errs)
	}
}

func TestValidateTripRequired(t *testing.T) {
	// Trip with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for trip with all required fields, got: %v", errs)
	}

	// Trip missing trip_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips[""] = &Trip{ID: "", RouteID: "route1", ServiceID: "service1"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for trip missing trip_id")
	}

	// Trip missing route_id
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed3.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed3.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "", ServiceID: "service1"}

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for trip missing route_id")
	}

	// Trip missing service_id
	feed4 := NewFeed()
	feed4.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed4.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed4.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed4.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: ""}

	errs = feed4.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for trip missing service_id")
	}
}

func TestValidateStopTimeRequired(t *testing.T) {
	// StopTime with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed.StopTimes = append(feed.StopTimes, &StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for stop_time with all required fields, got: %v", errs)
	}

	// StopTime missing trip_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed2.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed2.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed2.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed2.StopTimes = append(feed2.StopTimes, &StopTime{TripID: "", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for stop_time missing trip_id")
	}

	// StopTime missing stop_id
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed3.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed3.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed3.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed3.StopTimes = append(feed3.StopTimes, &StopTime{TripID: "trip1", StopID: "", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for stop_time missing stop_id")
	}
}

func TestValidateCalendarRequired(t *testing.T) {
	// Calendar with all required fields
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for calendar with all required fields, got: %v", errs)
	}

	// Calendar missing service_id
	feed2 := NewFeed()
	feed2.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed2.Calendars[""] = &Calendar{ServiceID: "", Monday: true, StartDate: "20240101", EndDate: "20241231"}

	errs = feed2.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for calendar missing service_id")
	}

	// Calendar missing start_date
	feed3 := NewFeed()
	feed3.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed3.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "", EndDate: "20241231"}

	errs = feed3.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for calendar missing start_date")
	}

	// Calendar missing end_date
	feed4 := NewFeed()
	feed4.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed4.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: ""}

	errs = feed4.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for calendar missing end_date")
	}
}

// ============================================================================
// Additional validation tests for completeness
// ============================================================================

func TestValidateEmptyFeed(t *testing.T) {
	// Empty feed should have at least one agency
	feed := NewFeed()

	errs := feed.Validate()
	if len(errs) == 0 {
		t.Error("Expected error for empty feed (no agencies)")
	}
}

func TestValidateValidFeed(t *testing.T) {
	// Complete valid feed
	feed := NewFeed()
	feed.Agencies["agency1"] = &Agency{ID: "agency1", Name: "Test Agency", URL: "http://test.com", Timezone: "America/New_York"}
	feed.Stops["stop1"] = &Stop{ID: "stop1", Name: "Stop 1", Lat: 40.0, Lon: -74.0}
	feed.Stops["stop2"] = &Stop{ID: "stop2", Name: "Stop 2", Lat: 40.1, Lon: -74.1}
	feed.Routes["route1"] = &Route{ID: "route1", AgencyID: "agency1", ShortName: "R1", Type: 3}
	feed.Calendars["service1"] = &Calendar{ServiceID: "service1", Monday: true, StartDate: "20240101", EndDate: "20241231"}
	feed.Trips["trip1"] = &Trip{ID: "trip1", RouteID: "route1", ServiceID: "service1"}
	feed.StopTimes = append(feed.StopTimes, &StopTime{TripID: "trip1", StopID: "stop1", StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:00"})
	feed.StopTimes = append(feed.StopTimes, &StopTime{TripID: "trip1", StopID: "stop2", StopSequence: 2, ArrivalTime: "08:10:00", DepartureTime: "08:10:00"})

	errs := feed.Validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid feed, got: %v", errs)
	}
}
