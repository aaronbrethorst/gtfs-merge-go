package gtfs

import (
	"fmt"
)

// ValidationError represents a validation error with context
type ValidationError struct {
	EntityType string
	EntityID   string
	Field      string
	Message    string
}

func (e *ValidationError) Error() string {
	if e.EntityID != "" {
		return fmt.Sprintf("%s '%s': %s", e.EntityType, e.EntityID, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.EntityType, e.Message)
}

// Validate checks the feed for GTFS compliance and referential integrity.
// Returns a slice of errors, or nil if the feed is valid.
func (f *Feed) Validate() []error {
	var errs []error

	// Validate that feed has at least one agency
	if len(f.Agencies) == 0 {
		errs = append(errs, &ValidationError{
			EntityType: "feed",
			Message:    "feed must have at least one agency",
		})
	}

	// Validate agencies (required fields)
	for _, agency := range f.Agencies {
		errs = append(errs, f.validateAgency(agency)...)
	}

	// Validate stops (required fields and parent_station reference)
	for _, stop := range f.Stops {
		errs = append(errs, f.validateStop(stop)...)
	}

	// Validate routes (required fields and agency reference)
	for _, route := range f.Routes {
		errs = append(errs, f.validateRoute(route)...)
	}

	// Validate calendars (required fields)
	for _, calendar := range f.Calendars {
		errs = append(errs, f.validateCalendar(calendar)...)
	}

	// Validate trips (required fields and route/service/shape references)
	for _, trip := range f.Trips {
		errs = append(errs, f.validateTrip(trip)...)
	}

	// Validate stop_times (required fields and trip/stop references)
	for _, stopTime := range f.StopTimes {
		errs = append(errs, f.validateStopTime(stopTime)...)
	}

	// Validate transfers (stop references)
	for _, transfer := range f.Transfers {
		errs = append(errs, f.validateTransfer(transfer)...)
	}

	// Validate frequencies (trip references)
	for _, frequency := range f.Frequencies {
		errs = append(errs, f.validateFrequency(frequency)...)
	}

	// Validate fare_attributes (agency reference)
	for _, fareAttr := range f.FareAttributes {
		errs = append(errs, f.validateFareAttribute(fareAttr)...)
	}

	// Validate fare_rules (fare and route references)
	for _, fareRule := range f.FareRules {
		errs = append(errs, f.validateFareRule(fareRule)...)
	}

	// Validate pathways (stop references)
	for _, pathway := range f.Pathways {
		errs = append(errs, f.validatePathway(pathway)...)
	}

	return errs
}

// validateAgency checks agency required fields
func (f *Feed) validateAgency(agency *Agency) []error {
	var errs []error

	if agency.Name == "" {
		errs = append(errs, &ValidationError{
			EntityType: "agency",
			EntityID:   string(agency.ID),
			Field:      "agency_name",
			Message:    "agency_name is required",
		})
	}

	if agency.URL == "" {
		errs = append(errs, &ValidationError{
			EntityType: "agency",
			EntityID:   string(agency.ID),
			Field:      "agency_url",
			Message:    "agency_url is required",
		})
	}

	if agency.Timezone == "" {
		errs = append(errs, &ValidationError{
			EntityType: "agency",
			EntityID:   string(agency.ID),
			Field:      "agency_timezone",
			Message:    "agency_timezone is required",
		})
	}

	return errs
}

// validateStop checks stop required fields and parent_station reference
func (f *Feed) validateStop(stop *Stop) []error {
	var errs []error

	if stop.ID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "stop",
			EntityID:   string(stop.ID),
			Field:      "stop_id",
			Message:    "stop_id is required",
		})
	}

	// stop_name is required for location_type 0, 1, 2
	if stop.LocationType <= 2 && stop.Name == "" {
		errs = append(errs, &ValidationError{
			EntityType: "stop",
			EntityID:   string(stop.ID),
			Field:      "stop_name",
			Message:    "stop_name is required for location_type 0, 1, or 2",
		})
	}

	// Validate parent_station reference if specified
	if stop.ParentStation != "" {
		if _, exists := f.Stops[stop.ParentStation]; !exists {
			errs = append(errs, &ValidationError{
				EntityType: "stop",
				EntityID:   string(stop.ID),
				Field:      "parent_station",
				Message:    fmt.Sprintf("stop references non-existent parent_station '%s'", stop.ParentStation),
			})
		}
	}

	return errs
}

// validateRoute checks route required fields and agency reference
func (f *Feed) validateRoute(route *Route) []error {
	var errs []error

	if route.ID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "route",
			EntityID:   string(route.ID),
			Field:      "route_id",
			Message:    "route_id is required",
		})
	}

	// At least one of route_short_name or route_long_name must be specified
	if route.ShortName == "" && route.LongName == "" {
		errs = append(errs, &ValidationError{
			EntityType: "route",
			EntityID:   string(route.ID),
			Field:      "route_short_name/route_long_name",
			Message:    "either route_short_name or route_long_name is required",
		})
	}

	// Validate agency_id reference
	if route.AgencyID != "" {
		if _, exists := f.Agencies[route.AgencyID]; !exists {
			errs = append(errs, &ValidationError{
				EntityType: "route",
				EntityID:   string(route.ID),
				Field:      "agency_id",
				Message:    fmt.Sprintf("route references non-existent agency '%s'", route.AgencyID),
			})
		}
	} else if len(f.Agencies) > 1 {
		// agency_id is required if there are multiple agencies
		errs = append(errs, &ValidationError{
			EntityType: "route",
			EntityID:   string(route.ID),
			Field:      "agency_id",
			Message:    "agency_id is required when multiple agencies exist",
		})
	}

	return errs
}

// validateCalendar checks calendar required fields
func (f *Feed) validateCalendar(calendar *Calendar) []error {
	var errs []error

	if calendar.ServiceID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "calendar",
			EntityID:   string(calendar.ServiceID),
			Field:      "service_id",
			Message:    "service_id is required",
		})
	}

	if calendar.StartDate == "" {
		errs = append(errs, &ValidationError{
			EntityType: "calendar",
			EntityID:   string(calendar.ServiceID),
			Field:      "start_date",
			Message:    "start_date is required",
		})
	}

	if calendar.EndDate == "" {
		errs = append(errs, &ValidationError{
			EntityType: "calendar",
			EntityID:   string(calendar.ServiceID),
			Field:      "end_date",
			Message:    "end_date is required",
		})
	}

	return errs
}

// validateTrip checks trip required fields and references
func (f *Feed) validateTrip(trip *Trip) []error {
	var errs []error

	if trip.ID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "trip",
			EntityID:   string(trip.ID),
			Field:      "trip_id",
			Message:    "trip_id is required",
		})
	}

	if trip.RouteID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "trip",
			EntityID:   string(trip.ID),
			Field:      "route_id",
			Message:    "route_id is required",
		})
	} else if _, exists := f.Routes[trip.RouteID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "trip",
			EntityID:   string(trip.ID),
			Field:      "route_id",
			Message:    fmt.Sprintf("trip references non-existent route '%s'", trip.RouteID),
		})
	}

	if trip.ServiceID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "trip",
			EntityID:   string(trip.ID),
			Field:      "service_id",
			Message:    "service_id is required",
		})
	} else {
		// Service can be in calendar or calendar_dates
		_, inCalendar := f.Calendars[trip.ServiceID]
		_, inCalendarDates := f.CalendarDates[trip.ServiceID]
		if !inCalendar && !inCalendarDates {
			errs = append(errs, &ValidationError{
				EntityType: "trip",
				EntityID:   string(trip.ID),
				Field:      "service_id",
				Message:    fmt.Sprintf("trip references non-existent service '%s'", trip.ServiceID),
			})
		}
	}

	// Validate shape_id reference if specified
	if trip.ShapeID != "" {
		if _, exists := f.Shapes[trip.ShapeID]; !exists {
			errs = append(errs, &ValidationError{
				EntityType: "trip",
				EntityID:   string(trip.ID),
				Field:      "shape_id",
				Message:    fmt.Sprintf("trip references non-existent shape '%s'", trip.ShapeID),
			})
		}
	}

	return errs
}

// validateStopTime checks stop_time required fields and references
func (f *Feed) validateStopTime(stopTime *StopTime) []error {
	var errs []error

	if stopTime.TripID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "stop_time",
			Field:      "trip_id",
			Message:    "trip_id is required",
		})
	} else if _, exists := f.Trips[stopTime.TripID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "stop_time",
			EntityID:   fmt.Sprintf("trip %s seq %d", stopTime.TripID, stopTime.StopSequence),
			Field:      "trip_id",
			Message:    fmt.Sprintf("stop_time references non-existent trip '%s'", stopTime.TripID),
		})
	}

	if stopTime.StopID == "" {
		errs = append(errs, &ValidationError{
			EntityType: "stop_time",
			EntityID:   fmt.Sprintf("trip %s seq %d", stopTime.TripID, stopTime.StopSequence),
			Field:      "stop_id",
			Message:    "stop_id is required",
		})
	} else if _, exists := f.Stops[stopTime.StopID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "stop_time",
			EntityID:   fmt.Sprintf("trip %s seq %d", stopTime.TripID, stopTime.StopSequence),
			Field:      "stop_id",
			Message:    fmt.Sprintf("stop_time references non-existent stop '%s'", stopTime.StopID),
		})
	}

	return errs
}

// validateTransfer checks transfer stop references
func (f *Feed) validateTransfer(transfer *Transfer) []error {
	var errs []error

	if _, exists := f.Stops[transfer.FromStopID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "transfer",
			Field:      "from_stop_id",
			Message:    fmt.Sprintf("transfer references non-existent from_stop_id '%s'", transfer.FromStopID),
		})
	}

	if _, exists := f.Stops[transfer.ToStopID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "transfer",
			Field:      "to_stop_id",
			Message:    fmt.Sprintf("transfer references non-existent to_stop_id '%s'", transfer.ToStopID),
		})
	}

	return errs
}

// validateFrequency checks frequency trip reference
func (f *Feed) validateFrequency(frequency *Frequency) []error {
	var errs []error

	if _, exists := f.Trips[frequency.TripID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "frequency",
			Field:      "trip_id",
			Message:    fmt.Sprintf("frequency references non-existent trip '%s'", frequency.TripID),
		})
	}

	return errs
}

// validateFareAttribute checks fare_attribute agency reference
func (f *Feed) validateFareAttribute(fareAttr *FareAttribute) []error {
	var errs []error

	// agency_id is optional, but if specified must be valid
	if fareAttr.AgencyID != "" {
		if _, exists := f.Agencies[fareAttr.AgencyID]; !exists {
			errs = append(errs, &ValidationError{
				EntityType: "fare_attribute",
				EntityID:   string(fareAttr.FareID),
				Field:      "agency_id",
				Message:    fmt.Sprintf("fare_attribute references non-existent agency '%s'", fareAttr.AgencyID),
			})
		}
	}

	return errs
}

// validateFareRule checks fare_rule references
func (f *Feed) validateFareRule(fareRule *FareRule) []error {
	var errs []error

	// fare_id is required and must be valid
	if _, exists := f.FareAttributes[fareRule.FareID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "fare_rule",
			Field:      "fare_id",
			Message:    fmt.Sprintf("fare_rule references non-existent fare_id '%s'", fareRule.FareID),
		})
	}

	// route_id is optional, but if specified must be valid
	if fareRule.RouteID != "" {
		if _, exists := f.Routes[fareRule.RouteID]; !exists {
			errs = append(errs, &ValidationError{
				EntityType: "fare_rule",
				Field:      "route_id",
				Message:    fmt.Sprintf("fare_rule references non-existent route_id '%s'", fareRule.RouteID),
			})
		}
	}

	return errs
}

// validatePathway checks pathway stop references
func (f *Feed) validatePathway(pathway *Pathway) []error {
	var errs []error

	if _, exists := f.Stops[pathway.FromStopID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "pathway",
			EntityID:   pathway.ID,
			Field:      "from_stop_id",
			Message:    fmt.Sprintf("pathway references non-existent from_stop_id '%s'", pathway.FromStopID),
		})
	}

	if _, exists := f.Stops[pathway.ToStopID]; !exists {
		errs = append(errs, &ValidationError{
			EntityType: "pathway",
			EntityID:   pathway.ID,
			Field:      "to_stop_id",
			Message:    fmt.Sprintf("pathway references non-existent to_stop_id '%s'", pathway.ToStopID),
		})
	}

	return errs
}
