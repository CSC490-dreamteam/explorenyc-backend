package postprocessing

//sanitizes and processes JSON from python microservice

import (
	"fmt"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

func ProcessRouteResponse(ppinput PostProcessorInput) (Itinerary, error) {
	output := ppinput.SolverOutput
	input := ppinput.SolverInput
	stopmap := ppinput.StopMap
	transittypematrix := ppinput.TransitTypeMatrix
	transitcostmatrix := ppinput.TransitCostMatrix

	entries := make([]ItineraryEntry, len(output.Route))

	//check if no solution was found
	if output.HasSolution == false {
		return Itinerary{}, fmt.Errorf("no solution found for given input")
	} //TODO, OUTPUT failure reason TO FRONTEND (forgot if this is even possible xd)

	for i, routeEntry := range output.Route {
		node := input.Nodes[routeEntry.NodeIndex]
		nodeAddress := stopmap[routeEntry.NodeIndex]

		var travelTimeToNext int64
		var transitModeToNext TransitType
		var travelCostToNext int64

		if i < len(output.Route)-1 {
			nextEntry := output.Route[i+1]
			fromIdx := routeEntry.NodeIndex
			toIdx := nextEntry.NodeIndex
			travelTimeToNext = input.TravelTimeMatrix[fromIdx][toIdx]
			transitModeToNext = transittypematrix[fromIdx][toIdx]
			travelCostToNext = transitcostmatrix[fromIdx][toIdx]
		}

		newEntry := ItineraryEntry{
			Name:                    node.Name,
			Address:                 nodeAddress,
			ArrivalTimeInMinutes:    routeEntry.ArrivalTimeInMinutes,
			DepartureTimeInMinutes:  routeEntry.DepartureTimeInMinutes,
			DurationAtStopInMinutes: routeEntry.DepartureTimeInMinutes - routeEntry.ArrivalTimeInMinutes,
			TravelTimeToNextStop:    travelTimeToNext,
			TransportToNextStop:     transitModeToNext,
			TransitCost:             travelCostToNext,
		}
		entries[i] = newEntry
	}

	//MAKE STOPMAP THAT MAPS STOPS BY index to THEIR ADDRESS in preprocessor

	return Itinerary{
		Entries:                 entries,
		DroppedStops:            ProcessDroppedStops(stopmap, output.DroppedStops),
		TotalTransitCostInCents: output.TotalCostInCents,
		TotalTimeInMinutes:      output.TotalTimeInMinutes,
		TotalCostInCents:        output.TotalCostInCents + 69, //TODO ADD RECREATION COSTS
		StartTimeInMinutes:      entries[0].ArrivalTimeInMinutes,
		EndTimeInMinutes:        entries[len(entries)-1].DepartureTimeInMinutes,
	}, nil

}

// TODO DONT COUNT IF THEY ARE CA DNIDATE GROUP FROM NICKS AI
// TODO GET REASON FOR EACH DROPPED STOP AND PASS TO FRONTEND
func ProcessDroppedStops(stopmap map[int]Address, droppedStopIndices []int) []Address {
	droppedStops := make([]Address, 0)
	for _, index := range droppedStopIndices {
		if addr, exists := stopmap[index]; exists {
			droppedStops = append(droppedStops, addr)
		}
	}
	return droppedStops
}
