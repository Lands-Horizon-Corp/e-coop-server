package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
)

const (
	cacheExpiration = 3 * time.Minute
	maxRetries      = 5
	retryDelay      = 2 * time.Second
)

type HorizonCache struct {
	config *HorizonConfig
	log    *HorizonLog
	client *redis.Client
}

func NewHorizonCache(
	config *HorizonConfig,
	log *HorizonLog,
) (*HorizonCache, error) {
	return &HorizonCache{
		config: config,
		log:    log,
	}, nil
}
func (hc *HorizonCache) Run() error {

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", hc.config.RedisHost, hc.config.RedisPort),
		Username: hc.config.RedisUsername,
		Password: hc.config.RedisPassword,
	}
	hc.client = redis.NewClient(opts)
	ctx := context.Background()

	var lastErr error
	for i := 1; i <= maxRetries; i++ {
		if err := hc.client.Ping(ctx).Err(); err != nil {
			lastErr = eris.Wrapf(err, "redis connection attempt %d failed", i)

			time.Sleep(retryDelay)
		} else {

			return nil
		}
	}

	finalErr := eris.Wrapf(lastErr, "failed to connect to Redis after %d attempts", maxRetries)
	hc.log.Log(LogEntry{
		Category: CategoryCache,
		Level:    LevelError,
		Message:  finalErr.Error(),
	})
	return finalErr
}

func (hc *HorizonCache) Stop() error {
	if hc.client == nil {
		return nil
	}
	if err := hc.client.Close(); err != nil {
		wrappedErr := eris.Wrap(err, "error closing Redis client")
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  wrappedErr.Error(),
		})
		return wrappedErr
	}

	hc.log.Log(LogEntry{
		Category: CategoryCache,
		Level:    LevelInfo,
		Message:  "Redis client closed",
	})
	return nil
}

func (hc *HorizonCache) Delete(key string) error {
	if hc.client == nil {
		return eris.New("redis client is not initialized")
	}
	ctx := context.Background()
	if err := hc.client.Del(ctx, key).Err(); err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to delete key %s: %v", key, err),
		})
		return eris.Wrap(err, "failed to delete key")
	}
	return nil
}

func (hc *HorizonCache) Exist(key string) bool {
	if hc.client == nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  "redis client is not initialized",
		})
		return false
	}
	ctx := context.Background()
	val, err := hc.client.Exists(ctx, key).Result()
	if err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to check existence of key %s: %v", key, err),
		})
		return false
	}
	return val > 0
}

func (hc *HorizonCache) Set(key string, data any) error {
	if hc.client == nil {
		return eris.New("redis client is not initialized")
	}
	ctx := context.Background()
	jsonData, err := json.Marshal(data)
	if err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to marshal data for key %s: %v", key, err),
		})
		return eris.Wrap(err, "failed to marshal data")
	}
	if err := hc.client.Set(ctx, key, jsonData, cacheExpiration).Err(); err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to set key %s: %v", key, err),
		})
		return eris.Wrap(err, "failed to set key")
	}
	return nil
}

func (hc *HorizonCache) Get(key string) (any, error) {
	if hc.client == nil {
		return nil, eris.New("redis client is not initialized")
	}
	ctx := context.Background()
	val, err := hc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to get key %s: %v", key, err),
		})
		return nil, eris.Wrap(err, "failed to get key")
	}

	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("failed to unmarshal value for key %s: %v", key, err),
		})
		return nil, eris.Wrap(err, "failed to unmarshal value")
	}
	return result, nil
}

func (hc *HorizonCache) Ping() error {
	if hc.client == nil {
		return eris.New("redis client is not initialized")
	}
	ctx := context.Background()
	if err := hc.client.Ping(ctx).Err(); err != nil {
		hc.log.Log(LogEntry{
			Category: CategoryCache,
			Level:    LevelError,
			Message:  fmt.Sprintf("Redis ping failed: %v", err),
		})
		return eris.Wrap(err, "redis ping failed")
	}
	hc.log.Log(LogEntry{
		Category: CategoryCache,
		Level:    LevelInfo,
		Message:  "Redis ping successful",
	})
	return nil
}
