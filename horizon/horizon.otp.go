package horizon

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type HorizonOTP struct {
	config   *HorizonConfig
	cache    *HorizonCache
	security *HorizonSecurity
}

func NewHorizonOTP(
	config *HorizonConfig,
	cache *HorizonCache,
	security *HorizonSecurity,
) (*HorizonOTP, error) {
	return &HorizonOTP{
		config:   config,
		cache:    cache,
		security: security,
	}, nil
}

func (ho *HorizonOTP) GenerateOTP(key string) (string, error) {
	hashed := ho.secured(key)

	max := big.NewInt(1_000_000)
	nBig, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate random OTP")
	}
	otp := fmt.Sprintf("%06d", nBig.Int64())
	expiration := time.Now().Add(2 * time.Minute).Unix()
	claims := jwt.MapClaims{
		"key": hashed,
		"otp": otp,
		"exp": expiration,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(ho.config.AppToken))
	if err != nil {
		return "", eris.Wrap(err, "failed to sign OTP JWT")
	}
	if err := ho.cache.Set(hashed, signed); err != nil {
		return "", eris.Wrap(err, "failed to store OTP token in cache")
	}
	ho.cache.log.Log(LogEntry{
		Category: CategoryOTP,
		Level:    LevelWarn,
		Message:  fmt.Sprintf("could not delete OTP key %s: %v", hashed, err),
		Fields: []zap.Field{
			zap.String("otp", otp),
			zap.Int64("expiration", expiration),
			zap.String("key", key),
		},
	})
	return otp, nil
}

func (ho *HorizonOTP) VerifyOTP(key string, value string) (bool, error) {
	hashed := ho.secured(key)
	raw, err := ho.cache.Get(hashed)
	if err != nil {
		return false, eris.Wrap(err, "failed to fetch OTP token from cache")
	}
	if raw == nil {
		return false, eris.New("no OTP found or expired")
	}
	tokenString, ok := raw.(string)
	if !ok {
		return false, eris.New("cached OTP token has invalid type")
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(ho.config.AppToken), nil
	})
	if err != nil || !token.Valid {
		return false, eris.Wrap(err, "invalid or expired OTP token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, eris.New("could not parse JWT claims")
	}

	if claims["otp"] != value {
		return false, eris.New("OTP does not match")
	}
	if err := ho.cache.Delete(hashed); err != nil {
		ho.cache.log.Log(LogEntry{
			Category: CategoryOTP,
			Level:    LevelWarn,
			Message:  fmt.Sprintf("could not delete OTP key %s: %v", hashed, err),
		})
	}
	return true, nil

}

func (ho *HorizonOTP) secured(key string) string {
	val := ho.security.Hash(key + ho.config.AppName)
	return string(val)
}
