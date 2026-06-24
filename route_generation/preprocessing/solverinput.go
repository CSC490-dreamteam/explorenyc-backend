package preprocessing

import (
	"fmt"
	"strings"
	"time"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// BuildSolverInput converts a resolved itinerary request into the payload
// the route solver expects.
func BuildSolverInput(req ItineraryRequest, places AcquiredPlaces, matrices CombinedMatrices) SolverInput {
	dayStart := ParseTimeIntoMinutes(req.EntryTime)
	dayEnd := ParseTimeIntoMinutes(req.ExitTime)

	addrs := places.Places
	numPlaces := len(addrs)
	needsSeparateEnd := places.EndIndex > 0 // -1 = open ended, 0 = round trip, otherwise it's a real end node

	var solverNodes []SolverNode

	//starting location
	solverNodes = append(solverNodes, SolverNode{
		ID:                "0",
		Name:              addrs[0].PlaceName,
		Latitude:          addrs[0].Lat,
		Longitude:         addrs[0].Lon,
		DurationInMinutes: 0, // start point, no dwell time
		TimeWindowStart:   dayStart,
		TimeWindowEnd:     dayEnd,
		Priority:          Mandatory,
		DropPenalty:       0,
	})

	for i, stop := range req.Stops {
		var timeWindowStart, timeWindowEnd int

		//set time window for that specific node for when arrival time can be set
		if stop.TimePreference != nil {
			preferred := ParseTimeIntoMinutes(*stop.TimePreference)
			timeWindowStart = preferred - 30 //30 min before preferred arrival time
			timeWindowEnd = preferred - 5    //5 min before actual time
		} else {
			// no preference = anytime during the day is fine
			timeWindowStart = dayStart
			timeWindowEnd = dayEnd
		}

		//setup priority and drop penalty based on whether stop is mandatory or not
		prio := WantToSee
		if stop.Mandatory {
			prio = Mandatory
		}

		var dropPenalty int
		switch prio {
		case Mandatory:
			dropPenalty = 0 //undroppable
		case WantToSee:
			dropPenalty = 8 //will be dropped if it screws over like 4-6 optional places?
		case Optional:
			dropPenalty = 2
		}

		solverNodes = append(solverNodes, SolverNode{ //we do i+1 to account for the start node we added at the beginning
			ID:                fmt.Sprintf("%d", i+1),
			Name:              addrs[i+1].PlaceName,
			Latitude:          addrs[i+1].Lat,
			Longitude:         addrs[i+1].Lon,
			DurationInMinutes: stop.Duration, //dwell time at the stop
			TimeWindowStart:   timeWindowStart,
			TimeWindowEnd:     timeWindowEnd,
			Priority:          prio,
			DropPenalty:       dropPenalty,
		})
	}

	//add end location (if one is set)
	if needsSeparateEnd {
		solverNodes = append(solverNodes, SolverNode{
			ID:                fmt.Sprintf("%d", numPlaces-1),
			Name:              addrs[numPlaces-1].PlaceName,
			Latitude:          addrs[numPlaces-1].Lat,
			Longitude:         addrs[numPlaces-1].Lon,
			DurationInMinutes: 0,
			TimeWindowStart:   dayStart,
			TimeWindowEnd:     dayEnd,
			Priority:          Mandatory,
			DropPenalty:       0,
		})
	}
	//roundtrip edgecase is covered by EndIndex already being 0

	return SolverInput{
		TripName:              req.TripName,
		Date:                  req.Date,
		Nodes:                 solverNodes,
		StartIndex:            0,
		EndIndex:              places.EndIndex,
		DayStartTimeInMinutes: dayStart,
		DayEndTimeInMinutes:   dayEnd,
		BudgetInCents:         5000, //PLACEHOLDER, $50 budget for transit costs
		TravelTimeMatrix:      matrices.TimeMinutes,
		CostMatrix:            matrices.CostCents,
		CandidateGroups:       []CandidateGroup{}, //TODO wait on nick
		RouteVariant:          Balanced,
		//empty so python doesnt get mad
		Precedences:   [][2]int{},
		ForcedEdges:   [][2]int{},
		ExcludedStops: []int{},
	}
}

// ParseTimeIntoMinutes turns "09:00 AM" into 540 (minutes from midnight).
func ParseTimeIntoMinutes(timeStr string) int {
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.ToUpper(timeStr)
	timeStr = strings.ReplaceAll(timeStr, "AM", " AM")
	timeStr = strings.ReplaceAll(timeStr, "PM", " PM")
	timeStr = strings.TrimSpace(timeStr)
	layouts := []string{"3:04 PM", "15:04", "15:04:05"}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return int(t.Hour()*60 + t.Minute())
		}
	}

	fmt.Printf("invalid time format: %s", timeStr)
	return -1
}
