package edges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type GoogleMaps struct{}


func (g GoogleMaps) AcquireWalkingTravelTime(stops []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(stops, "walking")
}

func (g GoogleMaps) AcquireCarTravelTime(stops []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(stops, "car")
}

func (g GoogleMaps) AcquireSubwayTravelTime(stops []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(stops, "subway")
}

// generic function to get travel time for any transit mode, API is mostly the same with tiny diffs
func (g GoogleMaps) acquireTravelTime(stops []Address, transitMode string) (EdgeWeights, error) {
	const endpoint = "https://routes.googleapis.com/distanceMatrix/v2:computeRouteMatrix"
	apiKey := os.Getenv("GOOGLE_MAPS_ROUTES_API_KEY")

	if apiKey == "" {
		return EdgeWeights{}, fmt.Errorf("GOOGLE_MAPS_ROUTES_API_KEY environment variable is not set")
	}

	//make "waypoints". how google wants the stops formatted for the API
	waypoints := make([]map[string]interface{}, len(stops))
	for i, stop := range stops {
		waypoints[i] = map[string]interface{}{
			"waypoint": map[string]interface{}{
				"location": map[string]interface{}{
					"latLng": map[string]float64{
						"latitude":  stop.Lat,
						"longitude": stop.Lon,
					},
				},
			},
		}
	}

	//make JSON body for the request
	body := map[string]interface{}{
		"origins":      waypoints,
		"destinations": waypoints,
	}

	switch transitMode {
	case "walking":
		body["travelMode"] = "WALK"
	case "car":
		body["travelMode"] = "DRIVE"
		body["routingPreference"] = "TRAFFIC_AWARE"
	case "subway":
		body["travelMode"] = "TRANSIT"
		body["transitPreferences"] = map[string]interface{}{
			"allowedTravelModes": []string{"SUBWAY"},
			"routingPreference":  "FEWER_TRANSFERS",
		}
	}
	requestBody, err := json.Marshal(body)

	if err != nil {
		return EdgeWeights{}, fmt.Errorf("failed to marshal request body: %v", err)
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	//request headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Goog-Api-Key", apiKey)
	request.Header.Set("X-Goog-FieldMask",
		"originIndex,destinationIndex,status,distanceMeters,duration,condition")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("failed to read response body: %w", err)
	}

	//uncomment to print the raw api body
	//fmt.Println("RAW RESPONSE:", string(responseBody))

	if response.StatusCode != http.StatusOK {
		return EdgeWeights{}, fmt.Errorf("non-OK HTTP status: %s", response.Status)
	}

	var edges []struct {
		OriginIndex      int    `json:"originIndex"`
		DestinationIndex int    `json:"destinationIndex"`
		DistanceMeters   int    `json:"distanceMeters"`
		Duration         string `json:"duration"`
		Condition        string `json:"condition"`
	}

	if err := json.Unmarshal(responseBody, &edges); err != nil {
		return EdgeWeights{}, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	// make N by N matrices
	n := len(stops)
	durations := make([][]int, n)
	distances := make([][]int, n)

	for i := 0; i < n; i++ {
		durations[i] = make([]int, n)
		distances[i] = make([]int, n)
	}

	//fill matrices with api response data
	for _, edge := range edges {

		//no route found from Google's end
		if edge.Condition == "ROUTE_NOT_FOUND" {
			continue
		}

		if edge.Duration == "" {
			return EdgeWeights{}, fmt.Errorf("empty duration for edge [%d][%d] with condition: %s",
				edge.OriginIndex, edge.DestinationIndex, edge.Condition)
		}

		seconds, err := strconv.Atoi(strings.TrimSuffix(edge.Duration, "s"))
		if err != nil {
			return EdgeWeights{}, fmt.Errorf("failed to parse duration for edge: %v", err)
		}
		durations[edge.OriginIndex][edge.DestinationIndex] = seconds
		distances[edge.OriginIndex][edge.DestinationIndex] = edge.DistanceMeters
	}

	return EdgeWeights{
		Nodes:     stops,
		Distances: distances,
		Durations: durations,
	}, nil
}

