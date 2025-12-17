package gtfs

import (
	"reflect"
	"testing"
)

// fieldSpec defines an expected field with its name and type
type fieldSpec struct {
	Name string
	Type string
}

// checkFields verifies that a struct type has all expected fields with correct types
func checkFields(t *testing.T, structType reflect.Type, expected []fieldSpec) {
	t.Helper()

	if structType.Kind() != reflect.Struct {
		t.Fatalf("expected struct type, got %v", structType.Kind())
	}

	// Build a map of actual fields
	actualFields := make(map[string]string)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		actualFields[field.Name] = field.Type.String()
	}

	// Check each expected field
	for _, spec := range expected {
		actualType, exists := actualFields[spec.Name]
		if !exists {
			t.Errorf("missing field %q", spec.Name)
			continue
		}
		if actualType != spec.Type {
			t.Errorf("field %q has type %q, expected %q", spec.Name, actualType, spec.Type)
		}
	}
}

func TestAgencyFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "gtfs.AgencyID"},
		{"Name", "string"},
		{"URL", "string"},
		{"Timezone", "string"},
		{"Lang", "string"},
		{"Phone", "string"},
		{"FareURL", "string"},
		{"Email", "string"},
	}

	checkFields(t, reflect.TypeOf(Agency{}), expected)
}

func TestStopFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "gtfs.StopID"},
		{"Code", "string"},
		{"Name", "string"},
		{"Desc", "string"},
		{"Lat", "float64"},
		{"Lon", "float64"},
		{"ZoneID", "string"},
		{"URL", "string"},
		{"LocationType", "int"},
		{"ParentStation", "gtfs.StopID"},
		{"Timezone", "string"},
		{"WheelchairBoarding", "int"},
		{"LevelID", "string"},
		{"PlatformCode", "string"},
	}

	checkFields(t, reflect.TypeOf(Stop{}), expected)
}

func TestRouteFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "gtfs.RouteID"},
		{"AgencyID", "gtfs.AgencyID"},
		{"ShortName", "string"},
		{"LongName", "string"},
		{"Desc", "string"},
		{"Type", "int"},
		{"URL", "string"},
		{"Color", "string"},
		{"TextColor", "string"},
		{"SortOrder", "*int"},
		{"ContinuousPickup", "int"},
		{"ContinuousDropOff", "int"},
	}

	checkFields(t, reflect.TypeOf(Route{}), expected)
}

func TestTripFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "gtfs.TripID"},
		{"RouteID", "gtfs.RouteID"},
		{"ServiceID", "gtfs.ServiceID"},
		{"Headsign", "string"},
		{"ShortName", "string"},
		{"DirectionID", "*int"},
		{"BlockID", "string"},
		{"ShapeID", "gtfs.ShapeID"},
		{"WheelchairAccessible", "int"},
		{"BikesAllowed", "int"},
	}

	checkFields(t, reflect.TypeOf(Trip{}), expected)
}

func TestStopTimeFields(t *testing.T) {
	expected := []fieldSpec{
		{"TripID", "gtfs.TripID"},
		{"ArrivalTime", "string"},
		{"DepartureTime", "string"},
		{"StopID", "gtfs.StopID"},
		{"StopSequence", "int"},
		{"StopHeadsign", "string"},
		{"PickupType", "int"},
		{"DropOffType", "int"},
		{"ContinuousPickup", "int"},
		{"ContinuousDropOff", "int"},
		{"ShapeDistTraveled", "*float64"},
		{"Timepoint", "*int"},
	}

	checkFields(t, reflect.TypeOf(StopTime{}), expected)
}

func TestCalendarFields(t *testing.T) {
	expected := []fieldSpec{
		{"ServiceID", "gtfs.ServiceID"},
		{"Monday", "bool"},
		{"Tuesday", "bool"},
		{"Wednesday", "bool"},
		{"Thursday", "bool"},
		{"Friday", "bool"},
		{"Saturday", "bool"},
		{"Sunday", "bool"},
		{"StartDate", "string"},
		{"EndDate", "string"},
	}

	checkFields(t, reflect.TypeOf(Calendar{}), expected)
}

func TestCalendarDateFields(t *testing.T) {
	expected := []fieldSpec{
		{"ServiceID", "gtfs.ServiceID"},
		{"Date", "string"},
		{"ExceptionType", "int"},
	}

	checkFields(t, reflect.TypeOf(CalendarDate{}), expected)
}

func TestShapePointFields(t *testing.T) {
	expected := []fieldSpec{
		{"ShapeID", "gtfs.ShapeID"},
		{"Lat", "float64"},
		{"Lon", "float64"},
		{"Sequence", "int"},
		{"DistTraveled", "*float64"},
	}

	checkFields(t, reflect.TypeOf(ShapePoint{}), expected)
}

func TestFrequencyFields(t *testing.T) {
	expected := []fieldSpec{
		{"TripID", "gtfs.TripID"},
		{"StartTime", "string"},
		{"EndTime", "string"},
		{"HeadwaySecs", "int"},
		{"ExactTimes", "int"},
	}

	checkFields(t, reflect.TypeOf(Frequency{}), expected)
}

func TestTransferFields(t *testing.T) {
	expected := []fieldSpec{
		{"FromStopID", "gtfs.StopID"},
		{"ToStopID", "gtfs.StopID"},
		{"TransferType", "int"},
		{"MinTransferTime", "int"},
	}

	checkFields(t, reflect.TypeOf(Transfer{}), expected)
}

func TestFareAttributeFields(t *testing.T) {
	expected := []fieldSpec{
		{"FareID", "gtfs.FareID"},
		{"Price", "float64"},
		{"CurrencyType", "string"},
		{"PaymentMethod", "int"},
		{"Transfers", "int"},
		{"AgencyID", "gtfs.AgencyID"},
		{"TransferDuration", "int"},
	}

	checkFields(t, reflect.TypeOf(FareAttribute{}), expected)
}

func TestFareRuleFields(t *testing.T) {
	expected := []fieldSpec{
		{"FareID", "gtfs.FareID"},
		{"RouteID", "gtfs.RouteID"},
		{"OriginID", "string"},
		{"DestinationID", "string"},
		{"ContainsID", "string"},
	}

	checkFields(t, reflect.TypeOf(FareRule{}), expected)
}

func TestFeedInfoFields(t *testing.T) {
	expected := []fieldSpec{
		{"PublisherName", "string"},
		{"PublisherURL", "string"},
		{"Lang", "string"},
		{"DefaultLang", "string"},
		{"StartDate", "string"},
		{"EndDate", "string"},
		{"Version", "string"},
		{"ContactEmail", "string"},
		{"ContactURL", "string"},
	}

	checkFields(t, reflect.TypeOf(FeedInfo{}), expected)
}

func TestAreaFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "gtfs.AreaID"},
		{"Name", "string"},
	}

	checkFields(t, reflect.TypeOf(Area{}), expected)
}

func TestPathwayFields(t *testing.T) {
	expected := []fieldSpec{
		{"ID", "string"},
		{"FromStopID", "gtfs.StopID"},
		{"ToStopID", "gtfs.StopID"},
		{"PathwayMode", "int"},
		{"IsBidirectional", "int"},
		{"Length", "float64"},
		{"TraversalTime", "int"},
		{"StairCount", "int"},
		{"MaxSlope", "float64"},
		{"MinWidth", "float64"},
		{"SignpostedAs", "string"},
		{"ReversedSignpostedAs", "string"},
	}

	checkFields(t, reflect.TypeOf(Pathway{}), expected)
}

// TestTypeAliasesExist verifies that all ID type aliases are defined
func TestTypeAliasesExist(t *testing.T) {
	// Test that type aliases compile and have the underlying type string
	var agencyID AgencyID = "test"
	var stopID StopID = "test"
	var routeID RouteID = "test"
	var tripID TripID = "test"
	var serviceID ServiceID = "test"
	var shapeID ShapeID = "test"
	var fareID FareID = "test"
	var areaID AreaID = "test"

	// Verify they can be converted back to string
	if string(agencyID) != "test" {
		t.Error("AgencyID should be convertible to string")
	}
	if string(stopID) != "test" {
		t.Error("StopID should be convertible to string")
	}
	if string(routeID) != "test" {
		t.Error("RouteID should be convertible to string")
	}
	if string(tripID) != "test" {
		t.Error("TripID should be convertible to string")
	}
	if string(serviceID) != "test" {
		t.Error("ServiceID should be convertible to string")
	}
	if string(shapeID) != "test" {
		t.Error("ShapeID should be convertible to string")
	}
	if string(fareID) != "test" {
		t.Error("FareID should be convertible to string")
	}
	if string(areaID) != "test" {
		t.Error("AreaID should be convertible to string")
	}
}
