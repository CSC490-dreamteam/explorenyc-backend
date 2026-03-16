package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	matrixaggregator "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/matrixaggregator"
	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathfinders"
	"github.com/CSC490-dreamteam/explorenyc-backend/route_generation/postprocessing"
)

type RouteRequest struct {
	Locations []string `json:"locations" binding:"required"`
}

type RouteResponse struct {
	URL    string   `json:"url"`
	Errors []string `json:"errors,omitempty"`
}

func apiKeyAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		expected := os.Getenv("API_KEY")
		if expected == "" {
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "API key not configed on server"})
			return
		}
		if context.GetHeader("X-API-Key") != expected {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
			return
		}
		context.Next()
	}
}

func main() {

	//loads .env is local, doesn't if on railway
	_ = godotenv.Load()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, //temp cors fix
		AllowMethods: []string{"POST", "GET", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "X-API-Key"},
	}))
	router.Use(apiKeyAuth())

	//env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pythonURL := os.Getenv("PYTHON_SERVICE_URL")
	if pythonURL == "" {
		fmt.Println("PYTHON_SERVICE_URL not set, defaulting to http://0.0.0.0:8000")
		pythonURL = "http://0.0.0.0:8000"
	}
	pythonURL = pythonURL + "/generate_route"

	//old brute force google routing endpoint
	router.POST("/GenerateRoute", func(context *gin.Context) {
		var req RouteRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		var stops []Stop
		var errors []string
		var mapProvider maps.Provider = maps.GoogleMaps{}

		for _, location := range req.Locations {

			addr, err := mapProvider.AcquireAddress(location)

			if err != nil {
				errors = append(errors, fmt.Sprintf("could not resolve '%s': %v", location, err))
				continue
			}
			stops = append(stops, Stop{
				Name:      location,
				Latitude:  addr.Lat,
				Longitude: addr.Lon,
			})
		}

		bestPath := pathfinders.BruteForcePathFinderWithDistance(stops)
		url := maps.GetGoogleMapsRouteExportURL(bestPath)

		for _, err := range errors {
			fmt.Printf("Error: %s\n", err)
		}

		context.JSON(http.StatusOK, RouteResponse{
			URL:    url,
			Errors: errors,
		})
	})

	router.POST("/GenerateItinerary", func(context *gin.Context) {
		//get json off frontend
		var ItineraryReq ItineraryRequest
		if err := context.ShouldBindJSON(&ItineraryReq); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		//parse frontend json

		var places []Address
		var errors []error
		var mapProvider maps.Provider = maps.GoogleMaps{}

		//insert start location
		startAddr, err := mapProvider.AcquireAddress(ItineraryReq.StartLocation)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not resolve start location '%s': %v", ItineraryReq.StartLocation, err)})
			return
		}

		//prep places for concurrency
		numStops := len(ItineraryReq.Stops)
		places = make([]Address, numStops+1) // +1 for the start location
		places[0] = startAddr
		stopErrors := make([]error, numStops)

		var stopGroup sync.WaitGroup
		stopGroup.Add(numStops)

		//concurrently iterate through stops and acquire addresses, storing errors in a separate slice
		for i, stop := range ItineraryReq.Stops {
			go func(index int, location string) {
				defer stopGroup.Done()

				addr, err := mapProvider.AcquireAddress(location)
				if err != nil {
					stopErrors[index] = fmt.Errorf("could not resolve '%s': %v", location, err)
					return
				}

				places[index+1] = addr
			}(i, stop.Location)
		}

		stopGroup.Wait()

		//append stop errors to main errors slice after concurrency is done
		for _, e := range stopErrors {
			if e != nil {
				errors = append(errors, e)
			}
		}

		//get edges between stops//

		//setup concurrency
		var edgeWeigthGroup sync.WaitGroup
		var walkingEdges, carEdges, subwayEdges EdgeWeights
		var walkingErr, carErr, subwayErr error

		//setup edge providers
		walkingDataProvider := edges.Mapbox{}
		carDataProvider := edges.Mapbox{}
		subwayDataProvider := edges.GoogleMaps{}

		edgeWeigthGroup.Add(3)

		go func() {
			defer edgeWeigthGroup.Done()
			walkingEdges, walkingErr = walkingDataProvider.AcquireWalkingTravelTime(places)
		}()

		go func() {
			defer edgeWeigthGroup.Done()
			subwayEdges, subwayErr = subwayDataProvider.AcquireSubwayTravelTime(places)
		}()

		go func() {
			defer edgeWeigthGroup.Done()
			carEdges, carErr = carDataProvider.AcquireCarTravelTime(places)
		}()

		edgeWeigthGroup.Wait()

		//acquire edgeweight errors after concurrency is done
		if walkingErr != nil || subwayErr != nil || carErr != nil {
			var errs []string
			if walkingErr != nil {
				errs = append(errs, fmt.Sprintf("failed to acquire walking travel times: %v", walkingErr))
			}
			if subwayErr != nil {
				errs = append(errs, fmt.Sprintf("failed to acquire subway travel times: %v", subwayErr))
			}
			if carErr != nil {
				errs = append(errs, fmt.Sprintf("failed to acquire car travel times: %v", carErr))
			}

			context.JSON(http.StatusInternalServerError, gin.H{
				"errors": errs,
			})
			return
		}

		fmt.Println("Edge Weights acquired")

		//combine weights using david's part

		transitconfig := CombineConfig{
			TimeValueCentsPerMinute:  25,
			WalkingMaxMinutes:        25,
			WalkingMaxDistanceMeters: 2000,
			SubwayFlatFareCents:      300,
			CarBaseFareCents:         250,
			CarCostPerMinuteCents:    12,
			CarCostPerKilometerCents: 50,
		}

		optimizedMatrices, err := matrixaggregator.CombineBestEdges(walkingEdges, subwayEdges, carEdges, transitconfig)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to combine best edges: %v", err))
			fmt.Println("combining error")
		}

		fmt.Println("Matrices Combined")

		//create python payload
		var solverNodes []SolverNode
		//add start and end
		solverNodes = append(solverNodes, SolverNode{
			ID:                "0",
			Name:              places[0].PlaceName,
			Latitude:          places[0].Lat,
			Longitude:         places[0].Lon,
			DurationInMinutes: 0, // start point, no dwell time
			TimeWindowStart:   parseTimeIntoMinutes(ItineraryReq.EntryTime),
			TimeWindowEnd:     parseTimeIntoMinutes(ItineraryReq.ExitTime),
			Priority:          Mandatory,
			DropPenalty:       0,
		})

		//map each place to its index for the post processor to use later when it gets the python output
		addressMapping := make(map[int]Address)
		for i, place := range places {
			addressMapping[i] = place
		}

		for i, stop := range ItineraryReq.Stops {
			//todo create string id or osmething

			var timeWindowStart int
			var timeWindowEnd int

			//set time window for that specific node for when arrival time can be set
			if stop.TimePreference != nil {
				preferred := parseTimeIntoMinutes(*stop.TimePreference)
				timeWindowStart = preferred - 30 //30 min before preferred arrival time
				timeWindowEnd = preferred - 5    //5 min before actual time
			} else {
				// no preference = anytime during the day is fine
				timeWindowStart = parseTimeIntoMinutes(ItineraryReq.EntryTime)
				timeWindowEnd = parseTimeIntoMinutes(ItineraryReq.ExitTime)
			}

			//setup priority and drop penalty based on whether stop is mandatory or not
			prio := WantToSee
			if stop.Mandatory {
				prio = Mandatory
			}

			var dropPenalty int = 0
			switch prio {
			case Mandatory:
				dropPenalty = 0 //undroppable
			case WantToSee:
				dropPenalty = 8 //will be dropped if it screws over like 4-6 optional places?
			case Optional:
				dropPenalty = 2
			}

			node := SolverNode{ //we do i+1 to account for the start node we added at the beginning
				ID:                fmt.Sprintf("%d", i+1),
				Name:              places[i+1].PlaceName,
				Latitude:          places[i+1].Lat,
				Longitude:         places[i+1].Lon,
				DurationInMinutes: 35,              //PLACEHOLDER time spent at that location
				TimeWindowStart:   timeWindowStart, //fix?
				TimeWindowEnd:     timeWindowEnd,   //fix
				Priority:          prio,
				DropPenalty:       dropPenalty,
			}
			solverNodes = append(solverNodes, node)
		}

		solverInput := SolverInput{
			Nodes:                 solverNodes,
			StartIndex:            0,
			EndIndex:              0,
			DayStartTimeInMinutes: parseTimeIntoMinutes(ItineraryReq.EntryTime),
			DayEndTimeInMinutes:   parseTimeIntoMinutes(ItineraryReq.ExitTime),
			BudgetInCents:         5000,                          //PLACEHOLDER, $50 budget for transit costs
			TravelTimeMatrix:      optimizedMatrices.TimeMinutes, //TODO david
			CostMatrix:            optimizedMatrices.CostCents,   //TODO placeholder, wait for david
			CandidateGroups:       []CandidateGroup{},            //TODO wait on nick
			RouteVariant:          Balanced,
			//empty so python doesnt get mad
			Precedences:   [][2]int{},
			ForcedEdges:   [][2]int{},
			ExcludedStops: []int{},
		}

		fmt.Printf("Solver input: %+v\n", solverInput)

		//send python payload

		pythonClient := &http.Client{Timeout: 10 * time.Second}

		pythonJSON, err := json.Marshal(solverInput)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to marshal solver input: %v", err)})
			return
		}

		pythonReq, err := http.NewRequest("POST", pythonURL, bytes.NewBuffer(pythonJSON))
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create request: %v", err)})
			fmt.Printf("failed to create request: %v", err)
			return
		}
		pythonReq.Header.Set("Content-Type", "application/json")

		pythonResp, err := pythonClient.Do(pythonReq)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to reach python service: %v", err)})
			fmt.Printf("failed to reach python service: %v", err)
			return
		}
		defer pythonResp.Body.Close()

		if pythonResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(pythonResp.Body)
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("python service returned %d: %s", pythonResp.StatusCode, string(body)),
			})
			return
		}

		//parse python response
		var solverOutput SolverOutput
		if err := json.NewDecoder(pythonResp.Body).Decode(&solverOutput); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to decode python response: %v", err)})
			return
		}

		fmt.Printf("Solver output: %+v\n", solverOutput)

		adaptedtransitTypeMatrix, err := ConvertModeMatrix(optimizedMatrices.Mode)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to convert mode matrix: %v", err)})
			return
		}

		//process python response with post processor
		PostProcessorInput := PostProcessorInput{
			SolverInput:       solverInput,
			SolverOutput:      solverOutput,
			StopMap:           addressMapping,
			TransitTypeMatrix: adaptedtransitTypeMatrix,
			TransitCostMatrix: optimizedMatrices.CostCents,
		}

		itinerary, err := postprocessing.ProcessRouteResponse(PostProcessorInput)

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to process route response: %v", err)})
			return
		}

		//print out errors
		if len(errors) > 0 {
			context.JSON(http.StatusOK, gin.H{
				"warnings":  fmt.Sprintf("itinerary generated with the following warnings: %v", errors),
				"itinerary": itinerary,
			})
			return
		}

		//send out the itinerary to the frontend
		context.JSON(http.StatusOK, itinerary)
	})

	router.Run(":" + port)

}

// turn 09:00 AM to 540
func parseTimeIntoMinutes(timeStr string) int {
	timeStr = strings.TrimSpace(timeStr)
	layouts := []string{"3:04 PM", "15:04", "15:04:05"}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return int(t.Hour()*60 + t.Minute())
		}
	}

	fmt.Printf("invalid time format: %s", timeStr)
	return -1
}

func ConvertModeMatrix(strMatrix [][]string) ([][]TransitType, error) {
	if len(strMatrix) == 0 {
		return make([][]TransitType, 0), nil
	}

	result := make([][]TransitType, len(strMatrix))

	for i, row := range strMatrix {
		result[i] = make([]TransitType, len(row))

		for j, val := range row {
			tt := parseTransitType(val)

			result[i][j] = tt
		}
	}

	return result, nil
}

func parseTransitType(s string) TransitType {
	switch s {
	case "WALKING":
		return Walking
	case "CAR":
		return Car
	case "SUBWAY":
		return Subway
	case "UNREACHABLE", "SELF":
		// Defaulting to Walking per your temporary requirement
		return Walking
	default:
		return 0
	}
}
