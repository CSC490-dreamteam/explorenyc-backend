package maps

//template for map data providers
type Provider interface {

	//getTravelTime(origin,destination) get edge between 2 addresses

}

type Address struct {
	Lat              float64
	Lon              float64
	Street           string
	City             string
	State            string
	Zip              string
	placeName        string
	formattedAddress string
}
