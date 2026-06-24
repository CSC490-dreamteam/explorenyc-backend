package processing

import (
	"fmt"
	"strings"
	"sync"

	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// AddressResolutionError means the start or a distinct end location
// couldn't be resolved, so there's nothing usable to build an itinerary
// from. Callers can use errors.As to map this to a 400, as opposed to a
// downstream failure.
type AddressResolutionError struct {
	Err error
}

func (e *AddressResolutionError) Error() string { return e.Err.Error() }
func (e *AddressResolutionError) Unwrap() error { return e.Err }

// AcquireAddresses turns the start/stop/end location strings on an
// ItineraryRequest into real addresses.
//
// The returned error is an *AddressResolutionError when fatal. The returned
// []error are non-fatal warnings for individual stops that failed to resolve.
func AcquireAddresses(req ItineraryRequest, provider maps.QueryProvider) (AcquiredPlaces, []error, error) {
	hasEnd := req.EndLocation != nil && *req.EndLocation != ""
	startEqualsEnd := hasEnd && strings.EqualFold(req.StartLocation, *req.EndLocation)
	needsSeparateEnd := hasEnd && !startEqualsEnd

	startAddr, err := provider.AcquireAddress(req.StartLocation)
	if err != nil {
		return AcquiredPlaces{}, nil, &AddressResolutionError{Err: fmt.Errorf("could not resolve start location '%s': %w", req.StartLocation, err)}
	}

	var endAddr Address
	if needsSeparateEnd {
		endAddr, err = provider.AcquireAddress(*req.EndLocation)
		if err != nil {
			return AcquiredPlaces{}, nil, &AddressResolutionError{Err: fmt.Errorf("could not resolve end location '%s': %w", *req.EndLocation, err)}
		}
	}

	numStops := len(req.Stops)
	numPlaces := numStops + 1 // +1 for the start location
	if needsSeparateEnd {
		numPlaces++ // +1 for the end location
	}

	places := make([]Address, numPlaces)
	places[0] = startAddr

	stopErrors := make([]error, numStops)
	var stopGroup sync.WaitGroup
	stopGroup.Add(numStops)

	for i, stop := range req.Stops {
		go func(index int, location string) {
			defer stopGroup.Done()

			addr, err := provider.AcquireAddress(location)
			if err != nil {
				stopErrors[index] = fmt.Errorf("could not resolve '%s': %w", location, err)
				return
			}

			places[index+1] = addr
		}(i, stop.Location)
	}

	stopGroup.Wait()

	var warnings []error
	for _, e := range stopErrors {
		if e != nil {
			warnings = append(warnings, e)
		}
	}

	if needsSeparateEnd {
		places[numPlaces-1] = endAddr
	}

	endIndex := -1 // open ended
	if needsSeparateEnd {
		endIndex = numPlaces - 1
	} else if startEqualsEnd {
		endIndex = 0 // round trip
	}

	return AcquiredPlaces{Places: places, EndIndex: endIndex}, warnings, nil
}

// EdgeProviders bundles the per-mode matrix providers used by AcquireEdgeWeights.
type EdgeProviders struct {
	Walking edges.WalkingMatrixProvider
	Car     edges.CarMatrixProvider
	Subway  edges.SubwayMatrixProvider
}

// AcquireEdgeWeights concurrently fetches EdgeWeights for each transit type
// selected (true) in transitTypes, ready to be passed to CombineBestEdges.
func AcquireEdgeWeights(places []Address, transitTypes map[string]bool, providers EdgeProviders) ([]EdgeWeights, []TransitType, []error) {
	var walkingEdges, carEdges, subwayEdges EdgeWeights
	var walkingErr, carErr, subwayErr error

	selectedCount := 0
	for _, selected := range transitTypes {
		if selected {
			selectedCount++
		}
	}

	var wg sync.WaitGroup
	wg.Add(selectedCount)

	if transitTypes["walking"] {
		go func() {
			defer wg.Done()
			walkingEdges, walkingErr = providers.Walking.AcquireWalkingTravelTime(places)
		}()
	}
	if transitTypes["subway"] {
		go func() {
			defer wg.Done()
			subwayEdges, subwayErr = providers.Subway.AcquireSubwayTravelTime(places)
		}()
	}
	if transitTypes["car"] {
		go func() {
			defer wg.Done()
			carEdges, carErr = providers.Car.AcquireCarTravelTime(places)
		}()
	}
	wg.Wait()

	var errs []error
	if transitTypes["walking"] && walkingErr != nil {
		errs = append(errs, fmt.Errorf("failed to get walking travel times: %w", walkingErr))
	}
	if transitTypes["subway"] && subwayErr != nil {
		errs = append(errs, fmt.Errorf("failed to get subway travel times: %w", subwayErr))
	}
	if transitTypes["car"] && carErr != nil {
		errs = append(errs, fmt.Errorf("failed to get car travel times: %w", carErr))
	}

	var edgeWeights []EdgeWeights
	var selectedTypes []TransitType
	if transitTypes["walking"] {
		edgeWeights = append(edgeWeights, walkingEdges)
		selectedTypes = append(selectedTypes, Walking)
	}
	if transitTypes["subway"] {
		edgeWeights = append(edgeWeights, subwayEdges)
		selectedTypes = append(selectedTypes, Subway)
	}
	if transitTypes["car"] {
		edgeWeights = append(edgeWeights, carEdges)
		selectedTypes = append(selectedTypes, Car)
	}

	return edgeWeights, selectedTypes, errs
}
