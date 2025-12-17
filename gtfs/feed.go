package gtfs

// Feed represents a complete GTFS feed
type Feed struct {
	Agencies       map[AgencyID]*Agency
	Stops          map[StopID]*Stop
	Routes         map[RouteID]*Route
	Trips          map[TripID]*Trip
	StopTimes      []*StopTime // Keyed by TripID+Sequence
	Calendars      map[ServiceID]*Calendar
	CalendarDates  map[ServiceID][]*CalendarDate
	Shapes         map[ShapeID][]*ShapePoint
	Frequencies    []*Frequency
	Transfers      []*Transfer
	FareAttributes map[FareID]*FareAttribute
	FareRules      []*FareRule
	FeedInfos      map[string]*FeedInfo // keyed by feed_id
	Areas          map[AreaID]*Area
	Pathways       []*Pathway

	// ColumnSets tracks which columns were present in each file when reading.
	// Key is the filename (e.g., "stop_times.txt"), value is set of column names.
	// Used during writing to only output columns that were present in source data.
	ColumnSets map[string]map[string]bool
}

// NewFeed creates an empty feed with all maps and slices initialized
func NewFeed() *Feed {
	return &Feed{
		Agencies:       make(map[AgencyID]*Agency),
		Stops:          make(map[StopID]*Stop),
		Routes:         make(map[RouteID]*Route),
		Trips:          make(map[TripID]*Trip),
		StopTimes:      make([]*StopTime, 0),
		Calendars:      make(map[ServiceID]*Calendar),
		CalendarDates:  make(map[ServiceID][]*CalendarDate),
		Shapes:         make(map[ShapeID][]*ShapePoint),
		Frequencies:    make([]*Frequency, 0),
		Transfers:      make([]*Transfer, 0),
		FareAttributes: make(map[FareID]*FareAttribute),
		FareRules:      make([]*FareRule, 0),
		FeedInfos:      make(map[string]*FeedInfo),
		Areas:          make(map[AreaID]*Area),
		Pathways:       make([]*Pathway, 0),
		ColumnSets:     make(map[string]map[string]bool),
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
