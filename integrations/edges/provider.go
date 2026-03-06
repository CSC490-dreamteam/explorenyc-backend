package edges

import (
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type WalkingMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireWalkingTravelTime(stops []Address) (EdgeWeights, error)
}

type CarMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireCarTravelTime(stops []Address) (EdgeWeights, error)
}

type SubwayMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireSubwayTravelTime(stops []Address) (EdgeWeights, error)
}

type BikeMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireBikeTravelTime(stops []Address) (EdgeWeights, error)
}
