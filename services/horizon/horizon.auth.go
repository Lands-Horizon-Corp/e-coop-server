package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// ClaimWithID represents an authentication claim that provides an ID.
type ClaimWithID interface {
	GetID() string
}

// AuthService provides authentication functionality with CSRF token management.
type AuthService[T ClaimWithID] interface {
	GetCSRF(ctx context.Context, c echo.Context) (T, error)

	ClearCSRF(ctx context.Context, c echo.Context)

	VerifyCSRF(ctx context.Context, token string) (T, error)

	SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error

	IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error)

	GetLoggedInUsers(ctx context.Context, c echo.Context) ([]T, error)

	LogoutOtherDevices(ctx context.Context, c echo.Context) error

	Key(token string) string

	Name() string
}

// AuthServiceImpl provides a Redis-backed implementation of AuthService.
type AuthServiceImpl[T ClaimWithID] struct {
	cache      CacheService
	name       string
	csrfHeader string
	ssl        bool
}

// NewAuthServiceImpl constructs a new AuthServiceImpl.
func NewAuthServiceImpl[T ClaimWithID](
	cache CacheService,
	name string,
	csrfHeader string,
	ssl bool,
) AuthService[T] {
	return &AuthServiceImpl[T]{
		cache:      cache,
		name:       name,
		csrfHeader: csrfHeader,
		ssl:        ssl,
	}
}

// mainKey returns the Redis key for a user session's claim.
func (h *AuthServiceImpl[T]) mainKey(userID, token string) string {
	return fmt.Sprintf("%s:csrf:%s:%s", h.name, userID, token)
}

// tokenToUserKey returns the Redis key for mapping a token to a user ID.
func (h *AuthServiceImpl[T]) tokenToUserKey(token string) string {
	return fmt.Sprintf("%s:csrf_token_to_user:%s", h.name, token)
}

// getTokenFromContext extracts the CSRF token from the request header or cookie.
func (h *AuthServiceImpl[T]) getTokenFromContext(c echo.Context) string {
	if token := c.Request().Header.Get(h.csrfHeader); token != "" {
		return token
	}
	if auth := c.Request().Header.Get("Authorization"); auth != "" {
		const bearer = "Bearer "
		if strings.HasPrefix(auth, bearer) {
			return strings.TrimSpace(auth[len(bearer):])
		}
	}
	if cookie, err := c.Cookie(h.csrfHeader); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if fv := c.FormValue(h.csrfHeader); fv != "" {
		return fv
	}
	if fv := c.FormValue("token"); fv != "" {
		return fv
	}

	return ""
}

// GetCSRF retrieves and validates the CSRF claim for the current session.
func (h *AuthServiceImpl[T]) GetCSRF(ctx context.Context, c echo.Context) (T, error) {
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
func (h *AuthServiceImpl[T]) SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	token, err := handlers.GenerateToken()
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
		Secure:   h.ssl,
		SameSite: http.SameSiteNoneMode,
	})

	return nil
}

// ClearCSRF removes the CSRF token and claim, and clears the cookie.
func (h *AuthServiceImpl[T]) ClearCSRF(ctx context.Context, c echo.Context) {
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
		Secure:   h.ssl,
		SameSite: http.SameSiteNoneMode,
	})
}

// IsLoggedInOnOtherDevice checks if the current user has valid CSRF tokens on other devices.
func (h *AuthServiceImpl[T]) IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error) {
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
		if exists, _ := h.cache.ArrExists(ctx, key); exists {
			return true, nil
		}
	}
	return false, nil
}

// GetLoggedInUsers returns all other users (excluding the current user) with at least one valid session.
func (h *AuthServiceImpl[T]) GetLoggedInUsers(ctx context.Context, c echo.Context) ([]T, error) {
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get current user claim")
	}
	currentToken := h.getTokenFromContext(c)
	if currentToken == "" {
		return nil, eris.New("CSRF token not found in request")
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
		token := parts[3]
		if token == currentToken {
			continue // Skip current session
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

// LogoutAllUsers logs out all sessions for the current user.
func (h *AuthServiceImpl[T]) LogoutAllUsers(ctx context.Context, c echo.Context) error {
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return eris.Wrap(err, "failed to get current user claim")
	}
	pattern := fmt.Sprintf("%s:csrf:%s:*", h.name, currentClaim.GetID())
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return eris.Wrap(err, "failed to get CSRF keys")
	}
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) < 4 {
			continue
		}
		err := h.cache.Delete(ctx, key)
		if err != nil {
			continue
		}
	}
	return nil
}

// VerifyCSRF validates a CSRF token and returns the associated claim if valid.
func (h *AuthServiceImpl[T]) VerifyCSRF(ctx context.Context, token string) (T, error) {
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
func (h *AuthServiceImpl[T]) LogoutOtherDevices(ctx context.Context, c echo.Context) error {
	currentClaim, err := h.GetCSRF(ctx, c)
	if err != nil {
		return eris.New("missing current session token")
	}
	pattern := fmt.Sprintf("%s:csrf:%s:*", h.name, currentClaim.GetID())
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

		// Delete both session key and token mapping
		if err := h.cache.Delete(ctx, key); err != nil {
			return eris.Wrapf(err, "failed to delete session key: %s", key)
		}
		if err := h.cache.Delete(ctx, h.tokenToUserKey(token)); err != nil {
			return eris.Wrapf(err, "failed to delete token mapping: %s", token)
		}
	}

	c.SetCookie(&http.Cookie{
		Name:     h.csrfHeader,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.ssl,
		SameSite: http.SameSiteNoneMode,
	})
	return nil
}

// Key returns the Redis key for token-to-user mapping.
func (h *AuthServiceImpl[T]) Key(token string) string {
	return h.tokenToUserKey(token)
}

// Name returns the service's configured name.
func (h *AuthServiceImpl[T]) Name() string {
	return h.name
}
