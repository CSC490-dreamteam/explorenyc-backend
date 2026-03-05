package postprocessing
//sanitizes and processes JSON from python microservice


import (
	"fmt"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)


func ProcessRouteResponse(input SolverInput,output SolverOutput ) (string, error) {

	//check if no solution was found
	if output.HasSolution == false {
		return "", fmt.Errorf("no solution found for given input")
	} //TODO, OUTPUT failure reason TO FRONTEND (forgot if this is even possible xd)

	entries := make([]ItineraryEntry, len(output.Route))

	for i, routeEntry := range output.Route {
		node := input.Nodes[routeEntry.NodeIndex]

		nodeAddress := Address{
			Latitude:  node.Latitude,
			Longitude: node.Longitude,
			Name:      node.Name,

		}

		newEntry := ItineraryEntry{
			


			//MAKE STOPMAP THAT MAPS STOPS BY NAME OT THEIR ADDRESS

}