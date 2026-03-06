package matrixAggregator

import (
	"fmt"
	"testing"

	"github.com/CSC490-dreamteam/explorenyc-backend/models"
)

func TestCombineBestEdges_LaymanDemo(t *testing.T) {
	// These are "made-up but plausible" NYC-ish travel times (minutes).
	// They are NOT meant to be perfect real-world numbers—just realistic enough
	// that the chosen modes make intuitive sense to a human reviewer.

	stops := []string{
		"Penn Station",
		"Central Park",
		"Ice Cream Shop",
		"Broadway",
	}

	// IMPORTANT V1 convention:
	// - Diagonal is 0 (same stop)
	// - Off-diagonal 0 means "unreachable / not available for this mode"
	//
	// NOTE: In real life walking is basically always reachable, but we include one
	// unreachable walking edge here to demonstrate the behavior.

	walk := [][]float64{
		{0, 35, 12, 28}, // Penn -> Central is far on foot
		{34, 0, 22, 18},
		{12, 21, 0, 17},
		{26, 18, 16, 0},
	}

	subway := [][]float64{
		{0, 18, 0, 11}, // Penn -> Ice Cream is "no direct subway" in this toy example
		{17, 0, 14, 12},
		{0, 14, 0, 9},
		{11, 12, 10, 0},
	}

	car := [][]float64{
		{0, 12, 7, 9},
		{13, 0, 10, 11},
		{8, 11, 0, 6},
		{9, 12, 7, 0},
	}

	// Tuning:
	// lambda controls how much we "value time" in $/min terms.
	//
	// With lambda=0.25, 10 minutes "costs" 2.5 score points even if $0.
	// WalkingMaxMinutes prevents the system from picking huge walking edges just because they're free.
	cfg := models.CombineConfig{
		LambdaDollarsPerMinute: 0.25,
		WalkingMaxMinutes:      25,

		// Cost assumptions:
		SubwayFlatFareDollars:   3.00,
		CarBaseFareDollars:      5.00,
		CarCostPerMinuteDollars: 0.20,
	}

	modes := []TransitMode{
		NewWalkingMode(walk),
		NewSubwayMode(subway),
		NewCarMode(car),
	}

	out, err := CombineBestEdges(modes, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// -------------------- Demo header --------------------
	t.Log("")
	t.Log("=======================================================================")
	t.Log(" MATRIX AGGREGATOR DEMO")
	t.Log("")
	t.Log(" What this does:")
	t.Log(" - You give 3 time matrices: WALKING, SUBWAY, CAR (minutes from Stop A -> Stop B)")
	t.Log(" - For each pair (A -> B), we choose ONE best option using a tunable tradeoff:")
	t.Log("       score = dollarCost + (lambda * minutes)")
	t.Log("   Lower score wins.")
	t.Log("")
	t.Log(" What we output:")
	t.Log(" - Mode Matrix: which transport we picked for each A -> B")
	t.Log(" - Time Matrix: travel time (minutes) for the chosen mode")
	t.Log(" - Cost Matrix: dollar cost for the chosen mode")
	t.Log("=======================================================================")

	t.Logf("Config: lambda=%.2f, walkCap=%d min, subwayFare=$%.2f, carBase=$%.2f, carPerMin=$%.2f",
		cfg.LambdaDollarsPerMinute,
		cfg.WalkingMaxMinutes,
		cfg.SubwayFlatFareDollars,
		cfg.CarBaseFareDollars,
		cfg.CarCostPerMinuteDollars,
	)

	// -------------------- Show inputs --------------------
	printPrettyFloatMatrixWithStops(t, "INPUT: WALKING time matrix (minutes)", stops, walk)
	printPrettyFloatMatrixWithStops(t, "INPUT: SUBWAY time matrix (minutes)", stops, subway)
	printPrettyFloatMatrixWithStops(t, "INPUT: CAR time matrix (minutes)", stops, car)

	// -------------------- Show outputs --------------------
	printPrettyStringMatrixWithStops(t, "OUTPUT: MODE matrix (chosen transport for each trip A -> B)", stops, out.Mode)
	printPrettyIntMatrixWithStops(t, "OUTPUT: TIME matrix (minutes for chosen mode)", stops, out.TimeMinutes)
	printPrettyFloatMatrixWithStops(t, "OUTPUT: COST matrix ($ for chosen mode)", stops, out.CostDollars)

	// -------------------- Minimal invariants --------------------
	validateOutputInvariants(t, out, cfg)
}

/* -------------------------------------------------------------------------
Pretty printing helpers (tables labeled by stop names)
------------------------------------------------------------------------- */

func printPrettyStringMatrixWithStops(t *testing.T, title string, stops []string, m [][]string) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	n := len(stops)

	// Header row
	header := fmt.Sprintf("%-14s |", "From \\ To")
	for j := 0; j < n; j++ {
		header += fmt.Sprintf(" %-14s |", shortStop(stops[j]))
	}
	t.Log(header)

	// Rows
	for i := 0; i < n; i++ {
		row := fmt.Sprintf("%-14s |", shortStop(stops[i]))
		for j := 0; j < n; j++ {
			row += fmt.Sprintf(" %-14s |", m[i][j])
		}
		t.Log(row)
	}
}

func printPrettyIntMatrixWithStops(t *testing.T, title string, stops []string, m [][]int) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	n := len(stops)

	header := fmt.Sprintf("%-14s |", "From \\ To")
	for j := 0; j < n; j++ {
		header += fmt.Sprintf(" %-14s |", shortStop(stops[j]))
	}
	t.Log(header)

	for i := 0; i < n; i++ {
		row := fmt.Sprintf("%-14s |", shortStop(stops[i]))
		for j := 0; j < n; j++ {
			row += fmt.Sprintf(" %-14d |", m[i][j])
		}
		t.Log(row)
	}
}

func printPrettyFloatMatrixWithStops(t *testing.T, title string, stops []string, m [][]float64) {
	t.Helper()
	t.Log("")
	t.Log("------------------------------------------------------------")
	t.Log(title)
	t.Log("------------------------------------------------------------")

	n := len(stops)

	header := fmt.Sprintf("%-14s |", "From \\ To")
	for j := 0; j < n; j++ {
		header += fmt.Sprintf(" %-14s |", shortStop(stops[j]))
	}
	t.Log(header)

	for i := 0; i < n; i++ {
		row := fmt.Sprintf("%-14s |", shortStop(stops[i]))
		for j := 0; j < n; j++ {
			row += fmt.Sprintf(" %-14.1f |", m[i][j])
		}
		t.Log(row)
	}
}

func shortStop(name string) string {
	// Keeps the table columns clean.
	// Adjust as you like, this is just to avoid mega wide tables.
	if len(name) <= 14 {
		return name
	}
	return name[:14]
}

/* -------------------------------------------------------------------------
Validation helpers
------------------------------------------------------------------------- */

func validateOutputInvariants(t *testing.T, out models.CombinedMatrices, cfg models.CombineConfig) {
	t.Helper()

	n := len(out.Mode)
	if n == 0 {
		t.Fatalf("output matrices should not be empty")
	}

	// shape checks
	if len(out.TimeMinutes) != n || len(out.CostDollars) != n {
		t.Fatalf("output matrices are mismatched sizes")
	}
	for i := 0; i < n; i++ {
		if len(out.Mode[i]) != n || len(out.TimeMinutes[i]) != n || len(out.CostDollars[i]) != n {
			t.Fatalf("output matrices must be NxN")
		}
	}

	// diagonal checks + mode/cost consistency
	for i := 0; i < n; i++ {
		if out.Mode[i][i] != "SELF" {
			t.Fatalf("diagonal mode expected SELF at (%d,%d), got %s", i, i, out.Mode[i][i])
		}
		if out.TimeMinutes[i][i] != 0 || out.CostDollars[i][i] != 0 {
			t.Fatalf("diagonal expected time=0 cost=0 at (%d,%d)", i, i)
		}
	}

	// basic mode/cost consistency
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			switch out.Mode[i][j] {
			case "WALKING":
				if out.CostDollars[i][j] != 0 {
					t.Fatalf("walking cost must be 0 at (%d,%d)", i, j)
				}
				if out.TimeMinutes[i][j] <= 0 {
					t.Fatalf("walking time must be >0 at (%d,%d)", i, j)
				}

			case "SUBWAY":
				if out.CostDollars[i][j] != cfg.SubwayFlatFareDollars {
					t.Fatalf("subway cost must be %.2f at (%d,%d)", cfg.SubwayFlatFareDollars, i, j)
				}
				if out.TimeMinutes[i][j] <= 0 {
					t.Fatalf("subway time must be >0 at (%d,%d)", i, j)
				}

			case "CAR":
				if out.CostDollars[i][j] < cfg.CarBaseFareDollars {
					t.Fatalf("car cost must be >= base fare at (%d,%d)", i, j)
				}
				if out.TimeMinutes[i][j] <= 0 {
					t.Fatalf("car time must be >0 at (%d,%d)", i, j)
				}

			case "UNREACHABLE":
				if out.TimeMinutes[i][j] != -1 || out.CostDollars[i][j] != -1 {
					t.Fatalf("unreachable must be time=-1 cost=-1 at (%d,%d)", i, j)
				}

			default:
				t.Fatalf("unexpected mode value at (%d,%d): %s", i, j, out.Mode[i][j])
			}
		}
	}
}
