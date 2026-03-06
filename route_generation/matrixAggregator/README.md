# matrixAggregator (V1)

This package combines multiple transit time matrices (walking/subway/car/etc.) into a single optimized graph.

For every directed edge `i -> j`, it selects the transit mode with the lowest score and returns:
- `TimeMinutes[i][j]` (rounded int minutes)
- `CostDollars[i][j]` (estimated $)
- `Mode[i][j]` ("WALKING", "SUBWAY", "CAR", "UNREACHABLE", "SELF")

## Scoring (tunable tradeoff)
We use a generalized cost score:

`score = moneyCost + (LambdaDollarsPerMinute * minutes)`

Lower score wins.

## Assumptions / Notes
- Matrices are `NxN` and share the same stop list in the same index order.
- Off-diagonal `0.0` means "unreachable edge" for a given transit type.
- Costs are treated as **per-edge additive** in V1:
    - walking: $0 (with a walking time cap)
    - subway: flat fare per edge
    - car: base + per-minute
- If later pricing becomes route-level (transfers/passes), per-edge mode selection becomes an approximation.

## Test
Run:
`go test -v ./route_generation/matrixAggregator`