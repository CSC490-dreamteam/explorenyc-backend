package matrixAggregator

import (
	"fmt"
	"math"

	"github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// CombineBestEdges combines multiple transportation modes (walking, subway, car, etc.)
// into one optimized set of matrices by selecting the best mode for each directed edge i->j.
//
// Output matrices are aligned:
//   - TimeMinutes[i][j] is the chosen travel time (rounded minutes)
//   - CostDollars[i][j] is the chosen per-edge dollar cost estimate
//   - Mode[i][j] is the chosen mode name ("WALKING", "SUBWAY", "CAR", etc.)
//
// Assumptions/Notes (we can change stuff later as needed):
//   1) All modes use NxN matrices with the SAME stop ordering.
//   2) Off-diagonal 0.0 minutes means "unreachable edge" for that mode.
//   3) Costs are treated as per-edge additive in V1.
//   4) Matrices are treated as directed (i->j may differ from j->i).
//
// Scoring (tunable tradeoff):
//   score = moneyCost + (LambdaDollarsPerMinute * minutes)
// Lower score wins.
func CombineBestEdges(modes []TransitMode, cfg models.CombineConfig) (models.CombinedMatrices, error) {

	if len(modes) == 0 {
		return models.CombinedMatrices{}, fmt.Errorf("no transit modes provided")
	}

	stopCount := modes[0].StopCount()

	if stopCount <= 0 {
		return models.CombinedMatrices{}, fmt.Errorf("invalid stop count")
	}

	// Ensure all matrices are same size
	for i, m := range modes {
		if m.StopCount() != stopCount {
			return models.CombinedMatrices{}, fmt.Errorf(
				"mode %d (%s) has different stop count", i, m.Name())
		}
	}

	timeOut := make([][]int, stopCount)
	costOut := make([][]float64, stopCount)
	modeOut := make([][]string, stopCount)

	for i := 0; i < stopCount; i++ {
		timeOut[i] = make([]int, stopCount)
		costOut[i] = make([]float64, stopCount)
		modeOut[i] = make([]string, stopCount)
	}

	for i := 0; i < stopCount; i++ {

		for j := 0; j < stopCount; j++ {

			// diagonal
			if i == j {
				timeOut[i][j] = 0
				costOut[i][j] = 0
				modeOut[i][j] = "SELF"
				continue
			}

			bestScore := math.Inf(1)
			bestMinutes := 0.0
			bestCost := 0.0
			bestMode := ""

			for _, mode := range modes {

				minutes, ok := mode.Minutes(i, j)

				if !ok {
					continue
				}

				// sanity check for upstream data issues
				if !isSaneMinutes(minutes) {
					continue
				}

				cost := mode.CostDollars(minutes, cfg)

				score := mode.Score(minutes, cost, cfg)

				if score < bestScore {
					bestScore = score
					bestMinutes = minutes
					bestCost = cost
					bestMode = mode.Name()
				}
			}

			if bestMode == "" {

				// no valid transit option
				timeOut[i][j] = -1
				costOut[i][j] = -1
				modeOut[i][j] = "UNREACHABLE"

				continue
			}

			rounded := int(math.Round(bestMinutes))

			if rounded == 0 {
				rounded = 1
			}

			timeOut[i][j] = rounded
			costOut[i][j] = bestCost
			modeOut[i][j] = bestMode
		}
	}

	return models.CombinedMatrices{
		TimeMinutes: timeOut,
		CostDollars: costOut,
		Mode:        modeOut,
	}, nil
}

/*
isSaneMinutes validates travel time values coming from upstream APIs.

This prevents bugs where a route API might return:
- NaN
- Infinity
- negative time
*/
func isSaneMinutes(minutes float64) bool {

	if math.IsNaN(minutes) || math.IsInf(minutes, 0) {
		return false
	}

	if minutes < 0 {
		return false
	}

	return true
}
