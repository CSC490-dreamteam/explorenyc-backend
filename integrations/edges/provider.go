package edges

import (
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type WalkingMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireWalkingTravelTime(stops []Stop) (EdgeWeights, error)
}

type CarMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireCarTravelTime(stops []Stop) (EdgeWeights, error)
}

type SubwayMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireSubwayTravelTime(stops []Stop) (EdgeWeights, error)
}

type BikeMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireBikeTravelTime(stops []Stop) (EdgeWeights, error)
}
