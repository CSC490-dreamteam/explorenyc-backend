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
				seconds := int(*matrixResponse.Durations[i][j])
				minutes := seconds / 60
				if minutes == 0 && seconds > 0 {
					minutes = 1
				}
				durations[i][j] = minutes
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

func (m Mapbox) AcquireWalkingLeg(origin Address, destination Address) (Leg, error) {
	return m.acquireLeg(origin, destination, Walking)
}

func (m Mapbox) AcquireCarLeg(origin Address, destination Address) (Leg, error) {
	return m.acquireLeg(origin, destination, Car)
}

func (m Mapbox) acquireLeg(origin Address, destination Address, transitType TransitType) (Leg, error) {

	var profile string
	switch transitType {
	case Walking:
		profile = "mapbox/walking"
	case Car:
		profile = "mapbox/driving-traffic"
	default:
		return Leg{}, fmt.Errorf("unsupported transit type: %d", transitType)
	}

	coords := fmt.Sprintf("%.6f,%.6f;%.6f,%.6f", origin.Lon, origin.Lat, destination.Lon, destination.Lat)

	apiKey := os.Getenv("MAPBOX_MATRIX_API_KEY")
	if apiKey == "" {
		return Leg{}, fmt.Errorf("MAPBOX_MATRIX_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf(
		"https://api.mapbox.com/directions/v5/%s/%s?geometries=polyline&overview=full&access_token=%s",
		profile,
		coords,
		apiKey,
	)

	resp, err := http.Get(url)

	if err != nil {
		return Leg{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Leg{}, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	var directionsResponse struct {
		Code   string `json:"code"`
		Routes []struct {
			Duration float64 `json:"duration"`
			Geometry string  `json:"geometry"`
			Legs     []struct {
			} `json:"legs"`
		} `json:"routes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&directionsResponse); err != nil {
		return Leg{}, fmt.Errorf("JSON decode error: %w", err)
	}

	if directionsResponse.Code != "Ok" {
		return Leg{}, fmt.Errorf("Mapbox API error: %s", directionsResponse.Code)
	}

	if len(directionsResponse.Routes) == 0 {
		return Leg{}, fmt.Errorf("no routes found")
	}

	route := directionsResponse.Routes[0]

	polyline := route.Geometry

	return Leg{
		TransportType: transitType,
		TravelTimes:   int(route.Duration/60) + 1, //convert to minutes
		TransitCosts:  0,
		Polylines:     []string{polyline},
	}, nil
}
