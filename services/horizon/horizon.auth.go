package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// AuthService defines the interface for CSRF token management.
// T must implement ClaimWithID to allow user identity comparison,
// which is useful for checking if a user is logged in on another device.
type AuthService[T ClaimWithID] interface {
	// GetCSRF retrieves the CSRF token from the request header,
	// looks up the associated claim from the cache, and returns it.
	// Returns an error if the token is missing or invalid.
	GetCSRF(ctx context.Context, c echo.Context) (T, error)

	// ClearCSRF removes the CSRF token from the cache using the token
	// from the request header. This is typically used during logout or session expiration.
	ClearCSRF(ctx context.Context, c echo.Context)

	// VerifyCSRF checks if a given CSRF token is valid and returns the associated claim.
	// Returns an error if the token is not found, expired, or malformed.
	VerifyCSRF(ctx context.Context, token string) (T, error)

	// SetCSRF generates a new CSRF token, stores the claim in the cache with the given expiry,
	// and adds the token to the response header.
	SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error

	// IsLoggedInOnOtherDevice checks if there are other active sessions (tokens)
	// in the cache that belong to the same user (same claim ID) but with a different token.
	// Useful for enforcing single-session policies or notifying users about concurrent logins.
	IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error)

	// Key returns the cache key format used to store the CSRF token in the cache.
	// This typically includes the service name and token.
	Key(token string) string

	// Name returns the name of the service or namespace used for namespacing the cache keys.
	Name() string
}

// ClaimWithID defines a claim that exposes a stable identifier
// to compare sessions across devices
type ClaimWithID interface {
	GetID() string
}

// HorizonAuthService stores state for CSRF using a cache backend
type HorizonAuthService[T ClaimWithID] struct {
	cache      CacheService
	name       string
	csrfHeader string
}

// NewHorizonAuthService constructs a new service
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

func (h *HorizonAuthService[T]) GetCSRF(ctx context.Context, c echo.Context) (T, error) {
	var zeroT T
	token := c.Request().Header.Get(h.csrfHeader)
	if token == "" {
		return zeroT, eris.New("CSRF token not found in request")
	}

	key := h.Key(token)
	val, err := h.cache.Get(ctx, key)
	if err != nil {
		return zeroT, eris.Wrap(err, "failed to get CSRF token")
	}

	data, ok := val.([]byte)
	if !ok {
		return zeroT, eris.New("invalid cache data format")
	}

	var claim T
	if err := json.Unmarshal(data, &claim); err != nil {
		return zeroT, eris.Wrap(err, "failed to unmarshal claim")
	}

	return claim, nil
}

func (h *HorizonAuthService[T]) ClearCSRF(ctx context.Context, c echo.Context) {
	token := c.Request().Header.Get(h.csrfHeader)
	if token == "" {
		return
	}

	key := h.Key(token)
	_ = h.cache.Delete(ctx, key)
}

func (h *HorizonAuthService[T]) VerifyCSRF(ctx context.Context, token string) (T, error) {
	var zeroT T
	key := h.Key(token)

	exists, err := h.cache.Exists(ctx, key)
	if err != nil || !exists {
		return zeroT, eris.New("invalid CSRF token")
	}
	val, err := h.cache.Get(ctx, key)
	if err != nil {
		return zeroT, eris.Wrap(err, "failed to verify CSRF token")
	}
	data, ok := val.([]byte)
	if !ok {
		return zeroT, eris.New("invalid cache data format")
	}
	var claim T
	if err := json.Unmarshal(data, &claim); err != nil {
		return zeroT, eris.Wrap(err, "invalid CSRF token claim type")
	}
	if IsZero(claim) {
		return zeroT, eris.New("invalid CSRF token claim type")
	}

	return claim, nil
}

func (h *HorizonAuthService[T]) SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	token, err := GenerateToken()
	if err != nil {
		return eris.Wrap(err, "failed to generate CSRF token")
	}

	data, err := json.Marshal(claim)
	if err != nil {
		return eris.Wrap(err, "failed to marshal claim")
	}

	key := h.Key(token)
	if err := h.cache.Set(ctx, key, data, expiry); err != nil {
		return eris.Wrap(err, "failed to set CSRF token")
	}
	c.Response().Header().Set(h.csrfHeader, token)
	return nil
}

func (h *HorizonAuthService[T]) IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error) {
	currentToken := c.Request().Header.Get(h.csrfHeader)
	if currentToken == "" {
		return false, eris.New("no CSRF token in request")
	}

	currentClaim, err := h.VerifyCSRF(ctx, currentToken)
	if err != nil {
		return false, eris.Wrap(err, "invalid current session")
	}

	pattern := fmt.Sprintf("%s:csrf:*", h.name)
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return false, eris.Wrap(err, "failed to get CSRF keys")
	}

	for _, key := range keys {
		if key == h.Key(currentToken) {
			continue
		}

		raw, err := h.cache.Get(ctx, key)
		if err != nil {
			continue
		}

		data, ok := raw.([]byte)
		if !ok {
			continue
		}

		var otherClaim T
		if err := json.Unmarshal(data, &otherClaim); err != nil {
			continue
		}

		if otherClaim.GetID() == currentClaim.GetID() {
			return true, nil
		}
	}

	return false, nil
}

func (h *HorizonAuthService[T]) Key(token string) string {
	return fmt.Sprintf("%s:csrf:%s", h.name, token)
}

func (h *HorizonAuthService[T]) Name() string {
	return h.name
}
