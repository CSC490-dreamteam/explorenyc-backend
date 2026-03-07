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
	Mode        [][]string
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

type ItineraryEntry struct {
	Name                    string //incase we call it like "Lunch" or "Museum" instead of the actual name of the stop
	Address                 Address
	ArrivalTimeInMinutes    int //clock time you get there, written as minutes from midnight, so 600 is 10 AM
	DepartureTimeInMinutes  int //clock time you leave, written as minutes from midnight, so 600 is 10 AM
	DurationAtStopInMinutes int //how long you spend at the stop, so DepartureTime - ArrivalTime
	TransportToNextStop     TransitType
	TravelTimeToNextStop    int
	TransitCost             int
}

type Itinerary struct {
	Entries                 []ItineraryEntry
	DroppedStops            []Address
	TotalTimeInMinutes      int
	TotalTransitCostInCents int
	TotalCostInCents        int
	StartTimeInMinutes      int
	EndTimeInMinutes        int
}

// ---------------- API Request Models ----------------

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

// ---------------- Post Processor ----------------

type PostProcessorInput struct {
	SolverInput       SolverInput
	SolverOutput      SolverOutput
	StopMap           map[int]Address
	TransitTypeMatrix [][]TransitType
	TransitCostMatrix [][]int
}

//////////////////
// solver types //
//////////////////

type Priority int

const (
	Mandatory Priority = iota //Mandatory = 0
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
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	DurationInMinutes int      `json:"duration_in_minutes"`
	TimeWindowStart   int      `json:"time_window_start"`
	TimeWindowEnd     int      `json:"time_window_end"`
	Priority          Priority `json:"Priority"`
	DropPenalty       int      `json:"drop_penalty"`
	CandidateGroupID  string   `json:"candidate_group_id,omitempty"`
}

// A group of stops where exactly one is picked by the route to be added in.
// i.e. Nicks algo suggests 5 taco places for lunch, one of these is selected to be added in the route.
type CandidateGroup struct {
	ID          string `json:"id"`
	StopIndices []int  `json:"stop_indices"`
}

type SolverInput struct {
	Nodes                 []SolverNode     `json:"nodes"`
	StartIndex            int              `json:"start_index"`
	EndIndex              int              `json:"end_index"`
	DayStartTimeInMinutes int              `json:"day_start_time_in_minutes"`
	DayEndTimeInMinutes   int              `json:"day_end_time_in_minutes"`
	BudgetInCents         int              `json:"budget_in_cents"`
	TravelTimeMatrix      [][]int          `json:"travel_time_matrix_in_minutes"`
	CostMatrix            [][]int          `json:"travel_cost_matrix_in_cents"`
	CandidateGroups       []CandidateGroup `json:"candidate_groups"`
	RouteVariant          RouteVariant     `json:"route_variant"`
	Precedences           [][2]int         `json:"precedences"`    //pairs where first must come before second
	ForcedEdges           [][2]int         `json:"forced_edges"`   //pairs where first must be immediately followed by second
	ExcludedStops         []int            `json:"excluded_stops"` //user rejected stops
}

// RouteEntry is a segment of a route.
type RouteEntry struct {
	NodeIndex              int `json:"node_index"`
	ArrivalTimeInMinutes   int `json:"arrival_time_in_minutes"`
	DepartureTimeInMinutes int `json:"departure_time_in_minutes"`
}

type SolverOutput struct {
	Route              []RouteEntry `json:"route"`
	DroppedStops       []int        `json:"dropped_stops"`
	TotalTimeInMinutes int          `json:"total_time_in_minutes"`
	TotalCostInCents   int          `json:"total_cost_in_cents"`
	Score              int          `json:"score"`
	HasSolution        bool         `json:"has_solution"`
}
