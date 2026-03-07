package matrixAggregator

import (
	"fmt"
	"testing"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

func TestCombineBestEdges_LaymanDemo_V2(t *testing.T) {
	stopNames := []string{
		"Penn Station",
		"Central Park",
		"Ice Cream Shop",
		"Broadway",
	}

	// We do not need real Address field values for this unit test.
	// We only need the slice length to match the matrix size.
	nodes := make([]Address, len(stopNames))

	// Walking: free, but capped later by config.
	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 35, 12, 28},
			{34, 0, 27, 18},
			{12, 26, 0, 17},
			{26, 18, 16, 0},
		},
		Distances: [][]int{
			{0, 2900, 900, 2300},
			{2800, 0, 2100, 1400},
			{900, 2000, 0, 1300},
			{2200, 1400, 1200, 0},
		},
	}

	// Subway: medium speed, flat fare, but not every edge is available.
	subwayEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 18, 0, 11},
			{17, 0, 0, 12},
			{0, 0, 0, 9},
			{11, 12, 10, 0},
		},
		Distances: [][]int{
			{0, 5200, 0, 3100},
			{5000, 0, 0, 2700},
			{0, 0, 0, 2100},
			{3000, 2800, 2200, 0},
		},
	}

	// Car: fastest in some places, but more expensive.
	carEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 12, 7, 9},
			{13, 0, 8, 11},
			{8, 9, 0, 6},
			{9, 12, 7, 0},
		},
		Distances: [][]int{
			{0, 4700, 1900, 3300},
			{4900, 0, 1500, 3600},
			{2000, 1700, 0, 1400},
			{3200, 3500, 1500, 0},
		},
	}

	combineConfig := CombineConfig{
		TimeValueCentsPerMinute:  25,
		WalkingMaxMinutes:        25,
		WalkingMaxDistanceMeters: 2000,
		SubwayFlatFareCents:      300,
		CarBaseFareCents:         250,
		CarCostPerMinuteCents:    12,
		CarCostPerKilometerCents: 50,
	}

	combinedMatrices, combineError := CombineBestEdges(
		walkingEdgeWeights,
		subwayEdgeWeights,
		carEdgeWeights,
		combineConfig,
	)
	if combineError != nil {
		t.Fatalf("unexpected error: %v", combineError)
	}

	t.Log("")
	t.Log("=======================================================================")
	t.Log(" MATRIX AGGREGATOR DEMO")
	t.Log("=======================================================================")
	t.Log("What goes in:")
	t.Log("- 3 EdgeWeights inputs: WALKING, SUBWAY, and CAR")
	t.Log("- Each input contains nodes, durations, and distances")
	t.Log("")
	t.Log("What comes out:")
	t.Log("- ONE chosen mode matrix")
	t.Log("- ONE chosen time matrix")
	t.Log("- ONE chosen cost matrix")
	t.Log("")
	t.Log("How the choice is made:")
	t.Log("comparisonPenalty = costCents + (timeValueCentsPerMinute * durationMinutes)")
	t.Log("Lower comparison penalty wins.")
	t.Log("=======================================================================")

	t.Logf(
		"Config: timeValue=%d cents/min, walkCap=%d min, walkDistanceCap=%d m, subwayFare=%d cents, carBase=%d cents, carPerMinute=%d cents, carPerKm=%d cents",
		combineConfig.TimeValueCentsPerMinute,
		combineConfig.WalkingMaxMinutes,
		combineConfig.WalkingMaxDistanceMeters,
		combineConfig.SubwayFlatFareCents,
		combineConfig.CarBaseFareCents,
		combineConfig.CarCostPerMinuteCents,
		combineConfig.CarCostPerKilometerCents,
	)

	printPrettyIntMatrixWithStops(t, "INPUT: Walking durations (minutes)", stopNames, walkingEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Subway durations (minutes)", stopNames, subwayEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Car durations (minutes)", stopNames, carEdgeWeights.Durations)

	printPrettyStringMatrixWithStops(t, "OUTPUT: Chosen mode for each trip", stopNames, combinedMatrices.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: Chosen travel time (minutes)", stopNames, combinedMatrices.TimeMinutes)
	printPrettyCentsMatrixWithStops(t, "OUTPUT: Chosen travel cost (stored as cents, shown as dollars)", stopNames, combinedMatrices.CostCents)

	validateOutputInvariants(t, combinedMatrices, combineConfig)
}

/* -------------------------------------------------------------------------
Pretty printing helpers
------------------------------------------------------------------------- */

func printPrettyStringMatrixWithStops(
	t *testing.T,
	title string,
	stopNames []string,
	matrix [][]string,
) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	headerRow := fmt.Sprintf("%-16s |", "From \\ To")
	for _, stopName := range stopNames {
		headerRow += fmt.Sprintf(" %-14s |", shortenStopName(stopName))
	}
	t.Log(headerRow)

	for fromStopIndex := 0; fromStopIndex < len(matrix); fromStopIndex++ {
		rowText := fmt.Sprintf("%-16s |", shortenStopName(stopNames[fromStopIndex]))
		for toStopIndex := 0; toStopIndex < len(matrix[fromStopIndex]); toStopIndex++ {
			rowText += fmt.Sprintf(" %-14s |", matrix[fromStopIndex][toStopIndex])
		}
		t.Log(rowText)
	}
}

func printPrettyIntMatrixWithStops(
	t *testing.T,
	title string,
	stopNames []string,
	matrix [][]int,
) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	headerRow := fmt.Sprintf("%-16s |", "From \\ To")
	for _, stopName := range stopNames {
		headerRow += fmt.Sprintf(" %-14s |", shortenStopName(stopName))
	}
	t.Log(headerRow)

	for fromStopIndex := 0; fromStopIndex < len(matrix); fromStopIndex++ {
		rowText := fmt.Sprintf("%-16s |", shortenStopName(stopNames[fromStopIndex]))
		for toStopIndex := 0; toStopIndex < len(matrix[fromStopIndex]); toStopIndex++ {
			rowText += fmt.Sprintf(" %-14d |", matrix[fromStopIndex][toStopIndex])
		}
		t.Log(rowText)
	}
}

func printPrettyCentsMatrixWithStops(
	t *testing.T,
	title string,
	stopNames []string,
	matrix [][]int,
) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	headerRow := fmt.Sprintf("%-16s |", "From \\ To")
	for _, stopName := range stopNames {
		headerRow += fmt.Sprintf(" %-14s |", shortenStopName(stopName))
	}
	t.Log(headerRow)

	for fromStopIndex := 0; fromStopIndex < len(matrix); fromStopIndex++ {
		rowText := fmt.Sprintf("%-16s |", shortenStopName(stopNames[fromStopIndex]))
		for toStopIndex := 0; toStopIndex < len(matrix[fromStopIndex]); toStopIndex++ {
			cellValue := formatCentsForHumans(matrix[fromStopIndex][toStopIndex])
			rowText += fmt.Sprintf(" %-14s |", cellValue)
		}
		t.Log(rowText)
	}
}

func shortenStopName(stopName string) string {
	if len(stopName) <= 14 {
		return stopName
	}
	return stopName[:14]
}

func formatCentsForHumans(costCents int) string {
	if costCents < 0 {
		return "UNREACHABLE"
	}
	return fmt.Sprintf("$%.2f", float64(costCents)/100.0)
}

/* -------------------------------------------------------------------------
Validation helpers
------------------------------------------------------------------------- */

func validateOutputInvariants(
	t *testing.T,
	combinedMatrices CombinedMatrices,
	combineConfig CombineConfig,
) {
	t.Helper()

	stopCount := len(combinedMatrices.Mode)
	if stopCount == 0 {
		t.Fatalf("output matrices should not be empty")
	}

	if len(combinedMatrices.TimeMinutes) != stopCount || len(combinedMatrices.CostCents) != stopCount {
		t.Fatalf("output matrices have mismatched sizes")
	}

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		if len(combinedMatrices.Mode[fromStopIndex]) != stopCount ||
			len(combinedMatrices.TimeMinutes[fromStopIndex]) != stopCount ||
			len(combinedMatrices.CostCents[fromStopIndex]) != stopCount {
			t.Fatalf("output matrices must be NxN")
		}

		if combinedMatrices.Mode[fromStopIndex][fromStopIndex] != selfModeName {
			t.Fatalf("diagonal mode must be %s", selfModeName)
		}

		if combinedMatrices.TimeMinutes[fromStopIndex][fromStopIndex] != 0 {
			t.Fatalf("diagonal time must be 0")
		}

		if combinedMatrices.CostCents[fromStopIndex][fromStopIndex] != 0 {
			t.Fatalf("diagonal cost must be 0")
		}
	}

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		for toStopIndex := 0; toStopIndex < stopCount; toStopIndex++ {
			if fromStopIndex == toStopIndex {
				continue
			}

			switch combinedMatrices.Mode[fromStopIndex][toStopIndex] {
			case walkingModeName:
				if combinedMatrices.CostCents[fromStopIndex][toStopIndex] != 0 {
					t.Fatalf("walking cost must be 0 at (%d,%d)", fromStopIndex, toStopIndex)
				}

			case subwayModeName:
				if combinedMatrices.CostCents[fromStopIndex][toStopIndex] != combineConfig.SubwayFlatFareCents {
					t.Fatalf("subway cost mismatch at (%d,%d)", fromStopIndex, toStopIndex)
				}

			case carModeName:
				if combinedMatrices.CostCents[fromStopIndex][toStopIndex] < combineConfig.CarBaseFareCents {
					t.Fatalf("car cost must be at least the base fare at (%d,%d)", fromStopIndex, toStopIndex)
				}

			case unreachableModeName:
				if combinedMatrices.TimeMinutes[fromStopIndex][toStopIndex] != -1 ||
					combinedMatrices.CostCents[fromStopIndex][toStopIndex] != -1 {
					t.Fatalf("unreachable edges must be stored as -1 / -1 at (%d,%d)", fromStopIndex, toStopIndex)
				}

			default:
				t.Fatalf("unexpected mode value at (%d,%d): %s", fromStopIndex, toStopIndex, combinedMatrices.Mode[fromStopIndex][toStopIndex])
			}
		}
	}
}
