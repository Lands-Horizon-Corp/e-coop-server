package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// ClaimWithID represents a generic claim that must have an ID.
// Used to associate CSRF/session data with users.
type ClaimWithID interface {
	GetID() string
}

// AuthService defines the contract for a generic authentication/session service.
// It is parameterized by a type T that implements the ClaimWithID interface,
// allowing the service to operate on custom claim types (e.g., user sessions, CSRF tokens).
type AuthService[T ClaimWithID] interface {
	// GetCSRF retrieves and validates the CSRF claim for the current session from the request context.
	GetCSRF(ctx context.Context, c echo.Context) (T, error)

	// ClearCSRF removes the CSRF token and associated claim from the session and clears relevant cookies.
	ClearCSRF(ctx context.Context, c echo.Context)

	// VerifyCSRF validates a specific CSRF token and returns the associated claim if valid.
	VerifyCSRF(ctx context.Context, token string) (T, error)

	// SetCSRF creates and stores a new CSRF claim/token, and sets the relevant headers/cookies on the response.
	SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error

	// IsLoggedInOnOtherDevice checks if the current user has any valid CSRF sessions on other devices/browsers.
	IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error)

	// GetLoggedInUsers returns all other users (excluding the current user) with at least one valid session.
	GetLoggedInUsers(ctx context.Context, c echo.Context) ([]T, error)

	// LogoutOtherDevices logs out all sessions for the specified user ID except the current session.
	LogoutOtherDevices(ctx context.Context, c echo.Context, id string) error

	// Key returns the storage key (e.g., Redis key) for mapping a token to a user.
	Key(token string) string

	// Name returns the name of the authentication service (used for key prefixing/namespacing).
	Name() string
}

// HorizonAuthService provides a Redis-backed implementation of AuthService.
type HorizonAuthService[T ClaimWithID] struct {
	cache      CacheService
	name       string
	csrfHeader string
}

// NewHorizonAuthService constructs a new HorizonAuthService.
func NewHorizonAuthService[T ClaimWithID](
	cache CacheService,
	name string,
	csrfHeader string,
) AuthService[T] {
	return &HorizonAuthService[T]{
		cache:      cache,
		name:       name,
		csrfHeader: csrfHeader,
	}
}

// mainKey returns the Redis key for a user session's claim.
func (h *HorizonAuthService[T]) mainKey(userID, token string) string {
	return fmt.Sprintf("%s:csrf:%s:%s", h.name, userID, token)
}

// tokenToUserKey returns the Redis key for mapping a token to a user ID.
func (h *HorizonAuthService[T]) tokenToUserKey(token string) string {
	return fmt.Sprintf("%s:csrf_token_to_user:%s", h.name, token)
}

// getTokenFromContext extracts the CSRF token from the request header or cookie.
func (h *HorizonAuthService[T]) getTokenFromContext(c echo.Context) string {
	if token := c.Request().Header.Get(h.csrfHeader); token != "" {
		return token
	}
	if cookie, err := c.Cookie(h.csrfHeader); err == nil {
		return cookie.Value
	}
	return ""
}

// GetCSRF retrieves and validates the CSRF claim for the current session.
func (h *HorizonAuthService[T]) GetCSRF(ctx context.Context, c echo.Context) (T, error) {
	var zeroT T
	token := h.getTokenFromContext(c)
	if token == "" {
		return zeroT, eris.New("CSRF token not found in request")
	}

	userIDBytes, err := h.cache.Get(ctx, h.tokenToUserKey(token))
	if err != nil || len(userIDBytes) == 0 {
		return zeroT, eris.New("CSRF token not found in cache")
	}
	userID := string(userIDBytes)

	val, err := h.cache.Get(ctx, h.mainKey(userID, token))
	if err != nil {
		return zeroT, eris.Wrap(err, "failed to get CSRF token")
	}

	var claim T
	if err := json.Unmarshal(val, &claim); err != nil {
		return zeroT, eris.Wrap(err, "failed to unmarshal claim")
	}
	return claim, nil
}

// SetCSRF creates a new CSRF token, stores the claim, sets headers and cookies.
func (h *HorizonAuthService[T]) SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	token, err := GenerateToken()
	if err != nil {
		return eris.Wrap(err, "failed to generate CSRF token")
	}

	userID := claim.GetID()
	data, err := json.Marshal(claim)
	if err != nil {
		return eris.Wrap(err, "failed to marshal claim")
	}

	mainKey := h.mainKey(userID, token)
	if err := h.cache.Set(ctx, mainKey, data, expiry); err != nil {
		return eris.Wrap(err, "failed to set CSRF token")
	}

	tokenUserKey := h.tokenToUserKey(token)
	if err := h.cache.Set(ctx, tokenUserKey, []byte(userID), expiry); err != nil {
		_ = h.cache.Delete(ctx, mainKey)
		return eris.Wrap(err, "failed to set token-user mapping")
	}

	c.Response().Header().Set(h.csrfHeader, token)
	c.SetCookie(&http.Cookie{
		Name:     h.csrfHeader,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(expiry),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// ClearCSRF removes the CSRF token and claim, and clears the cookie.
func (h *HorizonAuthService[T]) ClearCSRF(ctx context.Context, c echo.Context) {
	token := h.getTokenFromContext(c)
	if token == "" {
		return
	}

	tokenUserKey := h.tokenToUserKey(token)
	userIDBytes, err := h.cache.Get(ctx, tokenUserKey)
	if err == nil && len(userIDBytes) > 0 {
		userID := string(userIDBytes)
		_ = h.cache.Delete(ctx, h.mainKey(userID, token))
	}
	_ = h.cache.Delete(ctx, tokenUserKey)

	c.SetCookie(&http.Cookie{
		Name:     h.csrfHeader,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// IsLoggedInOnOtherDevice checks if the current user has valid CSRF tokens on other devices.
func (h *HorizonAuthService[T]) IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error) {
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return false, eris.Wrap(err, "could not retrieve CSRF claim")
	}
	currentToken := h.getTokenFromContext(c)
	if currentToken == "" {
		return false, eris.New("CSRF token not found in request")
	}

	pattern := fmt.Sprintf("%s:csrf:%s:*", h.name, currentClaim.GetID())
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return false, eris.Wrap(err, "failed to retrieve user sessions")
	}

	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 4 {
			continue
		}
		token := parts[3]
		if token == currentToken {
			continue
		}
		if exists, _ := h.cache.Exists(ctx, key); exists {
			return true, nil
		}
	}
	return false, nil
}

// GetLoggedInUsers returns all other users (excluding the current user) with at least one valid session.
func (h *HorizonAuthService[T]) GetLoggedInUsers(ctx context.Context, c echo.Context) ([]T, error) {
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get current user claim")
	}
	pattern := fmt.Sprintf("%s:csrf:%s:*", h.name, currentClaim.GetID())
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get CSRF keys")
	}
	uniqueUsers := []T{}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 4 {
			continue
		}
		val, err := h.cache.Get(ctx, key)
		if err != nil {
			continue
		}
		var claim T
		if err := json.Unmarshal(val, &claim); err != nil {
			continue
		}
		uniqueUsers = append(uniqueUsers, claim)
	}
	return uniqueUsers, nil
}

// VerifyCSRF validates a CSRF token and returns the associated claim if valid.
func (h *HorizonAuthService[T]) VerifyCSRF(ctx context.Context, token string) (T, error) {
	var zeroT T
	if token == "" {
		return zeroT, eris.New("empty CSRF token")
	}

	userIDBytes, err := h.cache.Get(ctx, h.tokenToUserKey(token))
	if err != nil || len(userIDBytes) == 0 {
		return zeroT, eris.New("invalid CSRF token")
	}
	userID := string(userIDBytes)

	val, err := h.cache.Get(ctx, h.mainKey(userID, token))
	if err != nil {
		return zeroT, eris.Wrap(err, "failed to verify CSRF token")
	}

	var claim T
	if err := json.Unmarshal(val, &claim); err != nil {
		return zeroT, eris.Wrap(err, "invalid CSRF token claim type")
	}
	return claim, nil
}

// LogoutOtherDevices logs out all other sessions for the user except the current one.
func (h *HorizonAuthService[T]) LogoutOtherDevices(ctx context.Context, c echo.Context, id string) error {
	currentToken := h.getTokenFromContext(c)
	if currentToken == "" {
		return eris.New("missing current session token")
	}
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return eris.New("missing current session token")
	}
	if currentClaim.GetID() != id {
		return eris.New("unauthorized to log out other users")
	}
	pattern := fmt.Sprintf("%s:csrf:%s:*", h.name, id)
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return eris.Wrap(err, "failed to retrieve user sessions")
	}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 4 {
			continue
		}
		token := parts[3]
		if token == currentToken {
			continue
		}
		_ = h.cache.Delete(ctx, key)
		_ = h.cache.Delete(ctx, h.tokenToUserKey(token))
	}
	return nil
}

// Key returns the Redis key for token-to-user mapping.
func (h *HorizonAuthService[T]) Key(token string) string {
	return h.tokenToUserKey(token)
}

// Name returns the service's configured name.
func (h *HorizonAuthService[T]) Name() string {
	return h.name
}
