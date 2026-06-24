package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	"github.com/CSC490-dreamteam/explorenyc-backend/data/cache"
	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/solver"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathfinders"
	"github.com/CSC490-dreamteam/explorenyc-backend/route_generation/processing"
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

	//init redis cache
	redisCache, err := cache.New()
	if err != nil {
		fmt.Printf("failed to init cache: %v\n", err)
		redisCache = &cache.Cache{} // no-op cache
	} else {
		defer redisCache.Close()
	}

	pythonURL := os.Getenv("PYTHON_SERVICE_URL")
	if pythonURL == "" {
		fmt.Println("PYTHON_SERVICE_URL not set, defaulting to http://0.0.0.0:8000")
		pythonURL = "http://0.0.0.0:8000"
	}
	pythonURL = pythonURL + "/generate_route"

	//setup context
	genCtx := processing.GenerationContext{
		Cache:       redisCache,
		MapProvider: maps.GoogleMaps{Cache: redisCache},
		EdgeProviders: processing.EdgeProviders{
			Walking: edges.Mapbox{Cache: redisCache},
			Car:     edges.Mapbox{Cache: redisCache},
			Subway:  edges.GoogleMaps{Cache: redisCache},
		},
		CombineConfig: CombineConfig{
			TimeValueCentsPerMinute:  25,
			WalkingMaxMinutes:        25,
			WalkingMaxDistanceMeters: 2000,
			SubwayFlatFareCents:      300,
			CarBaseFareCents:         250,
			CarCostPerMinuteCents:    12,
			CarCostPerKilometerCents: 50,
		},
		Solver: solver.NewPythonSolver(pythonURL),
	}

	//old brute force google routing endpoint
	router.POST("/GenerateRoute", func(context *gin.Context) {
		var req RouteRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		var stops []Stop
		var errors []string
		var mapProvider maps.QueryProvider = maps.GoogleMaps{Cache: redisCache}

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
		var itineraryReq ItineraryRequest
		if err := context.ShouldBindJSON(&itineraryReq); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		itinerary, warnings, err := processing.GenerateItinerary(genCtx, itineraryReq)
		if err != nil {
			var addrErr *processing.AddressResolutionError
			var unprocessable *solver.UnprocessableInputError
			switch {
			case errors.As(err, &addrErr):
				context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.As(err, &unprocessable):
				context.Data(http.StatusUnprocessableEntity, "application/json", unprocessable.Body)
			default:
				context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		//surface non-fatal warnings (e.g. a stop that couldn't be resolved) alongside the itinerary
		if len(warnings) > 0 {
			context.JSON(http.StatusOK, gin.H{
				"warnings":  fmt.Sprintf("itinerary generated with the following warnings: %v", warnings),
				"itinerary": itinerary,
			})
			return
		}

		context.JSON(http.StatusOK, itinerary)
	})

	router.Run(":" + port)

}

// func ConvertModeMatrix(strMatrix [][]string) ([][]TransitType, error) {
// 	if len(strMatrix) == 0 {
// 		return make([][]TransitType, 0), nil
// 	}

// 	result := make([][]TransitType, len(strMatrix))

// 	for i, row := range strMatrix {
// 		result[i] = make([]TransitType, len(row))

// 		for j, val := range row {
// 			tt := parseTransitType(val)

// 			result[i][j] = tt
// 		}
// 	}

// 	return result, nil
// }

// func parseTransitType(s string) TransitType {
// 	switch s {
// 	case "WALKING":
// 		return Walking
// 	case "CAR":
// 		return Car
// 	case "SUBWAY":
// 		return Subway
// 	case "UNREACHABLE", "SELF":
// 		// Defaulting to Walking per your temporary requirement
// 		return Walking
// 	default:
// 		return 0
// 	}
// }
