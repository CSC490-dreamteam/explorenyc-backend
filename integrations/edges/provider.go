package edges

import (
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type EdgeWeights struct {
	Nodes     []Stop
	durations [][]int
	distances [][]int
}

type WalkingMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	GetTravelTime(stops []Stop) (EdgeWeights, error)
}

type CarMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	GetTravelTime(stops []Stop) (EdgeWeights, error)
}

type SubwayMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	GetTravelTime(stops []Stop) (EdgeWeights, error)
}
