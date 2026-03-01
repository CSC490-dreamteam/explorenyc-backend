package edges

import (
	"fmt"
	"log"
	"testing"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"github.com/joho/godotenv"
)

func TestAcquireWalkingTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Stop{
		{Name: "Empire State Building", Latitude: 40.748441, Longitude: -73.985664},
		{Name: "Starbucks", Latitude: 40.756640, Longitude: -73.985900},
		{Name: "Central Park Zoo", Latitude: 40.767706, Longitude: -73.971991},
		{Name: "350 5th Ave", Latitude: 40.748721, Longitude: -73.984817},
	}

	var edgeProvider WalkingMatrixProvider = GoogleMaps{}
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
func TestAcquireCarTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Stop{
		{Name: "Empire State Building", Latitude: 40.748441, Longitude: -73.985664},
		{Name: "Starbucks", Latitude: 40.756640, Longitude: -73.985900},
		{Name: "Central Park Zoo", Latitude: 40.767706, Longitude: -73.971991},
		{Name: "350 5th Ave", Latitude: 40.748721, Longitude: -73.984817},
	}

	var edgeProvider CarMatrixProvider = GoogleMaps{}
	result, err := edgeProvider.AcquireCarTravelTime(stops)
	if err != nil {
		t.Fatalf("AcquireCarTravelTimeFromGoogle failed: %v", err)
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

func TestAcquireSubwayTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Stop{
		{Name: "Empire State Building", Latitude: 40.748441, Longitude: -73.985664},
		{Name: "Starbucks", Latitude: 40.756640, Longitude: -73.985900},
		{Name: "Central Park Zoo", Latitude: 40.767706, Longitude: -73.971991},
		{Name: "350 5th Ave", Latitude: 40.748721, Longitude: -73.984817},
	}

	var edgeProvider SubwayMatrixProvider = GoogleMaps{}
	result, err := edgeProvider.AcquireSubwayTravelTime(stops)
	if err != nil {
		t.Fatalf("AcquireSubwayTravelTimeFromGoogle failed: %v", err)
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
