// Package gtfs provides data structures and I/O for GTFS (General Transit Feed Specification) feeds.
package gtfs

// AgencyID is a unique identifier for an agency
type AgencyID string

// StopID is a unique identifier for a stop
type StopID string

// RouteID is a unique identifier for a route
type RouteID string

// TripID is a unique identifier for a trip
type TripID string

// ServiceID is a unique identifier for a service calendar
type ServiceID string

// ShapeID is a unique identifier for a shape
type ShapeID string

// FareID is a unique identifier for a fare attribute
type FareID string

// AreaID is a unique identifier for an area
type AreaID string

// Agency represents a transit agency (agency.txt)
type Agency struct {
	ID       AgencyID
	Name     string
	URL      string
	Timezone string
	Lang     string
	Phone    string
	FareURL  string
	Email    string
}

// Stop represents a stop location (stops.txt)
type Stop struct {
	ID                 StopID
	Code               string
	Name               string
	Desc               string
	Lat                float64
	Lon                float64
	ZoneID             string
	URL                string
	LocationType       int
	ParentStation      StopID
	Timezone           string
	WheelchairBoarding int
	LevelID            string
	PlatformCode       string
}

// Route represents a transit route (routes.txt)
type Route struct {
	ID                RouteID
	AgencyID          AgencyID
	ShortName         string
	LongName          string
	Desc              string
	Type              int
	URL               string
	Color             string
	TextColor         string
	SortOrder         int
	ContinuousPickup  int
	ContinuousDropOff int
}

// Trip represents a trip (trips.txt)
type Trip struct {
	ID                   TripID
	RouteID              RouteID
	ServiceID            ServiceID
	Headsign             string
	ShortName            string
	DirectionID          *int // Pointer to distinguish "not set" (nil) from "set to 0"
	BlockID              string
	ShapeID              ShapeID
	WheelchairAccessible int
	BikesAllowed         int
}

// StopTime represents a stop time (stop_times.txt)
type StopTime struct {
	TripID            TripID
	ArrivalTime       string // HH:MM:SS format, can exceed 24:00:00
	DepartureTime     string
	StopID            StopID
	StopSequence      int
	StopHeadsign      string
	PickupType        int
	DropOffType       int
	ContinuousPickup  int
	ContinuousDropOff int
	ShapeDistTraveled float64
	Timepoint         int
}

// Calendar represents a service calendar (calendar.txt)
type Calendar struct {
	ServiceID ServiceID
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate string // YYYYMMDD format
	EndDate   string
}

// CalendarDate represents a calendar exception (calendar_dates.txt)
type CalendarDate struct {
	ServiceID     ServiceID
	Date          string // YYYYMMDD format
	ExceptionType int    // 1=added, 2=removed
}

// ShapePoint represents a point in a shape (shapes.txt)
type ShapePoint struct {
	ShapeID      ShapeID
	Lat          float64
	Lon          float64
	Sequence     int
	DistTraveled float64
}

// Frequency represents frequency-based service (frequencies.txt)
type Frequency struct {
	TripID      TripID
	StartTime   string
	EndTime     string
	HeadwaySecs int
	ExactTimes  int
}

// Transfer represents a transfer rule (transfers.txt)
type Transfer struct {
	FromStopID      StopID
	ToStopID        StopID
	TransferType    int
	MinTransferTime int
}

// FareAttribute represents fare pricing (fare_attributes.txt)
type FareAttribute struct {
	FareID           FareID
	Price            float64
	CurrencyType     string
	PaymentMethod    int
	Transfers        int
	AgencyID         AgencyID
	TransferDuration int
}

// FareRule represents fare rules (fare_rules.txt)
type FareRule struct {
	FareID        FareID
	RouteID       RouteID
	OriginID      string
	DestinationID string
	ContainsID    string
}

// FeedInfo represents feed metadata (feed_info.txt)
type FeedInfo struct {
	PublisherName string
	PublisherURL  string
	Lang          string
	DefaultLang   string
	StartDate     string
	EndDate       string
	Version       string
	ContactEmail  string
	ContactURL    string
	FeedID        string
}

// Area represents a geographic area (areas.txt)
type Area struct {
	ID   AreaID
	Name string
}

// Pathway represents a station pathway (pathways.txt)
type Pathway struct {
	ID                   string
	FromStopID           StopID
	ToStopID             StopID
	PathwayMode          int
	IsBidirectional      int
	Length               float64
	TraversalTime        int
	StairCount           int
	MaxSlope             float64
	MinWidth             float64
	SignpostedAs         string
	ReversedSignpostedAs string
}
