package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"

	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathFinders"
)

func TestGrabAddressFromGoogleMaps(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env")
	}

	queries := []string{
		"Empire State Building",
		"Starbucks near Times Square",
		"Central Park Zoo",
		"350 5th Ave, New York, NY",
	}

	var mapProvider maps.Provider = maps.GoogleMaps{}

	for _, q := range queries {
		fmt.Println("---")
		fmt.Printf("Query: %s\n", q)

		addr, err := mapProvider.AcquireAddress(q)

		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("  PlaceName:        %s\n", addr.PlaceName)
		fmt.Printf("  FormattedAddress: %s\n", addr.FormattedAddress)
		fmt.Printf("  Street:           %s\n", addr.Street)
		fmt.Printf("  City:             %s\n", addr.City)
		fmt.Printf("  State:            %s\n", addr.State)
		fmt.Printf("  Zip:              %s\n", addr.Zip)
		fmt.Printf("  Lat:              %f\n", addr.Lat)
		fmt.Printf("  Lon:              %f\n", addr.Lon)
	}
}

func TestExportGoogleURL(t *testing.T) {

	var stops []Stop
	var mapProvider maps.Provider = maps.GoogleMaps{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env")
	}

	queries := []string{
		"Empire State Building",
		"Starbucks near Times Square",
		"Central Park Zoo",
		"350 5th Ave, New York, NY",
	}

	for _, q := range queries {
		fmt.Println("---")
		fmt.Printf("Query: %s\n", q)

		addr, err := mapProvider.AcquireAddress(q)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("  PlaceName:        %s\n", addr.PlaceName)
		fmt.Printf("  FormattedAddress: %s\n", addr.FormattedAddress)
		fmt.Printf("  Street:           %s\n", addr.Street)
		fmt.Printf("  City:             %s\n", addr.City)
		fmt.Printf("  State:            %s\n", addr.State)
		fmt.Printf("  Zip:              %s\n", addr.Zip)
		fmt.Printf("  Lat:              %f\n", addr.Lat)
		fmt.Printf("  Lon:              %f\n", addr.Lon)

		stops = append(stops, Stop{
			Name:      addr.PlaceName, // Use the resolved name
			Latitude:  addr.Lat,
			Longitude: addr.Lon,
		})
	}

	bestPath := pathfinders.BruteForcePathFinderWithDistance(stops)
	url := maps.GetGoogleMapsRouteExportURL(bestPath)
	fmt.Printf("Google Maps Route URL: %s\n", url)
}

func TestAcquireWalkingTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Stop{
		{Name: "Empire State Building", Latitude: 40.748441, Longitude: -73.985664},
		{Name: "Starbucks", Latitude: 40.756640, Longitude: -73.985900},
		{Name: "Central Park Zoo", Latitude: 40.767706, Longitude: -73.971991},
		{Name: "350 5th Ave", Latitude: 40.748721, Longitude: -73.984817},
	}

	var edgeProvider edges.WalkingMatrixProvider = edges.GoogleMaps{}
	result, err := edgeProvider.AcquireWalkingTravelTime(stops)
	if err != nil {
		t.Fatalf("AcquireWalkingTravelTimeFromGoogle failed: %v", err)
	}

	n := len(stops)

	//verify matrix dimensions
	if len(result.Durations) != n {
		t.Fatalf("expected %d rows in Durations, got %d", n, len(result.Durations))
	}
	if len(result.Distances) != n {
		t.Fatalf("expected %d rows in Distances, got %d", n, len(result.Distances))
	}

	//print out info
	fmt.Printf("--- Durations (seconds) ---")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d seconds\n", origin.Name, destination.Name, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.Name, destination.Name, result.Distances[i][j])
		}
	}

}
