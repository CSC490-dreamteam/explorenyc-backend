package matrixAggregator

import "github.com/CSC490-dreamteam/explorenyc-backend/models"

// SubwayMode implements TransitMode for subway.
//
// Cost model (placeholder):
// - Flat fare per edge: cfg.SubwayFlatFareDollars
//
// Notes:
// - If later subway pricing becomes route-level (transfers/passes),
//   this per-edge fare assumption becomes an approximation.
type SubwayMode struct {
	timeMatrix [][]float64
}

func NewSubwayMode(timeMatrix [][]float64) SubwayMode {
	return SubwayMode{timeMatrix: timeMatrix}
}

func (s SubwayMode) Name() string { return "SUBWAY" }

func (s SubwayMode) StopCount() int { return len(s.timeMatrix) }

func (s SubwayMode) Minutes(i, j int) (float64, bool) {
	min := s.timeMatrix[i][j]
	if min <= 0 {
		return 0, false
	}
	return min, true
}

func (s SubwayMode) CostDollars(minutes float64, cfg models.CombineConfig) float64 {
	_ = minutes
	return cfg.SubwayFlatFareDollars
}

func (s SubwayMode) Score(minutes float64, cost float64, cfg models.CombineConfig) float64 {
	return cost + cfg.LambdaDollarsPerMinute*minutes
}
