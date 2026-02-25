package maps

import (
	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
)

// template for map data providers
type Provider interface {
	AcquireAddress(query string) (Address, error)
}
