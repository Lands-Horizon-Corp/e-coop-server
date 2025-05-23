package horizon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4"
)

// AuthService defines the interface for CSRF token management
// T must implement ClaimWithID to allow user identity comparison
type AuthService[T ClaimWithID] interface {
	GetCSRF(ctx context.Context, c echo.Context) (T, error)
	ClearCSRF(ctx context.Context, c echo.Context)
	VerifyCSRF(ctx context.Context, token string) (T, error)
	SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error
	IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error)
	Key(token string) string
	Name() string
}

// ClaimWithID defines a claim that exposes a stable identifier
// to compare sessions across devices
type ClaimWithID interface {
	GetID() string
}

const (
	csrfTokenLength = 32
	csrfHeader      = "X-CSRF-Token"
)

// HorizonAuthService stores state for CSRF using a cache backend
type HorizonAuthService[T ClaimWithID] struct {
	cache CacheService
	name  string
}

// NewHorizonAuthService constructs a new service
func NewHorizonAuthService[T ClaimWithID](cache CacheService, name string) AuthService[T] {
	return &HorizonAuthService[T]{cache: cache, name: name}
}

func (h *HorizonAuthService[T]) GetCSRF(ctx context.Context, c echo.Context) (T, error) {
	var zeroT T
	token := c.Request().Header.Get(csrfHeader)
	if token == "" {
		return zeroT, errors.New("CSRF token not found in request")
	}

	key := h.Key(token)
	val, err := h.cache.Get(ctx, key)
	if err != nil {
		return zeroT, fmt.Errorf("failed to get CSRF token: %w", err)
	}

	// Deserialize from JSON
	data, ok := val.([]byte)
	if !ok {
		return zeroT, errors.New("invalid cache data format")
	}

	var claim T
	if err := json.Unmarshal(data, &claim); err != nil {
		return zeroT, fmt.Errorf("failed to unmarshal claim: %w", err)
	}

	return claim, nil
}

func (h *HorizonAuthService[T]) ClearCSRF(ctx context.Context, c echo.Context) {
	token := c.Request().Header.Get(csrfHeader)
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
		return zeroT, errors.New("invalid CSRF token")
	}
	val, err := h.cache.Get(ctx, key)
	if err != nil {
		return zeroT, fmt.Errorf("failed to verify CSRF token: %w", err)
	}
	data, ok := val.([]byte)
	if !ok {
		return zeroT, errors.New("invalid cache data format")
	}
	var claim T
	if err := json.Unmarshal(data, &claim); err != nil {
		return zeroT, fmt.Errorf("invalid CSRF token claim type: %w", err)
	}
	if IsZero(claim) {
		return zeroT, errors.New("invalid CSRF token claim type")
	}

	return claim, nil
}

func (h *HorizonAuthService[T]) SetCSRF(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	token, err := generateCSRFToken()
	if err != nil {
		return fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	// Serialize claim to JSON
	data, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("failed to marshal claim: %w", err)
	}

	key := h.Key(token)
	if err := h.cache.Set(ctx, key, data, expiry); err != nil {
		return fmt.Errorf("failed to set CSRF token: %w", err)
	}
	c.Response().Header().Set(csrfHeader, token)
	return nil
}

func (h *HorizonAuthService[T]) IsLoggedInOnOtherDevice(ctx context.Context, c echo.Context) (bool, error) {
	currentToken := c.Request().Header.Get(csrfHeader)
	if currentToken == "" {
		return false, errors.New("no CSRF token in request")
	}

	currentClaim, err := h.VerifyCSRF(ctx, currentToken)
	if err != nil {
		return false, fmt.Errorf("invalid current session: %w", err)
	}

	pattern := fmt.Sprintf("%s:csrf:*", h.name)
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return false, fmt.Errorf("failed to get CSRF keys: %w", err)
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

// Helper functions
func generateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
