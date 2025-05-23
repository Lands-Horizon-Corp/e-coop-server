package horizon_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon_test/horizon.auth_test.go

func createCachSetutService(t *testing.T) horizon.CacheService {
	ctx := context.Background()
	env := horizon.NewEnvironmentService("../../.env")
	cache := horizon.NewHorizonCache(
		env.GetString("REDIS_HOST", "localhost"),
		env.GetString("REDIS_PASSWORD", ""),
		env.GetString("REDIS_USERNAME", ""),
		env.GetInt("REDIS_PORT", 6379),
	)
	err := cache.Run(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		// Flush test database
		_ = cache.Delete(ctx, "*")
		_ = cache.Stop(ctx)
	})
	return cache
}

// Helper function to create echo.Context mock with header support
func newMockEchoContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Add middleware to propagate response headers to subsequent requests
	c.Request().Header = rec.Header()

	return c
}

// Test implementation of ClaimWithID
type MyClaim struct {
	UserID string
}

const (
	csrfHeader = "X-CSRF-Token"
)

func (m MyClaim) GetID() string { return m.UserID }

func TestAuthService(t *testing.T) {
	mockCache := createCachSetutService(t)
	service := horizon.NewHorizonAuthService[MyClaim](mockCache, "test", csrfHeader)

	ctx := context.Background()
	testClaim := MyClaim{UserID: "123"}

	t.Run("SetCSRF and GetCSRF", func(t *testing.T) {
		c := newMockEchoContext()
		err := service.SetCSRF(ctx, c, testClaim, time.Hour)
		assert.NoError(t, err)

		token := c.Response().Header().Get(csrfHeader)
		assert.NotEmpty(t, token)

		claim, err := service.GetCSRF(ctx, c)
		assert.NoError(t, err)
		assert.Equal(t, testClaim.UserID, claim.UserID)
	})

	t.Run("VerifyCSRF valid token", func(t *testing.T) {
		c := newMockEchoContext()
		_ = service.SetCSRF(ctx, c, testClaim, time.Hour)
		token := c.Response().Header().Get(csrfHeader)

		claim, err := service.VerifyCSRF(ctx, token)
		assert.NoError(t, err)
		assert.Equal(t, testClaim.UserID, claim.UserID)
	})

	t.Run("VerifyCSRF invalid token", func(t *testing.T) {
		_, err := service.VerifyCSRF(ctx, "invalid-token")
		assert.Error(t, err)
	})

	t.Run("ClearCSRF", func(t *testing.T) {
		c := newMockEchoContext()
		_ = service.SetCSRF(ctx, c, testClaim, time.Hour)
		token := c.Response().Header().Get(csrfHeader)

		service.ClearCSRF(ctx, c)

		exists, _ := mockCache.Exists(ctx, service.Key(token))
		assert.False(t, exists)
	})

	t.Run("Invalid_claim_type", func(t *testing.T) {

		badService := horizon.NewHorizonAuthService[MyClaim](mockCache, "test", csrfHeader)

		// Store invalid data
		key := badService.Key("test-token")
		ctx := context.Background()

		// Store raw string instead of JSON bytes
		err := mockCache.Set(ctx, key, []byte("invalid-claim-type"), time.Hour)
		require.NoError(t, err)

		_, err = badService.VerifyCSRF(ctx, "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid CSRF token claim type")
	})

	t.Run("IsLoggedInOnOtherDevice - different device", func(t *testing.T) {
		// First session
		c1 := newMockEchoContext()
		err := service.SetCSRF(ctx, c1, testClaim, time.Hour)
		require.NoError(t, err)

		// Second session
		c2 := newMockEchoContext()
		err = service.SetCSRF(ctx, c2, testClaim, time.Hour)
		require.NoError(t, err)

		// Verify from first session context
		loggedIn, err := service.IsLoggedInOnOtherDevice(ctx, c1)
		assert.NoError(t, err)
		assert.True(t, loggedIn)

		// Cleanup
		service.ClearCSRF(ctx, c1)
		service.ClearCSRF(ctx, c2)
	})

	t.Run("Key generation", func(t *testing.T) {
		token := "test-token"
		expected := "test:csrf:test-token"
		assert.Equal(t, expected, service.Key(token))
	})

	t.Run("GetCSRF missing token", func(t *testing.T) {
		c := newMockEchoContext()
		_, err := service.GetCSRF(ctx, c)
		assert.Error(t, err)
	})

	t.Run("Invalid_claim_type", func(t *testing.T) {
		badCache := createCachSetutService(t)
		assert.NotNil(t, badCache)
		badService := horizon.NewHorizonAuthService[MyClaim](badCache, "test", csrfHeader)

		key := badService.Key("test-token")
		ctx := context.Background()

		invalidData := []byte(`{"invalid": "data"}`)
		err := badCache.Set(ctx, key, invalidData, time.Hour)
		require.NoError(t, err)

		_, err = badService.VerifyCSRF(ctx, "test-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid CSRF token claim type")
	})
}
