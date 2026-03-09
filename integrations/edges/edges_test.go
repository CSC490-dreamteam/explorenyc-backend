package edges

import (
	"fmt"
	"log"
	"testing"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"github.com/joho/godotenv"
)

// google
func TestAcquireWalkingTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
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
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}

}
func TestAcquireCarTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
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
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}

}

func TestAcquireSubwayTravelTimeFromGoogle(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
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
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}

}

// mapbox
func TestAcquireWalkingTravelTimeFromMapbox(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
	}

	var edgeProvider WalkingMatrixProvider = Mapbox{}
	result, err := edgeProvider.AcquireWalkingTravelTime(stops)
	if err != nil {
		t.Fatalf("AcquireWalkingTravelTimeFromMapbox failed: %v", err)
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
	fmt.Printf("--- Durations (seconds) ---\n")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---\n")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}
}

func TestAcquireCarTravelTimeFromMapbox(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	stops := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
	}

	var edgeProvider CarMatrixProvider = Mapbox{}
	result, err := edgeProvider.AcquireCarTravelTime(stops)
	if err != nil {
		t.Fatalf("AcquireCarTravelTimeFromMapbox failed: %v", err)
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
	fmt.Printf("--- Durations (seconds) ---\n")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---\n")
	for i, origin := range stops {
		for j, destination := range stops {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}
}

// tomtom
func TestAcquireCarTravelTimeFromTomTom(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	addrs := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
	}

	var edgeProvider CarMatrixProvider = Tomtom{}
	result, err := edgeProvider.AcquireCarTravelTime(addrs)
	if err != nil {
		t.Fatalf("AcquireCarTravelTimeFromTomTom failed: %v", err)
	}

	n := len(addrs)

	//verify matrix dimensions
	if len(result.Durations) != n {
		t.Fatalf("expected %d rows in Durations, got %d", n, len(result.Durations))
	}
	if len(result.Distances) != n {
		t.Fatalf("expected %d rows in Distances, got %d", n, len(result.Distances))
	}

	//print out info
	fmt.Printf("--- Durations (seconds) ---\n")
	for i, origin := range addrs {
		for j, destination := range addrs {
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---\n")
	for i, origin := range addrs {
		for j, destination := range addrs {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}
}

func TestAcquireWalkingTravelTimeFromTomTom(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env")
	}

	addrs := []Address{
		{PlaceName: "Empire State Building", Lat: 40.748441, Lon: -73.985664},
		{PlaceName: "Starbucks", Lat: 40.756640, Lon: -73.985900},
		{PlaceName: "Central Park Zoo", Lat: 40.767706, Lon: -73.971991},
		{PlaceName: "350 5th Ave", Lat: 40.748721, Lon: -73.984817},
	}

	var edgeProvider WalkingMatrixProvider = Tomtom{}
	result, err := edgeProvider.AcquireWalkingTravelTime(addrs)
	if err != nil {
		t.Fatalf("AcquireWalkingTravelTimeFromTomTom failed: %v", err)
	}

	n := len(addrs)

	//verify matrix dimensions
	if len(result.Durations) != n {
		t.Fatalf("expected %d rows in Durations, got %d", n, len(result.Durations))
	}
	if len(result.Distances) != n {
		t.Fatalf("expected %d rows in Distances, got %d", n, len(result.Distances))
	}

	//print out info
	fmt.Printf("--- Durations (seconds) ---\n")
	for i, origin := range addrs {
		for j, destination := range addrs {
			fmt.Printf("  %s -> %s: %d seconds\n", origin.PlaceName, destination.PlaceName, result.Durations[i][j])
		}
	}

	fmt.Printf("--- Distances (meters) ---\n")
	for i, origin := range addrs {
		for j, destination := range addrs {
			fmt.Printf("  %s -> %s: %d meters\n", origin.PlaceName, destination.PlaceName, result.Distances[i][j])
		}
	}
}
