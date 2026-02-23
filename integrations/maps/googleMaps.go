package maps

import (
	"fmt"
	"strings"

	pathfinders "github.com/CSC490-dreamteam/explorenyc-backend/route_generation/pathFinders"
)

//FOR NOW JUST PROGRAM RAW WITHOUT INTERFACE SO WE KNOW WHAT WE ARE DOING

func GetGoogleMapsRouteExportURL(path pathfinders.Path) string {

	//check if stops slice is empty
	if len(path.StopOrder) == 0 {
		return ""
	}
	baseURL := "https://www.google.com/maps/dir/"

	//get coords off stop slice
	var coords []string

	for _, stop := range path.StopOrder {
		coords = append(coords, fmt.Sprintf("%g,%g", stop.Latitude, stop.Longitude))
	}

	//join coords with slashes
	urlPath := strings.Join(coords, "/")

	return baseURL + urlPath

}

func GrabAddressFromGoogleMaps(query string) (Address, error) {
	const endpoint = "https://places.googleapis.com/v1/places:searchText"

}
