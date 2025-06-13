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

// go test -v
// Cache defines the interface for Redis operations
type CacheService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	Flush(ctx context.Context) error
}

type HorizonCache struct {
	host     string
	password string
	username string
	port     int
	client   *redis.Client
	prefix   string
}

func NewHorizonCache(host, password, username string, port int) CacheService {
	return &HorizonCache{
		host:     host,
		password: password,
		username: username,
		port:     port,
		client:   nil,
		prefix:   "",
	}
}

func (h *HorizonCache) applyPrefix(key string) string {
	return h.prefix + key
}

func (h *HorizonCache) Flush(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	return eris.Wrap(h.client.FlushAll(ctx).Err(), "failed to flush Redis")
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
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	pattern := h.prefix + "*"
	keys, err := h.Keys(ctx, pattern)
	if err != nil {
		return eris.Wrap(err, "failed to fetch keys for cleanup")
	}
	for _, key := range keys {
		if err := h.Delete(ctx, key); err != nil {
			return eris.Wrapf(err, "failed to delete key: %s", key)
		}
	}

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

	// Add prefix to the key
	prefixedKey := h.applyPrefix(key)

	// Get raw bytes directly
	val, err := h.client.Get(ctx, prefixedKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, eris.Wrap(err, "failed to get key")
}

func (h *HorizonCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}

	// Add prefix to the key
	prefixedKey := h.applyPrefix(key)

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
		h.client.Set(ctx, prefixedKey, data, ttl).Err(),
		"failed to set key",
	)
}

func (h *HorizonCache) Exists(ctx context.Context, key string) (bool, error) {
	if h.client == nil {
		return false, eris.New("redis client is not initialized")
	}

	// Add prefix to the key
	prefixedKey := h.applyPrefix(key)

	val, err := h.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (h *HorizonCache) Delete(ctx context.Context, key string) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}

	// Add prefix to the key
	prefixedKey := h.applyPrefix(key)

	return h.client.Del(ctx, prefixedKey).Err()
}

func (h *HorizonCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	// Add prefix to the pattern
	prefixedPattern := h.applyPrefix(pattern)

	var cursor uint64
	var keys []string
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = h.client.Scan(ctx, cursor, prefixedPattern, 100).Result()
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
