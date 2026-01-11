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

type Cache struct {
	host     string
	password string
	username string
	port     int
	client   *redis.Client
	prefix   string
}

func NewHorizonCache(host, password, username string, port int) *Cache {
	return &Cache{
		host:     host,
		password: password,
		username: username,
		port:     port,
		client:   nil,
		prefix:   "",
	}
}

func (h *Cache) applyPrefix(key string) string {
	return h.prefix + key
}

func (h *Cache) Flush(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	return eris.Wrap(h.client.FlushAll(ctx).Err(), "failed to flush Redis")
}

func (h *Cache) Run(ctx context.Context) error {
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

func (h *Cache) Stop(ctx context.Context) error {
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

func (h *Cache) Ping(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	if err := h.client.Ping(ctx).Err(); err != nil {
		return eris.Wrap(err, "redis ping failed")
	}
	return nil
}

func (h *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	val, err := h.client.Get(ctx, prefixedKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, eris.Wrap(err, "failed to get key")
}

func (h *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}

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

func (h *Cache) Exists(ctx context.Context, key string) (bool, error) {
	if h.client == nil {
		return false, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	val, err := h.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (h *Cache) Delete(ctx context.Context, key string) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	prefixedKey := h.applyPrefix(key)
	return h.client.Del(ctx, prefixedKey).Err()
}

func (h *Cache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}
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

func (h *Cache) ZAdd(ctx context.Context, key string, score float64, member any) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	z := redis.Z{
		Score:  score,
		Member: member,
	}

	return eris.Wrap(
		h.client.ZAdd(ctx, prefixedKey, z).Err(),
		"failed to add member to sorted set",
	)
}

func (h *Cache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	result, err := h.client.ZRange(ctx, prefixedKey, start, stop).Result()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get range from sorted set")
	}

	return result, nil
}

func (h *Cache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	if h.client == nil {
		return nil, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	result, err := h.client.ZRangeWithScores(ctx, prefixedKey, start, stop).Result()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get range with scores from sorted set")
	}

	return result, nil
}

func (h *Cache) ZCard(ctx context.Context, key string) (int64, error) {
	if h.client == nil {
		return 0, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	result, err := h.client.ZCard(ctx, prefixedKey).Result()
	if err != nil {
		return 0, eris.Wrap(err, "failed to get sorted set cardinality")
	}

	return result, nil
}

func (h *Cache) ZRem(ctx context.Context, key string, members ...any) (int64, error) {
	if h.client == nil {
		return 0, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	result, err := h.client.ZRem(ctx, prefixedKey, members...).Result()
	if err != nil {
		return 0, eris.Wrap(err, "failed to remove members from sorted set")
	}

	return result, nil
}

func (h *Cache) ZRemRangeByScore(ctx context.Context, key string, min, max string) (int64, error) {
	if h.client == nil {
		return 0, eris.New("redis client is not initialized")
	}

	prefixedKey := h.applyPrefix(key)

	result, err := h.client.ZRemRangeByScore(ctx, prefixedKey, min, max).Result()
	if err != nil {
		return 0, eris.Wrap(err, "failed to remove members by score from sorted set")
	}

	return result, nil
}
