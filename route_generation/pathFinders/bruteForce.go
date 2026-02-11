package pathfinders

import "math"

//possible algo

type Stop struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type Path struct {
	stopOrder []Stop
	totalTime int //in seconds or soemthing?
}

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
		stopOrder: fastestStops,
		totalTime: fastestTime,
	}
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
