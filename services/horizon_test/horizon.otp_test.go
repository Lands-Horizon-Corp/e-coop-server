package horizon_test

import (
	"context"
	"testing"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test ./services/horizon_test/horizon.otp_test.go
// --- Setup helpers ---
func setupSecurityUtilsOTP() horizon.SecurityService {
	env := horizon.NewEnvironmentService("../../.env")
	token := env.GetByteSlice("APP_TOKEN", "")
	return horizon.NewSecurityService(
		env.GetUint32("PASSWORD_MEMORY", 65536),  // memory (e.g., 64MB)
		env.GetUint32("PASSWORD_ITERATIONS", 3),  // iterations
		env.GetUint8("PASSWORD_PARALLELISM", 2),  // parallelism
		env.GetUint32("PASSWORD_SALT_LENTH", 16), // salt length in bytes
		env.GetUint32("PASSWORD_KEY_LENGTH", 32), // key length in bytes
		token,
	)
}

func setupHorizonOTP() horizon.OTPService {
	env := horizon.NewEnvironmentService("../../.env")
	cache := horizon.NewHorizonCache(
		env.GetString("REDIS_HOST", ""),
		env.GetString("REDIS_PASSWORD", ""),
		env.GetString("REDIS_USERNAME", ""),
		env.GetInt("REDIS_PORT", 6379),
	)
	cache.Run(context.Background())
	if err := cache.Ping(context.Background()); err != nil {
		panic(err)
	}
	security := setupSecurityUtilsOTP()
	return horizon.NewHorizonOTP([]byte("secret"), cache, security)
}

// --- Tests ---

func TestGenerateAndVerifyOTP(t *testing.T) {
	ctx := context.Background()
	service := setupHorizonOTP()

	key := "test:otp:user@example.com"

	// Generate OTP
	code, err := service.Generate(ctx, key)
	assert.NoError(t, err)
	assert.Len(t, code, 6)

	// Verify OTP
	valid, err := service.Verify(ctx, key, code)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Invalid OTP
	invalid, err := service.Verify(ctx, key, "000000")
	assert.NoError(t, err)
	assert.False(t, invalid)
}

func TestRevokeOTP(t *testing.T) {
	ctx := context.Background()
	service := setupHorizonOTP()

	key := "test:otp:revoke@example.com"

	// Generate OTP first
	_, err := service.Generate(ctx, key)
	assert.NoError(t, err)

	// Revoke it
	err = service.Revoke(ctx, key)
	assert.NoError(t, err)

	// Should fail verification after revoke
	ok, err := service.Verify(ctx, key, "anycode")
	assert.Error(t, err)
	assert.False(t, ok)
}
