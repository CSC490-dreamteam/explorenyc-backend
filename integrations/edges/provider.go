package edges

import (
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

type WalkingMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireWalkingTravelTime(addrs []Address) (EdgeWeights, error)
}

type CarMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireCarTravelTime(addrs []Address) (EdgeWeights, error)
}

type SubwayMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireSubwayTravelTime(addrs []Address) (EdgeWeights, error)
}

type BikeMatrixProvider interface {

	//retrieves matrices of travel times between addresses, used for edge generation
	AcquireBikeTravelTime(addrs []Address) (EdgeWeights, error)
}

type SubwayLegProvider interface {
	AcquireSubwayLegs(origin Address, destination Address) ([]Leg, error)
}

type WalkingLegProvider interface {
	AcquireWalkingLeg(origin Address, destination Address) (Leg, error)
}

type CarLegProvider interface {
	AcquireCar(origin Address, destination Address) (Leg, error)
}
