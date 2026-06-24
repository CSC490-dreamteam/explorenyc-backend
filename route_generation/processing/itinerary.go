package processing

import (
	"github.com/CSC490-dreamteam/explorenyc-backend/data/cache"
	maps "github.com/CSC490-dreamteam/explorenyc-backend/integrations/maps"
	"github.com/CSC490-dreamteam/explorenyc-backend/integrations/solver"
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"github.com/CSC490-dreamteam/explorenyc-backend/route_generation/postprocessing"
	"github.com/CSC490-dreamteam/explorenyc-backend/route_generation/preprocessing"
)

// GenerationContext bundles the dependencies shared by every itinerary
// generation request, so handlers build it once and pass it in.
type GenerationContext struct {
	Cache         *cache.Cache
	MapProvider   maps.QueryProvider
	EdgeProviders preprocessing.EdgeProviders
	CombineConfig CombineConfig
	Solver        solver.RouteSolver
}

// GenerateItinerary runs the full pipeline for one itinerary request:
// resolve addresses -> acquire edge weights -> combine into one graph ->
// build the solver payload -> solve -> post-process into an Itinerary.
//
// The returned []error are non-fatal warnings (e.g. a stop that couldn't be
// resolved) that the caller may still want to surface to the user.
func GenerateItinerary(ctx GenerationContext, req ItineraryRequest) (Itinerary, []error, error) {
	resolved, warnings, err := ResolveAddresses(req, ctx.MapProvider)
	if err != nil {
		return Itinerary{}, nil, err
	}

	edgeWeights, selectedTypes, edgeErrs := preprocessing.AcquireEdgeWeights(resolved.Places, req.TransitTypes, ctx.EdgeProviders)
	warnings = append(warnings, edgeErrs...)

	matrices, err := preprocessing.CombineBestEdges(edgeWeights, selectedTypes, ctx.CombineConfig)
	if err != nil {
		warnings = append(warnings, err)
	}

	addressMapping := make(map[int]Address, len(resolved.Places))
	for i, place := range resolved.Places {
		addressMapping[i] = place
	}

	solverInput := preprocessing.BuildSolverInput(req, resolved, matrices)

	solverOutput, err := ctx.Solver.Solve(solverInput)
	if err != nil {
		return Itinerary{}, warnings, err
	}

	itinerary, err := postprocessing.ProcessRouteResponse(PostProcessorInput{
		SolverInput:       solverInput,
		SolverOutput:      solverOutput,
		StopMap:           addressMapping,
		TransitTypeMatrix: matrices.Mode,
		TransitCostMatrix: matrices.CostCents,
	}, *ctx.Cache)
	if err != nil {
		return Itinerary{}, warnings, err
	}

	return itinerary, warnings, nil
}
