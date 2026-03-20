package postprocessing

//sanitizes and processes JSON from python microservice

import (
	"context"
	"fmt"

	edges "github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"golang.org/x/sync/errgroup"
)

func ProcessRouteResponse(ppinput PostProcessorInput) (Itinerary, error) {
	output := ppinput.SolverOutput
	input := ppinput.SolverInput
	addressmap := ppinput.StopMap
	transittypematrix := ppinput.TransitTypeMatrix
	transitcostmatrix := ppinput.TransitCostMatrix

	entries := make([]ItineraryEntry, len(output.Route))

	//check if no solution was found
	if output.HasSolution == false {
		return Itinerary{}, fmt.Errorf("no solution found for given input, params are trash")
	} //TODO, OUTPUT failure reason TO FRONTEND (forgot if this is even possible xd)

	//get subway edges and get legs
	subwayLegs := make([][]Leg, len(output.Route))
	subwayGroup, context := errgroup.WithContext(context.Background())

	for i := 0; i < len(output.Route)-1; i++ {
		current := output.Route[i]
		next := output.Route[i+1]
		fromIdx := current.NodeIndex
		toIdx := next.NodeIndex

		if transittypematrix[fromIdx][toIdx] != Subway {
			continue
		}
		originAddr := addressmap[fromIdx]
		destAddr := addressmap[toIdx]

		//concurrently get legs
		subwayGroup.Go(func() error {
			//hardcode google for now
			legs, err := GetSubwayLegs(edges.GoogleMaps{}, context, originAddr, destAddr)
			if err != nil {
				legs = []Leg{{
					TransportTypes: Subway,
					TravelTimes:    input.TravelTimeMatrix[fromIdx][toIdx],
					TransitCosts:   transitcostmatrix[fromIdx][toIdx],
				}}
			}
			subwayLegs[i] = legs
			return nil
		})
	}

	if err := subwayGroup.Wait(); err != nil {
		return Itinerary{}, fmt.Errorf("failed to fetch subway legs: %w", err)
	}

	//loop over items and make itinerary
	for i, routeEntry := range output.Route {
		node := input.Nodes[routeEntry.NodeIndex]
		nodeAddress := addressmap[routeEntry.NodeIndex]

		var legs []Leg

		if i < len(output.Route)-1 {
			if subwayLegs[i] != nil {
				legs = subwayLegs[i]
			} else {
				fromIdx := routeEntry.NodeIndex
				toIdx := output.Route[i+1].NodeIndex
				legs = []Leg{{
					TransportTypes: transittypematrix[fromIdx][toIdx],
					TravelTimes:    input.TravelTimeMatrix[fromIdx][toIdx],
					TransitCosts:   transitcostmatrix[fromIdx][toIdx],
				}}
			}
		}

		entries[i] = ItineraryEntry{
			Name:                    node.Name,
			Address:                 nodeAddress,
			ArrivalTimeInMinutes:    routeEntry.ArrivalTimeInMinutes,
			DepartureTimeInMinutes:  routeEntry.DepartureTimeInMinutes,
			DurationAtStopInMinutes: routeEntry.DepartureTimeInMinutes - routeEntry.ArrivalTimeInMinutes,
			Legs:                    legs,
		}
	}

	return Itinerary{ //return big boy official JSON
		Entries:                 entries,
		DroppedStops:            ProcessDroppedStops(addressmap, output.DroppedStops),
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

func GetSubwayLegs(provider edges.SubwayLegProvider, ctx context.Context, origin Address, destination Address) ([]Leg, error) {
	return provider.AcquireSubwayLegs(origin, destination)
}
