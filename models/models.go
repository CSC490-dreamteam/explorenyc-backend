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

// ---------------- Matrix Aggregator Models  ----------------

// CombinedMatrices is returned by the matrixAggregator.
//
// Interpretation:
// - Mode[i][j] tells you which transit mode was chosen for edge i -> j.
// - TimeMinutes[i][j] is the travel time (rounded minutes) for that chosen mode.
// - CostDollars[i][j] is the dollar cost estimate for that chosen mode.
//
// Special values:
// - Mode[i][i] = "NONE" (diagonal)
// - Mode[i][j] = "UNREACHABLE" when no mode has a valid path for i -> j
//   (TimeMinutes and CostDollars are set to -1 in that case)
type CombinedMatrices struct {
	TimeMinutes [][]int
	CostDollars [][]float64
	Mode        [][]string
}

// CombineConfig controls the tunable tradeoff between time and money.
//
// Scoring framework (lower is better):
//   score = moneyCost + (LambdaDollarsPerMinute * minutes)
//
// This is a standard "generalized cost" approach that lets us tune
// how much we value time vs saving money.
//
// Notes:
//
// 1) Costs are treated as *per-edge additive*:
//    - Walking: $0 per edge (with a walking cap)
//    - Subway: flat fare per edge (e.g. $3)
//    - Car: base + per-minute per edge
//
//    If later pricing becomes route-level (subway transfer rules, daily passes,
//    surge across the route, etc.) then picking the best mode per edge becomes
//    an approximation. Route-level cost should then be handled in the pathfinding step.
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
