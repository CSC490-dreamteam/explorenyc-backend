package preprocessing

import . "github.com/CSC490-dreamteam/explorenyc-backend/models"

// transitCostFunction calculates the cost in cents for one edge of one transit type.
type transitCostFunction func(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int

// getTransitCostFunction returns the registered cost function for a transit type.
func getTransitCostFunction(transitType TransitType) transitCostFunction {
	return transitCostFunctions()[transitType]
}

// hasTransitCostFunction checks whether a transit type has a cost function registered.
// This is used during validation to reject unsupported enum values early.
func hasTransitCostFunction(transitType TransitType) bool {
	_, exists := transitCostFunctions()[transitType]
	return exists
}

// transitCostFunctions returns the map of supported cost calculators.
// Using a map here matches the new spec's move toward dynamic transit handling.
func transitCostFunctions() map[TransitType]transitCostFunction {
	return map[TransitType]transitCostFunction{
		Walking: calculateWalkingCostInCents,
		Car:     calculateCarCostInCents,
		Subway:  calculateSubwayCostInCents,
	}
}

// calculateWalkingCostInCents keeps walking free.
func calculateWalkingCostInCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	_ = durationMinutes
	_ = distanceMeters
	_ = combineConfig
	return 0
}

// calculateSubwayCostInCents treats subway as a flat fare per edge.
func calculateSubwayCostInCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	_ = durationMinutes
	_ = distanceMeters
	return combineConfig.SubwayFlatFareCents
}

// calculateCarCostInCents uses both duration and distance.
//
// Formula:
//
//	base fare
//
// + per-minute charge
// + per-kilometer charge
//
// Distance is stored in meters, so integer division is used for conversion.
func calculateCarCostInCents(
	durationMinutes int,
	distanceMeters int,
	combineConfig CombineConfig,
) int {
	distanceCostInCents := (distanceMeters * combineConfig.CarCostPerKilometerCents) / 1000
	durationCostInCents := durationMinutes * combineConfig.CarCostPerMinuteCents

	return combineConfig.CarBaseFareCents + durationCostInCents + distanceCostInCents
}
