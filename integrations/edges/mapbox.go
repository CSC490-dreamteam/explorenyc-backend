package edges

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type Mapbox struct{}

func (m Mapbox) AcquireCarTravelTime(Addrs []Address) (EdgeWeights, error) {
	return m.acquireTravelTime(Addrs, "mapbox/driving-traffic")
}

func (m Mapbox) AcquireWalkingTravelTime(Addrs []Address) (EdgeWeights, error) {
	return m.acquireTravelTime(Addrs, "mapbox/walking")
}

func (m Mapbox) AcquireBikeTravelTime(Addrs []Address) (EdgeWeights, error) {
	return m.acquireTravelTime(Addrs, "mapbox/cycling")
}

func (m Mapbox) acquireTravelTime(Addrs []Address, profile string) (EdgeWeights, error) {
	// check if this is a traffic-aware profile (only driving-traffic has traffic data)
	isTrafficProfile := strings.Contains(profile, "traffic")

	// validate stop count for traffic profile
	if isTrafficProfile && len(Addrs) > 10 {
		return EdgeWeights{}, fmt.Errorf("traffic-aware profile supports maximum 10 Addrs, got %d", len(Addrs))
	}

	//build coordinates string (lon,lat format)
	var coords []string
	for _, stop := range Addrs {
		coords = append(coords, fmt.Sprintf("%.6f,%.6f", stop.Lon, stop.Lat))
	}
	coordinates := strings.Join(coords, ";")

	//get API key
	apiKey := os.Getenv("MAPBOX_MATRIX_API_KEY")
	if apiKey == "" {
		return EdgeWeights{}, fmt.Errorf("MAPBOX_MATRIX_API_KEY environment variable is not set")
	}

	// make URL
	url := fmt.Sprintf(
		"https://api.mapbox.com/directions-matrix/v1/%s/%s?access_token=%s&annotations=distance,duration",
		profile,
		coordinates,
		apiKey,
	)

	// make request
	resp, err := http.Get(url)
	if err != nil {
		return EdgeWeights{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		return EdgeWeights{}, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	// parse response
	var matrixResponse struct {
		Code      string       `json:"code"` //http code such as 200,404,401 etc
		Durations [][]*float64 `json:"durations"`
		Distances [][]*float64 `json:"distances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&matrixResponse); err != nil {
		return EdgeWeights{}, fmt.Errorf("JSON decode error: %w", err)
	}

	// verify API success
	if matrixResponse.Code != "Ok" {
		return EdgeWeights{}, fmt.Errorf("Mapbox API error: %s", matrixResponse.Code)
	}

	// Initialize matrices
	n := len(Addrs)
	durations := make([][]int, n)
	distances := make([][]int, n)
	for i := 0; i < n; i++ {
		durations[i] = make([]int, n)
		distances[i] = make([]int, n)
	}

	// populate matrices
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			// handle duration (convert to int seconds)
			if matrixResponse.Durations[i][j] != nil {
				durations[i][j] = int(*matrixResponse.Durations[i][j])
			}

			// handle distance (convert to int meters)
			if matrixResponse.Distances[i][j] != nil {
				distances[i][j] = int(*matrixResponse.Distances[i][j])
			}
		}
	}

	return EdgeWeights{
		Nodes:     Addrs,
		Distances: distances,
		Durations: durations,
	}, nil

}
