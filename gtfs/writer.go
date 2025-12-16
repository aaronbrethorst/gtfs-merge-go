package gtfs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strconv"
)

// WriteToPath writes a GTFS feed to a zip file at the given path.
func WriteToPath(feed *Feed, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if err := WriteToZip(feed, f); err != nil {
		return err
	}

	return nil
}

// WriteToZip writes a GTFS feed to a zip archive.
func WriteToZip(feed *Feed, w io.Writer) error {
	zw := zip.NewWriter(w)
	defer func() { _ = zw.Close() }()

	// Write required files
	if err := writeAgencies(zw, feed); err != nil {
		return fmt.Errorf("writing agency.txt: %w", err)
	}
	if err := writeStops(zw, feed); err != nil {
		return fmt.Errorf("writing stops.txt: %w", err)
	}
	if err := writeRoutes(zw, feed); err != nil {
		return fmt.Errorf("writing routes.txt: %w", err)
	}
	if err := writeTrips(zw, feed); err != nil {
		return fmt.Errorf("writing trips.txt: %w", err)
	}
	if err := writeStopTimes(zw, feed); err != nil {
		return fmt.Errorf("writing stop_times.txt: %w", err)
	}

	// Write calendar files (at least one required)
	if len(feed.Calendars) > 0 {
		if err := writeCalendars(zw, feed); err != nil {
			return fmt.Errorf("writing calendar.txt: %w", err)
		}
	}
	if len(feed.CalendarDates) > 0 {
		if err := writeCalendarDates(zw, feed); err != nil {
			return fmt.Errorf("writing calendar_dates.txt: %w", err)
		}
	}

	// Write optional files (only if data exists)
	if len(feed.Shapes) > 0 {
		if err := writeShapes(zw, feed); err != nil {
			return fmt.Errorf("writing shapes.txt: %w", err)
		}
	}
	if len(feed.Frequencies) > 0 {
		if err := writeFrequencies(zw, feed); err != nil {
			return fmt.Errorf("writing frequencies.txt: %w", err)
		}
	}
	if len(feed.Transfers) > 0 {
		if err := writeTransfers(zw, feed); err != nil {
			return fmt.Errorf("writing transfers.txt: %w", err)
		}
	}
	if len(feed.FareAttributes) > 0 {
		if err := writeFareAttributes(zw, feed); err != nil {
			return fmt.Errorf("writing fare_attributes.txt: %w", err)
		}
	}
	if len(feed.FareRules) > 0 {
		if err := writeFareRules(zw, feed); err != nil {
			return fmt.Errorf("writing fare_rules.txt: %w", err)
		}
	}
	if feed.FeedInfo != nil {
		if err := writeFeedInfo(zw, feed); err != nil {
			return fmt.Errorf("writing feed_info.txt: %w", err)
		}
	}
	if len(feed.Areas) > 0 {
		if err := writeAreas(zw, feed); err != nil {
			return fmt.Errorf("writing areas.txt: %w", err)
		}
	}
	if len(feed.Pathways) > 0 {
		if err := writePathways(zw, feed); err != nil {
			return fmt.Errorf("writing pathways.txt: %w", err)
		}
	}

	return zw.Close()
}

// Helper functions for formatting values
func formatInt(v int) string {
	return strconv.Itoa(v)
}

func formatFloat(v float64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatBool(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

// writeAgencies writes agency.txt
func writeAgencies(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("agency.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"agency_id", "agency_name", "agency_url", "agency_timezone", "agency_lang", "agency_phone", "agency_fare_url", "agency_email"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, a := range feed.Agencies {
		record := []string{
			string(a.ID),
			a.Name,
			a.URL,
			a.Timezone,
			a.Lang,
			a.Phone,
			a.FareURL,
			a.Email,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeStops writes stops.txt
func writeStops(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("stops.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"stop_id", "stop_code", "stop_name", "stop_desc", "stop_lat", "stop_lon", "zone_id", "stop_url", "location_type", "parent_station", "stop_timezone", "wheelchair_boarding", "level_id", "platform_code"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, s := range feed.Stops {
		record := []string{
			string(s.ID),
			s.Code,
			s.Name,
			s.Desc,
			strconv.FormatFloat(s.Lat, 'f', -1, 64),
			strconv.FormatFloat(s.Lon, 'f', -1, 64),
			s.ZoneID,
			s.URL,
			formatInt(s.LocationType),
			string(s.ParentStation),
			s.Timezone,
			formatInt(s.WheelchairBoarding),
			s.LevelID,
			s.PlatformCode,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeRoutes writes routes.txt
func writeRoutes(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("routes.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"route_id", "agency_id", "route_short_name", "route_long_name", "route_desc", "route_type", "route_url", "route_color", "route_text_color", "route_sort_order", "continuous_pickup", "continuous_drop_off"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, r := range feed.Routes {
		record := []string{
			string(r.ID),
			string(r.AgencyID),
			r.ShortName,
			r.LongName,
			r.Desc,
			formatInt(r.Type),
			r.URL,
			r.Color,
			r.TextColor,
			formatInt(r.SortOrder),
			formatInt(r.ContinuousPickup),
			formatInt(r.ContinuousDropOff),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeTrips writes trips.txt
func writeTrips(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("trips.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"trip_id", "route_id", "service_id", "trip_headsign", "trip_short_name", "direction_id", "block_id", "shape_id", "wheelchair_accessible", "bikes_allowed"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, t := range feed.Trips {
		record := []string{
			string(t.ID),
			string(t.RouteID),
			string(t.ServiceID),
			t.Headsign,
			t.ShortName,
			formatInt(t.DirectionID),
			t.BlockID,
			string(t.ShapeID),
			formatInt(t.WheelchairAccessible),
			formatInt(t.BikesAllowed),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeStopTimes writes stop_times.txt
func writeStopTimes(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("stop_times.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence", "stop_headsign", "pickup_type", "drop_off_type", "continuous_pickup", "continuous_drop_off", "shape_dist_traveled", "timepoint"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, st := range feed.StopTimes {
		record := []string{
			string(st.TripID),
			st.ArrivalTime,
			st.DepartureTime,
			string(st.StopID),
			formatInt(st.StopSequence),
			st.StopHeadsign,
			formatInt(st.PickupType),
			formatInt(st.DropOffType),
			formatInt(st.ContinuousPickup),
			formatInt(st.ContinuousDropOff),
			formatFloat(st.ShapeDistTraveled),
			formatInt(st.Timepoint),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeCalendars writes calendar.txt
func writeCalendars(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("calendar.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, c := range feed.Calendars {
		record := []string{
			string(c.ServiceID),
			formatBool(c.Monday),
			formatBool(c.Tuesday),
			formatBool(c.Wednesday),
			formatBool(c.Thursday),
			formatBool(c.Friday),
			formatBool(c.Saturday),
			formatBool(c.Sunday),
			c.StartDate,
			c.EndDate,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeCalendarDates writes calendar_dates.txt
func writeCalendarDates(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("calendar_dates.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"service_id", "date", "exception_type"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, dates := range feed.CalendarDates {
		for _, cd := range dates {
			record := []string{
				string(cd.ServiceID),
				cd.Date,
				formatInt(cd.ExceptionType),
			}
			if err := csvw.WriteRecord(record); err != nil {
				return err
			}
		}
	}

	return csvw.Flush()
}

// writeShapes writes shapes.txt
func writeShapes(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("shapes.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence", "shape_dist_traveled"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, points := range feed.Shapes {
		for _, sp := range points {
			record := []string{
				string(sp.ShapeID),
				strconv.FormatFloat(sp.Lat, 'f', -1, 64),
				strconv.FormatFloat(sp.Lon, 'f', -1, 64),
				formatInt(sp.Sequence),
				formatFloat(sp.DistTraveled),
			}
			if err := csvw.WriteRecord(record); err != nil {
				return err
			}
		}
	}

	return csvw.Flush()
}

// writeFrequencies writes frequencies.txt
func writeFrequencies(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("frequencies.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"trip_id", "start_time", "end_time", "headway_secs", "exact_times"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, f := range feed.Frequencies {
		record := []string{
			string(f.TripID),
			f.StartTime,
			f.EndTime,
			formatInt(f.HeadwaySecs),
			formatInt(f.ExactTimes),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeTransfers writes transfers.txt
func writeTransfers(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("transfers.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"from_stop_id", "to_stop_id", "transfer_type", "min_transfer_time"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, t := range feed.Transfers {
		record := []string{
			string(t.FromStopID),
			string(t.ToStopID),
			formatInt(t.TransferType),
			formatInt(t.MinTransferTime),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeFareAttributes writes fare_attributes.txt
func writeFareAttributes(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("fare_attributes.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"fare_id", "price", "currency_type", "payment_method", "transfers", "agency_id", "transfer_duration"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, fa := range feed.FareAttributes {
		record := []string{
			string(fa.FareID),
			strconv.FormatFloat(fa.Price, 'f', 2, 64),
			fa.CurrencyType,
			formatInt(fa.PaymentMethod),
			formatInt(fa.Transfers),
			string(fa.AgencyID),
			formatInt(fa.TransferDuration),
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeFareRules writes fare_rules.txt
func writeFareRules(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("fare_rules.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"fare_id", "route_id", "origin_id", "destination_id", "contains_id"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, fr := range feed.FareRules {
		record := []string{
			string(fr.FareID),
			string(fr.RouteID),
			fr.OriginID,
			fr.DestinationID,
			fr.ContainsID,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writeFeedInfo writes feed_info.txt
func writeFeedInfo(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("feed_info.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"feed_publisher_name", "feed_publisher_url", "feed_lang", "default_lang", "feed_start_date", "feed_end_date", "feed_version", "feed_contact_email", "feed_contact_url"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	fi := feed.FeedInfo
	record := []string{
		fi.PublisherName,
		fi.PublisherURL,
		fi.Lang,
		fi.DefaultLang,
		fi.StartDate,
		fi.EndDate,
		fi.Version,
		fi.ContactEmail,
		fi.ContactURL,
	}
	if err := csvw.WriteRecord(record); err != nil {
		return err
	}

	return csvw.Flush()
}

// writeAreas writes areas.txt
func writeAreas(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("areas.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"area_id", "area_name"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, a := range feed.Areas {
		record := []string{
			string(a.ID),
			a.Name,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}

// writePathways writes pathways.txt
func writePathways(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("pathways.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)
	header := []string{"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional", "length", "traversal_time", "stair_count", "max_slope", "min_width", "signposted_as", "reversed_signposted_as"}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, p := range feed.Pathways {
		record := []string{
			p.ID,
			string(p.FromStopID),
			string(p.ToStopID),
			formatInt(p.PathwayMode),
			formatInt(p.IsBidirectional),
			formatFloat(p.Length),
			formatInt(p.TraversalTime),
			formatInt(p.StairCount),
			formatFloat(p.MaxSlope),
			formatFloat(p.MinWidth),
			p.SignpostedAs,
			p.ReversedSignpostedAs,
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}
