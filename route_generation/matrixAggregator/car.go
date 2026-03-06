package matrixAggregator

import "github.com/CSC490-dreamteam/explorenyc-backend/models"

// CarMode implements TransitMode for car.
//
// Cost model (placeholder):
// - cost = cfg.CarBaseFareDollars + cfg.CarCostPerMinuteDollars * minutes
//
// Notes:
// - This is intentionally simple because we only have time matrices rn.
// - If later we have distance, tolls, surge, or a real pricing API,
//   we can replace CostDollars() with a more realistic formula.
type CarMode struct {
	timeMatrix [][]float64
}

func NewCarMode(timeMatrix [][]float64) CarMode {
	return CarMode{timeMatrix: timeMatrix}
}

func (c CarMode) Name() string { return "CAR" }

func (c CarMode) StopCount() int { return len(c.timeMatrix) }

func (c CarMode) Minutes(i, j int) (float64, bool) {
	min := c.timeMatrix[i][j]
	if min <= 0 {
		return 0, false
	}
	return min, true
}

func (c CarMode) CostDollars(minutes float64, cfg models.CombineConfig) float64 {
	return cfg.CarBaseFareDollars + cfg.CarCostPerMinuteDollars*minutes
}

func (c CarMode) Score(minutes float64, cost float64, cfg models.CombineConfig) float64 {
	return cost + cfg.LambdaDollarsPerMinute*minutes
}
