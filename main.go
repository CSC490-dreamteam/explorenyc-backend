package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathFinders"
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

	router.POST("/GenerateRoute", func(context *gin.Context) {
		var req RouteRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		var stops []Stop
		var errors []string

		for _, location := range req.Locations {
			//addr, err := maps.GrabAddressFromOSM(location)
			addr, err := maps.GrabAddressFromGoogleMaps(location)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)

}
