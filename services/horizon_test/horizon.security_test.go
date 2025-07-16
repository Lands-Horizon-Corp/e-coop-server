package horizon_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test ./services/horizon_test/horizon.security_test.go
func setupSecurityUtils() horizon.SecurityService {
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

func TestGenerateUUID(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()

	uuid, err := sec.GenerateUUID(ctx)
	assert.NoError(t, err)
	assert.Len(t, uuid, 36) // UUID v4 standard length
}

func TestHashAndVerifyPassword(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()
	password := "MySecurePassword!@#"

	hashed, err := sec.HashPassword(ctx, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)

	isValid, err := sec.VerifyPassword(ctx, hashed, password)
	assert.NoError(t, err)
	assert.True(t, isValid)

	isValidWrong, err := sec.VerifyPassword(ctx, hashed, "WrongPassword")
	assert.NoError(t, err)
	assert.False(t, isValidWrong)
}

func TestEncryptAndDecrypt(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()

	plaintext := "Confidential Info"

	encrypted, err := sec.Encrypt(ctx, plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// base64 decode to check if it's encoded properly
	_, err = base64.StdEncoding.DecodeString(encrypted)
	assert.NoError(t, err)

	decrypted, err := sec.Decrypt(ctx, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestGenerateUUIDv5(t *testing.T) {
	sec := setupSecurityUtils()
	ctx := context.Background()

	name := "example.com"
	uuid5, err := sec.GenerateUUIDv5(ctx, name)
	assert.NoError(t, err)
	assert.Len(t, uuid5, 36)

	parsed, err := uuid.Parse(uuid5)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Version(5), parsed.Version())

	_, err = sec.GenerateUUIDv5(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty", err.Error())
}
