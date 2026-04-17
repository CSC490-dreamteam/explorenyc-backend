package preprocessing

import (
	"fmt"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// dynamicTransitInput pairs one transit enum with its matching EdgeWeights input.
type dynamicTransitInput struct {
	transitType TransitType
	edgeWeights EdgeWeights
}

// evaluatedEdge stores one candidate result for one directed edge under one transit type.
// used internally while comparing all provided transit options for a given edge.
type evaluatedEdge struct {
	transitType              TransitType
	durationMinutes          int
	distanceMeters           int
	costCents                int
	comparisonPenaltyInCents int
	valid                    bool
}

// CombineBestEdges combines a dynamic list of transit graphs into one optimized graph.
//
// Inputs:
//   - edgeWeightsByTransit: one EdgeWeights struct per selected transit type
//   - transitTypes: the enum corresponding to each EdgeWeights input
//   - combineConfig: config used for the time-vs-cost tradeoff and pricing rules
//
// IMPORTANT:
//   - edgeWeightsByTransit and transitTypes must be the same length
//   - both slices must be in matching order
//   - all EdgeWeights inputs must describe the same ordered list of nodes
//
// Output:
//   - TimeMinutes[from][to] is the chosen travel time
//   - CostCents[from][to] is the chosen travel cost
//   - Mode[from][to] is the chosen TransitType enum
//
// Diagonal behavior:
//   - The diagonal is intentionally left at the zero value because the routing engine
//     ignores self-to-self edges and this was explicitly approved.
func CombineBestEdges(
	edgeWeightsByTransit []EdgeWeights,
	transitTypes []TransitType,
	combineConfig CombineConfig,
) (CombinedMatrices, error) {
	if validationError := validateAggregatorInputs(edgeWeightsByTransit, transitTypes); validationError != nil {
		return CombinedMatrices{}, validationError
	}

	stopCount := len(edgeWeightsByTransit[0].Nodes)

	timeOutput, costOutput, modeOutput := initializeOutputMatrices(stopCount)
	transitInputs := buildDynamicTransitInputs(edgeWeightsByTransit, transitTypes)

	// If walking is the ONLY selected transit type, we bypass the walking caps.
	// This prevents the route from failing just because every walking edge exceeds
	// the normal walking cap. The cap still applies in mixed-mode scenarios.
	onlyWalkingSelected := isWalkingOnlySelection(transitTypes)

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		for toStopIndex := 0; toStopIndex < stopCount; toStopIndex++ {
			if fromStopIndex == toStopIndex {
				continue
			}

			selectedEdge, selectionError := chooseBestEdge(
				fromStopIndex,
				toStopIndex,
				transitInputs,
				combineConfig,
				onlyWalkingSelected,
			)
			if selectionError != nil {
				return CombinedMatrices{}, selectionError
			}

			//intentionally writes directly into the output matrices.
			timeOutput[fromStopIndex][toStopIndex] = selectedEdge.durationMinutes
			costOutput[fromStopIndex][toStopIndex] = selectedEdge.costCents
			modeOutput[fromStopIndex][toStopIndex] = selectedEdge.transitType
		}
	}

	return CombinedMatrices{
		TimeMinutes: timeOutput,
		CostCents:   costOutput,
		Mode:        modeOutput,
	}, nil
}

// buildDynamicTransitInputs pairs the two input slices into one shared structure.
// keeps the main aggregation loop cleaner and avoids repeated parallel indexing.
func buildDynamicTransitInputs(
	edgeWeightsByTransit []EdgeWeights,
	transitTypes []TransitType,
) []dynamicTransitInput {
	transitInputs := make([]dynamicTransitInput, len(transitTypes))

	for inputIndex, transitType := range transitTypes {
		transitInputs[inputIndex] = dynamicTransitInput{
			transitType: transitType,
			edgeWeights: edgeWeightsByTransit[inputIndex],
		}
	}

	return transitInputs
}

// isWalkingOnlySelection returns true only when the caller selected exactly one
// transit type and that type is Walking.
// the walking-cap override only in the "walking only" case.
func isWalkingOnlySelection(transitTypes []TransitType) bool {
	return len(transitTypes) == 1 && transitTypes[0] == Walking
}

// validateAggregatorInputs ensures the dynamic inputs are valid and comparable.
func validateAggregatorInputs(
	edgeWeightsByTransit []EdgeWeights,
	transitTypes []TransitType,
) error {
	if len(edgeWeightsByTransit) == 0 {
		return fmt.Errorf("edgeWeightsByTransit cannot be empty")
	}

	if len(transitTypes) == 0 {
		return fmt.Errorf("transitTypes cannot be empty")
	}

	if len(edgeWeightsByTransit) != len(transitTypes) {
		return fmt.Errorf(
			"edgeWeightsByTransit length %d does not match transitTypes length %d",
			len(edgeWeightsByTransit),
			len(transitTypes),
		)
	}

	if duplicateError := validateNoDuplicateTransitTypes(transitTypes); duplicateError != nil {
		return duplicateError
	}

	for _, transitType := range transitTypes {
		if !hasTransitCostFunction(transitType) {
			return fmt.Errorf("unsupported transit type: %v", transitType)
		}
	}

	for inputIndex, edgeWeights := range edgeWeightsByTransit {
		if shapeError := validateEdgeWeightsShape(edgeWeights, transitTypes[inputIndex]); shapeError != nil {
			return shapeError
		}
	}

	if comparabilityError := validateComparableNodeCounts(edgeWeightsByTransit, transitTypes); comparabilityError != nil {
		return comparabilityError
	}

	return nil
}

// validateNoDuplicateTransitTypes rejects duplicate transit types.
// each EdgeWeights input is expected to represent one unique transport mode.
func validateNoDuplicateTransitTypes(transitTypes []TransitType) error {
	seenTransitTypes := make(map[TransitType]bool)

	for _, transitType := range transitTypes {
		if seenTransitTypes[transitType] {
			return fmt.Errorf("duplicate transit type provided: %v", transitType)
		}
		seenTransitTypes[transitType] = true
	}

	return nil
}

// validateEdgeWeightsShape checks that one EdgeWeights input is internally NxN
// and that its matrices match the node count.
func validateEdgeWeightsShape(edgeWeights EdgeWeights, transitType TransitType) error {
	stopCount := len(edgeWeights.Nodes)

	if stopCount == 0 {
		return fmt.Errorf("edge weights for transit type %v contain no nodes", transitType)
	}

	if len(edgeWeights.Durations) != stopCount {
		return fmt.Errorf(
			"duration row count %d does not match node count %d for transit type %v",
			len(edgeWeights.Durations),
			stopCount,
			transitType,
		)
	}

	if len(edgeWeights.Distances) != stopCount {
		return fmt.Errorf(
			"distance row count %d does not match node count %d for transit type %v",
			len(edgeWeights.Distances),
			stopCount,
			transitType,
		)
	}

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		if len(edgeWeights.Durations[fromStopIndex]) != stopCount {
			return fmt.Errorf(
				"duration row %d has length %d, expected %d for transit type %v",
				fromStopIndex,
				len(edgeWeights.Durations[fromStopIndex]),
				stopCount,
				transitType,
			)
		}

		if len(edgeWeights.Distances[fromStopIndex]) != stopCount {
			return fmt.Errorf(
				"distance row %d has length %d, expected %d for transit type %v",
				fromStopIndex,
				len(edgeWeights.Distances[fromStopIndex]),
				stopCount,
				transitType,
			)
		}
	}

	return nil
}

// validateComparableNodeCounts ensures all selected transit graphs describe the same
// number of nodes. As with the previous iteration, this assumes upstream preserves
// the same node ordering across the provided transit types.
func validateComparableNodeCounts(
	edgeWeightsByTransit []EdgeWeights,
	transitTypes []TransitType,
) error {
	expectedStopCount := len(edgeWeightsByTransit[0].Nodes)

	for inputIndex := 1; inputIndex < len(edgeWeightsByTransit); inputIndex++ {
		if len(edgeWeightsByTransit[inputIndex].Nodes) != expectedStopCount {
			return fmt.Errorf(
				"node count %d for transit type %v does not match expected node count %d from transit type %v",
				len(edgeWeightsByTransit[inputIndex].Nodes),
				transitTypes[inputIndex],
				expectedStopCount,
				transitTypes[0],
			)
		}
	}

	return nil
}

// initializeOutputMatrices allocates the three output matrices.
// diagonal is intentionally left as the zero value and not separately initialized.
func initializeOutputMatrices(stopCount int) ([][]int, [][]int, [][]TransitType) {
	timeOutput := make([][]int, stopCount)
	costOutput := make([][]int, stopCount)
	modeOutput := make([][]TransitType, stopCount)

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		timeOutput[fromStopIndex] = make([]int, stopCount)
		costOutput[fromStopIndex] = make([]int, stopCount)
		modeOutput[fromStopIndex] = make([]TransitType, stopCount)
	}

	return timeOutput, costOutput, modeOutput
}

//chooseBestEdge evaluates all selected transit types for one directed edge and returns the candidate with the lowest comparison penalty.

// if no valid candidate exists, this function returns an error.
func chooseBestEdge(
	fromStopIndex int,
	toStopIndex int,
	transitInputs []dynamicTransitInput,
	combineConfig CombineConfig,
	onlyWalkingSelected bool,
) (evaluatedEdge, error) {
	bestEdge := evaluatedEdge{
		valid: false,
	}

	for _, transitInput := range transitInputs {
		candidateEdge := evaluateEdgeForTransit(
			fromStopIndex,
			toStopIndex,
			transitInput,
			combineConfig,
			onlyWalkingSelected,
		)

		if !candidateEdge.valid {
			continue
		}

		if !bestEdge.valid || candidateEdge.comparisonPenaltyInCents < bestEdge.comparisonPenaltyInCents {
			bestEdge = candidateEdge
		}
	}

	if !bestEdge.valid {
		return evaluatedEdge{}, fmt.Errorf(
			"no valid transit edge found from stop %d to stop %d",
			fromStopIndex,
			toStopIndex,
		)
	}

	return bestEdge, nil
}

// evaluateEdgeForTransit computes one candidate edge for one transit type.
func evaluateEdgeForTransit(
	fromStopIndex int,
	toStopIndex int,
	transitInput dynamicTransitInput,
	combineConfig CombineConfig,
	onlyWalkingSelected bool,
) evaluatedEdge {
	durationMinutes := transitInput.edgeWeights.Durations[fromStopIndex][toStopIndex]
	distanceMeters := transitInput.edgeWeights.Distances[fromStopIndex][toStopIndex]

	if !isSaneTravelValue(durationMinutes) || !isSaneTravelValue(distanceMeters) {
		return evaluatedEdge{valid: false}
	}

	//off-diagonal duration 0 means unreachable for that transit type.
	if durationMinutes <= 0 {
		return evaluatedEdge{valid: false}
	}

	if edgeExceedsTransitCaps(transitInput.transitType, durationMinutes, distanceMeters, combineConfig, onlyWalkingSelected) {
		return evaluatedEdge{valid: false}
	}

	costFunction := getTransitCostFunction(transitInput.transitType)
	costCents := costFunction(durationMinutes, distanceMeters, combineConfig)

	comparisonPenaltyInCents := calculateComparisonPenaltyInCents(
		durationMinutes,
		costCents,
		combineConfig,
	)

	return evaluatedEdge{
		transitType:              transitInput.transitType,
		durationMinutes:          durationMinutes,
		distanceMeters:           distanceMeters,
		costCents:                costCents,
		comparisonPenaltyInCents: comparisonPenaltyInCents,
		valid:                    true,
	}
}

// edgeExceedsTransitCaps applies transit-specific filtering.
// to stay close to the previous iteration, only walking uses caps right now.
// walking is the only selected transit type, walking caps are ignored
// multiple transit types are selected, walking caps still apply as normal
func edgeExceedsTransitCaps(
	transitType TransitType,
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
	onlyWalkingSelected bool,
) bool {
	if transitType != Walking {
		return false
	}

	if onlyWalkingSelected {
		return false
	}

	if combineConfig.WalkingMaxMinutes > 0 && durationMinutes > combineConfig.WalkingMaxMinutes {
		return true
	}

	if combineConfig.WalkingMaxDistanceMeters > 0 && distanceMeters > combineConfig.WalkingMaxDistanceMeters {
		return true
	}

	return false
}

// calculateComparisonPenaltyInCents computes the "lower is better" comparison value.
// the name uses "penalty" instead of "score" because smaller values are better.
func calculateComparisonPenaltyInCents(
	durationMinutes int,
	costCents int,
	combineConfig CombineConfig,
) int {
	return costCents + (combineConfig.TimeValueCentsPerMinute * durationMinutes)
}

// isSaneTravelValue rejects impossible negative values from upstream data.
func isSaneTravelValue(value int) bool {
	return value >= 0
}
