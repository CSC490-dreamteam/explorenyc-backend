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

	//get transit edges and legs
	subwayLegs := make([][]Leg, len(output.Route))
	walkingLegs := make([][]Leg, len(output.Route))
	carLegs := make([][]Leg, len(output.Route))
	group, context := errgroup.WithContext(context.Background())

	for i := 0; i < len(output.Route)-1; i++ {
		current := output.Route[i]
		next := output.Route[i+1]
		fromIdx := current.NodeIndex
		toIdx := next.NodeIndex

		transitType := transittypematrix[fromIdx][toIdx]
		originAddr := addressmap[fromIdx]
		destAddr := addressmap[toIdx]

		//concurrently get legs for all transit types
		switch transitType {
		case Subway:
			group.Go(func() error {
				legs, err := GetSubwayLegs(edges.GoogleMaps{}, context, originAddr, destAddr)
				if err != nil {
					legs = []Leg{{
						TransportType: Subway,
						TravelTimes:   input.TravelTimeMatrix[fromIdx][toIdx],
						TransitCosts:  transitcostmatrix[fromIdx][toIdx],
					}}
				}
				subwayLegs[i] = legs
				return nil
			})
		case Walking:
			group.Go(func() error {
				leg, err := GetWalkingLeg(edges.Mapbox{}, context, originAddr, destAddr)
				if err != nil {
					fmt.Printf("Walking Error (idx %d): %v\n", i, err)
					leg = Leg{
						TransportType: Walking,
						TravelTimes:   input.TravelTimeMatrix[fromIdx][toIdx],
						TransitCosts:  transitcostmatrix[fromIdx][toIdx],
					}
				}
				walkingLegs[i] = []Leg{leg}
				return nil
			})
		case Car:
			group.Go(func() error {
				leg, err := GetCarLeg(edges.Mapbox{}, context, originAddr, destAddr)
				if err != nil {
					leg = Leg{
						TransportType: Car,
						TravelTimes:   input.TravelTimeMatrix[fromIdx][toIdx],
						TransitCosts:  transitcostmatrix[fromIdx][toIdx],
					}
				}
				carLegs[i] = []Leg{leg}
				return nil
			})
		}
	}

	if err := group.Wait(); err != nil {
		return Itinerary{}, fmt.Errorf("failed to fetch subway legs: %w", err)
	}

	//loop over items and make itinerary
	for i, routeEntry := range output.Route {
		node := input.Nodes[routeEntry.NodeIndex]
		nodeAddress := addressmap[routeEntry.NodeIndex]

		var legs []Leg

		//maybe a better way to do this.. idk its 1AM lol
		if i < len(output.Route)-1 {
			if subwayLegs[i] != nil {
				legs = subwayLegs[i]
			} else if walkingLegs[i] != nil {
				legs = walkingLegs[i]
			} else if carLegs[i] != nil {
				legs = carLegs[i]
			} else {
				fromIdx := routeEntry.NodeIndex
				toIdx := output.Route[i+1].NodeIndex
				legs = []Leg{{
					TransportType: transittypematrix[fromIdx][toIdx],
					TravelTimes:   input.TravelTimeMatrix[fromIdx][toIdx],
					TransitCosts:  transitcostmatrix[fromIdx][toIdx],
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
		Stops:                   entries,
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

func GetWalkingLeg(provider edges.WalkingLegProvider, ctx context.Context, origin Address, destination Address) (Leg, error) {
	return provider.AcquireWalkingLeg(origin, destination)
}

func GetCarLeg(provider edges.CarLegProvider, ctx context.Context, origin Address, destination Address) (Leg, error) {
	return provider.AcquireCarLeg(origin, destination)
}
