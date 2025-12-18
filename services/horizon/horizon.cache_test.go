package horizon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHorizonCache(t *testing.T) {
	ctx := context.Background()

	env := NewEnvironmentService("../../.env")

	redisHost := env.GetString("REDIS_HOST", "")
	redisPassword := env.GetString("REDIS_PASSWORD", "")
	redisUsername := env.GetString("REDIS_USERNAME", "")
	redisPort := env.GetInt("REDIS_PORT", 0)

	cache := NewHorizonCache(redisHost, redisPassword, redisUsername, redisPort)

	err := cache.Run(ctx)
	assert.NoError(t, err, "Start should not return an error")

	err = cache.Ping(ctx)
	assert.NoError(t, err, "Ping should not return an error")

	key := "test-key"
	value := map[string]string{"foo": "bar"}
	ttl := 2 * time.Second

	err = cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err, "Set should not return an error")

	got, err := cache.Get(ctx, key)
	assert.NoError(t, err, "Get should not return an error")
	assert.NotNil(t, got, "Get should return a value")

	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err, "Exists should not return an error")
	assert.True(t, exists, "Key should exist")

	time.Sleep(ttl + time.Second)
	exists, _ = cache.Exists(ctx, key)
	assert.False(t, exists, "Key should have expired")

	_ = cache.Set(ctx, key, value, ttl)
	err = cache.Delete(ctx, key)
	assert.NoError(t, err, "Delete should not return an error")
	exists, _ = cache.Exists(ctx, key)
	assert.False(t, exists, "Key should not exist after deletion")

	err = cache.Stop(ctx)
	assert.NoError(t, err, "Stop should not return an error")
}
