package pathfinders

import "math"

//possible algo

type Stop struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type Path struct {
	StopOrder []Stop
	TotalTime int //in seconds or soemthing?
}

//bruteforce path calculator with travel time
func BruteForcePathFinder(stops []Stop, travelTimeMatrix [][]int) Path {
	//do brute force algo to find best path

	numStops := len(stops)

	//turn stop sinto indices to use for shuffling
	indices := make([]int, numStops)
	for i := 0; i < numStops; i++ {
		indices[i] = i
	}

	var fastestOrder []int       //stores the currently found fastest path order
	fastestTime := math.MaxInt64 //default to max for starting value

	permutations := findPermutations(indices)

	//start the brute forcing
	for _, perm := range permutations {
		currentTime := calculatePathTime(perm, travelTimeMatrix)

		//if a better time is found, update best time
		if currentTime < fastestTime {
			fastestTime = currentTime
			fastestOrder = make([]int, len(perm))
			copy(fastestOrder, perm)
		}
	}

	//translate indices back into structs and return the best path
	fastestStops := make([]Stop, len(fastestOrder))

	for index, stopNum := range fastestOrder {
		fastestStops[index] = stops[stopNum]
	}
	return Path{
		StopOrder: fastestStops,
		TotalTime: fastestTime,
	}
}

//brute forces a calculation of th ebest route based on travel distance regardless of traffic/ transit time
func BruteForcePathFinderWithDistance(stops []Stop) Path {

	//build distance matrix
	numStops := len(stops)
	distanceMatrix := make([][]float64, numStops)
	for i := range distanceMatrix {
		distanceMatrix[i] = make([]float64, numStops)
		for j := 0; j < numStops; j++ {
			if i == j {
				distanceMatrix[i][j] = 0
			} else {
				distanceMatrix[i][j] = calculateEdgeDistance(
					stops[i].Latitude, stops[i].Longitude,
					stops[j].Latitude, stops[j].Longitude,
				)
			}
		}
	}

	//turn stops into indices for easier calculating
	indices := make([]int, numStops)
	for i := 0; i < numStops; i++ {
		indices[i] = i
	}

	var fastestOrder []int
	fastestDistance := math.MaxFloat64

	permutations := findPermutations(indices)

	//find shortest distance for that perm
	for _, perm := range permutations {
		currentDistance := calculatePathDistance(perm, distanceMatrix)
		if currentDistance < fastestDistance {
			fastestDistance = currentDistance
			fastestOrder = make([]int, len(perm))
			copy(fastestOrder, perm)
		}
	}

	//make slice with the route order to return
	fastestStops := make([]Stop, len(fastestOrder))
	for index, stopNum := range fastestOrder {
		fastestStops[index] = stops[stopNum]
	}

	//return the order
	return Path{
		StopOrder: fastestStops,
		TotalTime: int(fastestDistance), //temporary TODO FIX LATER
	}
}

//straight line distance between 2 points
func calculateEdgeDistance(lat1, lon1, lat2, lon2 float64) float64 {
	latDiff := lat2 - lat1
	lonDiff := lon2 - lon1
	return math.Sqrt(latDiff*latDiff + lonDiff*lonDiff)
}

//calc length of an entire route given the order of stops and the distance matrix
func calculatePathDistance(order []int, distanceMatrix [][]float64) float64 {
	total := 0.0
	for i := 0; i < len(order)-1; i++ {
		from := order[i]
		to := order[i+1]
		total += distanceMatrix[from][to]
	}
	return total
}

//takes a list of numbers and outputs the different permutations of them
func findPermutations(nums []int) [][]int {
	var result [][]int
	var backtrack func(start int)

	backtrack = func(start int) {
		if start == len(nums) {
			permutation := make([]int, len(nums))
			copy(permutation, nums)
			result = append(result, permutation)
			return
		}
		for i := start; i < len(nums); i++ {
			nums[start], nums[i] = nums[i], nums[start]
			backtrack(start + 1)
			nums[start], nums[i] = nums[i], nums[start]
		}
	}

	backtrack(0)
	return result
}

//calculates the total travel time for a sequence of given stops
func calculatePathTime(order []int, travelTimeMatrix [][]int) int {
	totalTime := 0
	for i := 0; i < len(order)-1; i++ {

		from := order[i]
		to := order[i+1]

		totalTime += travelTimeMatrix[from][to]
	}
	return totalTime
}
