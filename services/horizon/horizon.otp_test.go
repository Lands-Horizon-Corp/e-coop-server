package horizon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSecurityUtilsOTP() SecurityService {
	env := NewEnvironmentService("../../.env")
	token := env.GetByteSlice("APP_TOKEN", "")
	return NewSecurityService(
		env.GetUint32("PASSWORD_MEMORY", 65536),  // memory (e.g., 64MB)
		env.GetUint32("PASSWORD_ITERATIONS", 3),  // iterations
		env.GetUint8("PASSWORD_PARALLELISM", 2),  // parallelism
		env.GetUint32("PASSWORD_SALT_LENTH", 16), // salt length in bytes
		env.GetUint32("PASSWORD_KEY_LENGTH", 32), // key length in bytes
		token,
	)
}

func setupHorizonOTP() OTPService {
	env := NewEnvironmentService("../../.env")
	cache := NewHorizonCache(
		env.GetString("REDIS_HOST", ""),
		env.GetString("REDIS_PASSWORD", ""),
		env.GetString("REDIS_USERNAME", ""),
		env.GetInt("REDIS_PORT", 6379),
	)
	if err := cache.Run(context.Background()); err != nil {
		panic(err)
	}
	if err := cache.Ping(context.Background()); err != nil {
		panic(err)
	}
	security := setupSecurityUtilsOTP()
	return NewHorizonOTP([]byte("secret"), cache, security)
}


func TestGenerateOTP(t *testing.T) {
	otp := setupHorizonOTP()
	ctx := context.Background()
	key := t.Name()

	t.Run("successful generation", func(t *testing.T) {
		code, err := otp.Generate(ctx, key)
		require.NoError(t, err)
		assert.Len(t, code, 6, "OTP should be 6 digits")

		valid, err := otp.Verify(ctx, key, code)
		assert.True(t, valid)
		assert.NoError(t, err)
	})

	t.Run("replaces existing OTP", func(t *testing.T) {
		code1, err := otp.Generate(ctx, key)
		require.NoError(t, err)

		code2, err := otp.Generate(ctx, key)
		require.NoError(t, err)
		assert.NotEqual(t, code1, code2, "New OTP should be different")

		valid, err := otp.Verify(ctx, key, code1)
		assert.False(t, valid)
		assert.NoError(t, err) // Updated assertion

		valid, err = otp.Verify(ctx, key, code2)
		assert.True(t, valid)
		assert.NoError(t, err)
	})
}

func TestVerifyOTP(t *testing.T) {
	otp := setupHorizonOTP()
	ctx := context.Background()
	key := t.Name()
	code, _ := otp.Generate(ctx, key)

	t.Run("valid code", func(t *testing.T) {
		valid, err := otp.Verify(ctx, key, code)
		assert.True(t, valid)
		assert.NoError(t, err)

		valid, err = otp.Verify(ctx, key, code)
		assert.False(t, valid)
		assert.Error(t, err)
	})

	t.Run("invalid code", func(t *testing.T) {
		_, err := otp.Generate(ctx, key) // Reset state
		require.NoError(t, err)

		valid, err := otp.Verify(ctx, key, "000000")
		assert.False(t, valid)
		assert.NoError(t, err)

		valid, err = otp.Verify(ctx, key, "111111")
		assert.False(t, valid)
		assert.NoError(t, err)

		valid, err = otp.Verify(ctx, key, "222222")
		assert.False(t, valid)
		assert.ErrorContains(t, err, "maximum verification attempts reached")

		valid, err = otp.Verify(ctx, key, code)
		assert.False(t, valid)
		assert.Error(t, err)
	})
}

func TestRevokeOTP(t *testing.T) {
	otp := setupHorizonOTP()
	ctx := context.Background()
	key := t.Name()

	t.Run("successful revocation", func(t *testing.T) {
		code, _ := otp.Generate(ctx, key)
		err := otp.Revoke(ctx, key)
		assert.NoError(t, err)

		valid, err := otp.Verify(ctx, key, code)
		assert.False(t, valid)
		assert.Error(t, err)
	})

	t.Run("revoke non-existent OTP", func(t *testing.T) {
		err := otp.Revoke(ctx, "non-existent-key")
		assert.NoError(t, err)
	})
}

func TestEdgeCases(t *testing.T) {
	otp := setupHorizonOTP()
	ctx := context.Background()

	t.Run("verify before generation", func(t *testing.T) {
		valid, err := otp.Verify(ctx, "uninitialized-key", "anycode")
		assert.False(t, valid)
		assert.ErrorContains(t, err, "OTP not found or expired")
	})

}
