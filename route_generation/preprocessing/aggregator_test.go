package preprocessing

//to run test go test -v ./route_generation/preprocessing/matrixaggregator

import (
	"fmt"
	"testing"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

func TestCombineBestEdgesWalkingOnly(t *testing.T) {
	stopNames := []string{
		"Penn Station",
		"Central Park",
		"Broadway",
	}

	nodes := makeDemoNodes(stopNames)

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 12, 20},
			{13, 0, 11},
			{19, 10, 0},
		},
		Distances: [][]int{
			{0, 900, 1500},
			{950, 0, 800},
			{1450, 750, 0},
		},
	}

	combineConfig := demoCombineConfig()

	combinedMatrices, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeights},
		[]TransitType{Walking},
		combineConfig,
	)
	if combineError != nil {
		t.Fatalf("unexpected error: %v", combineError)
	}

	printDemoHeader(t, "walking only")
	printPrettyIntMatrixWithStops(t, "INPUT: Walking durations (minutes)", stopNames, walkingEdgeWeights.Durations)
	printPrettyTransitTypeMatrixWithStops(t, "OUTPUT: Chosen transit type matrix", stopNames, combinedMatrices.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: Chosen time matrix", stopNames, combinedMatrices.TimeMinutes)
	printPrettyCentsMatrixWithStops(t, "OUTPUT: Chosen cost matrix", stopNames, combinedMatrices.CostCents)

	validateBasicOutputInvariants(t, combinedMatrices)

	//because only walking was provided, every non-diagonal edge should resolve to Walking.
	for fromStopIndex := 0; fromStopIndex < len(stopNames); fromStopIndex++ {
		for toStopIndex := 0; toStopIndex < len(stopNames); toStopIndex++ {
			if fromStopIndex == toStopIndex {
				continue
			}

			if combinedMatrices.Mode[fromStopIndex][toStopIndex] != Walking {
				t.Fatalf("expected walking mode at (%d,%d)", fromStopIndex, toStopIndex)
			}

			if combinedMatrices.CostCents[fromStopIndex][toStopIndex] != 0 {
				t.Fatalf("walking cost should be 0 at (%d,%d)", fromStopIndex, toStopIndex)
			}
		}
	}
}

func TestCombineBestEdgesWalkingAndSubway(t *testing.T) {
	stopNames := []string{
		"Penn Station",
		"Central Park",
		"Broadway",
	}

	nodes := makeDemoNodes(stopNames)

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 34, 18},
			{35, 0, 22},
			{19, 23, 0},
		},
		Distances: [][]int{
			{0, 2800, 1400},
			{2900, 0, 1800},
			{1450, 1900, 0},
		},
	}

	subwayEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 15, 10},
			{14, 0, 12},
			{11, 13, 0},
		},
		Distances: [][]int{
			{0, 5000, 2800},
			{5100, 0, 3000},
			{2900, 3200, 0},
		},
	}

	combineConfig := demoCombineConfig()

	combinedMatrices, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeights, subwayEdgeWeights},
		[]TransitType{Walking, Subway},
		combineConfig,
	)
	if combineError != nil {
		t.Fatalf("unexpected error: %v", combineError)
	}

	printDemoHeader(t, "walking + subway")
	printPrettyIntMatrixWithStops(t, "INPUT: Walking durations (minutes)", stopNames, walkingEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Subway durations (minutes)", stopNames, subwayEdgeWeights.Durations)
	printPrettyTransitTypeMatrixWithStops(t, "OUTPUT: Chosen transit type matrix", stopNames, combinedMatrices.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: Chosen time matrix", stopNames, combinedMatrices.TimeMinutes)
	printPrettyCentsMatrixWithStops(t, "OUTPUT: Chosen cost matrix", stopNames, combinedMatrices.CostCents)

	validateBasicOutputInvariants(t, combinedMatrices)

	//penn station -> central park walking exceeds the configured walking caps, so subway should win on that edge.
	if combinedMatrices.Mode[0][1] != Subway {
		t.Fatalf("expected subway from Penn Station to Central Park")
	}
}

func TestCombineBestEdgesCarThenWalkingOrderStillWorks(t *testing.T) {
	stopNames := []string{
		"Penn Station",
		"Central Park",
		"Broadway",
	}

	nodes := makeDemoNodes(stopNames)

	carEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 10, 7},
			{11, 0, 8},
			{8, 9, 0},
		},
		Distances: [][]int{
			{0, 4200, 1800},
			{4300, 0, 2000},
			{1900, 2100, 0},
		},
	}

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 12, 9},
			{13, 0, 10},
			{9, 11, 0},
		},
		Distances: [][]int{
			{0, 950, 700},
			{1000, 0, 800},
			{720, 850, 0},
		},
	}

	combineConfig := demoCombineConfig()

	combinedMatrices, combineError := CombineBestEdges(
		[]EdgeWeights{carEdgeWeights, walkingEdgeWeights},
		[]TransitType{Car, Walking},
		combineConfig,
	)
	if combineError != nil {
		t.Fatalf("unexpected error: %v", combineError)
	}

	printDemoHeader(t, "car + walking (reversed input order)")
	printPrettyIntMatrixWithStops(t, "INPUT: Car durations (minutes)", stopNames, carEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Walking durations (minutes)", stopNames, walkingEdgeWeights.Durations)
	printPrettyTransitTypeMatrixWithStops(t, "OUTPUT: Chosen transit type matrix", stopNames, combinedMatrices.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: Chosen time matrix", stopNames, combinedMatrices.TimeMinutes)
	printPrettyCentsMatrixWithStops(t, "OUTPUT: Chosen cost matrix", stopNames, combinedMatrices.CostCents)

	validateBasicOutputInvariants(t, combinedMatrices)

	//the order of the input slices should not break the logic.
	//since walking is still cheap and within caps, it should win on short edges.
	if combinedMatrices.Mode[0][2] != Walking {
		t.Fatalf("expected walking from Penn Station to Broadway")
	}
}

func TestCombineBestEdgesAllTransitTypes(t *testing.T) {
	stopNames := []string{
		"Penn Station",
		"Central Park",
		"Broadway",
		"Museum",
	}

	nodes := makeDemoNodes(stopNames)

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 35, 18, 40},
			{34, 0, 20, 14},
			{19, 21, 0, 24},
			{39, 15, 23, 0},
		},
		Distances: [][]int{
			{0, 2800, 1400, 3200},
			{2700, 0, 1600, 1100},
			{1450, 1700, 0, 2000},
			{3100, 1050, 1950, 0},
		},
	}

	subwayEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 15, 10, 19},
			{14, 0, 11, 8},
			{11, 12, 0, 10},
			{20, 9, 11, 0},
		},
		Distances: [][]int{
			{0, 5200, 2800, 6000},
			{5100, 0, 3000, 1900},
			{2900, 3100, 0, 2500},
			{6100, 2000, 2600, 0},
		},
	}

	carEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 12, 8, 14},
			{13, 0, 9, 7},
			{8, 9, 0, 8},
			{15, 8, 9, 0},
		},
		Distances: [][]int{
			{0, 4300, 1800, 5000},
			{4400, 0, 2000, 1300},
			{1900, 2100, 0, 1600},
			{5100, 1400, 1700, 0},
		},
	}

	combineConfig := demoCombineConfig()

	combinedMatrices, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeights, subwayEdgeWeights, carEdgeWeights},
		[]TransitType{Walking, Subway, Car},
		combineConfig,
	)
	if combineError != nil {
		t.Fatalf("unexpected error: %v", combineError)
	}

	printDemoHeader(t, "walking + subway + car")
	printPrettyIntMatrixWithStops(t, "INPUT: Walking durations (minutes)", stopNames, walkingEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Subway durations (minutes)", stopNames, subwayEdgeWeights.Durations)
	printPrettyIntMatrixWithStops(t, "INPUT: Car durations (minutes)", stopNames, carEdgeWeights.Durations)
	printPrettyTransitTypeMatrixWithStops(t, "OUTPUT: Chosen transit type matrix", stopNames, combinedMatrices.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: Chosen time matrix", stopNames, combinedMatrices.TimeMinutes)
	printPrettyCentsMatrixWithStops(t, "OUTPUT: Chosen cost matrix", stopNames, combinedMatrices.CostCents)

	validateBasicOutputInvariants(t, combinedMatrices)
}

func TestCombineBestEdgesRejectsMismatchedSliceLengths(t *testing.T) {
	nodes := make([]Address, 2)

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 10},
			{11, 0},
		},
		Distances: [][]int{
			{0, 900},
			{950, 0},
		},
	}

	_, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeights},
		[]TransitType{Walking, Subway},
		demoCombineConfig(),
	)
	if combineError == nil {
		t.Fatalf("expected an error for mismatched slice lengths")
	}
}

func TestCombineBestEdgesRejectsDuplicateTransitTypes(t *testing.T) {
	nodes := make([]Address, 2)

	walkingEdgeWeightsA := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 10},
			{11, 0},
		},
		Distances: [][]int{
			{0, 900},
			{950, 0},
		},
	}

	walkingEdgeWeightsB := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 12},
			{13, 0},
		},
		Distances: [][]int{
			{0, 1000},
			{1100, 0},
		},
	}

	_, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeightsA, walkingEdgeWeightsB},
		[]TransitType{Walking, Walking},
		demoCombineConfig(),
	)
	if combineError == nil {
		t.Fatalf("expected an error for duplicate transit types")
	}
}

func TestCombineBestEdgesRejectsUnsupportedTransitTypes(t *testing.T) {
	nodes := make([]Address, 2)

	walkingEdgeWeights := EdgeWeights{
		Nodes: nodes,
		Durations: [][]int{
			{0, 10},
			{11, 0},
		},
		Distances: [][]int{
			{0, 900},
			{950, 0},
		},
	}

	unsupportedTransitType := TransitType(999)

	_, combineError := CombineBestEdges(
		[]EdgeWeights{walkingEdgeWeights},
		[]TransitType{unsupportedTransitType},
		demoCombineConfig(),
	)
	if combineError == nil {
		t.Fatalf("expected an error for unsupported transit type")
	}
}

//demo helpers lol

func demoCombineConfig() CombineConfig {
	return CombineConfig{
		TimeValueCentsPerMinute:  25,
		WalkingMaxMinutes:        25,
		WalkingMaxDistanceMeters: 2000,
		SubwayFlatFareCents:      300,
		CarBaseFareCents:         250,
		CarCostPerMinuteCents:    12,
		CarCostPerKilometerCents: 50,
	}
}

func makeDemoNodes(stopNames []string) []Address {
	nodes := make([]Address, len(stopNames))

	for stopIndex, stopName := range stopNames {
		nodes[stopIndex] = Address{
			PlaceName:        stopName,
			FormattedAddress: stopName,
		}
	}

	return nodes
}

func printDemoHeader(t *testing.T, testName string) {
	t.Helper()
	t.Log("")
	t.Log("=======================================================================")
	t.Logf(" MATRIX AGGREGATOR DEMO: %s", testName)
	t.Log("=======================================================================")
	t.Log("The aggregator receives a dynamic list of transit graphs plus a matching")
	t.Log("list of TransitType enums, then chooses the lowest-penalty option per edge.")
	t.Log("=======================================================================")
}

func printPrettyIntMatrixWithStops(
	t *testing.T,
	title string,
	stopNames []string,
	matrix [][]int,
) {
	t.Helper()
	t.Log("")
	t.Log(title)

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
	t.Log(title)

	headerRow := fmt.Sprintf("%-16s |", "From \\ To")
	for _, stopName := range stopNames {
		headerRow += fmt.Sprintf(" %-14s |", shortenStopName(stopName))
	}
	t.Log(headerRow)

	for fromStopIndex := 0; fromStopIndex < len(matrix); fromStopIndex++ {
		rowText := fmt.Sprintf("%-16s |", shortenStopName(stopNames[fromStopIndex]))
		for toStopIndex := 0; toStopIndex < len(matrix[fromStopIndex]); toStopIndex++ {
			rowText += fmt.Sprintf(" %-14s |", formatCentsForHumans(matrix[fromStopIndex][toStopIndex]))
		}
		t.Log(rowText)
	}
}

func printPrettyTransitTypeMatrixWithStops(
	t *testing.T,
	title string,
	stopNames []string,
	matrix [][]TransitType,
) {
	t.Helper()
	t.Log("")
	t.Log(title)

	headerRow := fmt.Sprintf("%-16s |", "From \\ To")
	for _, stopName := range stopNames {
		headerRow += fmt.Sprintf(" %-14s |", shortenStopName(stopName))
	}
	t.Log(headerRow)

	for fromStopIndex := 0; fromStopIndex < len(matrix); fromStopIndex++ {
		rowText := fmt.Sprintf("%-16s |", shortenStopName(stopNames[fromStopIndex]))
		for toStopIndex := 0; toStopIndex < len(matrix[fromStopIndex]); toStopIndex++ {
			if fromStopIndex == toStopIndex {
				rowText += fmt.Sprintf(" %-14s |", "Diagonal")
				continue
			}
			rowText += fmt.Sprintf(" %-14s |", transitTypeToLabel(matrix[fromStopIndex][toStopIndex]))
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
	return fmt.Sprintf("$%.2f", float64(costCents)/100.0)
}

func transitTypeToLabel(transitType TransitType) string {
	switch transitType {
	case Walking:
		return "Walking"
	case Car:
		return "Car"
	case Subway:
		return "Subway"
	default:
		return fmt.Sprintf("Transit(%d)", transitType)
	}
}

func validateBasicOutputInvariants(t *testing.T, combinedMatrices CombinedMatrices) {
	t.Helper()

	stopCount := len(combinedMatrices.TimeMinutes)
	if stopCount == 0 {
		t.Fatalf("output matrices should not be empty")
	}

	if len(combinedMatrices.CostCents) != stopCount || len(combinedMatrices.Mode) != stopCount {
		t.Fatalf("output matrices have mismatched row counts")
	}

	for fromStopIndex := 0; fromStopIndex < stopCount; fromStopIndex++ {
		if len(combinedMatrices.TimeMinutes[fromStopIndex]) != stopCount ||
			len(combinedMatrices.CostCents[fromStopIndex]) != stopCount ||
			len(combinedMatrices.Mode[fromStopIndex]) != stopCount {
			t.Fatalf("output matrices must be NxN")
		}
	}
}
