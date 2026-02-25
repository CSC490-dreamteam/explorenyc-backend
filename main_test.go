package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"

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

	for _, q := range queries {
		fmt.Println("---")
		fmt.Printf("Query: %s\n", q)

		addr, err := maps.GrabAddressFromGoogleMaps(q)
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

		addr, err := maps.GrabAddressFromGoogleMaps(q)
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
