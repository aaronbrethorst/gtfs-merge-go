package merge

import (
	"errors"
	"fmt"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
)

// ErrNoInputFeeds indicates no input feeds were provided
var ErrNoInputFeeds = errors.New("at least one input feed is required")

// Merger orchestrates the merging of multiple GTFS feeds
type Merger struct {
	debug bool
}

// New creates a new Merger with default strategies
func New(opts ...Option) *Merger {
	m := &Merger{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// MergeFiles merges multiple GTFS files into one output file.
// IMPORTANT: Input feeds are processed in REVERSE order (newest/last first).
// This ensures entities from newer feeds are added first and older duplicates
// are potentially dropped.
func (m *Merger) MergeFiles(inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 {
		return ErrNoInputFeeds
	}

	// Read all feeds
	feeds := make([]*gtfs.Feed, 0, len(inputPaths))
	for _, path := range inputPaths {
		feed, err := gtfs.ReadFromPath(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		feeds = append(feeds, feed)
	}

	// Merge feeds
	merged, err := m.MergeFeeds(feeds)
	if err != nil {
		return err
	}

	// Write output
	return gtfs.WriteToPath(merged, outputPath)
}

// MergeFeeds merges multiple Feed objects into a single Feed.
// IMPORTANT: Feeds are processed in REVERSE order (last element first).
func (m *Merger) MergeFeeds(feeds []*gtfs.Feed) (*gtfs.Feed, error) {
	if len(feeds) == 0 {
		return nil, ErrNoInputFeeds
	}

	// Start with an empty target feed
	target := gtfs.NewFeed()

	// Process feeds in reverse order (last feed first, which gets no prefix)
	for i := len(feeds) - 1; i >= 0; i-- {
		feedIndex := len(feeds) - 1 - i // 0 for last feed, 1 for second-to-last, etc.
		prefix := GetPrefixForIndex(feedIndex)

		ctx := NewMergeContext(feeds[i], target)
		ctx.Prefix = prefix

		if err := m.mergeFeed(ctx); err != nil {
			return nil, fmt.Errorf("merging feed %d: %w", i, err)
		}
	}

	return target, nil
}

// mergeFeed merges a single source feed into the target
func (m *Merger) mergeFeed(ctx *MergeContext) error {
	// Merge entities in dependency order:
	// 1. Agencies (no dependencies)
	m.mergeAgencies(ctx)

	// 2. Areas (no dependencies)
	m.mergeAreas(ctx)

	// 3. Stops (references: parent_station)
	m.mergeStops(ctx)

	// 4. Service Calendars (no dependencies)
	m.mergeCalendars(ctx)
	m.mergeCalendarDates(ctx)

	// 5. Routes (references: agency_id)
	m.mergeRoutes(ctx)

	// 6. Shapes (no dependencies)
	m.mergeShapes(ctx)

	// 7. Trips (references: route_id, service_id, shape_id)
	m.mergeTrips(ctx)

	// 8. Stop Times (references: trip_id, stop_id)
	m.mergeStopTimes(ctx)

	// 9. Frequencies (references: trip_id)
	m.mergeFrequencies(ctx)

	// 10. Transfers (references: from_stop_id, to_stop_id)
	m.mergeTransfers(ctx)

	// 11. Pathways (references: from_stop_id, to_stop_id)
	m.mergePathways(ctx)

	// 12. Fare Attributes (references: agency_id)
	m.mergeFareAttributes(ctx)

	// 13. Fare Rules (references: fare_id, route_id)
	m.mergeFareRules(ctx)

	// 14. Feed Info (no dependencies)
	m.mergeFeedInfo(ctx)

	return nil
}

// mergeAgencies merges agencies from source to target
func (m *Merger) mergeAgencies(ctx *MergeContext) {
	for _, agency := range ctx.Source.Agencies {
		newID := gtfs.AgencyID(ctx.Prefix + string(agency.ID))
		ctx.AgencyIDMapping[agency.ID] = newID

		newAgency := &gtfs.Agency{
			ID:       newID,
			Name:     agency.Name,
			URL:      agency.URL,
			Timezone: agency.Timezone,
			Lang:     agency.Lang,
			Phone:    agency.Phone,
			FareURL:  agency.FareURL,
			Email:    agency.Email,
		}
		ctx.Target.Agencies[newID] = newAgency
	}
}

// mergeAreas merges areas from source to target
func (m *Merger) mergeAreas(ctx *MergeContext) {
	for _, area := range ctx.Source.Areas {
		newID := gtfs.AreaID(ctx.Prefix + string(area.ID))
		ctx.AreaIDMapping[area.ID] = newID

		newArea := &gtfs.Area{
			ID:   newID,
			Name: area.Name,
		}
		ctx.Target.Areas[newID] = newArea
	}
}

// mergeStops merges stops from source to target
func (m *Merger) mergeStops(ctx *MergeContext) {
	for _, stop := range ctx.Source.Stops {
		newID := gtfs.StopID(ctx.Prefix + string(stop.ID))
		ctx.StopIDMapping[stop.ID] = newID

		// Handle parent_station reference
		parentStation := stop.ParentStation
		if parentStation != "" {
			if mappedParent, ok := ctx.StopIDMapping[parentStation]; ok {
				parentStation = mappedParent
			} else {
				parentStation = gtfs.StopID(ctx.Prefix + string(parentStation))
			}
		}

		newStop := &gtfs.Stop{
			ID:                 newID,
			Code:               stop.Code,
			Name:               stop.Name,
			Desc:               stop.Desc,
			Lat:                stop.Lat,
			Lon:                stop.Lon,
			ZoneID:             stop.ZoneID,
			URL:                stop.URL,
			LocationType:       stop.LocationType,
			ParentStation:      parentStation,
			Timezone:           stop.Timezone,
			WheelchairBoarding: stop.WheelchairBoarding,
			LevelID:            stop.LevelID,
			PlatformCode:       stop.PlatformCode,
		}
		ctx.Target.Stops[newID] = newStop
	}
}

// mergeCalendars merges calendars from source to target
func (m *Merger) mergeCalendars(ctx *MergeContext) {
	for _, cal := range ctx.Source.Calendars {
		newID := gtfs.ServiceID(ctx.Prefix + string(cal.ServiceID))
		ctx.ServiceIDMapping[cal.ServiceID] = newID

		newCal := &gtfs.Calendar{
			ServiceID: newID,
			Monday:    cal.Monday,
			Tuesday:   cal.Tuesday,
			Wednesday: cal.Wednesday,
			Thursday:  cal.Thursday,
			Friday:    cal.Friday,
			Saturday:  cal.Saturday,
			Sunday:    cal.Sunday,
			StartDate: cal.StartDate,
			EndDate:   cal.EndDate,
		}
		ctx.Target.Calendars[newID] = newCal
	}
}

// mergeCalendarDates merges calendar dates from source to target
func (m *Merger) mergeCalendarDates(ctx *MergeContext) {
	for serviceID, dates := range ctx.Source.CalendarDates {
		newServiceID := ctx.ServiceIDMapping[serviceID]
		if newServiceID == "" {
			// Service may only be defined in calendar_dates, not calendar
			newServiceID = gtfs.ServiceID(ctx.Prefix + string(serviceID))
			ctx.ServiceIDMapping[serviceID] = newServiceID
		}

		for _, date := range dates {
			newDate := &gtfs.CalendarDate{
				ServiceID:     newServiceID,
				Date:          date.Date,
				ExceptionType: date.ExceptionType,
			}
			ctx.Target.CalendarDates[newServiceID] = append(ctx.Target.CalendarDates[newServiceID], newDate)
		}
	}
}

// mergeRoutes merges routes from source to target
func (m *Merger) mergeRoutes(ctx *MergeContext) {
	for _, route := range ctx.Source.Routes {
		newID := gtfs.RouteID(ctx.Prefix + string(route.ID))
		ctx.RouteIDMapping[route.ID] = newID

		// Map agency reference
		agencyID := route.AgencyID
		if agencyID != "" {
			if mappedAgency, ok := ctx.AgencyIDMapping[agencyID]; ok {
				agencyID = mappedAgency
			}
		}

		newRoute := &gtfs.Route{
			ID:                newID,
			AgencyID:          agencyID,
			ShortName:         route.ShortName,
			LongName:          route.LongName,
			Desc:              route.Desc,
			Type:              route.Type,
			URL:               route.URL,
			Color:             route.Color,
			TextColor:         route.TextColor,
			SortOrder:         route.SortOrder,
			ContinuousPickup:  route.ContinuousPickup,
			ContinuousDropOff: route.ContinuousDropOff,
		}
		ctx.Target.Routes[newID] = newRoute
	}
}

// mergeShapes merges shapes from source to target
func (m *Merger) mergeShapes(ctx *MergeContext) {
	for shapeID, points := range ctx.Source.Shapes {
		newID := gtfs.ShapeID(ctx.Prefix + string(shapeID))
		ctx.ShapeIDMapping[shapeID] = newID

		for _, point := range points {
			newPoint := &gtfs.ShapePoint{
				ShapeID:      newID,
				Lat:          point.Lat,
				Lon:          point.Lon,
				Sequence:     point.Sequence,
				DistTraveled: point.DistTraveled,
			}
			ctx.Target.Shapes[newID] = append(ctx.Target.Shapes[newID], newPoint)
		}
	}
}

// mergeTrips merges trips from source to target
func (m *Merger) mergeTrips(ctx *MergeContext) {
	for _, trip := range ctx.Source.Trips {
		newID := gtfs.TripID(ctx.Prefix + string(trip.ID))
		ctx.TripIDMapping[trip.ID] = newID

		// Map references
		routeID := trip.RouteID
		if mappedRoute, ok := ctx.RouteIDMapping[routeID]; ok {
			routeID = mappedRoute
		}

		serviceID := trip.ServiceID
		if mappedService, ok := ctx.ServiceIDMapping[serviceID]; ok {
			serviceID = mappedService
		}

		shapeID := trip.ShapeID
		if shapeID != "" {
			if mappedShape, ok := ctx.ShapeIDMapping[shapeID]; ok {
				shapeID = mappedShape
			}
		}

		newTrip := &gtfs.Trip{
			ID:                   newID,
			RouteID:              routeID,
			ServiceID:            serviceID,
			Headsign:             trip.Headsign,
			ShortName:            trip.ShortName,
			DirectionID:          trip.DirectionID,
			BlockID:              trip.BlockID,
			ShapeID:              shapeID,
			WheelchairAccessible: trip.WheelchairAccessible,
			BikesAllowed:         trip.BikesAllowed,
		}
		ctx.Target.Trips[newID] = newTrip
	}
}

// mergeStopTimes merges stop times from source to target
func (m *Merger) mergeStopTimes(ctx *MergeContext) {
	for _, st := range ctx.Source.StopTimes {
		// Map references
		tripID := st.TripID
		if mappedTrip, ok := ctx.TripIDMapping[tripID]; ok {
			tripID = mappedTrip
		}

		stopID := st.StopID
		if mappedStop, ok := ctx.StopIDMapping[stopID]; ok {
			stopID = mappedStop
		}

		newST := &gtfs.StopTime{
			TripID:            tripID,
			ArrivalTime:       st.ArrivalTime,
			DepartureTime:     st.DepartureTime,
			StopID:            stopID,
			StopSequence:      st.StopSequence,
			StopHeadsign:      st.StopHeadsign,
			PickupType:        st.PickupType,
			DropOffType:       st.DropOffType,
			ContinuousPickup:  st.ContinuousPickup,
			ContinuousDropOff: st.ContinuousDropOff,
			ShapeDistTraveled: st.ShapeDistTraveled,
			Timepoint:         st.Timepoint,
		}
		ctx.Target.StopTimes = append(ctx.Target.StopTimes, newST)
	}
}

// mergeFrequencies merges frequencies from source to target
func (m *Merger) mergeFrequencies(ctx *MergeContext) {
	for _, freq := range ctx.Source.Frequencies {
		// Map trip reference
		tripID := freq.TripID
		if mappedTrip, ok := ctx.TripIDMapping[tripID]; ok {
			tripID = mappedTrip
		}

		newFreq := &gtfs.Frequency{
			TripID:      tripID,
			StartTime:   freq.StartTime,
			EndTime:     freq.EndTime,
			HeadwaySecs: freq.HeadwaySecs,
			ExactTimes:  freq.ExactTimes,
		}
		ctx.Target.Frequencies = append(ctx.Target.Frequencies, newFreq)
	}
}

// mergeTransfers merges transfers from source to target
func (m *Merger) mergeTransfers(ctx *MergeContext) {
	for _, transfer := range ctx.Source.Transfers {
		// Map stop references
		fromStopID := transfer.FromStopID
		if mappedStop, ok := ctx.StopIDMapping[fromStopID]; ok {
			fromStopID = mappedStop
		}

		toStopID := transfer.ToStopID
		if mappedStop, ok := ctx.StopIDMapping[toStopID]; ok {
			toStopID = mappedStop
		}

		newTransfer := &gtfs.Transfer{
			FromStopID:      fromStopID,
			ToStopID:        toStopID,
			TransferType:    transfer.TransferType,
			MinTransferTime: transfer.MinTransferTime,
		}
		ctx.Target.Transfers = append(ctx.Target.Transfers, newTransfer)
	}
}

// mergePathways merges pathways from source to target
func (m *Merger) mergePathways(ctx *MergeContext) {
	for _, pathway := range ctx.Source.Pathways {
		// Map stop references
		fromStopID := pathway.FromStopID
		if mappedStop, ok := ctx.StopIDMapping[fromStopID]; ok {
			fromStopID = mappedStop
		}

		toStopID := pathway.ToStopID
		if mappedStop, ok := ctx.StopIDMapping[toStopID]; ok {
			toStopID = mappedStop
		}

		newPathway := &gtfs.Pathway{
			ID:                   ctx.Prefix + pathway.ID,
			FromStopID:           fromStopID,
			ToStopID:             toStopID,
			PathwayMode:          pathway.PathwayMode,
			IsBidirectional:      pathway.IsBidirectional,
			Length:               pathway.Length,
			TraversalTime:        pathway.TraversalTime,
			StairCount:           pathway.StairCount,
			MaxSlope:             pathway.MaxSlope,
			MinWidth:             pathway.MinWidth,
			SignpostedAs:         pathway.SignpostedAs,
			ReversedSignpostedAs: pathway.ReversedSignpostedAs,
		}
		ctx.Target.Pathways = append(ctx.Target.Pathways, newPathway)
	}
}

// mergeFareAttributes merges fare attributes from source to target
func (m *Merger) mergeFareAttributes(ctx *MergeContext) {
	for _, fare := range ctx.Source.FareAttributes {
		newID := gtfs.FareID(ctx.Prefix + string(fare.FareID))
		ctx.FareIDMapping[fare.FareID] = newID

		// Map agency reference
		agencyID := fare.AgencyID
		if agencyID != "" {
			if mappedAgency, ok := ctx.AgencyIDMapping[agencyID]; ok {
				agencyID = mappedAgency
			}
		}

		newFare := &gtfs.FareAttribute{
			FareID:           newID,
			Price:            fare.Price,
			CurrencyType:     fare.CurrencyType,
			PaymentMethod:    fare.PaymentMethod,
			Transfers:        fare.Transfers,
			AgencyID:         agencyID,
			TransferDuration: fare.TransferDuration,
		}
		ctx.Target.FareAttributes[newID] = newFare
	}
}

// mergeFareRules merges fare rules from source to target
func (m *Merger) mergeFareRules(ctx *MergeContext) {
	for _, rule := range ctx.Source.FareRules {
		// Map references
		fareID := rule.FareID
		if mappedFare, ok := ctx.FareIDMapping[fareID]; ok {
			fareID = mappedFare
		}

		routeID := rule.RouteID
		if routeID != "" {
			if mappedRoute, ok := ctx.RouteIDMapping[routeID]; ok {
				routeID = mappedRoute
			}
		}

		newRule := &gtfs.FareRule{
			FareID:        fareID,
			RouteID:       routeID,
			OriginID:      rule.OriginID,
			DestinationID: rule.DestinationID,
			ContainsID:    rule.ContainsID,
		}
		ctx.Target.FareRules = append(ctx.Target.FareRules, newRule)
	}
}

// mergeFeedInfo merges feed info from source to target
func (m *Merger) mergeFeedInfo(ctx *MergeContext) {
	if ctx.Source.FeedInfo != nil {
		// For now, just take the last feed info (first processed)
		// Future: merge versions and date ranges
		if ctx.Target.FeedInfo == nil {
			ctx.Target.FeedInfo = &gtfs.FeedInfo{
				PublisherName: ctx.Source.FeedInfo.PublisherName,
				PublisherURL:  ctx.Source.FeedInfo.PublisherURL,
				Lang:          ctx.Source.FeedInfo.Lang,
				DefaultLang:   ctx.Source.FeedInfo.DefaultLang,
				StartDate:     ctx.Source.FeedInfo.StartDate,
				EndDate:       ctx.Source.FeedInfo.EndDate,
				Version:       ctx.Source.FeedInfo.Version,
				ContactEmail:  ctx.Source.FeedInfo.ContactEmail,
				ContactURL:    ctx.Source.FeedInfo.ContactURL,
			}
		}
	}
}
