package main

type Stop struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type Path struct {
	StopOrder []Stop
	TotalTime int //in seconds or soemthing?
}

type Address struct {
	Lat         float64
	Lon         float64
	Street      string
	City        string
	State       string
	Zip         string
	DisplayName string
}
