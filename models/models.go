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
	Durations [][]int //travel times both ways: A -> B as well as B -> A, since they may differ
	Distances [][]int //distances both ways: A -> B as well as B -> A, since they may differ
}

// ---------------- Matrix Aggregator Models  ----------------

// CombinedMatrices is returned by the matrixAggregator.
//
// Interpretation:
// - Mode[i][j] tells you which transit mode was chosen for edge i -> j.
// - TimeMinutes[i][j] is the travel time (rounded minutes) for that chosen mode.
// - CostDollars[i][j] is the dollar cost estimate for that chosen mode.
//
// Special values:
//   - Mode[i][i] = "NONE" (diagonal)
//   - Mode[i][j] = "UNREACHABLE" when no mode has a valid path for i -> j
//     (TimeMinutes and CostDollars are set to -1 in that case)
type CombinedMatrices struct {
	TimeMinutes [][]int
	CostDollars [][]int
	Mode        [][]TransitType
}

// CombineConfig controls the tunable tradeoff between time and money.
//
// Scoring framework (lower is better):
//
//	score = moneyCost + (LambdaDollarsPerMinute * minutes)
//
// This is a standard "generalized cost" approach that lets us tune
// how much we value time vs saving money.
//
// Notes:
//
// 1) Costs are treated as *per-edge additive*:
//
//   - Walking: $0 per edge (with a walking cap)
//
//   - Subway: flat fare per edge (e.g. $3)
//
//   - Car: base + per-minute per edge
//
//     If later pricing becomes route-level (subway transfer rules, daily passes,
//     surge across the route, etc.) then picking the best mode per edge becomes
//     an approximation. Route-level cost should then be handled in the pathfinding step.
//
// 2) Enforced a cap to avoid "walk everywhere" results because its free lol.
type CombineConfig struct {
	// LambdaDollarsPerMinute is the "tradeoff knob":
	// - Higher => prioritize time more (minutes become expensive)
	// - Lower  => prioritize money more (minutes matter less)
	//
	// Example: 0.25 $/min ≈ $15/hour.
	LambdaDollarsPerMinute float64

	// WalkingMaxMinutes is a hard cap: if walking time exceeds this value,
	// walking is considered NOT allowed for that edge in V1.
	WalkingMaxMinutes int

	// SubwayFlatFareDollars is assumed to be the per-edge subway fare.
	SubwayFlatFareDollars float64

	// CarBaseFareDollars is a pickup-like base fee per edge (Uber-ish).
	CarBaseFareDollars float64

	// CarCostPerMinuteDollars is a simple time-based estimate per edge.
	// (We do not have distance/surge/tolls rn, easy to upgrade later.)
	CarCostPerMinuteDollars float64
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
	TransitCostMatrix [][]int
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

// a group of stops where exactly one is picked by the route to be added in
// i.e Nicks algo suggests 5 taco places for lunch, one of these is selected to be added in the route
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
	//maybe these 3 are overkill idk
	Precedences   [][2]int `json:"precedences"`    //list of pairs of stop indices where the first must come before the second in the route
	ForcedEdges   [][2]int `json:"forced_edges"`   //list of pairs of stop indices where the first must be immediately followed by the second in the route
	ExcludedStops []int    `json:"excluded_stops"` //user rejected stops
}

// segment of a route
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
