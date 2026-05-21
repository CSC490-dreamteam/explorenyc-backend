package edges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/CSC490-dreamteam/explorenyc-backend/data/cache"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type GoogleMaps struct {
	Cache *cache.Cache
}

func (g GoogleMaps) AcquireWalkingTravelTime(Addrs []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(Addrs, Walking)
}

func (g GoogleMaps) AcquireCarTravelTime(Addrs []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(Addrs, Car)
}

func (g GoogleMaps) AcquireSubwayTravelTime(Addrs []Address) (EdgeWeights, error) {
	return g.acquireTravelTime(Addrs, Subway)
}

// generic function to get travel time for any transit mode, API is mostly the same with tiny diffs
func (g GoogleMaps) acquireTravelTime(Addrs []Address, transitMode TransitType) (EdgeWeights, error) {
	const endpoint = "https://routes.googleapis.com/distanceMatrix/v2:computeRouteMatrix"
	apiKey := os.Getenv("GOOGLE_MAPS_ROUTES_API_KEY")

	if apiKey == "" {
		return EdgeWeights{}, fmt.Errorf("GOOGLE_MAPS_ROUTES_API_KEY environment variable is not set")
	}

	// make N by N matrices
	n := len(Addrs)
	durations := make([][]int, n)
	distances := make([][]int, n)

	for i := 0; i < n; i++ {
		durations[i] = make([]int, n)
		distances[i] = make([]int, n)
	}

	//cache lookups
	type pair struct{ i, j int }
	var missingPairs []pair

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			ev, err := g.Cache.GetEdgeValue(Addrs[i], transitMode, Addrs[j])
			if err != nil {
				// log but don't block — treat as miss
				slog.Warn("cache edge get error", "err", err)
			}
			if ev != nil {
				durations[i][j] = ev.DurationMinutes
				distances[i][j] = ev.DistanceM
			} else {
				missingPairs = append(missingPairs, pair{i, j})
			}
		}
	}

	// all hits — return immediately, no Google API call
	if len(missingPairs) == 0 {
		return EdgeWeights{Nodes: Addrs, Durations: durations, Distances: distances}, nil
	}

	// build set of missing pairs for Google API
	originSet := make(map[int]bool)
	destSet := make(map[int]bool)
	for _, p := range missingPairs {
		originSet[p.i] = true
		destSet[p.j] = true
	}
	//make sorted list of indices
	originIndices := make([]int, 0, len(originSet))
	for idx := range originSet {
		originIndices = append(originIndices, idx)
	}
	sort.Ints(originIndices)

	destIndices := make([]int, 0, len(destSet))
	for idx := range destSet {
		destIndices = append(destIndices, idx)
	}
	sort.Ints(destIndices)

	// translate Google's response indices back to the original array indices so we can place results in the correct matrix positions.
	origToReal := make(map[int]int)
	for si, realIdx := range originIndices {
		origToReal[si] = realIdx
	}

	destToReal := make(map[int]int)
	for si, realIdx := range destIndices {
		destToReal[si] = realIdx
	}

	//make "waypoints". (how google wants the stops formatted for the API)
	makeWaypoints := func(indices []int) []map[string]interface{} {
		waypoints := make([]map[string]interface{}, len(indices))
		for subsetindex, realIdx := range indices {
			waypoints[subsetindex] = map[string]interface{}{
				"waypoint": map[string]interface{}{
					"location": map[string]interface{}{
						"latLng": map[string]float64{
							"latitude":  Addrs[realIdx].Lat,
							"longitude": Addrs[realIdx].Lon,
						},
					},
				},
			}
		}
		return waypoints
	}

	//make JSON body for the request
	body := map[string]interface{}{
		"origins":      makeWaypoints(originIndices),
		"destinations": makeWaypoints(destIndices),
	}

	switch transitMode {
	case Walking:
		body["travelMode"] = "WALK"
	case Car:
		body["travelMode"] = "DRIVE"
		body["routingPreference"] = "TRAFFIC_AWARE"
	case Subway:
		body["travelMode"] = "TRANSIT"
		body["transitPreferences"] = map[string]interface{}{
			"allowedTravelModes": []string{"SUBWAY"},
			"routingPreference":  "FEWER_TRANSFERS",
		}
	default:
		return EdgeWeights{}, fmt.Errorf("unsupported transit mode: %d", transitMode)
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
		//convert to minutes
		minutes := seconds / 60
		if minutes == 0 && seconds > 0 {
			minutes = 1
		}

		// translate subset indices back to original array positions
		realI := origToReal[edge.OriginIndex]
		realJ := destToReal[edge.DestinationIndex]

		durations[realI][realJ] = minutes
		distances[realI][realJ] = edge.DistanceMeters

		// async cache write that doesn't block the response to the client
		go func() {
			ev := &cache.EdgeValue{
				DurationMinutes: minutes,
				DistanceM:       edge.DistanceMeters,
				CostCents:       0, // populated later by cost calculator
				Legs:            nil,
			}
			if err := g.Cache.SetEdgeValue(Addrs[realI], transitMode, Addrs[realJ], ev); err != nil {
				slog.Warn("cache edge set error", "err", err)
			}
		}()
	}

	return EdgeWeights{
		Nodes:     Addrs,
		Distances: distances,
		Durations: durations,
	}, nil
}

func (g GoogleMaps) AcquireSubwayLegs(origin Address, destination Address) ([]Leg, error) {

	const endpoint = "https://routes.googleapis.com/directions/v2:computeRoutes"
	apiKey := os.Getenv("GOOGLE_MAPS_ROUTES_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_ROUTES_API_KEY environment variable is not set")
	}

	//early return for cache hit
	edgeval, err := g.Cache.GetEdgeValue(origin, Subway, destination)
	if err != nil {
		slog.Warn("cache edge get error", "err", err)
	}
	if edgeval != nil && edgeval.Legs != nil {
		return edgeval.Legs, nil
	}

	body := map[string]interface{}{
		"origin": map[string]interface{}{
			"location": map[string]interface{}{
				"latLng": map[string]interface{}{
					"latitude":  origin.Lat,
					"longitude": origin.Lon,
				},
			},
		},
		"destination": map[string]interface{}{
			"location": map[string]interface{}{
				"latLng": map[string]interface{}{
					"latitude":  destination.Lat,
					"longitude": destination.Lon,
				},
			},
		},
		"travelMode": "TRANSIT",
		"transitPreferences": map[string]interface{}{
			"allowedTravelModes": []string{"SUBWAY"},
		},
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Goog-Api-Key", apiKey)
	request.Header.Set("X-Goog-FieldMask",
		"routes.legs.steps.transitDetails,routes.legs.steps.staticDuration,routes.legs.steps.travelMode,routes.legs.steps.polyline.encodedPolyline")
	//transit details, duration, travel type, , polyline

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status: %s", response.Status)
	}

	//pull out response
	type LegsAPIResponse struct {
		Routes []struct {
			Legs []struct {
				Steps []struct {
					TravelMode     string `json:"travelMode"`
					StaticDuration string `json:"staticDuration"`
					Polyline       struct {
						EncodedPolyline string `json:"encodedPolyline"`
					} `json:"polyline"`
				} `json:"steps"`
			} `json:"legs"`
		} `json:"routes"`
	}

	var result LegsAPIResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	if len(result.Routes) == 0 || len(result.Routes[0].Legs) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	//what google calls the legs
	steps := result.Routes[0].Legs[0].Steps

	//create legs to return
	legs := make([]Leg, 0, len(steps))

	for _, step := range steps {
		var transportType TransitType
		var transitCosts int

		if step.TravelMode == "WALK" {
			transportType = Walking
			transitCosts = 0
		} else {
			transportType = Subway
			transitCosts = 300 //in cents
		}

		//parse duration like "120s" to int
		durationStr := strings.TrimSuffix(step.StaticDuration, "s")
		seconds, _ := strconv.Atoi(durationStr)
		travelTime := seconds / 60
		if travelTime == 0 && seconds > 0 {
			travelTime = 1
		}

		// merge with previous leg if same transport type
		if len(legs) > 0 && legs[len(legs)-1].TransportType == transportType {
			legs[len(legs)-1].TravelTimes += travelTime
			legs[len(legs)-1].TransitCosts += transitCosts
			legs[len(legs)-1].Polylines = append(legs[len(legs)-1].Polylines, step.Polyline.EncodedPolyline)
		} else {
			legs = append(legs, Leg{
				TransportType: transportType,
				TravelTimes:   travelTime,
				TransitCosts:  transitCosts,
				Polylines:     []string{step.Polyline.EncodedPolyline},
			})
		}
	}

	// cache write via goroutine to not block main thread
	go func() {
		// If edge existed without legs, update it. If it didn't exist at all, create it.
		if edgeval == nil {
			edgeval = &cache.EdgeValue{}
		}
		edgeval.Legs = legs
		if err := g.Cache.SetEdgeValue(origin, Subway, destination, edgeval); err != nil {
			slog.Warn("cache edge legs set error", "err", err)
		}
	}()

	return legs, nil
}
