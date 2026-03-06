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
	Durations                [][]int //travel times both ways: A -> B as well as B -> A, since they may differ
	Distances                [][]int //distances both ways: A -> B as well as B -> A, since they may differ
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
	Name                    string //incase we call it like "Lunch" or "Museum" instead of the actual name of the stop
	Address                 Address
	ArrivalTimeInMinutes    int64 //clock time you get there, written as minutes from midnight, so 600 is 10 AM
	DepartureTimeInMinutes  int64 //clock time you leave, written as minutes from midnight, so 600 is 10 AM
	DurationAtStopInMinutes int64 //how long you spend at the stop, so DepartureTime - ArrivalTime
	TransportToNextStop     TransitType
	TravelTimeToNextStop    int64
	TransitCost             int64
}

type Itinerary struct {
	Entries                 []ItineraryEntry
	DroppedStops            []Address
	TotalTimeInMinutes      int64
	TotalTransitCostInCents int64
	TotalCostInCents        int64
	StartTimeInMinutes      int64
	EndTimeInMinutes        int64
}

type StopRequest struct {
	Location       string  `json:"location"`
	Mandatory      bool    `json:"mandatory"`
	TimePreference *string `json:"timePreference"` //pointer so it can be null?
}

type ItineraryRequest struct {
	TripName      string        `json:"tripName"`
	Date          string        `json:"date"`
	EntryTime     string        `json:"entryTime"` //09:00 AM
	ExitTime      string        `json:"exitTime"`  //09:00 PM
	StartLocation string        `json:"startLocation"`
	EndLocation   *string       `json:"endLocation"`
	Stops         []StopRequest `json:"stops"`
}

type PostProcessorInput struct {
	SolverInput       SolverInput
	SolverOutput      SolverOutput
	StopMap           map[int]Address
	TransitTypeMatrix [][]TransitType
	TransitCostMatrix [][]int64
}

//////////////////
// solver types //
//////////////////

type Priority int

const (
	Mandatory Priority = iota //iota means the types are automatically assigned values starting from 0, so Mandatory = 0,  MustSee = 2, Optional = 1
	WantToSee                 //higher priority but still optional
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
