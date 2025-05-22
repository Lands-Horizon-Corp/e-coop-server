package horizon

import (
	"context"
	"fmt"
	"time"
)

// OTPService manages one-time password generation and validation (6 digits)
type OTPService interface {
	// Generate creates a new OTP code for a key
	Generate(ctx context.Context, key string) (string, error)

	// Verify checks a code against the stored OTP
	Verify(ctx context.Context, key, code string) (bool, error)

	// Revoke invalidates an existing OTP code
	Revoke(ctx context.Context, key string) error
}

type HorizonOTP struct {
	secret   []byte
	cache    CacheService
	security SecurityService
}

// NewHorizonOTP creates a new OTPService instance
func NewHorizonOTP(secret []byte, cache CacheService, security SecurityService) OTPService {
	return &HorizonOTP{
		secret:   secret,
		cache:    cache,
		security: security,
	}
}

// Generate implements OTPService.
func (h *HorizonOTP) Generate(ctx context.Context, key string) (string, error) {
	h.cache.Delete(ctx, key)
	random, err := GenerateRandomDigits(6)
	if err != nil {
		return "", err
	}
	result := fmt.Sprint(random)
	hash, err := h.security.HashPassword(ctx, result)
	if err != nil {
		return "", err
	}
	if err := h.cache.Set(ctx, key, hash, 5*time.Minute); err != nil {
		return "", err
	}
	return result, nil
}

// Revoke implements OTPService.
func (h *HorizonOTP) Revoke(ctx context.Context, key string) error {
	if err := h.cache.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

// Verify implements OTPService.
func (h *HorizonOTP) Verify(ctx context.Context, key string, code string) (bool, error) {
	cachedCode, err := h.cache.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if cachedCode == nil {
		return false, fmt.Errorf("code not found for key: %s", key)
	}

	cachedStr, ok := cachedCode.(string)
	if !ok {
		return false, fmt.Errorf("cached code is not a string for key: %s", key)
	}
	return h.security.VerifyPassword(ctx, cachedStr, code)
}
