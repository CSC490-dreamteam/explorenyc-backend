package main

import (
	"fmt"

	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

func main() {

	//fmt.Print("Hello, NYC!\n")

	//sample locations to test with
	locations := []string{
		"Empire State Building",
		//"wall street ny",
		//"starbucks soho ny",
		"Times Square new york ny",
		//"grand central station",
		"Penn station New york ny",
		"macy's new york ny",
	}

	var stops []pathfinders.Stop

	//iterate through all given locations
	for _, location := range locations {
		addr, err := maps.GrabAddressFromOSM(location) //grab location data from OSM API
		if err != nil {
			fmt.Printf("Error for '%s': %v\n", location, err)
			continue
		}

		//print address info
		fmt.Printf("%s:\n%s, %s, %s %s\n\n",
			location,
			addr.Street,
			addr.City,
			addr.State,
			addr.Zip,
		)

		//add stops to slice with coords
		stops = append(stops, pathfinders.Stop{
			Name:      location,
			Latitude:  addr.Lat,
			Longitude: addr.Lon,
		})
	}

	// Find optimal path
	bestPath := pathfinders.BruteForcePathFinderWithDistance(stops)

	//fmt.Printf("Best path order: %v\n", bestPath)
	fmt.Print("Here is a google maps link!")
	fmt.Print(maps.GetGoogleMapsRouteExportURL(bestPath))

}
