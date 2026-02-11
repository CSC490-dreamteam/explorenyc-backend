package pathfinders

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

// func BruteForcePathFinder(stops []Stop) Path {
// 	//do brute force algo to find best path

// 	numStops := len(stops)

// }
