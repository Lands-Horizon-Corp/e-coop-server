package horizon_test

import (
	"context"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test -v ./services/horizon_test/horizon.cache_test.go

func TestHorizonCache(t *testing.T) {
	ctx := context.Background()

	env := horizon.NewEnvironmentService("../../.env")

	redisHost := env.GetString("REDIS_HOST", "")
	redisPassword := env.GetString("REDIS_PASSWORD", "")
	redisUsername := env.GetString("REDIS_USERNAME", "")
	redisPort := env.GetInt("REDIS_PORT", 0)

	cache := horizon.NewHorizonCache(redisHost, redisPassword, redisUsername, redisPort)

	// Start the Redis connection
	err := cache.Run(ctx)
	assert.NoError(t, err, "Start should not return an error")

	// Ping Redis
	err = cache.Ping(ctx)
	assert.NoError(t, err, "Ping should not return an error")

	key := "test-key"
	value := map[string]string{"foo": "bar"}
	ttl := 2 * time.Second

	// Set value
	err = cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err, "Set should not return an error")

	// Get value
	got, err := cache.Get(ctx, key)
	assert.NoError(t, err, "Get should not return an error")
	assert.NotNil(t, got, "Get should return a value")

	// Check if key exists
	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err, "Exists should not return an error")
	assert.True(t, exists, "Key should exist")

	// Wait for TTL to expire
	time.Sleep(ttl + time.Second)
	exists, _ = cache.Exists(ctx, key)
	assert.False(t, exists, "Key should have expired")

	// Set again and delete
	_ = cache.Set(ctx, key, value, ttl)
	err = cache.Delete(ctx, key)
	assert.NoError(t, err, "Delete should not return an error")
	exists, _ = cache.Exists(ctx, key)
	assert.False(t, exists, "Key should not exist after deletion")

	// Stop the Redis client
	err = cache.Stop(ctx)
	assert.NoError(t, err, "Stop should not return an error")
}
