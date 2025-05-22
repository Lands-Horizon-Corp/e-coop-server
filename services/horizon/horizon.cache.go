package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
)

// Cache defines the interface for Redis operations
type CacheService interface {
	// Start initializes the Redis connection pool
	Run(ctx context.Context) error

	// Stop gracefully shuts down all Redis connections
	Stop(ctx context.Context) error

	// Ping checks Redis server health
	Ping(ctx context.Context) error

	// Get retrieves a value by key from Redis
	Get(ctx context.Context, key string) (any, error)

	// Set stores a value with TTL expiration
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error
}

type HorizonCache struct {
	host     string
	password string
	username string
	port     int
	client   *redis.Client
}

func NewHorizonCache(host, password, username string, port int) CacheService {
	return &HorizonCache{
		host:     host,
		password: password,
		username: username,
		port:     port,
		client:   nil,
	}
}
func (h *HorizonCache) Run(ctx context.Context) error {
	h.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", h.host, h.port),
		Username: h.username,
		Password: h.password,
		DB:       0,
	})

	if err := h.client.Ping(ctx).Err(); err != nil {
		return eris.Wrap(err, "failed to ping Redis server")
	}
	return nil
}
func (h *HorizonCache) Stop(ctx context.Context) error {
	return h.client.Close()
}
func (h *HorizonCache) Ping(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	if err := h.client.Ping(ctx).Err(); err != nil {
		return eris.Wrap(err, "redis ping failed")
	}
	return nil
}
func (h *HorizonCache) Get(ctx context.Context, key string) (any, error) {
	if h.client == nil {
		return "", eris.New("redis client is not initialized")
	}
	val, err := h.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, eris.Wrap(err, "failed to get key")
	}
	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal value")
	}
	return result, nil
}

func (h *HorizonCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return eris.Wrap(err, "failed to marshal data")
	}
	if err := h.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return eris.Wrap(err, "failed to set key")
	}
	return nil
}

func (h *HorizonCache) Exists(ctx context.Context, key string) (bool, error) {
	if h.client == nil {
		return false, eris.New("redis client is not initialized")
	}
	val, err := h.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (h *HorizonCache) Delete(ctx context.Context, key string) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	return h.client.Del(ctx, key).Err()
}
