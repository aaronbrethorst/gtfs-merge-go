package gtfs

// Feed represents a complete GTFS feed
type Feed struct {
	Agencies       map[AgencyID]*Agency
	AgencyOrder    []AgencyID // Tracks insertion order for deterministic output
	Stops          map[StopID]*Stop
	StopOrder      []StopID // Tracks insertion order for deterministic output
	Routes         map[RouteID]*Route
	RouteOrder     []RouteID // Tracks insertion order for deterministic output
	Trips          map[TripID]*Trip
	TripOrder      []TripID // Tracks insertion order for deterministic output
	StopTimes      []*StopTime // Keyed by TripID+Sequence (already ordered)
	Calendars      map[ServiceID]*Calendar
	CalendarOrder  []ServiceID // Tracks insertion order for deterministic output
	CalendarDates  map[ServiceID][]*CalendarDate
	CalendarDateOrder []ServiceID // Tracks insertion order for deterministic output
	Shapes         map[ShapeID][]*ShapePoint
	ShapeOrder     []ShapeID // Tracks insertion order for deterministic output
	Frequencies    []*Frequency // Already ordered
	Transfers      []*Transfer  // Already ordered
	FareAttributes map[FareID]*FareAttribute
	FareAttrOrder  []FareID // Tracks insertion order for deterministic output
	FareRules      []*FareRule // Already ordered
	FeedInfos      map[string]*FeedInfo // keyed by feed_id
	FeedInfoOrder  []string // Tracks insertion order for deterministic output
	Areas          map[AreaID]*Area
	AreaOrder      []AreaID // Tracks insertion order for deterministic output
	Pathways       []*Pathway // Already ordered

	// ColumnSets tracks which columns were present in each file when reading.
	// Key is the filename (e.g., "stop_times.txt"), value is set of column names.
	// Used during writing to only output columns that were present in source data.
	ColumnSets map[string]map[string]bool
}

// NewFeed creates an empty feed with all maps and slices initialized
func NewFeed() *Feed {
	return &Feed{
		Agencies:          make(map[AgencyID]*Agency),
		AgencyOrder:       make([]AgencyID, 0),
		Stops:             make(map[StopID]*Stop),
		StopOrder:         make([]StopID, 0),
		Routes:            make(map[RouteID]*Route),
		RouteOrder:        make([]RouteID, 0),
		Trips:             make(map[TripID]*Trip),
		TripOrder:         make([]TripID, 0),
		StopTimes:         make([]*StopTime, 0),
		Calendars:         make(map[ServiceID]*Calendar),
		CalendarOrder:     make([]ServiceID, 0),
		CalendarDates:     make(map[ServiceID][]*CalendarDate),
		CalendarDateOrder: make([]ServiceID, 0),
		Shapes:            make(map[ShapeID][]*ShapePoint),
		ShapeOrder:        make([]ShapeID, 0),
		Frequencies:       make([]*Frequency, 0),
		Transfers:         make([]*Transfer, 0),
		FareAttributes:    make(map[FareID]*FareAttribute),
		FareAttrOrder:     make([]FareID, 0),
		FareRules:         make([]*FareRule, 0),
		FeedInfos:         make(map[string]*FeedInfo),
		FeedInfoOrder:     make([]string, 0),
		Areas:             make(map[AreaID]*Area),
		AreaOrder:         make([]AreaID, 0),
		Pathways:          make([]*Pathway, 0),
		ColumnSets:        make(map[string]map[string]bool),
	}
}

// AddColumnSet adds a set of columns for a given filename
func (f *Feed) AddColumnSet(filename string, columns []string) {
	if f.ColumnSets == nil {
		f.ColumnSets = make(map[string]map[string]bool)
	}
	colSet := make(map[string]bool)
	for _, col := range columns {
		colSet[col] = true
	}
	f.ColumnSets[filename] = colSet
}

// MergeColumnSets merges column sets from another feed using union.
// Columns present in ANY feed are kept (to match Java behavior).
// This ensures we don't lose data when feeds have different optional columns.
func (f *Feed) MergeColumnSets(other *Feed) {
	if other.ColumnSets == nil {
		return
	}
	if f.ColumnSets == nil {
		// First feed: copy all column sets
		f.ColumnSets = make(map[string]map[string]bool)
		for filename, cols := range other.ColumnSets {
			f.ColumnSets[filename] = make(map[string]bool)
			for col := range cols {
				f.ColumnSets[filename][col] = true
			}
		}
		return
	}
	// Subsequent feeds: union with existing columns
	for filename, otherCols := range other.ColumnSets {
		if existingCols, exists := f.ColumnSets[filename]; exists {
			// Add any columns from other feed that we don't have
			for col := range otherCols {
				existingCols[col] = true
			}
		} else {
			// File doesn't exist in target yet, add it
			f.ColumnSets[filename] = make(map[string]bool)
			for col := range otherCols {
				f.ColumnSets[filename][col] = true
			}
		}
	}
}

// HasColumn checks if a column was present for a given file
func (f *Feed) HasColumn(filename, column string) bool {
	if f.ColumnSets == nil {
		return true // Default to including column if no tracking
	}
	colSet, exists := f.ColumnSets[filename]
	if !exists {
		return true // Default to including column if file not tracked
	}
	return colSet[column]
}

// AddAgency adds an agency to both the map and order slice
func (f *Feed) AddAgency(a *Agency) {
	f.Agencies[a.ID] = a
	f.AgencyOrder = append(f.AgencyOrder, a.ID)
}

// AddStop adds a stop to both the map and order slice
func (f *Feed) AddStop(s *Stop) {
	f.Stops[s.ID] = s
	f.StopOrder = append(f.StopOrder, s.ID)
}

// AddRoute adds a route to both the map and order slice
func (f *Feed) AddRoute(r *Route) {
	f.Routes[r.ID] = r
	f.RouteOrder = append(f.RouteOrder, r.ID)
}

// AddTrip adds a trip to both the map and order slice
func (f *Feed) AddTrip(t *Trip) {
	f.Trips[t.ID] = t
	f.TripOrder = append(f.TripOrder, t.ID)
}

// AddCalendar adds a calendar to both the map and order slice
func (f *Feed) AddCalendar(c *Calendar) {
	f.Calendars[c.ServiceID] = c
	f.CalendarOrder = append(f.CalendarOrder, c.ServiceID)
}

// AddCalendarDate adds a calendar date to the map and tracks order for the service ID
func (f *Feed) AddCalendarDate(cd *CalendarDate) {
	// Track order only for first occurrence of this service_id
	if _, exists := f.CalendarDates[cd.ServiceID]; !exists {
		f.CalendarDateOrder = append(f.CalendarDateOrder, cd.ServiceID)
	}
	f.CalendarDates[cd.ServiceID] = append(f.CalendarDates[cd.ServiceID], cd)
}

// AddFareAttribute adds a fare attribute to both the map and order slice
func (f *Feed) AddFareAttribute(fa *FareAttribute) {
	f.FareAttributes[fa.FareID] = fa
	f.FareAttrOrder = append(f.FareAttrOrder, fa.FareID)
}

// AddFeedInfo adds a feed info to both the map and order slice
func (f *Feed) AddFeedInfo(fi *FeedInfo) {
	// Track order only for new entries
	if _, exists := f.FeedInfos[fi.FeedID]; !exists {
		f.FeedInfoOrder = append(f.FeedInfoOrder, fi.FeedID)
	}
	f.FeedInfos[fi.FeedID] = fi
}

// AddArea adds an area to both the map and order slice
func (f *Feed) AddArea(a *Area) {
	f.Areas[a.ID] = a
	f.AreaOrder = append(f.AreaOrder, a.ID)
}

// AddShape adds a shape point to the map and tracks order for the shape ID
func (f *Feed) AddShape(sp *ShapePoint) {
	// Track order only for first occurrence of this shape_id
	if _, exists := f.Shapes[sp.ShapeID]; !exists {
		f.ShapeOrder = append(f.ShapeOrder, sp.ShapeID)
	}
	f.Shapes[sp.ShapeID] = append(f.Shapes[sp.ShapeID], sp)
}

// SyncOrderSlices populates all order slices from their corresponding maps.
// This is useful for tests that add entries directly to maps without using AddX methods.
// Note: The order will be non-deterministic since Go maps don't preserve insertion order.
func (f *Feed) SyncOrderSlices() {
	// Agencies
	f.AgencyOrder = make([]AgencyID, 0, len(f.Agencies))
	for id := range f.Agencies {
		f.AgencyOrder = append(f.AgencyOrder, id)
	}

	// Stops
	f.StopOrder = make([]StopID, 0, len(f.Stops))
	for id := range f.Stops {
		f.StopOrder = append(f.StopOrder, id)
	}

	// Routes
	f.RouteOrder = make([]RouteID, 0, len(f.Routes))
	for id := range f.Routes {
		f.RouteOrder = append(f.RouteOrder, id)
	}

	// Trips
	f.TripOrder = make([]TripID, 0, len(f.Trips))
	for id := range f.Trips {
		f.TripOrder = append(f.TripOrder, id)
	}

	// Calendars
	f.CalendarOrder = make([]ServiceID, 0, len(f.Calendars))
	for id := range f.Calendars {
		f.CalendarOrder = append(f.CalendarOrder, id)
	}

	// CalendarDates
	f.CalendarDateOrder = make([]ServiceID, 0, len(f.CalendarDates))
	for id := range f.CalendarDates {
		f.CalendarDateOrder = append(f.CalendarDateOrder, id)
	}

	// Shapes
	f.ShapeOrder = make([]ShapeID, 0, len(f.Shapes))
	for id := range f.Shapes {
		f.ShapeOrder = append(f.ShapeOrder, id)
	}

	// FareAttributes
	f.FareAttrOrder = make([]FareID, 0, len(f.FareAttributes))
	for id := range f.FareAttributes {
		f.FareAttrOrder = append(f.FareAttrOrder, id)
	}

	// FeedInfos
	f.FeedInfoOrder = make([]string, 0, len(f.FeedInfos))
	for id := range f.FeedInfos {
		f.FeedInfoOrder = append(f.FeedInfoOrder, id)
	}

	// Areas
	f.AreaOrder = make([]AreaID, 0, len(f.Areas))
	for id := range f.Areas {
		f.AreaOrder = append(f.AreaOrder, id)
	}
}
