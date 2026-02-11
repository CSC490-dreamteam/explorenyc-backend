package route_generation

//file for inteface for each TSP algo

//import "context"

//context is something for if the server sends a cancellation error

type PathFinder interface {
	FindPath()
}

type PathFinderInput struct {
	//stops
	//edges
	//constraints?? criteria?

}
