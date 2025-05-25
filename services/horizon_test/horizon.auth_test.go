package horizon_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = cache.Run(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
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
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claim := TestClaimUserCSRF{UserID: "123", Email: "test@example.com"}
	expiry := 30 * time.Minute

	// Test SetCSRF
	err := service.SetCSRF(ctx, c, claim, expiry)
	require.NoError(t, err)

	// Verify response headers
	token := rec.Header().Get("X-CSRF-Token")
	assert.NotEmpty(t, token)

	// Verify cookie properties
	var csrfCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "X-CSRF-Token" {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.Equal(t, token, csrfCookie.Value)
	assert.Equal(t, "/", csrfCookie.Path)
	assert.True(t, csrfCookie.Expires.After(time.Now().Add(29*time.Minute)), "Cookie should have correct expiry")
	assert.True(t, csrfCookie.HttpOnly, "Cookie should be HTTP only")
	assert.True(t, csrfCookie.Secure, "Cookie should be secure")

	// Test GetCSRF
	req.Header.Set("X-CSRF-Token", token)
	retrievedClaim, err := service.GetCSRF(ctx, c)
	require.NoError(t, err)
	assert.Equal(t, claim, retrievedClaim)

	// Verify cache entries
	mainKey := fmt.Sprintf("test:csrf:%s:%s", claim.UserID, token)
	tokenUserKey := fmt.Sprintf("test:csrf_token_to_user:%s", token)

	exists, err := cache.Exists(ctx, mainKey)
	require.NoError(t, err)
	assert.True(t, exists, "Main key should exist")

	exists, err = cache.Exists(ctx, tokenUserKey)
	require.NoError(t, err)
	assert.True(t, exists, "Token-user mapping key should exist")

	// Verify token-to-user mapping
	userID, err := cache.Get(ctx, tokenUserKey)
	require.NoError(t, err)
	assert.Equal(t, claim.UserID, string(userID))
}

func TestVerifyCSRF(t *testing.T) {
	ctx, service, cache := setupTest(t)
	validToken := "valid_token"
	invalidToken := "invalid_token"

	// Valid case
	claim := TestClaimUserCSRF{UserID: "789", Email: "verify@test.com"}
	mainKey := fmt.Sprintf("test:csrf:%s:%s", claim.UserID, validToken)
	tokenUserKey := fmt.Sprintf("test:csrf_token_to_user:%s", validToken)

	data, _ := json.Marshal(claim)
	require.NoError(t, cache.Set(ctx, mainKey, data, time.Hour))
	require.NoError(t, cache.Set(ctx, tokenUserKey, []byte(claim.UserID), time.Hour))

	t.Run("Valid token", func(t *testing.T) {
		result, err := service.VerifyCSRF(ctx, validToken)
		require.NoError(t, err)
		assert.Equal(t, claim, result)
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := service.VerifyCSRF(ctx, invalidToken)
		assert.Error(t, err)
	})

	t.Run("Expired token", func(t *testing.T) {
		expiredKey := fmt.Sprintf("test:csrf:%s:expired_token", "expired_user")
		expiredTokenUserKey := fmt.Sprintf("test:csrf_token_to_user:%s", "expired_token")

		require.NoError(t, cache.Set(ctx, expiredKey, []byte("data"), time.Millisecond))
		require.NoError(t, cache.Set(ctx, expiredTokenUserKey, []byte("expired_user"), time.Millisecond))

		time.Sleep(2 * time.Millisecond)

		_, err := service.VerifyCSRF(ctx, "expired_token")
		assert.Error(t, err)
	})
}

func TestGetLoggedInUsers1(t *testing.T) {
	ctx, service, cache := setupTest(t)
	e := echo.New()

	// Create current user session
	currentUser := TestClaimUserCSRF{UserID: "current", Email: "current@test.com"}
	currentReq := httptest.NewRequest(http.MethodGet, "/", nil)
	currentRec := httptest.NewRecorder()
	currentC := e.NewContext(currentReq, currentRec)
	require.NoError(t, service.SetCSRF(ctx, currentC, currentUser, time.Hour))

	// Create other users
	users := []TestClaimUserCSRF{
		{UserID: "userA", Email: "a@test.com"},
		{UserID: "userB", Email: "b@test.com"},
		{UserID: "userC", Email: "c@test.com"},
	}

	// Create one session per user
	for _, user := range users {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, service.SetCSRF(ctx, c, user, time.Hour))
	}

	t.Run("Exclude expired sessions", func(t *testing.T) {
		// Add expired user
		expiredUser := TestClaimUserCSRF{UserID: "expired", Email: "expired@test.com"}
		expiredKey := fmt.Sprintf("test:csrf:%s:expired_token", expiredUser.UserID)
		expiredData, _ := json.Marshal(expiredUser)
		require.NoError(t, cache.Set(ctx, expiredKey, expiredData, time.Millisecond))

		time.Sleep(2 * time.Millisecond)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		userIDs := make(map[string]struct{})
		for _, u := range result {
			userIDs[u.UserID] = struct{}{}
		}
		assert.NotContains(t, userIDs, "expired")
	})
}

func TestClearCSRF(t *testing.T) {
	ctx, service, cache := setupTest(t)
	e := echo.New()

	// Setup initial token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	claim := TestClaimUserCSRF{UserID: "456", Email: "clear@test.com"}
	require.NoError(t, service.SetCSRF(ctx, c, claim, 30*time.Minute))
	token := rec.Header().Get("X-CSRF-Token")

	// Create new request/response for clear operation
	clearReq := httptest.NewRequest(http.MethodGet, "/", nil)
	clearRec := httptest.NewRecorder()
	clearC := e.NewContext(clearReq, clearRec)
	clearReq.Header.Set("X-CSRF-Token", token)

	// Clear CSRF
	service.ClearCSRF(ctx, clearC)

	// Verify response cookies
	clearedCookies := clearRec.Result().Cookies()
	var clearedCookie *http.Cookie
	for _, cookie := range clearedCookies {
		if cookie.Name == "X-CSRF-Token" {
			clearedCookie = cookie
			break
		}
	}
	require.NotNil(t, clearedCookie, "Cleared cookie should exist")
	assert.Empty(t, clearedCookie.Value)
	assert.True(t, clearedCookie.Expires.Unix() <= time.Now().Unix())

	// Verify cache entries
	mainKey := fmt.Sprintf("test:csrf:%s:%s", claim.UserID, token)
	exists, err := cache.Exists(ctx, mainKey)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIsLoggedInOnOtherDevice(t *testing.T) {
	ctx, service, cache := setupTest(t)
	e := echo.New()

	// Setup main session
	mainReq := httptest.NewRequest(http.MethodGet, "/", nil)
	mainRec := httptest.NewRecorder()
	mainC := e.NewContext(mainReq, mainRec)
	user1 := TestClaimUserCSRF{UserID: "user1", Email: "user1@test.com"}
	require.NoError(t, service.SetCSRF(ctx, mainC, user1, time.Hour))
	token := mainRec.Header().Get("X-CSRF-Token")

	t.Run("Different user sessions", func(t *testing.T) {
		// Create different user session
		user2Req := httptest.NewRequest(http.MethodGet, "/", nil)
		user2Rec := httptest.NewRecorder()
		user2C := e.NewContext(user2Req, user2Rec)
		user2 := TestClaimUserCSRF{UserID: "user2", Email: "user2@test.com"}
		require.NoError(t, service.SetCSRF(ctx, user2C, user2, time.Hour))

		// Verify session creation
		user2Token := user2Rec.Header().Get("X-CSRF-Token")
		user2Key := fmt.Sprintf("test:csrf:%s:%s", user2.UserID, user2Token)
		exists, err := cache.Exists(ctx, user2Key)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check for other devices
		checkReq := httptest.NewRequest(http.MethodGet, "/", nil)
		checkReq.Header.Set("X-CSRF-Token", token)
		checkC := e.NewContext(checkReq, mainRec)
		loggedIn, err := service.IsLoggedInOnOtherDevice(ctx, checkC)
		require.NoError(t, err)
		assert.False(t, loggedIn)
	})
}

func TestGetLoggedInUsers2(t *testing.T) {
	ctx, service, cache := setupTest(t)
	e := echo.New()

	// Create current user session
	currentUser := TestClaimUserCSRF{UserID: "current", Email: "current@test.com"}
	currentReq := httptest.NewRequest(http.MethodGet, "/", nil)
	currentRec := httptest.NewRecorder()
	currentC := e.NewContext(currentReq, currentRec)
	require.NoError(t, service.SetCSRF(ctx, currentC, currentUser, time.Hour))

	t.Run("Basic functionality", func(t *testing.T) {
		// Create 3 unique users
		users := []TestClaimUserCSRF{
			{UserID: "userA", Email: "a@test.com"},
			{UserID: "userB", Email: "b@test.com"},
			{UserID: "userC", Email: "c@test.com"},
		}

		// Create one session per user
		for _, user := range users {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			require.NoError(t, service.SetCSRF(ctx, c, user, time.Hour))
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		assert.Len(t, result, 3, "Should return exactly 3 users")
		userIDs := make(map[string]struct{})
		for _, u := range result {
			userIDs[u.UserID] = struct{}{}
		}
		assert.Contains(t, userIDs, "userA")
		assert.Contains(t, userIDs, "userB")
		assert.Contains(t, userIDs, "userC")
		assert.NotContains(t, userIDs, "current")
	})

	t.Run("Exclude users with expired sessions", func(t *testing.T) {
		// Add expired user
		expiredUser := TestClaimUserCSRF{UserID: "expired", Email: "expired@test.com"}
		expiredKey := fmt.Sprintf("test:csrf:%s:expired_token", expiredUser.UserID)
		expiredData, _ := json.Marshal(expiredUser)
		require.NoError(t, cache.Set(ctx, expiredKey, expiredData, time.Millisecond))

		time.Sleep(2 * time.Millisecond)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		userIDs := make(map[string]struct{})
		for _, u := range result {
			userIDs[u.UserID] = struct{}{}
		}
		assert.NotContains(t, userIDs, "expired")
	})

	t.Run("Handle multiple sessions per user", func(t *testing.T) {
		// Create user with multiple sessions
		multiUser := TestClaimUserCSRF{UserID: "multi", Email: "multi@test.com"}

		// First session
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)
		require.NoError(t, service.SetCSRF(ctx, c1, multiUser, time.Hour))

		// Second session
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		require.NoError(t, service.SetCSRF(ctx, c2, multiUser, time.Hour))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		// Should only appear once
		count := 0
		for _, u := range result {
			if u.UserID == "multi" {
				count++
			}
		}
		assert.Equal(t, 1, count, "User should appear only once")
	})

	t.Run("Skip malformed session data", func(t *testing.T) {
		// Create valid user
		validUser := TestClaimUserCSRF{UserID: "valid", Email: "valid@test.com"}
		validReq := httptest.NewRequest(http.MethodGet, "/", nil)
		validRec := httptest.NewRecorder()
		validC := e.NewContext(validReq, validRec)
		require.NoError(t, service.SetCSRF(ctx, validC, validUser, time.Hour))

		// Create malformed entry
		malformedKey := fmt.Sprintf("test:csrf:%s:malformed", "badUser")
		require.NoError(t, cache.Set(ctx, malformedKey, []byte("{invalid json}"), time.Hour))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		// Should only contain valid user
		foundValid := false
		for _, u := range result {
			if u.UserID == "valid" {
				foundValid = true
			}
			assert.NotEqual(t, "badUser", u.UserID, "Malformed entry should be skipped")
		}
		assert.True(t, foundValid, "Valid user should be present")
	})

	t.Run("Handle mixed valid and invalid sessions", func(t *testing.T) {
		// Create user with both valid and expired sessions
		mixedUser := TestClaimUserCSRF{UserID: "mixed", Email: "mixed@test.com"}

		// Valid session
		validReq := httptest.NewRequest(http.MethodGet, "/", nil)
		validRec := httptest.NewRecorder()
		validC := e.NewContext(validReq, validRec)
		require.NoError(t, service.SetCSRF(ctx, validC, mixedUser, time.Hour))

		// Expired session
		expiredKey := fmt.Sprintf("test:csrf:%s:expired_token", mixedUser.UserID)
		expiredData, _ := json.Marshal(mixedUser)
		require.NoError(t, cache.Set(ctx, expiredKey, expiredData, time.Millisecond))
		time.Sleep(2 * time.Millisecond)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentRec.Header().Get("X-CSRF-Token"))
		c := e.NewContext(req, currentRec)

		result, err := service.GetLoggedInUsers(ctx, c)
		require.NoError(t, err)

		// Should still return the user because at least one valid session exists
		found := false
		for _, u := range result {
			if u.UserID == "mixed" {
				found = true
				break
			}
		}
		assert.True(t, found, "User should be present with at least one valid session")
	})
}

// go -v run -tags=test ./services/horizon_test/horizon.auth_test.go

func TestLogoutOtherDevices(t *testing.T) {
	ctx, service, cache := setupTest(t)
	e := echo.New()
	userID := "test-user-123"

	t.Run("Unauthorized access", func(t *testing.T) {
		cache.Flush(ctx) // Ensure clean state
		// Create valid session
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		claim := TestClaimUserCSRF{UserID: userID, Email: "user@test.com"}
		require.NoError(t, service.SetCSRF(ctx, c, claim, time.Hour))

		// Try to logout different user
		err := service.LogoutOtherDevices(ctx, c, "different-user")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing current session token")
	})

	t.Run("Successfully logout other devices", func(t *testing.T) {
		cache.Flush(ctx)
		// Create 3 sessions for the user
		var currentToken string
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			claim := TestClaimUserCSRF{UserID: userID, Email: fmt.Sprintf("session%d@test.com", i)}
			require.NoError(t, service.SetCSRF(ctx, c, claim, time.Hour))

			if i == 0 {
				currentToken = rec.Header().Get("X-CSRF-Token")
			}
		}

		// Verify 3 sessions exist
		pattern := fmt.Sprintf("test:csrf:%s:*", userID)
		keys, err := cache.Keys(ctx, pattern)
		require.NoError(t, err)
		assert.Len(t, keys, 3)

		// Create request with current session
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", currentToken)
		c := e.NewContext(req, nil)

		// Execute logout
		err = service.LogoutOtherDevices(ctx, c, userID)
		require.NoError(t, err)

		// Verify remaining sessions
		keysAfter, err := cache.Keys(ctx, pattern)
		require.NoError(t, err)
		assert.Len(t, keysAfter, 1, "Should only keep current session")

		// Verify token mappings
		tokenUserKeys, _ := cache.Keys(ctx, "test:csrf_token_to_user:*")
		assert.Len(t, tokenUserKeys, 1)
	})

	t.Run("No other sessions to logout", func(t *testing.T) {
		cache.Flush(ctx)
		// Create single session
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		claim := TestClaimUserCSRF{UserID: userID, Email: "single@test.com"}
		require.NoError(t, service.SetCSRF(ctx, c, claim, time.Hour))

		// Use correct token in header
		token := rec.Header().Get("X-CSRF-Token")
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("X-CSRF-Token", token)
		c2 := e.NewContext(req2, nil)

		// Execute logout
		err := service.LogoutOtherDevices(ctx, c2, userID)
		require.NoError(t, err)

		// Verify session still exists
		pattern := fmt.Sprintf("test:csrf:%s:*", userID)
		keys, _ := cache.Keys(ctx, pattern)
		assert.Len(t, keys, 1)
	})

	t.Run("Mixed user sessions", func(t *testing.T) {
		cache.Flush(ctx)
		// Create 2 sessions for target user
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)
		claim1 := TestClaimUserCSRF{UserID: userID, Email: "mixed1@test.com"}
		require.NoError(t, service.SetCSRF(ctx, c1, claim1, time.Hour))
		token1 := rec1.Header().Get("X-CSRF-Token")

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		claim2 := TestClaimUserCSRF{UserID: userID, Email: "mixed2@test.com"}
		require.NoError(t, service.SetCSRF(ctx, c2, claim2, time.Hour))

		// Create another user's session
		reqOther := httptest.NewRequest(http.MethodGet, "/", nil)
		recOther := httptest.NewRecorder()
		cOther := e.NewContext(reqOther, recOther)
		claimOther := TestClaimUserCSRF{UserID: "other-user", Email: "other@test.com"}
		require.NoError(t, service.SetCSRF(ctx, cOther, claimOther, time.Hour))

		// Execute logout on target user, using a valid session token in header
		reqLogout := httptest.NewRequest(http.MethodGet, "/", nil)
		reqLogout.Header.Set("X-CSRF-Token", token1)
		cLogout := e.NewContext(reqLogout, nil)
		err := service.LogoutOtherDevices(ctx, cLogout, userID)
		require.NoError(t, err)

		// Verify target user sessions
		targetPattern := fmt.Sprintf("test:csrf:%s:*", userID)
		targetKeys, _ := cache.Keys(ctx, targetPattern)
		assert.Len(t, targetKeys, 1, "Target user should have 1 session remaining")

		// Verify other user's session remains
		otherPattern := "test:csrf:other-user:*"
		otherKeys, _ := cache.Keys(ctx, otherPattern)
		assert.Len(t, otherKeys, 1, "Other user's session should remain untouched")
	})

	t.Run("Invalid session token", func(t *testing.T) {
		cache.Flush(ctx)
		// Create request with invalid token
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-CSRF-Token", "invalid-token")
		c := e.NewContext(req, nil)

		err := service.LogoutOtherDevices(ctx, c, userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing current session token")
	})
}
