package maps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//FOR NOW JUST PROGRAM RAW WITHOUT INTERFACE SO WE KNOW WHAT WE ARE DOING

type Address struct {
	Lat         float64
	Lon         float64
	Street      string
	City        string
	State       string
	Zip         string
	DisplayName string
}

func GrabAddressFromOSM(query string) (Address, error) {
	const endpoint = "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", string(query))
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("limit", "1")

	requestURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := http.Get(requestURL)
	if err != nil {
		return Address{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer response.Body.Close()

	//read the response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Address{}, fmt.Errorf("failed to read response body: %w", err)
	}

	//bad http code error
	if response.StatusCode != http.StatusOK {
		return Address{}, fmt.Errorf("non-OK HTTP status: %s", response.Status)
	}

	var results []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"`
		Address     struct {
			HouseNumber string `json:"house_number"`
			Road        string `json:"road"`
			City        string `json:"city"`
			State       string `json:"state"`
			Postcode    string `json:"postcode"`
		} `json:"address"`
	}

	//formatting  error
	if err := json.Unmarshal(body, &results); err != nil {
		return Address{}, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	//empty results error
	if len(results) == 0 {
		return Address{}, fmt.Errorf("no results found for query: %s", query)
	}

	//get first result
	result := results[0]

	//convert coordinates to float64
	var lat, lon float64
	fmt.Sscanf(result.Lat, "%f", &lat)
	fmt.Sscanf(result.Lon, "%f", &lon)

	//set street
	street := result.Address.HouseNumber
	if street != "" && result.Address.Road != "" {
		street += " "
	}
	street += result.Address.Road

	return Address{
		Lat:         lat,
		Lon:         lon,
		Street:      street,
		City:        result.Address.City,
		State:       result.Address.State,
		Zip:         result.Address.Postcode,
		DisplayName: result.DisplayName,
	}, nil

}
