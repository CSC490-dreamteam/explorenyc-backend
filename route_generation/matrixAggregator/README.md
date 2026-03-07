# Matrix Aggregator

The Matrix Aggregator combines multiple transportation graphs (walking, subway, car) into a single optimized graph that can be used by the routing service.

For each directed edge `(A → B)`, the aggregator evaluates all available transportation modes and selects the one with the lowest **comparison penalty**.

The result is returned as three aligned matrices representing the best travel option between every pair of nodes.

# Inputs

The aggregator receives routing data in the form of three `EdgeWeights` structs, one for each transportation type (walking, subway, car).

```go
type EdgeWeights struct {
    Nodes     []Address
    Durations [][]int
    Distances [][]int
}
```

## Nodes

`Nodes` is an ordered list of locations used to index the matrices.

```go
Nodes []Address
```

Each node contains address and geographic information.

```go
type Address struct {
    Lat              float64
    Lon              float64
    Street           string
    City             string
    State            string
    Zip              string
    PlaceName        string
    FormattedAddress string
}
```

All input matrices must use the **same node ordering**.


# Durations Matrix

```go
Durations [][]int
```

This matrix represents travel time in **minutes** between nodes.

Example:

```
    A   B   C
A   0  10  20
B  12   0  15
C  18  14   0
```

These matrices are **directed**, meaning travel from A → B may differ from B → A.

# Distances Matrix

```go
Distances [][]int
```

This matrix represents travel distance in **meters** between nodes.

Distance values are primarily used for calculating car travel costs.

# Output

The aggregator returns a struct containing three matrices:

```go
type CombinedMatrices struct {
    TimeMinutes [][]int
    CostCents   [][]int
    Mode        [][]string
}
```

## Time Matrix

Travel time in **minutes** for the selected transportation mode.

```
TimeMinutes[from][to]
```

## Cost Matrix

Estimated cost stored in **integer cents**.

```
CostCents[from][to]
```

Using cents avoids floating point rounding issues.

Example:

```
$3.00 → 300
```

## Mode Matrix

The transportation mode selected for each edge.

Possible values include:

```
WALKING
SUBWAY
CAR
UNREACHABLE
SELF
```

Example:

```
Mode[2][3] = "WALKING"
```

Which means:

```
TimeMinutes[2][3] = walking travel time
CostCents[2][3] = 0
```

# Algorithm

For each directed edge `(A → B)`, the aggregator computes a **comparison penalty** for each transportation mode.

```
comparisonPenalty = costCents + (timeValueCentsPerMinute * durationMinutes)
```

The transportation mode with the **lowest penalty** is selected.

# Cost Models

## Walking

Walking is free:

```
cost = 0
```

Walking is limited by configurable caps:

* `WalkingMaxMinutes`
* `WalkingMaxDistanceMeters`

If either limit is exceeded, walking is not considered for that edge.


## Subway

Subway cost is treated as a **flat fare per edge**.

```
cost = SubwayFlatFareCents
```

This is an approximation since real subway pricing is route-based.

## Car

Car cost is calculated using both time and distance.

```
cost =
    baseFare
  + (minutes * perMinuteCost)
  + (distanceMeters * perKmCost / 1000)
```

# Changes from last iteration

### EdgeWeights Struct Input

The aggregator now accepts `EdgeWeights` structs directly instead of raw matrices. This matches how routing data is retrieved from the Google routing service.

### Costs Stored as Integer Cents

All monetary values are stored as **integer cents** instead of floating-point dollars. This avoids floating point precision issues and follows common backend financial practices.

### Distance-Based Pricing

Distance values are now used in the car cost model to produce more realistic estimates.

### Simplified Architecture

The interface-heavy design was reworked. The aggregator now processes the three `EdgeWeights` inputs directly, simplifying integration with the routing service.

### Naming

Several variables were renamed.

Examples:

```
i, j → fromStopIndex, toStopIndex
score → comparisonPenalty
```

### Reduced Nested Logic

Large nested loops were refactored into smaller helper functions to improve readability and maintainability.

### Input Validation

The aggregator now validates that:

* matrices are NxN
* node counts match across all modes
* matrix dimensions are consistent


# Running the Tests

From the project root directory run:

```
go test -v ./route_generation/matrixAggregator
```

This will execute the matrix aggregator tests and print readable output showing the input matrices and the final optimized matrices.