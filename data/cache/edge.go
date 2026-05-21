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

func formEdgeKey(originLat, originLon float64, mode string, destLat, destLon float64) string {
	oHash := geohash.EncodeWithPrecision(originLat, originLon, geohashPrecision)
	dHash := geohash.EncodeWithPrecision(destLat, destLon, geohashPrecision)
	return fmt.Sprintf("%s%s:%s:%s", edgeKeyPrefix, oHash, mode, dHash)
}

func (c *Cache) GetEdgeValue(originLat, originLon float64, transit TransitType, destLat, destLon float64) (*EdgeValue, error) {
	if c.client == nil {
		return nil, nil
	}

	val, err := c.client.Get(context.Background(), formEdgeKey(originLat, originLon, transit.String(), destLat, destLon)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("cache: edge get: %w", err)
	}

	var result EdgeValue
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, fmt.Errorf("cache: edge unmarshal: %w", err)
	}
	return &result, nil
}

func (c *Cache) SetEdgeValue(originLat, originLon float64, transit TransitType, destLat, destLon float64, edge *EdgeValue) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(edge)
	if err != nil {
		return fmt.Errorf("cache: edge marshal: %w", err)
	}

	if err := c.client.Set(context.Background(), formEdgeKey(originLat, originLon, transit.String(), destLat, destLon), data, edgeTTL).Err(); err != nil {
		return fmt.Errorf("cache: edge set: %w", err)
	}
	return nil
}
