package edges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type Tomtom struct{}

func (t Tomtom) AcquireCarTravelTime(Addrs []Address) (EdgeWeights, error) {
	return t.acquireTravelTime(Addrs, "car")
}

func (t Tomtom) AcquireWalkingTravelTime(Addrs []Address) (EdgeWeights, error) {
	return t.acquireTravelTime(Addrs, "pedestrian")
}

func (t Tomtom) acquireTravelTime(Addrs []Address, travelMode string) (EdgeWeights, error) {
	//validate matrix size (200 cell limit = 14x14 max for square matrices)
	n := len(Addrs)
	if n > 14 {
		return EdgeWeights{}, fmt.Errorf("TomTom synchronous API supports maximum 14 stops (got %d), use asynchronous API for larger matrices", n)
	}

	//get API key
	TOMTOM_MATRIX_API_KEY := os.Getenv("TOMTOM_MATRIX_API_KEY")
	if TOMTOM_MATRIX_API_KEY == "" {
		return EdgeWeights{}, fmt.Errorf("TOMTOM_MATRIX_API_KEY environment variable is not set")
	}

	//build request payload
	payload := map[string]interface{}{
		"origins":      t.makePoints(Addrs),
		"destinations": t.makePoints(Addrs),
		"options": map[string]interface{}{
			"travelMode": travelMode,
			"routeType":  "fastest", // only option for synchronous API
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("failed to marshal request body: %v", err)
	}

	//make request
	url := fmt.Sprintf("https://api.tomtom.com/routing/matrix/2?key=%s", TOMTOM_MATRIX_API_KEY)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	//check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return EdgeWeights{}, fmt.Errorf("non-OK HTTP status: %s, body: %s", resp.Status, string(body))
	}

	//parse response
	var tomtomResponse struct {
		Data []struct {
			OriginIndex      int `json:"originIndex"`
			DestinationIndex int `json:"destinationIndex"`
			RouteSummary     *struct {
				LengthInMeters      int `json:"lengthInMeters"`
				TravelTimeInSeconds int `json:"travelTimeInSeconds"`
			} `json:"routeSummary,omitempty"`
			DetailedError *struct {
				Code       string `json:"code"`
				Message    string `json:"message"`
				StatusCode int    `json:"statusCode"`
			} `json:"detailedError,omitempty"`
		} `json:"data"`
		Statistics struct {
			TotalCount int `json:"totalCount"`
			Successes  int `json:"successes"`
			Failures   int `json:"failures"`
		} `json:"statistics"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, &tomtomResponse); err != nil {
		return EdgeWeights{}, fmt.Errorf("JSON decode error: %w, raw response: %s", err, string(body))
	}

	//verify API success
	if tomtomResponse.Statistics.Failures > 0 {
		//try to give details from the first failed cell
		for _, cell := range tomtomResponse.Data {
			if cell.DetailedError != nil {
				return EdgeWeights{}, fmt.Errorf(
					"TomTom API error for route [%d→%d]: %s (%s)",
					cell.OriginIndex, cell.DestinationIndex,
					cell.DetailedError.Message, cell.DetailedError.Code,
				)
			}
		}
		//fallback if no detailed error found
		return EdgeWeights{}, fmt.Errorf("TomTom API: %d route(s) failed", tomtomResponse.Statistics.Failures)
	}

	//initialize matrices
	durations := make([][]int, n)
	distances := make([][]int, n)
	for i := 0; i < n; i++ {
		durations[i] = make([]int, n)
		distances[i] = make([]int, n)
	}

	//populate matrices
	for _, cell := range tomtomResponse.Data {
		i := cell.OriginIndex
		j := cell.DestinationIndex
		if cell.RouteSummary != nil {
			seconds := cell.RouteSummary.TravelTimeInSeconds
			minutes := seconds / 60
			if minutes == 0 && seconds > 0 {
				minutes = 1
			}
			durations[i][j] = minutes
			distances[i][j] = cell.RouteSummary.LengthInMeters
		}
	}
	return EdgeWeights{
		Nodes:     Addrs,
		Distances: distances,
		Durations: durations,
	}, nil
}

func (t Tomtom) makePoints(Addrs []Address) []map[string]interface{} {
	points := make([]map[string]interface{}, len(Addrs))
	for i, Addr := range Addrs {
		points[i] = map[string]interface{}{
			"point": map[string]float64{
				"latitude":  Addr.Lat,
				"longitude": Addr.Lon,
			},
		}
	}
	return points
}
