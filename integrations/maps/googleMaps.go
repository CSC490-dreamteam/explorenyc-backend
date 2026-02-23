package maps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	apiKey := os.Getenv("GOOGLE_MAPS_PLACES_API_KEY")

	if apiKey == "" {
		return Address{}, fmt.Errorf("GOOGLE_MAPS_PLACES_API_KEY environment variable is not set")
	}

	//build JSON request body
	requestBody, err := json.Marshal(map[string]string{"textQuery": query})
	if err != nil {
		return Address{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	//actual api endpoint request
	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return Address{}, fmt.Errorf("failed to create request: %w", err)
	}

	//endpoint headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Goog-Api-Key", apiKey)
	request.Header.Set("X-Goog-FieldMask",
		"places.displayName,places.addressComponents,places.location") //what data we get and what we google bills us

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return Address{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Address{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return Address{}, fmt.Errorf("non-OK HTTP status: %s — %s", response.Status, string(body))
	}

	var result struct {
		Places []struct {
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			FormattedAddress string `json:"formattedAddress"`
			Location         struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			AddressComponents []struct {
				LongText  string   `json:"longText"`
				ShortText string   `json:"shortText"`
				Types     []string `json:"types"`
			} `json:"addressComponents"`
		} `json:"places"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return Address{}, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	if len(result.Places) == 0 {
		return Address{}, fmt.Errorf("no results found for query: %s", query)
	}

	place := result.Places[0] //just grabs the first one for now

	//divvy up data to their respective parts
	var streetNumber, route, city, state, zip string
	for _, component := range place.AddressComponents {
		for _, text := range component.Types {
			switch text {
			case "street_number":
				streetNumber = component.LongText
			case "route":
				route = component.LongText
			case "locality":
				city = component.LongText
			case "administrative_area_level_1":
				state = component.ShortText //"NY" instead of "New York"
			case "postal_code":
				zip = component.LongText
			}
		}
	}

	//fuses together street number and street name nicely
	street := streetNumber
	if street != "" && route != "" {
		street += " "
	}
	street += route

	return Address{
		Lat:              place.Location.Latitude,
		Lon:              place.Location.Longitude,
		Street:           street,
		City:             city,
		State:            state,
		Zip:              zip,
		placeName:        place.DisplayName.Text,
		formattedAddress: place.FormattedAddress,
	}, nil

}
