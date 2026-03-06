package matrixAggregator

import "github.com/CSC490-dreamteam/explorenyc-backend/models"

// TransitMode is the extension point.
// To add a new transit type later (ex: BIKING, UBER), create a new file that
// defines a struct implementing this interface, then include it in the modes slice.
//
// This keeps the main aggreator clean and extensible.
type TransitMode interface {
	// Name is written into the output Mode matrix. Treat like an enum string.
	Name() string

	// StopCount returns N for an NxN matrix. Aggregator validates consistency.
	StopCount() int

	// Minutes returns the time (in minutes) for edge i->j.
	// ok=false means this mode cannot travel i->j (unreachable / missing data).
	Minutes(i, j int) (minutes float64, ok bool)

	// CostDollars returns the per-edge dollar cost estimate for this mode.
	// V1 assumes per-edge additivity (review notes in models.CombineConfig).
	CostDollars(minutes float64, cfg models.CombineConfig) float64

	// Score returns a comparable score across ALL modes.
	// Lower score = better.
	Score(minutes float64, cost float64, cfg models.CombineConfig) float64
}
