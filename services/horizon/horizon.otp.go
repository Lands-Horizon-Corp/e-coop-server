package horizon

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rotisserie/eris"
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
	key = h.key(ctx, key)
	keyCount := h.keyCount(ctx, key)

	// Default count is 0
	count := 0

	// Try to get existing count from cache
	countStr, err := h.cache.Get(ctx, keyCount)
	if err == nil && string(countStr) != "" {
		count, err = strconv.Atoi(string(countStr))
		if err != nil {
			return "", eris.Wrap(err, "invalid count value in cache")
		}
	}

	// If count >= 3, block further attempts
	if count >= 3 {
		return "", eris.New("maximum attempts reached, please wait 5 minutes")
	}

	// Increment and store count
	count++
	if err := h.cache.Set(ctx, keyCount, strconv.Itoa(count), 5*time.Minute); err != nil {
		return "", eris.Wrap(err, "failed to set attempt count")
	}

	h.cache.Delete(ctx, key)
	random, err := GenerateRandomDigits(6)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate OTP")
	}
	result := fmt.Sprint(random)
	hash, err := h.security.HashPassword(ctx, result)
	if err != nil {
		return "", eris.Wrap(err, "failed to hash OTP")
	}
	if err := h.cache.Set(ctx, key, hash, 5*time.Minute); err != nil {
		return "", eris.Wrap(err, "failed to store OTP in cache")
	}
	return result, nil
}

// Revoke implements OTPService.
func (h *HorizonOTP) Revoke(ctx context.Context, key string) error {
	key = h.key(ctx, key)
	key, err := h.security.GenerateUUIDv5(ctx, key)
	if err != nil {
		return err
	}
	if err := h.cache.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

// Verify implements OTPService.
func (h *HorizonOTP) Verify(ctx context.Context, key string, code string) (bool, error) {
	key = h.key(ctx, key)
	cachedCode, err := h.cache.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if cachedCode == nil {
		return false, fmt.Errorf("code not found for key: %s", key)
	}
	return h.security.VerifyPassword(ctx, string(cachedCode), code)
}

func (h *HorizonOTP) key(ctx context.Context, key string) string {
	key, err := h.security.GenerateUUIDv5(ctx, key)
	if err != nil {
		return fmt.Sprintf("otp-%s", key)
	}
	return fmt.Sprintf("otp-%s", key)
}

func (h *HorizonOTP) keyCount(ctx context.Context, key string) string {
	key, err := h.security.GenerateUUIDv5(ctx, key)
	if err != nil {
		return fmt.Sprintf("otp-count-%s", key)
	}
	return fmt.Sprintf("otp-count-%s", key)
}
