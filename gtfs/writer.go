package gtfs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"sort"
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
	if len(feed.FeedInfos) > 0 {
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

// formatOptionalInt formats an optional integer field.
// Returns empty string for 0 (the GTFS default for optional int fields).
func formatOptionalInt(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

// formatIntPtr formats a pointer to int.
// Returns empty string for nil, otherwise formats the integer value.
func formatIntPtr(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

func formatFloat(v float64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// formatFloatPtr formats a pointer to float64.
// Returns empty string for nil, otherwise formats the value with 6 decimal places.
func formatFloatPtr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 6, 64)
}

// formatPriceFloat formats a price with 6 decimal places, including when the value is 0
func formatPriceFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 6, 64)
}

func formatBool(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

// columnChecker tracks which optional columns have non-default values.
// Used to match Java behavior: only output columns that have actual data.
type columnChecker struct {
	columns map[string]bool // column name -> has non-default value
}

func newColumnChecker(optionalCols []string) *columnChecker {
	checker := &columnChecker{
		columns: make(map[string]bool),
	}
	for _, col := range optionalCols {
		checker.columns[col] = false
	}
	return checker
}

func (c *columnChecker) markNonDefault(col string) {
	c.columns[col] = true
}

func (c *columnChecker) hasNonDefaultValue(col string) bool {
	if hasValue, exists := c.columns[col]; exists {
		return hasValue
	}
	return true // If not tracked as optional, include by default
}

func (c *columnChecker) allFound() bool {
	for _, found := range c.columns {
		if !found {
			return false
		}
	}
	return true
}

// writeAgencies writes agency.txt
func writeAgencies(zw *zip.Writer, feed *Feed) error {
	w, err := zw.Create("agency.txt")
	if err != nil {
		return err
	}

	csvw := NewCSVWriter(w)

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Agency) string
	}
	allCols := []colDef{
		{"agency_id", func(a *Agency) string { return string(a.ID) }},
		{"agency_name", func(a *Agency) string { return a.Name }},
		{"agency_url", func(a *Agency) string { return a.URL }},
		{"agency_timezone", func(a *Agency) string { return a.Timezone }},
		{"agency_lang", func(a *Agency) string { return a.Lang }},
		{"agency_phone", func(a *Agency) string { return a.Phone }},
		{"agency_fare_url", func(a *Agency) string { return a.FareURL }},
		{"agency_email", func(a *Agency) string { return a.Email }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("agency.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, a := range feed.Agencies {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(a)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Stop) string
	}
	allCols := []colDef{
		{"stop_id", func(s *Stop) string { return string(s.ID) }},
		{"stop_code", func(s *Stop) string { return s.Code }},
		{"stop_name", func(s *Stop) string { return s.Name }},
		{"stop_desc", func(s *Stop) string { return s.Desc }},
		{"stop_lat", func(s *Stop) string { return strconv.FormatFloat(s.Lat, 'f', -1, 64) }},
		{"stop_lon", func(s *Stop) string { return strconv.FormatFloat(s.Lon, 'f', -1, 64) }},
		{"zone_id", func(s *Stop) string { return s.ZoneID }},
		{"stop_url", func(s *Stop) string { return s.URL }},
		{"location_type", func(s *Stop) string { return formatOptionalInt(s.LocationType) }},
		{"parent_station", func(s *Stop) string { return string(s.ParentStation) }},
		{"stop_timezone", func(s *Stop) string { return s.Timezone }},
		{"wheelchair_boarding", func(s *Stop) string { return formatOptionalInt(s.WheelchairBoarding) }},
		{"level_id", func(s *Stop) string { return s.LevelID }},
		{"platform_code", func(s *Stop) string { return s.PlatformCode }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"stop_id": true, "stop_name": true, "stop_lat": true, "stop_lon": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{
		"stop_code", "stop_desc", "zone_id", "stop_url", "location_type",
		"parent_station", "stop_timezone", "wheelchair_boarding", "level_id", "platform_code",
	}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, s := range feed.Stops {
		if s.Code != "" {
			checker.markNonDefault("stop_code")
		}
		if s.Desc != "" {
			checker.markNonDefault("stop_desc")
		}
		if s.ZoneID != "" {
			checker.markNonDefault("zone_id")
		}
		if s.URL != "" {
			checker.markNonDefault("stop_url")
		}
		if s.LocationType != 0 {
			checker.markNonDefault("location_type")
		}
		if s.ParentStation != "" {
			checker.markNonDefault("parent_station")
		}
		if s.Timezone != "" {
			checker.markNonDefault("stop_timezone")
		}
		if s.WheelchairBoarding != 0 {
			checker.markNonDefault("wheelchair_boarding")
		}
		if s.LevelID != "" {
			checker.markNonDefault("level_id")
		}
		if s.PlatformCode != "" {
			checker.markNonDefault("platform_code")
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("stops.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, s := range feed.Stops {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(s)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Route) string
	}
	allCols := []colDef{
		{"route_id", func(r *Route) string { return string(r.ID) }},
		{"agency_id", func(r *Route) string { return string(r.AgencyID) }},
		{"route_short_name", func(r *Route) string { return r.ShortName }},
		{"route_long_name", func(r *Route) string { return r.LongName }},
		{"route_desc", func(r *Route) string { return r.Desc }},
		{"route_type", func(r *Route) string { return formatInt(r.Type) }},
		{"route_url", func(r *Route) string { return r.URL }},
		{"route_color", func(r *Route) string { return r.Color }},
		{"route_text_color", func(r *Route) string { return r.TextColor }},
		{"route_sort_order", func(r *Route) string { return formatIntPtr(r.SortOrder) }},
		{"continuous_pickup", func(r *Route) string { return formatOptionalInt(r.ContinuousPickup) }},
		{"continuous_drop_off", func(r *Route) string { return formatOptionalInt(r.ContinuousDropOff) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"route_id": true, "route_type": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{
		"agency_id", "route_short_name", "route_long_name", "route_desc",
		"route_url", "route_color", "route_text_color", "route_sort_order",
		"continuous_pickup", "continuous_drop_off",
	}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, r := range feed.Routes {
		if r.AgencyID != "" {
			checker.markNonDefault("agency_id")
		}
		if r.ShortName != "" {
			checker.markNonDefault("route_short_name")
		}
		if r.LongName != "" {
			checker.markNonDefault("route_long_name")
		}
		if r.Desc != "" {
			checker.markNonDefault("route_desc")
		}
		if r.URL != "" {
			checker.markNonDefault("route_url")
		}
		if r.Color != "" {
			checker.markNonDefault("route_color")
		}
		if r.TextColor != "" {
			checker.markNonDefault("route_text_color")
		}
		if r.SortOrder != nil {
			checker.markNonDefault("route_sort_order")
		}
		if r.ContinuousPickup != 0 {
			checker.markNonDefault("continuous_pickup")
		}
		if r.ContinuousDropOff != 0 {
			checker.markNonDefault("continuous_drop_off")
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("routes.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, r := range feed.Routes {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(r)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Trip) string
	}
	allCols := []colDef{
		{"trip_id", func(t *Trip) string { return string(t.ID) }},
		{"route_id", func(t *Trip) string { return string(t.RouteID) }},
		{"service_id", func(t *Trip) string { return string(t.ServiceID) }},
		{"trip_headsign", func(t *Trip) string { return t.Headsign }},
		{"trip_short_name", func(t *Trip) string { return t.ShortName }},
		{"direction_id", func(t *Trip) string { return formatIntPtr(t.DirectionID) }},
		{"block_id", func(t *Trip) string { return t.BlockID }},
		{"shape_id", func(t *Trip) string { return string(t.ShapeID) }},
		{"wheelchair_accessible", func(t *Trip) string { return formatOptionalInt(t.WheelchairAccessible) }},
		{"bikes_allowed", func(t *Trip) string { return formatOptionalInt(t.BikesAllowed) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"trip_id": true, "route_id": true, "service_id": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{
		"trip_headsign", "trip_short_name", "direction_id", "block_id",
		"shape_id", "wheelchair_accessible", "bikes_allowed",
	}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, t := range feed.Trips {
		if t.Headsign != "" {
			checker.markNonDefault("trip_headsign")
		}
		if t.ShortName != "" {
			checker.markNonDefault("trip_short_name")
		}
		if t.DirectionID != nil {
			checker.markNonDefault("direction_id")
		}
		if t.BlockID != "" {
			checker.markNonDefault("block_id")
		}
		if t.ShapeID != "" {
			checker.markNonDefault("shape_id")
		}
		if t.WheelchairAccessible != 0 {
			checker.markNonDefault("wheelchair_accessible")
		}
		if t.BikesAllowed != 0 {
			checker.markNonDefault("bikes_allowed")
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("trips.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, t := range feed.Trips {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(t)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*StopTime) string
	}
	allCols := []colDef{
		{"trip_id", func(st *StopTime) string { return string(st.TripID) }},
		{"arrival_time", func(st *StopTime) string { return st.ArrivalTime }},
		{"departure_time", func(st *StopTime) string { return st.DepartureTime }},
		{"stop_id", func(st *StopTime) string { return string(st.StopID) }},
		{"stop_sequence", func(st *StopTime) string { return formatInt(st.StopSequence) }},
		{"stop_headsign", func(st *StopTime) string { return st.StopHeadsign }},
		{"pickup_type", func(st *StopTime) string { return formatOptionalInt(st.PickupType) }},
		{"drop_off_type", func(st *StopTime) string { return formatOptionalInt(st.DropOffType) }},
		{"continuous_pickup", func(st *StopTime) string { return formatOptionalInt(st.ContinuousPickup) }},
		{"continuous_drop_off", func(st *StopTime) string { return formatOptionalInt(st.ContinuousDropOff) }},
		{"shape_dist_traveled", func(st *StopTime) string { return formatFloatPtr(st.ShapeDistTraveled) }},
		{"timepoint", func(st *StopTime) string { return formatIntPtr(st.Timepoint) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"trip_id": true, "arrival_time": true, "departure_time": true,
		"stop_id": true, "stop_sequence": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{
		"stop_headsign", "pickup_type", "drop_off_type",
		"continuous_pickup", "continuous_drop_off", "shape_dist_traveled", "timepoint",
	}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, st := range feed.StopTimes {
		if st.StopHeadsign != "" {
			checker.markNonDefault("stop_headsign")
		}
		if st.PickupType != 0 {
			checker.markNonDefault("pickup_type")
		}
		if st.DropOffType != 0 {
			checker.markNonDefault("drop_off_type")
		}
		if st.ContinuousPickup != 0 {
			checker.markNonDefault("continuous_pickup")
		}
		if st.ContinuousDropOff != 0 {
			checker.markNonDefault("continuous_drop_off")
		}
		if st.ShapeDistTraveled != nil {
			checker.markNonDefault("shape_dist_traveled")
		}
		if st.Timepoint != nil {
			checker.markNonDefault("timepoint")
		}
		// Early termination: stop if all optional columns have data
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("stop_times.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, st := range feed.StopTimes {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(st)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Calendar) string
	}
	allCols := []colDef{
		{"service_id", func(c *Calendar) string { return string(c.ServiceID) }},
		{"monday", func(c *Calendar) string { return formatBool(c.Monday) }},
		{"tuesday", func(c *Calendar) string { return formatBool(c.Tuesday) }},
		{"wednesday", func(c *Calendar) string { return formatBool(c.Wednesday) }},
		{"thursday", func(c *Calendar) string { return formatBool(c.Thursday) }},
		{"friday", func(c *Calendar) string { return formatBool(c.Friday) }},
		{"saturday", func(c *Calendar) string { return formatBool(c.Saturday) }},
		{"sunday", func(c *Calendar) string { return formatBool(c.Sunday) }},
		{"start_date", func(c *Calendar) string { return c.StartDate }},
		{"end_date", func(c *Calendar) string { return c.EndDate }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("calendar.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, c := range feed.Calendars {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(c)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*CalendarDate) string
	}
	allCols := []colDef{
		{"service_id", func(cd *CalendarDate) string { return string(cd.ServiceID) }},
		{"date", func(cd *CalendarDate) string { return cd.Date }},
		{"exception_type", func(cd *CalendarDate) string { return formatInt(cd.ExceptionType) }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("calendar_dates.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, dates := range feed.CalendarDates {
		for _, cd := range dates {
			record := make([]string, len(activeCols))
			for i, col := range activeCols {
				record[i] = col.getter(cd)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*ShapePoint) string
	}
	allCols := []colDef{
		{"shape_id", func(sp *ShapePoint) string { return string(sp.ShapeID) }},
		{"shape_pt_lat", func(sp *ShapePoint) string { return strconv.FormatFloat(sp.Lat, 'f', -1, 64) }},
		{"shape_pt_lon", func(sp *ShapePoint) string { return strconv.FormatFloat(sp.Lon, 'f', -1, 64) }},
		{"shape_pt_sequence", func(sp *ShapePoint) string { return formatInt(sp.Sequence) }},
		{"shape_dist_traveled", func(sp *ShapePoint) string { return formatFloatPtr(sp.DistTraveled) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"shape_id": true, "shape_pt_lat": true, "shape_pt_lon": true, "shape_pt_sequence": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{"shape_dist_traveled"}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, points := range feed.Shapes {
		for _, sp := range points {
			if sp.DistTraveled != nil {
				checker.markNonDefault("shape_dist_traveled")
			}
			if checker.allFound() {
				break
			}
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("shapes.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	// Sort shape IDs for deterministic output order
	shapeIDs := make([]ShapeID, 0, len(feed.Shapes))
	for id := range feed.Shapes {
		shapeIDs = append(shapeIDs, id)
	}
	sort.Slice(shapeIDs, func(i, j int) bool {
		return string(shapeIDs[i]) < string(shapeIDs[j])
	})

	for _, shapeID := range shapeIDs {
		points := feed.Shapes[shapeID]
		for _, sp := range points {
			record := make([]string, len(activeCols))
			for i, col := range activeCols {
				record[i] = col.getter(sp)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Frequency) string
	}
	allCols := []colDef{
		{"trip_id", func(f *Frequency) string { return string(f.TripID) }},
		{"start_time", func(f *Frequency) string { return f.StartTime }},
		{"end_time", func(f *Frequency) string { return f.EndTime }},
		{"headway_secs", func(f *Frequency) string { return formatInt(f.HeadwaySecs) }},
		{"exact_times", func(f *Frequency) string { return formatOptionalInt(f.ExactTimes) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"trip_id": true, "start_time": true, "end_time": true, "headway_secs": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{"exact_times"}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, f := range feed.Frequencies {
		if f.ExactTimes != 0 {
			checker.markNonDefault("exact_times")
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("frequencies.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, f := range feed.Frequencies {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(f)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Transfer) string
	}
	allCols := []colDef{
		{"from_stop_id", func(t *Transfer) string { return string(t.FromStopID) }},
		{"to_stop_id", func(t *Transfer) string { return string(t.ToStopID) }},
		{"transfer_type", func(t *Transfer) string { return formatOptionalInt(t.TransferType) }},
		{"min_transfer_time", func(t *Transfer) string { return formatOptionalInt(t.MinTransferTime) }},
		{"from_route_id", func(t *Transfer) string { return string(t.FromRouteID) }},
		{"to_route_id", func(t *Transfer) string { return string(t.ToRouteID) }},
		{"from_trip_id", func(t *Transfer) string { return string(t.FromTripID) }},
		{"to_trip_id", func(t *Transfer) string { return string(t.ToTripID) }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"from_stop_id": true, "to_stop_id": true,
	}

	// Filter columns: include if required OR present in source data
	// Match Java behavior: include columns if they were in any source feed, even if all values are default
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("transfers.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, t := range feed.Transfers {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(t)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*FareAttribute) string
	}
	allCols := []colDef{
		{"fare_id", func(fa *FareAttribute) string { return string(fa.FareID) }},
		{"price", func(fa *FareAttribute) string { return strconv.FormatFloat(fa.Price, 'f', 6, 64) }},
		{"currency_type", func(fa *FareAttribute) string { return fa.CurrencyType }},
		{"payment_method", func(fa *FareAttribute) string { return formatInt(fa.PaymentMethod) }}, // Always output, 0 is valid
		{"transfers", func(fa *FareAttribute) string { return formatOptionalInt(fa.Transfers) }},  // Empty means unlimited
		{"agency_id", func(fa *FareAttribute) string { return string(fa.AgencyID) }},
		{"transfer_duration", func(fa *FareAttribute) string { return formatOptionalInt(fa.TransferDuration) }},
		{"youth_price", func(fa *FareAttribute) string { return formatPriceFloat(fa.YouthPrice) }},   // Always output with 6 decimals
		{"senior_price", func(fa *FareAttribute) string { return formatPriceFloat(fa.SeniorPrice) }}, // Always output with 6 decimals
	}

	// Required columns are always included
	// Note: youth_price and senior_price are always included to match Java behavior
	requiredCols := map[string]bool{
		"fare_id": true, "price": true, "currency_type": true,
		"youth_price": true, "senior_price": true,
	}

	// Filter columns: include if required OR present in source data
	// Match Java behavior: include columns if they were in any source feed, even if all values are default
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("fare_attributes.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, fa := range feed.FareAttributes {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(fa)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*FareRule) string
	}
	allCols := []colDef{
		{"fare_id", func(fr *FareRule) string { return string(fr.FareID) }},
		{"route_id", func(fr *FareRule) string { return string(fr.RouteID) }},
		{"origin_id", func(fr *FareRule) string { return fr.OriginID }},
		{"destination_id", func(fr *FareRule) string { return fr.DestinationID }},
		{"contains_id", func(fr *FareRule) string { return fr.ContainsID }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("fare_rules.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, fr := range feed.FareRules {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(fr)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*FeedInfo) string
	}
	allCols := []colDef{
		{"feed_publisher_name", func(fi *FeedInfo) string { return fi.PublisherName }},
		{"feed_publisher_url", func(fi *FeedInfo) string { return fi.PublisherURL }},
		{"feed_lang", func(fi *FeedInfo) string { return fi.Lang }},
		{"default_lang", func(fi *FeedInfo) string { return fi.DefaultLang }},
		{"feed_start_date", func(fi *FeedInfo) string { return fi.StartDate }},
		{"feed_end_date", func(fi *FeedInfo) string { return fi.EndDate }},
		{"feed_version", func(fi *FeedInfo) string { return fi.Version }},
		{"feed_contact_email", func(fi *FeedInfo) string { return fi.ContactEmail }},
		{"feed_contact_url", func(fi *FeedInfo) string { return fi.ContactURL }},
		{"feed_id", func(fi *FeedInfo) string { return fi.FeedID }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("feed_info.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	// Collect and sort feed_ids for consistent output
	ids := make([]string, 0, len(feed.FeedInfos))
	for id := range feed.FeedInfos {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		// Sort numerically if possible, else lexicographically
		ni, ei := strconv.Atoi(ids[i])
		nj, ej := strconv.Atoi(ids[j])
		if ei == nil && ej == nil {
			return ni < nj
		}
		return ids[i] < ids[j]
	})

	// Write all FeedInfo entries
	for _, id := range ids {
		fi := feed.FeedInfos[id]
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(fi)
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Area) string
	}
	allCols := []colDef{
		{"area_id", func(a *Area) string { return string(a.ID) }},
		{"area_name", func(a *Area) string { return a.Name }},
	}

	// Filter to only columns present in source data
	var activeCols []colDef
	for _, col := range allCols {
		if feed.HasColumn("areas.txt", col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, a := range feed.Areas {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(a)
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

	// Define all possible columns in order, with their getters
	type colDef struct {
		name   string
		getter func(*Pathway) string
	}
	allCols := []colDef{
		{"pathway_id", func(p *Pathway) string { return p.ID }},
		{"from_stop_id", func(p *Pathway) string { return string(p.FromStopID) }},
		{"to_stop_id", func(p *Pathway) string { return string(p.ToStopID) }},
		{"pathway_mode", func(p *Pathway) string { return formatInt(p.PathwayMode) }},
		{"is_bidirectional", func(p *Pathway) string { return formatInt(p.IsBidirectional) }},
		{"length", func(p *Pathway) string { return formatFloat(p.Length) }},
		{"traversal_time", func(p *Pathway) string { return formatOptionalInt(p.TraversalTime) }},
		{"stair_count", func(p *Pathway) string { return formatOptionalInt(p.StairCount) }},
		{"max_slope", func(p *Pathway) string { return formatFloat(p.MaxSlope) }},
		{"min_width", func(p *Pathway) string { return formatFloat(p.MinWidth) }},
		{"signposted_as", func(p *Pathway) string { return p.SignpostedAs }},
		{"reversed_signposted_as", func(p *Pathway) string { return p.ReversedSignpostedAs }},
	}

	// Required columns are always included
	requiredCols := map[string]bool{
		"pathway_id": true, "from_stop_id": true, "to_stop_id": true,
		"pathway_mode": true, "is_bidirectional": true,
	}

	// Optional columns: only include if at least one row has non-default value
	optionalCols := []string{
		"length", "traversal_time", "stair_count", "max_slope", "min_width",
		"signposted_as", "reversed_signposted_as",
	}
	checker := newColumnChecker(optionalCols)

	// Pre-scan to find which optional columns have non-default values
	for _, p := range feed.Pathways {
		if p.Length != 0.0 {
			checker.markNonDefault("length")
		}
		if p.TraversalTime != 0 {
			checker.markNonDefault("traversal_time")
		}
		if p.StairCount != 0 {
			checker.markNonDefault("stair_count")
		}
		if p.MaxSlope != 0.0 {
			checker.markNonDefault("max_slope")
		}
		if p.MinWidth != 0.0 {
			checker.markNonDefault("min_width")
		}
		if p.SignpostedAs != "" {
			checker.markNonDefault("signposted_as")
		}
		if p.ReversedSignpostedAs != "" {
			checker.markNonDefault("reversed_signposted_as")
		}
		if checker.allFound() {
			break
		}
	}

	// Filter columns: include if required OR (in source AND has non-default value)
	var activeCols []colDef
	for _, col := range allCols {
		if requiredCols[col.name] {
			activeCols = append(activeCols, col)
		} else if feed.HasColumn("pathways.txt", col.name) && checker.hasNonDefaultValue(col.name) {
			activeCols = append(activeCols, col)
		}
	}

	// Build header from active columns
	header := make([]string, len(activeCols))
	for i, col := range activeCols {
		header[i] = col.name
	}
	if err := csvw.WriteHeader(header); err != nil {
		return err
	}

	for _, p := range feed.Pathways {
		record := make([]string, len(activeCols))
		for i, col := range activeCols {
			record[i] = col.getter(p)
		}
		if err := csvw.WriteRecord(record); err != nil {
			return err
		}
	}

	return csvw.Flush()
}
