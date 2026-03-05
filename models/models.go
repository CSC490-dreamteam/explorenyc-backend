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
	Nodes     []Address
	Durations [][]int // travel times in minutes
	Distances [][]int // travel distances in meters
}

// ---------------- Matrix Aggregator Models  ----------------

// EdgeWeights stores the graph data for one transportation type.
// Nodes is the ordered list of addresses used by the matrices.
// Durations and Distances are directed matrices, meaning A->B can differ from B->A.
// CombinedMatrices is the final output of the matrix aggregator.

// For every directed edge from one stop to another:
// - TimeMinutes[from][to] tells us how long the chosen mode takes
// - CostCents[from][to] tells us the estimated cost of the chosen mode in cents
// - Mode[from][to] tells us which transportation type was selected
type CombinedMatrices struct {
	TimeMinutes [][]int
	CostCents   [][]int
	Mode        [][]string
}

// CombineConfig controls how the aggregator compares one transportation mode
// against another.

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
	// for that edge in .
	WalkingMaxMinutes        int
	WalkingMaxDistanceMeters int
	// Subway cost model (assumption: flat fare per edge).
	SubwayFlatFareCents int
	// Car cost model.
	// We now have both duration and distance available from EdgeWeights,
	// so car pricing can use both.
	CarBaseFareCents         int
	CarCostPerMinuteCents    int
	CarCostPerKilometerCents int
}

type TransitType int

const (
	Walking TransitType = iota
	Car
	Subway
	//Biking TODO?
	//Uber TODO
)

type ItineraryEntry struct {
	Address                Address
	ArrivalTimeInMinutes   int64
	DepartureTimeInMinutes int64
	TransportToNextStop    TransitType
	TravelTimeToNextStop   int64
	TransitCost            int64
}

//////////////////
// solver types //
//////////////////

type Priority int

const (
	Mandatory Priority = iota //iota means the types are automatically assigned values starting from 0, so Mandatory = 0,  MustSee = 2, Optional = 1
	WantToSee
	Optional
)

type RouteVariant int

const (
	TimeOptimized RouteVariant = iota
	CostOptimized
	Balanced
)

type SolverNode struct {
	ID                string
	Name              string
	Latitude          float64
	Longitude         float64
	DurationInMinutes int64
	TimeWindowStart   int64
	TimeWindowEnd     int64
	Priority          Priority
	DropPenalty       int64
	CandidateGroupID  string
}

// a group of stops where exactly one is picked by the route to be added in
// i.e Nicks algo suggests 5 taco places for lunch, one of these is selected to be added in the route
type CandidateGroup struct {
	ID          string
	StopIndices []int
}

type SolverInput struct {
	Nodes                 []SolverNode
	StartIndex            int
	EndIndex              int
	DayStartTimeInMinutes int64
	DayEndTimeInMinutes   int64
	BudgetInCents         int64
	TravelTimeMatrix      [][]int64
	CostMatrix            [][]int64
	CandidateGroups       []CandidateGroup
	RouteVariant          RouteVariant
	//maybe these 3 are overkill idk
	Precedences   [][2]int //list of pairs of stop indices where the first must come before the second in the route
	ForcedEdges   [][2]int //list of pairs of stop indices where the first must be immediately followed by the second in the route
	ExcludedStops []int    //user rejected stops
}

// segment of a route
type RouteEntry struct {
	NodeIndex              int
	ArrivalTimeInMinutes   int64
	DepartureTimeInMinutes int64
}

type SolverOutput struct {
	Route              []RouteEntry
	DroppedStops       []int
	TotalTimeInMinutes int64
	TotalCostInCents   int64
	Score              int64
	HasSolution        bool
}
