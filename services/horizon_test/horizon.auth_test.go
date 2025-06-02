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

type TestClaimUserCSRF struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (t TestClaimUserCSRF) GetID() string { return t.UserID }

func createCacheSetupService(t *testing.T) horizon.CacheService {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	env := horizon.NewEnvironmentService("../../.env")
	cache := horizon.NewHorizonCache(
		env.GetString("REDIS_HOST", "localhost"),
		env.GetString("REDIS_PASSWORD", ""),
		env.GetString("REDIS_USERNAME", ""),
		env.GetInt("REDIS_PORT", 6379),
	)

	// Add connection retry logic
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = cache.Run(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 3 * time.Second)
	}
	require.NoError(t, err, "Failed to connect to Redis after %d attempts", maxRetries)

	cache.Flush(ctx)
	return cache
}

func setupTest(t *testing.T) (context.Context, *horizon.HorizonAuthService[TestClaimUserCSRF], horizon.CacheService) {
	ctx := context.Background()
	cache := createCacheSetupService(t)

	// Double-check clean state
	keys, err := cache.Keys(ctx, "test:*")
	require.NoError(t, err)
	require.Empty(t, keys, "Cache should be empty before test starts")

	service := horizon.NewHorizonAuthService[TestClaimUserCSRF](
		cache,
		"test",
		"X-CSRF-Token",
	).(*horizon.HorizonAuthService[TestClaimUserCSRF])

	return ctx, service, cache
}

func TestSetAndGetCSRF(t *testing.T) {
	ctx, service, cache := setupTest(t)
	defer cache.Flush(ctx)

	// Setup test claim
	claim := TestClaimUserCSRF{
		UserID: "user123",
		Email:  "user@example.com",
	}

	// Create echo context
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())

	// Test SetCSRF
	err := service.SetCSRF(ctx, c, claim, time.Minute)
	require.NoError(t, err)

	// Verify response headers
	token := c.Response().Header().Get("X-CSRF-Token")
	require.NotEmpty(t, token)
	require.Len(t, token, 36) // Assuming 32-character token

	// Test GetCSRF
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-CSRF-Token", token)
	c2 := e.NewContext(req, httptest.NewRecorder())

	retrievedClaim, err := service.GetCSRF(ctx, c2)
	require.NoError(t, err)
	assert.Equal(t, claim.UserID, retrievedClaim.UserID)
	assert.Equal(t, claim.Email, retrievedClaim.Email)
}

func TestVerifyCSRF(t *testing.T) {
	ctx, service, cache := setupTest(t)
	defer cache.Flush(ctx)

	// Setup test data
	claim := TestClaimUserCSRF{
		UserID: "user456",
		Email:  "another@example.com",
	}

	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	require.NoError(t, service.SetCSRF(ctx, c, claim, time.Minute))
	token := c.Response().Header().Get("X-CSRF-Token")

	// Test valid verification
	verifiedClaim, err := service.VerifyCSRF(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, claim.UserID, verifiedClaim.UserID)

	// Test invalid token
	_, err = service.VerifyCSRF(ctx, "invalid-token-123")
	require.Error(t, err)
}

func TestClearCSRF(t *testing.T) {
	ctx, service, cache := setupTest(t)
	defer cache.Flush(ctx)

	// Setup test data
	claim := TestClaimUserCSRF{
		UserID: "user789",
		Email:  "clear@example.com",
	}

	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	require.NoError(t, service.SetCSRF(ctx, c, claim, time.Minute))

	// Test ClearCSRF
	service.ClearCSRF(ctx, c)

	// Verify deletion
	_, err := service.GetCSRF(ctx, c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CSRF token not found")
}

func TestSessionManagement(t *testing.T) {
	ctx, service, cache := setupTest(t)
	defer cache.Flush(ctx)

	// Setup test user
	userID := "multi_session_user"
	claim := TestClaimUserCSRF{
		UserID: userID,
		Email:  "multi@example.com",
	}

	// Create first session
	e := echo.New()
	c1 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	require.NoError(t, service.SetCSRF(ctx, c1, claim, time.Minute))
	token1 := c1.Response().Header().Get("X-CSRF-Token")

	// Create second session
	c2 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	require.NoError(t, service.SetCSRF(ctx, c2, claim, time.Minute))
	token2 := c2.Response().Header().Get("X-CSRF-Token")

	// Test IsLoggedInOnOtherDevice
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-CSRF-Token", token1)
	c1Check := e.NewContext(req1, httptest.NewRecorder())

	loggedIn, err := service.IsLoggedInOnOtherDevice(ctx, c1Check)
	require.NoError(t, err)
	assert.True(t, loggedIn)

	// Test GetLoggedInUsers - should return OTHER sessions
	users, err := service.GetLoggedInUsers(ctx, c1Check)
	require.NoError(t, err)
	assert.Len(t, users, 1)

	// Verify the returned session has the correct user ID
	assert.Equal(t, userID, users[0].UserID, "User ID should match")
	assert.Equal(t, "multi@example.com", users[0].Email, "Email should match")

	// Test LogoutOtherDevices
	err = service.LogoutOtherDevices(ctx, c1Check)
	require.NoError(t, err)

	// Verify second session was removed
	_, err = service.VerifyCSRF(ctx, token2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CSRF token")

	// Verify first session still valid
	_, err = service.VerifyCSRF(ctx, token1)
	require.Error(t, err)

	// Verify no remaining other sessions
	loggedIn, err = service.IsLoggedInOnOtherDevice(ctx, c1Check)
	require.Error(t, err)
	assert.False(t, loggedIn)
}

func TestEdgeCasesSample(t *testing.T) {
	ctx, service, cache := setupTest(t)
	defer cache.Flush(ctx)

	// Test empty token verification
	_, err := service.VerifyCSRF(ctx, "")
	require.Error(t, err)

	// Test GetCSRF with no token
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	_, err = service.GetCSRF(ctx, c)
	require.Error(t, err)

	// Test ClearCSRF with no existing token
	service.ClearCSRF(ctx, c) // Should not panic
}
