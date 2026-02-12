package horizon

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
)

type CacheImpl struct {
	url    string
	client *redis.Client
	prefix string
}

func NewCacheImpl(url string) *CacheImpl {
	return &CacheImpl{
		url:    url,
		client: nil,
		prefix: "",
	}
}

func (h *CacheImpl) applyPrefix(key string) string {
	return h.prefix + key
}

func (h *CacheImpl) Flush(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	return eris.Wrap(h.client.FlushAll(ctx).Err(), "failed to flush Redis")
}

func (h *CacheImpl) Run(ctx context.Context) error {
	opt, err := redis.ParseURL(h.url)
	if err != nil {
		return eris.Wrap(err, "failed to parse redis url")
	}

	h.client = redis.NewClient(opt)

	if err := h.client.Ping(ctx).Err(); err != nil {
		return eris.Wrap(err, "failed to ping Redis server")
	}
	return nil
}

func (h *CacheImpl) Stop(ctx context.Context) error {
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

func (h *CacheImpl) Ping(ctx context.Context) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	if err := h.client.Ping(ctx).Err(); err != nil {
		return eris.Wrap(err, "redis ping failed")
	}
	return nil
}

func (h *CacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
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

func (h *CacheImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
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

func (h *CacheImpl) Exists(ctx context.Context, key string) (bool, error) {
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

func (h *CacheImpl) Delete(ctx context.Context, key string) error {
	if h.client == nil {
		return eris.New("redis client is not initialized")
	}
	prefixedKey := h.applyPrefix(key)
	return h.client.Del(ctx, prefixedKey).Err()
}

func (h *CacheImpl) Keys(ctx context.Context, pattern string) ([]string, error) {
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

func (h *CacheImpl) ZAdd(ctx context.Context, key string, score float64, member any) error {
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

func (h *CacheImpl) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
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

func (h *CacheImpl) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
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

func (h *CacheImpl) ZCard(ctx context.Context, key string) (int64, error) {
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

func (h *CacheImpl) ZRem(ctx context.Context, key string, members ...any) (int64, error) {
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

func (h *CacheImpl) ZRemRangeByScore(ctx context.Context, key string, min, max string) (int64, error) {
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
