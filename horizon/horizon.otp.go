package horizon

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/rotisserie/eris"
)

type OTPImpl struct {
	secret   []byte
	cache    *CacheImpl
	security *SecurityImpl
	secured  bool
}

func NewHorizonOTP(secret []byte, cache *CacheImpl, security *SecurityImpl, secured bool) *OTPImpl {
	return &OTPImpl{
		secret:   secret,
		cache:    cache,
		security: security,
		secured:  secured,
	}
}

func (h *OTPImpl) Generate(ctx context.Context, key string) (string, error) {
	otpKey := h.key(key)
	countKey := h.keyCount(key)

	if err := h.Revoke(ctx, key); err != nil {
		return "", eris.Wrap(err, "failed to revoke existing OTP")
	}
	random, err := helpers.GenerateDigitCode(6)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate OTP")
	}
	if !h.secured {
		log.Printf("[OTP MOCK MODE] key=%s | code=%s", key, random)
	}
	hash, err := h.security.HashPassword(fmt.Sprint(random))
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
	return random, nil
}

func (h *OTPImpl) Verify(ctx context.Context, key, code string) (bool, error) {
	otpKey := h.key(key)
	countKey := h.keyCount(key)
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
		_ = h.Revoke(ctx, key)
		return false, eris.New("maximum verification attempts reached")
	}
	match, err := h.security.VerifyPassword(string(cachedHash), code)
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

func (h *OTPImpl) Revoke(ctx context.Context, key string) error {
	if err := h.cache.Delete(ctx, h.key(key)); err != nil {
		return eris.Wrapf(err, "failed to delete OTP for key: %s", key)
	}
	if err := h.cache.Delete(ctx, h.keyCount(key)); err != nil {
		return eris.Wrapf(err, "failed to delete count for key: %s", key)
	}
	return nil
}

func (h *OTPImpl) key(key string) string {
	hashedKey, err := h.security.GenerateUUIDv5(key)
	if err != nil {
		return fmt.Sprintf("otp-%s", key)
	}
	return fmt.Sprintf("otp-%s", hashedKey)
}

func (h *OTPImpl) keyCount(key string) string {
	hashedKey, err := h.security.GenerateUUIDv5(key)
	if err != nil {
		return fmt.Sprintf("otp-count-%s", key)
	}
	return fmt.Sprintf("otp-count-%s", hashedKey)
}
