package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"github.com/mmcloughlin/geohash"
	"github.com/redis/go-redis/v9"
)

const (
	edgeKeyPrefix         = "edge:"
	edgeTTL               = 7 * 24 * time.Hour //7 days
	geohashPrecision uint = 8                  // ~38m x 19m, tight enough for NYC
)

type EdgeValue struct {
	DurationMinutes int   `json:"duration_minutes"`
	DistanceM       int   `json:"distance_m"`
	CostCents       int   `json:"cost_cents"`
	Legs            []Leg `json:"legs"` // nil pre-solver, populated post-solver
}

func formEdgeKey(origin Address, mode string, destination Address) string {
	oHash := geohash.EncodeWithPrecision(origin.Lat, origin.Lon, geohashPrecision)
	dHash := geohash.EncodeWithPrecision(destination.Lat, destination.Lon, geohashPrecision)
	return fmt.Sprintf("%s%s:%s:%s", edgeKeyPrefix, oHash, mode, dHash)
}

func (c *Cache) GetEdgeValue(origin Address, transit TransitType, destination Address) (*EdgeValue, error) {
	if c.client == nil {
		return nil, nil
	}

	val, err := c.client.Get(context.Background(), formEdgeKey(origin, transit.String(), destination)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("cache error: edge get: %w", err)
	}

	var result EdgeValue
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, fmt.Errorf("cache error: edge unmarshal: %w", err)
	}
	return &result, nil
}

func (c *Cache) SetEdgeValue(origin Address, transit TransitType, destination Address, edge *EdgeValue) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(edge)
	if err != nil {
		return fmt.Errorf("cache error: edge marshal: %w", err)
	}

	if err := c.client.Set(context.Background(), formEdgeKey(origin, transit.String(), destination), data, edgeTTL).Err(); err != nil {
		return fmt.Errorf("cache error: edge set: %w", err)
	}
	return nil
}
