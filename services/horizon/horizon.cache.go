package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with TTL expiration
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error

	Keys(ctx context.Context, pattern string) ([]string, error)
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
func (h *HorizonCache) Get(ctx context.Context, key string) ([]byte, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	// Get raw bytes directly
	val, err := h.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, eris.Wrap(err, "failed to get key")
}

func (h *HorizonCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case int:
		data = []byte(strconv.Itoa(v))
	case int8:
		data = []byte(strconv.FormatInt(int64(v), 10))
	case int16:
		data = []byte(strconv.FormatInt(int64(v), 10))
	case int32:
		data = []byte(strconv.FormatInt(int64(v), 10))
	case int64:
		data = []byte(strconv.FormatInt(v, 10))
	case uint:
		data = []byte(strconv.FormatUint(uint64(v), 10))
	case uint8:
		data = []byte(strconv.FormatUint(uint64(v), 10))
	case uint16:
		data = []byte(strconv.FormatUint(uint64(v), 10))
	case uint32:
		data = []byte(strconv.FormatUint(uint64(v), 10))
	case uint64:
		data = []byte(strconv.FormatUint(v, 10))
	case float32:
		data = []byte(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		data = []byte(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		data = []byte(strconv.FormatBool(v))
	default:
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			return eris.Wrap(err, "failed to marshal value")
		}
	}

	return eris.Wrap(
		h.client.Set(ctx, key, data, ttl).Err(),
		"failed to set key",
	)
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

func (h *HorizonCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	var cursor uint64
	var keys []string
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = h.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, eris.Wrap(err, "failed to scan keys")
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
