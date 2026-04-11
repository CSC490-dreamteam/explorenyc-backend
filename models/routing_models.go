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

// EdgeWeights stores the graph data for one transportation type.
// Nodes is the ordered list of addresses used by the matrices.
// Durations and Distances are directed matrices, meaning A->B can differ from B->A.
type EdgeWeights struct {
	Nodes     []Address
	Durations [][]int // travel times in minutes
	Distances [][]int // travel distances in meters
}

// ---------------- Matrix Aggregator Models ----------------

// CombinedMatrices is the final output of the matrix aggregator.
//
// For every directed edge from one stop to another:
//   - TimeMinutes[from][to] tells us how long the chosen mode takes
//   - CostCents[from][to] tells us the estimated cost of the chosen mode in cents
//   - Mode[from][to] tells us which transportation type was selected
//
// Special values:
//   - Mode[i][i] = "NONE" (diagonal)
//   - Mode[i][j] = "UNREACHABLE" when no mode has a valid path for i -> j
//     (TimeMinutes and CostCents are set to -1 in that case)
type CombinedMatrices struct {
	TimeMinutes [][]int
	CostCents   [][]int
	Mode        [][]TransitType
}

// CombineConfig controls how the aggregator compares one transportation mode
// against another.
//
// Notes:
//   - Money is stored in cents instead of dollars
//   - We compare edges using a "lower is better" value called comparison penalty:
//     comparisonPenalty = costCents + (TimeValueCentsPerMinute * durationMinutes)
//   - Walking is still free, walking cap to avoid choosing unrealistically
//     long walking edges.
type CombineConfig struct {
	// TimeValueCentsPerMinute is the tunable tradeoff knob between time and money.
	// Higher values make faster routes more attractive.
	TimeValueCentsPerMinute int
	// Walking caps. If either cap is exceeded, walking is considered invalid
	// for that edge.
	WalkingMaxMinutes        int
	WalkingMaxDistanceMeters int
	// Subway cost model (assumption: flat fare per edge).
	SubwayFlatFareCents int
	// Car cost model.
	// We have both duration and distance available from EdgeWeights,
	// so car pricing can use both.
	CarBaseFareCents         int
	CarCostPerMinuteCents    int
	CarCostPerKilometerCents int
}

// ---------------- Itinerary Models ----------------

type TransitType int

const (
	Walking TransitType = iota
	Car
	Subway
	//Biking TODO?
	//Uber TODO
)

var stringToTransit = map[string]TransitType{
	"walking": Walking,
	"car":     Car,
	"subway":  Subway,
}
