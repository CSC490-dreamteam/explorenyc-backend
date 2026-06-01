package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New() (*Cache, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return &Cache{client: nil}, nil
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("cache error: parse REDIS_URL: %w", err)
	}

	client := redis.NewClient(opts)
	return &Cache{client: client}, nil
}

func (c *Cache) Ping() error {
	if c.client == nil {
		return fmt.Errorf("cache error: no redis client configured")
	}
	return c.client.Ping(context.Background()).Err()
}

func (c *Cache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
