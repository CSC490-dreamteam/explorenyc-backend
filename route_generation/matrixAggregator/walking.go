package matrixAggregator

import (
	"math"

	"github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// WalkingMode implements TransitMode for walking.
//
// Unreachable Convention:
// - Off-diagonal time value of 0.0 means "unreachable / not provided".
// - Diagonal is handled by aggregator (i==j).
//
// Walking Cap:
// - If walk time > cfg.WalkingMaxMinutes, walking is DISALLOWED for that edge.
//   We implement that by returning a huge score.
type WalkingMode struct {
	timeMatrix [][]float64
}

func NewWalkingMode(timeMatrix [][]float64) WalkingMode {
	return WalkingMode{timeMatrix: timeMatrix}
}

func (w WalkingMode) Name() string { return "WALKING" }

func (w WalkingMode) StopCount() int { return len(w.timeMatrix) }

func (w WalkingMode) Minutes(i, j int) (float64, bool) {
	min := w.timeMatrix[i][j]
	// assumption: 0.0 off-diagonal = unreachable
	if min <= 0 {
		return 0, false
	}
	return min, true
}

func (w WalkingMode) CostDollars(minutes float64, cfg models.CombineConfig) float64 {
	_ = minutes
	_ = cfg
	return 0.0
}

func (w WalkingMode) Score(minutes float64, cost float64, cfg models.CombineConfig) float64 {
	// Hard cap to prevent "walk everywhere"
	if cfg.WalkingMaxMinutes > 0 && minutes > float64(cfg.WalkingMaxMinutes) {
		// Using +Inf effectively disqualifies walking without complicating the aggregator.
		return math.Inf(1)
	}
	return cost + cfg.LambdaDollarsPerMinute*minutes
}
