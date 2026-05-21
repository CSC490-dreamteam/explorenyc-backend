package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	. "github.com/CSC490-dreamteam/explorenyc-backend/models"
	"github.com/redis/go-redis/v9"
)

const (
	geocodeKeyPrefix = "geo:"
	geocodeTTL       = 30 * 24 * time.Hour //30 days
)

var multiSpace = regexp.MustCompile(`\s+`)

// makes string all lowercase,
// trims leading and ending spaces,
// removes punctuation,
// and consolidates multiples spaces into single spaces
func normalizeAddress(addr string) string {
	s := strings.ToLower(strings.TrimSpace(addr))
	s = strings.Map(func(r rune) rune {
		switch r {
		case ',', '.', '#', '!', '?', ';', ':':
			return -1
		default:
			return r
		}
	}, s)
	return strings.TrimSpace(multiSpace.ReplaceAllString(s, " "))
}

// makes the key when given a plain string such as 'empire state building'
func formGeocodeKey(addrString string) string {
	hash := sha256.Sum256([]byte(normalizeAddress(addrString)))
	return fmt.Sprintf("%s%x", geocodeKeyPrefix, hash)
}

// gets redis value
func (c *Cache) GetGeocodeValue(addrString string) (*Address, error) {
	if c.client == nil {
		return nil, nil
	}

	val, err := c.client.Get(context.Background(), formGeocodeKey(addrString)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("cache error: geocode get: %w", err)
	}

	var result Address
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, fmt.Errorf("cache error: geocode unmarshal: %w", err)
	}
	return &result, nil
}

// stores redis value
func (c *Cache) SetGeocodeValue(addrString string, address *Address) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(address)
	if err != nil {
		return fmt.Errorf("cache error: geocode marshal: %w", err)
	}

	if err := c.client.Set(context.Background(), formGeocodeKey(addrString), data, geocodeTTL).Err(); err != nil {
		return fmt.Errorf("cache error: geocode set: %w", err)
	}
	return nil
}
