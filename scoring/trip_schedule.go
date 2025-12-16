package scoring

import (
	"strconv"
	"strings"

	"github.com/aaronbrethorst/gtfs-merge-go/gtfs"
	"github.com/aaronbrethorst/gtfs-merge-go/strategy"
)

// TripScheduleOverlapScorer scores trips by schedule similarity.
// Computes overlap of time windows [first_stop_departure, last_stop_arrival].
// Uses interval overlap formula: (overlap/interval_a + overlap/interval_b) / 2
type TripScheduleOverlapScorer struct{}

// Score returns a similarity score based on schedule overlap.
func (s *TripScheduleOverlapScorer) Score(ctx *strategy.MergeContext, source, target *gtfs.Trip) float64 {
	sourceStart, sourceEnd := getTripTimeWindow(ctx.Source, source.ID)
	targetStart, targetEnd := getTripTimeWindow(ctx.Target, target.ID)

	// Convert to float64 for IntervalOverlapScore
	return IntervalOverlapScore(
		float64(sourceStart), float64(sourceEnd),
		float64(targetStart), float64(targetEnd),
	)
}

// getTripTimeWindow returns the start and end times for a trip in seconds.
// Start is the first stop's departure time, end is the last stop's arrival time.
// Returns (0, 0) if trip has no stop times.
func getTripTimeWindow(feed *gtfs.Feed, tripID gtfs.TripID) (start, end int) {
	var stopTimes []*gtfs.StopTime
	for _, st := range feed.StopTimes {
		if st.TripID == tripID {
			stopTimes = append(stopTimes, st)
		}
	}

	if len(stopTimes) == 0 {
		return 0, 0
	}

	// Find first and last stop by sequence
	var firstStop, lastStop *gtfs.StopTime
	for _, st := range stopTimes {
		if firstStop == nil || st.StopSequence < firstStop.StopSequence {
			firstStop = st
		}
		if lastStop == nil || st.StopSequence > lastStop.StopSequence {
			lastStop = st
		}
	}

	// Use departure time for start, arrival time for end
	start = parseGTFSTime(firstStop.DepartureTime)
	end = parseGTFSTime(lastStop.ArrivalTime)

	return start, end
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS) to seconds since midnight.
// GTFS times can exceed 24:00:00 for overnight trips.
// Returns 0 for invalid or empty times.
func parseGTFSTime(timeStr string) int {
	if timeStr == "" {
		return 0
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}

	return hours*3600 + minutes*60 + seconds
}
