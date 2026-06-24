package preprocessing

import (
	"fmt"
	"sync"

	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/edges"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

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
