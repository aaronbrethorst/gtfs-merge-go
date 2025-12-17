package gtfs

// ParseAgency parses a CSVRow into an Agency struct.
func ParseAgency(row *CSVRow) *Agency {
	return &Agency{
		ID:       AgencyID(row.Get("agency_id")),
		Name:     row.Get("agency_name"),
		URL:      row.Get("agency_url"),
		Timezone: row.Get("agency_timezone"),
		Lang:     row.Get("agency_lang"),
		Phone:    row.Get("agency_phone"),
		FareURL:  row.Get("agency_fare_url"),
		Email:    row.Get("agency_email"),
	}
}

// ParseStop parses a CSVRow into a Stop struct.
func ParseStop(row *CSVRow) *Stop {
	return &Stop{
		ID:                 StopID(row.Get("stop_id")),
		Code:               row.Get("stop_code"),
		Name:               row.Get("stop_name"),
		Desc:               row.Get("stop_desc"),
		Lat:                row.GetFloat("stop_lat"),
		Lon:                row.GetFloat("stop_lon"),
		ZoneID:             row.Get("zone_id"),
		URL:                row.Get("stop_url"),
		LocationType:       row.GetInt("location_type"),
		ParentStation:      StopID(row.Get("parent_station")),
		Timezone:           row.Get("stop_timezone"),
		WheelchairBoarding: row.GetInt("wheelchair_boarding"),
		LevelID:            row.Get("level_id"),
		PlatformCode:       row.Get("platform_code"),
	}
}

// ParseRoute parses a CSVRow into a Route struct.
func ParseRoute(row *CSVRow) *Route {
	return &Route{
		ID:                RouteID(row.Get("route_id")),
		AgencyID:          AgencyID(row.Get("agency_id")),
		ShortName:         row.Get("route_short_name"),
		LongName:          row.Get("route_long_name"),
		Desc:              row.Get("route_desc"),
		Type:              row.GetInt("route_type"),
		URL:               row.Get("route_url"),
		Color:             row.Get("route_color"),
		TextColor:         row.Get("route_text_color"),
		SortOrder:         row.GetInt("route_sort_order"),
		ContinuousPickup:  row.GetInt("continuous_pickup"),
		ContinuousDropOff: row.GetInt("continuous_drop_off"),
	}
}

// ParseTrip parses a CSVRow into a Trip struct.
func ParseTrip(row *CSVRow) *Trip {
	return &Trip{
		ID:                   TripID(row.Get("trip_id")),
		RouteID:              RouteID(row.Get("route_id")),
		ServiceID:            ServiceID(row.Get("service_id")),
		Headsign:             row.Get("trip_headsign"),
		ShortName:            row.Get("trip_short_name"),
		DirectionID:          row.GetIntPtr("direction_id"),
		BlockID:              row.Get("block_id"),
		ShapeID:              ShapeID(row.Get("shape_id")),
		WheelchairAccessible: row.GetInt("wheelchair_accessible"),
		BikesAllowed:         row.GetInt("bikes_allowed"),
	}
}

// ParseStopTime parses a CSVRow into a StopTime struct.
func ParseStopTime(row *CSVRow) *StopTime {
	return &StopTime{
		TripID:            TripID(row.Get("trip_id")),
		ArrivalTime:       row.Get("arrival_time"),
		DepartureTime:     row.Get("departure_time"),
		StopID:            StopID(row.Get("stop_id")),
		StopSequence:      row.GetInt("stop_sequence"),
		StopHeadsign:      row.Get("stop_headsign"),
		PickupType:        row.GetInt("pickup_type"),
		DropOffType:       row.GetInt("drop_off_type"),
		ContinuousPickup:  row.GetInt("continuous_pickup"),
		ContinuousDropOff: row.GetInt("continuous_drop_off"),
		ShapeDistTraveled: row.GetFloat("shape_dist_traveled"),
		Timepoint:         row.GetInt("timepoint"),
	}
}

// ParseCalendar parses a CSVRow into a Calendar struct.
func ParseCalendar(row *CSVRow) *Calendar {
	return &Calendar{
		ServiceID: ServiceID(row.Get("service_id")),
		Monday:    row.GetBool("monday"),
		Tuesday:   row.GetBool("tuesday"),
		Wednesday: row.GetBool("wednesday"),
		Thursday:  row.GetBool("thursday"),
		Friday:    row.GetBool("friday"),
		Saturday:  row.GetBool("saturday"),
		Sunday:    row.GetBool("sunday"),
		StartDate: row.Get("start_date"),
		EndDate:   row.Get("end_date"),
	}
}

// ParseCalendarDate parses a CSVRow into a CalendarDate struct.
func ParseCalendarDate(row *CSVRow) *CalendarDate {
	return &CalendarDate{
		ServiceID:     ServiceID(row.Get("service_id")),
		Date:          row.Get("date"),
		ExceptionType: row.GetInt("exception_type"),
	}
}

// ParseShapePoint parses a CSVRow into a ShapePoint struct.
func ParseShapePoint(row *CSVRow) *ShapePoint {
	return &ShapePoint{
		ShapeID:      ShapeID(row.Get("shape_id")),
		Lat:          row.GetFloat("shape_pt_lat"),
		Lon:          row.GetFloat("shape_pt_lon"),
		Sequence:     row.GetInt("shape_pt_sequence"),
		DistTraveled: row.GetFloat("shape_dist_traveled"),
	}
}

// ParseFrequency parses a CSVRow into a Frequency struct.
func ParseFrequency(row *CSVRow) *Frequency {
	return &Frequency{
		TripID:      TripID(row.Get("trip_id")),
		StartTime:   row.Get("start_time"),
		EndTime:     row.Get("end_time"),
		HeadwaySecs: row.GetInt("headway_secs"),
		ExactTimes:  row.GetInt("exact_times"),
	}
}

// ParseTransfer parses a CSVRow into a Transfer struct.
func ParseTransfer(row *CSVRow) *Transfer {
	return &Transfer{
		FromStopID:      StopID(row.Get("from_stop_id")),
		ToStopID:        StopID(row.Get("to_stop_id")),
		TransferType:    row.GetInt("transfer_type"),
		MinTransferTime: row.GetInt("min_transfer_time"),
	}
}

// ParseFareAttribute parses a CSVRow into a FareAttribute struct.
func ParseFareAttribute(row *CSVRow) *FareAttribute {
	return &FareAttribute{
		FareID:           FareID(row.Get("fare_id")),
		Price:            row.GetFloat("price"),
		CurrencyType:     row.Get("currency_type"),
		PaymentMethod:    row.GetInt("payment_method"),
		Transfers:        row.GetInt("transfers"),
		AgencyID:         AgencyID(row.Get("agency_id")),
		TransferDuration: row.GetInt("transfer_duration"),
	}
}

// ParseFareRule parses a CSVRow into a FareRule struct.
func ParseFareRule(row *CSVRow) *FareRule {
	return &FareRule{
		FareID:        FareID(row.Get("fare_id")),
		RouteID:       RouteID(row.Get("route_id")),
		OriginID:      row.Get("origin_id"),
		DestinationID: row.Get("destination_id"),
		ContainsID:    row.Get("contains_id"),
	}
}

// ParseFeedInfo parses a CSVRow into a FeedInfo struct.
func ParseFeedInfo(row *CSVRow) *FeedInfo {
	return &FeedInfo{
		PublisherName: row.Get("feed_publisher_name"),
		PublisherURL:  row.Get("feed_publisher_url"),
		Lang:          row.Get("feed_lang"),
		DefaultLang:   row.Get("default_lang"),
		StartDate:     row.Get("feed_start_date"),
		EndDate:       row.Get("feed_end_date"),
		Version:       row.Get("feed_version"),
		ContactEmail:  row.Get("feed_contact_email"),
		ContactURL:    row.Get("feed_contact_url"),
		FeedID:        row.Get("feed_id"),
	}
}

// ParseArea parses a CSVRow into an Area struct.
func ParseArea(row *CSVRow) *Area {
	return &Area{
		ID:   AreaID(row.Get("area_id")),
		Name: row.Get("area_name"),
	}
}

// ParsePathway parses a CSVRow into a Pathway struct.
func ParsePathway(row *CSVRow) *Pathway {
	return &Pathway{
		ID:                   row.Get("pathway_id"),
		FromStopID:           StopID(row.Get("from_stop_id")),
		ToStopID:             StopID(row.Get("to_stop_id")),
		PathwayMode:          row.GetInt("pathway_mode"),
		IsBidirectional:      row.GetInt("is_bidirectional"),
		Length:               row.GetFloat("length"),
		TraversalTime:        row.GetInt("traversal_time"),
		StairCount:           row.GetInt("stair_count"),
		MaxSlope:             row.GetFloat("max_slope"),
		MinWidth:             row.GetFloat("min_width"),
		SignpostedAs:         row.Get("signposted_as"),
		ReversedSignpostedAs: row.Get("reversed_signposted_as"),
	}
}
