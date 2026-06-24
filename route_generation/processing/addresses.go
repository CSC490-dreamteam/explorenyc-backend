package processing

import (
	"fmt"
	"strings"
	"sync"

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

// ResolveAddresses turns the start/stop/end location strings on an
// ItineraryRequest into real addresses.
//
// The returned error is an *AddressResolutionError when fatal. The returned
// []error are non-fatal warnings for individual stops that failed to resolve.
func ResolveAddresses(req ItineraryRequest, provider maps.QueryProvider) (AcquiredPlaces, []error, error) {
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