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
	FeedInfo       *FeedInfo
	Areas          map[AreaID]*Area
	Pathways       []*Pathway
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
		FeedInfo:       nil,
		Areas:          make(map[AreaID]*Area),
		Pathways:       make([]*Pathway, 0),
	}
}
