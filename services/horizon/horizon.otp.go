package horizon

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/rotisserie/eris"
)

type OTPService interface {
	Generate(ctx context.Context, key string) (string, error)
	Verify(ctx context.Context, key, code string) (bool, error)
	Revoke(ctx context.Context, key string) error
}

type OTP struct {
	secret   []byte
	cache    CacheService
	security SecurityService
	secured  bool
}

func NewHorizonOTP(secret []byte, cache CacheService, security SecurityService, secured bool) OTPService {
	return &OTP{
		secret:   secret,
		cache:    cache,
		security: security,
		secured:  secured,
	}
}

func (h *OTP) Generate(ctx context.Context, key string) (string, error) {
	otpKey := h.key(ctx, key)
	countKey := h.keyCount(ctx, key)

	if err := h.Revoke(ctx, key); err != nil {
		return "", eris.Wrap(err, "failed to revoke existing OTP")
	}

	random, err := handlers.GenerateRandomDigits(6)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate OTP")
	}
	code := fmt.Sprint(random)

	// 🔐 NOT SECURED → LOG OTP (LOCAL / DEV / LAN MODE)
	if !h.secured {
		log.Printf(
			"[OTP MOCK MODE] key=%s | code=%s",
			key,
			code,
		)
	}

	hash, err := h.security.HashPassword(ctx, code)
	if err != nil {
		return "", eris.Wrap(err, "failed to hash OTP")
	}
	if err := h.cache.Set(ctx, otpKey, hash, 5*time.Minute); err != nil {
		return "", eris.Wrap(err, "failed to store OTP")
	}

	if err := h.cache.Set(ctx, countKey, "0", 5*time.Minute); err != nil {
		if delErr := h.cache.Delete(ctx, otpKey); delErr != nil {
			return "", eris.Wrapf(err, "failed to initialize attempt count; also failed to cleanup OTP: %v", delErr)
		}
		return "", eris.Wrap(err, "failed to initialize attempt count")
	}

	return code, nil
}

func (h *OTP) Verify(ctx context.Context, key, code string) (bool, error) {
	otpKey := h.key(ctx, key)
	countKey := h.keyCount(ctx, key)

	cachedHash, err := h.cache.Get(ctx, otpKey)
	if err != nil {
		return false, eris.Wrap(err, "error retrieving OTP")
	}
	if cachedHash == nil {
		return false, eris.New("OTP not found or expired")
	}

	countStr, err := h.cache.Get(ctx, countKey)
	if err != nil {
		return false, eris.Wrap(err, "error retrieving attempt count")
	}
	count := 0
	if countStr != nil {
		count, err = strconv.Atoi(string(countStr))
		if err != nil {
			return false, eris.Wrap(err, "invalid count format")
		}
	}

	if count >= 3 {
		_ = h.Revoke(ctx, key) // Revoke if limit reached
		return false, eris.New("maximum verification attempts reached")
	}

	match, err := h.security.VerifyPassword(ctx, string(cachedHash), code)
	if err != nil {
		return false, eris.Wrap(err, "verification failed")
	}

	if !match {
		count++
		if err := h.cache.Set(ctx, countKey, strconv.Itoa(count), 5*time.Minute); err != nil {
			return false, eris.Wrap(err, "failed to update attempt count")
		}
		if count >= 3 {
			_ = h.Revoke(ctx, key)
			return false, eris.New("maximum verification attempts reached")
		}
		return false, nil
	}
	_ = h.Revoke(ctx, key)
	return true, nil
}

func (h *OTP) Revoke(ctx context.Context, key string) error {
	otpKey := h.key(ctx, key)
	countKey := h.keyCount(ctx, key)
	if err := h.cache.Delete(ctx, otpKey); err != nil {
		return eris.Wrapf(err, "failed to delete OTP for key: %s", key)
	}
	if err := h.cache.Delete(ctx, countKey); err != nil {
		return eris.Wrapf(err, "failed to delete count for key: %s", key)
	}
	return nil
}

func (h *OTP) key(ctx context.Context, key string) string {
	hashedKey, err := h.security.GenerateUUIDv5(ctx, key)
	if err != nil {
		return fmt.Sprintf("otp-%s", key)
	}
	return fmt.Sprintf("otp-%s", hashedKey)
}

func (h *OTP) keyCount(ctx context.Context, key string) string {
	hashedKey, err := h.security.GenerateUUIDv5(ctx, key)
	if err != nil {
		return fmt.Sprintf("otp-count-%s", key)
	}
	return fmt.Sprintf("otp-count-%s", hashedKey)
}
