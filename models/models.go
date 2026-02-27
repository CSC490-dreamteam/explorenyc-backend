package models

type Stop struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type Path struct {
	StopOrder []Stop
	TotalTime int //in seconds or something?
}

type Address struct {
	Lat              float64
	Lon              float64
	Street           string
	City             string
	State            string
	Zip              string
	PlaceName        string // "Empire State Building"
	FormattedAddress string // "Empire State Building, 20 W 34th St., New York, NY 10001"
}

type EdgeWeights struct {
	Nodes     []Stop
	Durations [][]int //travel times both ways: A -> B as well as B -> A, since they may differ
	Distances [][]int //distances both ways: A -> B as well as B -> A, since they may differ
}
