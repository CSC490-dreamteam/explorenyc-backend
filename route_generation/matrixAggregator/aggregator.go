package matrixAggregator

import (
	"fmt"
	"math"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

const (
	walkingModeName     = "WALKING"
	subwayModeName      = "SUBWAY"
	carModeName         = "CAR"
	selfModeName        = "SELF"
	unreachableModeName = "UNREACHABLE"
)

// transportModeInput is an internal helper that lets us keep the aggregator logic
// generic without using the previous interface-based design.
//
// This keeps the main function simpler while still making the mode comparison logic
// reusable and easy to extend later.
type transportModeInput struct {
	modeName string

	edgeWeights EdgeWeights

	// If a cap is 0, it is ignored.
	maxDurationMinutes int
	maxDistanceMeters  int

	calculateCostCents func(durationMinutes int, distanceMeters int, combineConfig CombineConfig) int
}

// evaluatedEdge stores the computed result for one candidate mode on one edge.
type evaluatedEdge struct {
	modeName               string
	durationMinutes        int
	distanceMeters         int
	costCents              int
	comparisonPenaltyCents int
	valid                  bool
}

// CombineBestEdges combines walking, subway, and car EdgeWeights into one optimized
// set of output matrices.
//
// Output matrices are aligned by index:
// - TimeMinutes[from][to] gives the chosen duration
// - CostCents[from][to] gives the chosen cost in cents
// - Mode[from][to] gives the chosen transport mode name
//
//	behavior changes from last timee:
//	1. The function now accepts 3 EdgeWeights structs directly
//	2. Money is stored in cents as ints instead of float dollars.
//	3. Distances are now available and used in the car cost model.
//	4. The old interface design was removed.
func CombineBestEdges(
	walkingEdgeWeights EdgeWeights,
	subwayEdgeWeights EdgeWeights,
	carEdgeWeights EdgeWeights,
	combineConfig CombineConfig,
) (CombinedMatrices, error) {

	if validationError := validateComparableEdgeWeights(
		walkingEdgeWeights,
		subwayEdgeWeights,
		carEdgeWeights,
	); validationError != nil {
		return CombinedMatrices{}, validationError
	}

	stopCount := len(walkingEdgeWeights.Nodes)

	timeOutput, costOutput, modeOutput := initializeOutputMatrices(stopCount)
	writeDiagonalOutputs(timeOutput, costOutput, modeOutput)

	transportModes := buildTransportModes(
		walkingEdgeWeights,
		subwayEdgeWeights,
		carEdgeWeights,
		combineConfig,
	)

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		for toStopIndex := 0; toStopIndex < stopCount; toStopIndex++ {
			if fromStopIndex == toStopIndex {
				continue
			}

			selectedEdge := chooseBestEdge(
				fromStopIndex,
				toStopIndex,
				transportModes,
				combineConfig,
			)

			writeSelectedEdgeOutputs(
				fromStopIndex,
				toStopIndex,
				selectedEdge,
				timeOutput,
				costOutput,
				modeOutput,
			)
		}
	}

	return CombinedMatrices{
		TimeMinutes: timeOutput,
		CostCents:   costOutput,
		Mode:        modeOutput,
	}, nil
}

// buildTransportModes converts the three input EdgeWeights structs into a common
// internal shape so the selection logic can be reused for each mode.
func buildTransportModes(
	walkingEdgeWeights EdgeWeights,
	subwayEdgeWeights EdgeWeights,
	carEdgeWeights EdgeWeights,
	combineConfig CombineConfig,
) []transportModeInput {
	_ = combineConfig

	return []transportModeInput{
		{
			modeName:           walkingModeName,
			edgeWeights:        walkingEdgeWeights,
			maxDurationMinutes: combineConfig.WalkingMaxMinutes,
			maxDistanceMeters:  combineConfig.WalkingMaxDistanceMeters,
			calculateCostCents: calculateWalkingCostCents,
		},
		{
			modeName:           subwayModeName,
			edgeWeights:        subwayEdgeWeights,
			calculateCostCents: calculateSubwayCostCents,
		},
		{
			modeName:           carModeName,
			edgeWeights:        carEdgeWeights,
			calculateCostCents: calculateCarCostCents,
		},
	}
}

// validateComparableEdgeWeights ensures all three EdgeWeights structs are compatible.
//
// IMPORTANT:
// We validate matrix sizes and node counts here.
// We do NOT deeply compare whether each Address value is identical across all three
// inputs, because Address may not be directly comparable depending on its definition.
// This function assumes upstream preserves the same node ordering across all modes.
func validateComparableEdgeWeights(
	walkingEdgeWeights EdgeWeights,
	subwayEdgeWeights EdgeWeights,
	carEdgeWeights EdgeWeights,
) error {
	if shapeError := validateEdgeWeightsShape(walkingEdgeWeights, walkingModeName); shapeError != nil {
		return shapeError
	}
	if shapeError := validateEdgeWeightsShape(subwayEdgeWeights, subwayModeName); shapeError != nil {
		return shapeError
	}
	if shapeError := validateEdgeWeightsShape(carEdgeWeights, carModeName); shapeError != nil {
		return shapeError
	}

	expectedStopCount := len(walkingEdgeWeights.Nodes)

	if len(subwayEdgeWeights.Nodes) != expectedStopCount {
		return fmt.Errorf(
			"%s node count %d does not match %s node count %d",
			subwayModeName,
			len(subwayEdgeWeights.Nodes),
			walkingModeName,
			expectedStopCount,
		)
	}

	if len(carEdgeWeights.Nodes) != expectedStopCount {
		return fmt.Errorf(
			"%s node count %d does not match %s node count %d",
			carModeName,
			len(carEdgeWeights.Nodes),
			walkingModeName,
			expectedStopCount,
		)
	}

	return nil
}

// validateEdgeWeightsShape ensures one EdgeWeights struct is internally well-formed.
func validateEdgeWeightsShape(edgeWeights EdgeWeights, modeName string) error {
	stopCount := len(edgeWeights.Nodes)

	if stopCount == 0 {
		return fmt.Errorf("%s EdgeWeights has no nodes", modeName)
	}

	if len(edgeWeights.Durations) != stopCount {
		return fmt.Errorf(
			"%s duration row count %d does not match node count %d",
			modeName,
			len(edgeWeights.Durations),
			stopCount,
		)
	}

	if len(edgeWeights.Distances) != stopCount {
		return fmt.Errorf(
			"%s distance row count %d does not match node count %d",
			modeName,
			len(edgeWeights.Distances),
			stopCount,
		)
	}

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		if len(edgeWeights.Durations[fromStopIndex]) != stopCount {
			return fmt.Errorf(
				"%s duration row %d has length %d, expected %d",
				modeName,
				fromStopIndex,
				len(edgeWeights.Durations[fromStopIndex]),
				stopCount,
			)
		}

		if len(edgeWeights.Distances[fromStopIndex]) != stopCount {
			return fmt.Errorf(
				"%s distance row %d has length %d, expected %d",
				modeName,
				fromStopIndex,
				len(edgeWeights.Distances[fromStopIndex]),
				stopCount,
			)
		}
	}

	return nil
}

// initializeOutputMatrices allocates the three matrices returned by the aggregator.
func initializeOutputMatrices(stopCount int) ([][]int, [][]int, [][]string) {
	timeOutput := make([][]int, stopCount)
	costOutput := make([][]int, stopCount)
	modeOutput := make([][]string, stopCount)

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		timeOutput[fromStopIndex] = make([]int, stopCount)
		costOutput[fromStopIndex] = make([]int, stopCount)
		modeOutput[fromStopIndex] = make([]string, stopCount)
	}

	return timeOutput, costOutput, modeOutput
}

// writeDiagonalOutputs initializes all diagonal cells as SELF / 0 / 0.
func writeDiagonalOutputs(
	timeOutput [][]int,
	costOutput [][]int,
	modeOutput [][]string,
) {
	for stopIndex := 0; stopIndex < len(timeOutput); stopIndex++ {
		timeOutput[stopIndex][stopIndex] = 0
		costOutput[stopIndex][stopIndex] = 0
		modeOutput[stopIndex][stopIndex] = selfModeName
	}
}

// chooseBestEdge evaluates all available transport modes for one directed edge
// and returns the one with the lowest comparison penalty.
func chooseBestEdge(
	fromStopIndex int,
	toStopIndex int,
	transportModes []transportModeInput,
	combineConfig CombineConfig,
) evaluatedEdge {
	bestEdge := evaluatedEdge{
		valid:                  false,
		comparisonPenaltyCents: math.MaxInt,
	}

	for _, mode := range transportModes {
		candidateEdge := evaluateEdgeForMode(
			fromStopIndex,
			toStopIndex,
			mode,
			combineConfig,
		)

		if !candidateEdge.valid {
			continue
		}

		if candidateEdge.comparisonPenaltyCents < bestEdge.comparisonPenaltyCents {
			bestEdge = candidateEdge
		}
	}

	return bestEdge
}

// evaluateEdgeForMode computes one transport mode's result for a single edge.
func evaluateEdgeForMode(
	fromStopIndex int,
	toStopIndex int,
	mode transportModeInput,
	combineConfig CombineConfig,
) evaluatedEdge {
	durationSeconds := mode.edgeWeights.Durations[fromStopIndex][toStopIndex]
	durationMinutes := secondsToMinutesCeil(durationSeconds)

	distanceMeters := mode.edgeWeights.Distances[fromStopIndex][toStopIndex]

	if !isSaneTravelValue(durationMinutes) || !isSaneTravelValue(distanceMeters) {
		return evaluatedEdge{valid: false}
	}

	// Off-diagonal duration 0 means unreachable for that mode.
	if durationMinutes <= 0 {
		return evaluatedEdge{valid: false}
	}

	if edgeExceedsCaps(durationMinutes, distanceMeters, mode) {
		return evaluatedEdge{valid: false}
	}

	costCents := mode.calculateCostCents(durationMinutes, distanceMeters, combineConfig)
	comparisonPenaltyCents := calculateComparisonPenaltyCents(
		durationMinutes,
		costCents,
		combineConfig,
	)

	return evaluatedEdge{
		modeName:               mode.modeName,
		durationMinutes:        durationMinutes,
		distanceMeters:         distanceMeters,
		costCents:              costCents,
		comparisonPenaltyCents: comparisonPenaltyCents,
		valid:                  true,
	}
}

// edgeExceedsCaps applies optional per-mode caps.
// Right now this is mainly used for walking so free walking does not dominate
// extremely long trips.
func edgeExceedsCaps(
	durationMinutes int,
	distanceMeters int,
	mode transportModeInput,
) bool {
	if mode.maxDurationMinutes > 0 && durationMinutes > mode.maxDurationMinutes {
		return true
	}

	if mode.maxDistanceMeters > 0 && distanceMeters > mode.maxDistanceMeters {
		return true
	}

	return false
}

// writeSelectedEdgeOutputs writes either the selected edge values or UNREACHABLE.
func writeSelectedEdgeOutputs(
	fromStopIndex int,
	toStopIndex int,
	selectedEdge evaluatedEdge,
	timeOutput [][]int,
	costOutput [][]int,
	modeOutput [][]string,
) {
	if !selectedEdge.valid {
		timeOutput[fromStopIndex][toStopIndex] = -1
		costOutput[fromStopIndex][toStopIndex] = -1
		modeOutput[fromStopIndex][toStopIndex] = unreachableModeName
		return
	}

	timeOutput[fromStopIndex][toStopIndex] = selectedEdge.durationMinutes
	costOutput[fromStopIndex][toStopIndex] = selectedEdge.costCents
	modeOutput[fromStopIndex][toStopIndex] = selectedEdge.modeName
}

// calculateComparisonPenaltyCents computes the "lower is better" comparison value.
//
// renamed this away from "score"
func calculateComparisonPenaltyCents(
	durationMinutes int,
	costCents int,
	combineConfig CombineConfig,
) int {
	return costCents + (combineConfig.TimeValueCentsPerMinute * durationMinutes)
}

// calculateWalkingCostCents keeps walking free.
func calculateWalkingCostCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	_ = durationMinutes
	_ = distanceMeters
	_ = combineConfig
	return 0
}

// calculateSubwayCostCents treats subway as a flat per-edge fare.
func calculateSubwayCostCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	_ = durationMinutes
	_ = distanceMeters
	return combineConfig.SubwayFlatFareCents
}

// calculateCarCostCents uses both duration and distance now that both are available.
// Formula:
//
//	base fare
//
// + per-minute charge
// + per-kilometer charge
// Distance is stored in meters, so we convert using integer math.
func calculateCarCostCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	distanceCostCents := (distanceMeters * combineConfig.CarCostPerKilometerCents) / 1000
	durationCostCents := durationMinutes * combineConfig.CarCostPerMinuteCents

	return combineConfig.CarBaseFareCents + durationCostCents + distanceCostCents
}

// isSaneTravelValue rejects impossible matrix values coming from upstream data.
func isSaneTravelValue(value int) bool {
	return value >= 0
}

func secondsToMinutesCeil(seconds int) int {
	return (seconds + 59) / 60
}
