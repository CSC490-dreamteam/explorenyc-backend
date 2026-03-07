package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)

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

		var stops []Address
		var errors []string
		var mapProvider maps.Provider = maps.GoogleMaps{}

		//get location data for each stop requested
		for _, location := range ItineraryReq.Stops {

			addr, err := mapProvider.AcquireAddress(location.Location)

			if err != nil {
				errors = append(errors, fmt.Sprintf("could not resolve '%s': %v", location, err))
				continue
			}
			stops = append(stops, Address{
				Lat:              addr.Lat,
				Lon:              addr.Lon,
				Street:           addr.Street,
				City:             addr.City,
				State:            addr.State,
				Zip:              addr.Zip,
				PlaceName:        addr.PlaceName,
				FormattedAddress: addr.FormattedAddress,
			})
		}

		//get edges between stops//

		//setup data providers
		walkingDataProvider := edges.GoogleMaps{}
		carDataProvider := edges.GoogleMaps{}
		subwayDataProvider := edges.GoogleMaps{}

		//get edge weights for each transit mode
		walkingEdges, err := walkingDataProvider.AcquireWalkingTravelTime(stops)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to acquire walking travel times: %v", err))
		}
		carEdges, err := carDataProvider.AcquireCarTravelTime(stops)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to acquire car travel times: %v", err))
		}
		subwayEdges, err := subwayDataProvider.AcquireSubwayTravelTime(stops)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to acquire subway travel times: %v", err))
		}

		//combine weights using david's part
		//TODO

		optimizedMatrices := CombineBestEdges(walkingEdges, carEdges, subwayEdges) //placeholder for now, just uses walking edges

		//create python payload

		var solverNodes []SolverNode

		for i, stop := range ItineraryReq.Stops {
			//todo create string id or osmething

			var timeWindowStart int
			var timeWindowEnd int

			//set time window for that specific node
			if stop.TimePreference != nil {
				timeWindowStart = parseTimeIntoMinutes(*stop.TimePreference)
				timeWindowEnd = timeWindowStart + 90 //PLACEHOLDER, just gives 1.5 hr time window for now
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

			node := SolverNode{
				ID:                fmt.Sprintf("%d", i),
				Name:              stops[i].PlaceName,
				Latitude:          stops[i].Lat,
				Longitude:         stops[i].Lon,
				DurationInMinutes: 90,              //PLACEHOLDER
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
			EndIndex:              1,
			DayStartTimeInMinutes: parseTimeIntoMinutes(ItineraryReq.EntryTime),
			DayEndTimeInMinutes:   parseTimeIntoMinutes(ItineraryReq.ExitTime),
			BudgetInCents:         5000,                          //PLACEHOLDER, $50 budget for transit costs
			TravelTimeMatrix:      optimizedMatrices.TimeMinutes, //TODO david
			CostMatrix:            optimizedMatrices.CostDollars, //TODO placeholder, wait for david
			CandidateGroups:       nil,                           //TODO wait on nick
			RouteVariant:          Balanced,
		}

		//send python payload

		//parse python response

		//process python response with post processor
		PostProcessorInput := PostProcessorInput{
			SolverInput:       solverInput,
			SolverOutput:      SolverOutput{},                //TODO get real output from python
			StopMap:           make(map[int]Address),         //TODO make stopmap that maps stop indices to their address for post processor
			TransitTypeMatrix: optimizedMatrices.Mode,        //TODO get from david
			TransitCostMatrix: optimizedMatrices.CostDollars, //TODO get from david
		}

		itinerary, err := postprocessing.ProcessRouteResponse(PostProcessorInput)

		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to process route response: %v", err)})
			return
		}

		//send out the itinerary to the frontend
		context.JSON(http.StatusOK, itinerary)
	})

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

	fmt.Errorf("invalid time format: %s", timeStr)
	return -1
}
