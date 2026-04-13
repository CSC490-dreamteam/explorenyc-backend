package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathfinders"
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

	var mapProvider maps.QueryProvider = maps.GoogleMaps{}

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
	var mapProvider maps.QueryProvider = maps.GoogleMaps{}

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
func inputTime(s string) *string {
	return &s
}
func TestLocalGenerateItinerary(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Fatal("error loading .env")
	}

	requestBody := ItineraryRequest{
		TripName:      "NYC Adventure",
		Date:          "2023-10-27",
		EntryTime:     "11:00 AM",
		ExitTime:      "09:00 PM",
		StartLocation: "Penn Station, New York, NY",
		TransitTypes: map[string]bool{
			"walking": true,
			"car":     false,
			"subway":  true,
		},
		Stops: []StopRequest{
			{
				Location:       "Times Square",
				TimePreference: nil,
				Mandatory:      true,
				Duration:       60, //spend an hour there
			},
			{
				Location:       "Central Park",
				TimePreference: inputTime("4:00 PM"),
				Mandatory:      true,
				Duration:       60, //spend an hour there
			},
			{
				Location:       "burp castle new york ny",
				TimePreference: nil,
				Mandatory:      true,
				Duration:       60, //spend an hour there
			},
			{
				Location:       "Holcombe Rucker Park",
				TimePreference: nil,
				Mandatory:      true,
				Duration:       130,
			},
		},
	}

	jsonValue, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/GenerateItinerary", port), bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("API_KEY"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed (is your server running?): %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	fmt.Println(string(body))
}
