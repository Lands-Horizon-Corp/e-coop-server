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

const (
	otpTTL      = 2 * time.Minute
	maxOTPTries = 3
)

func (ho *HorizonOTP) GenerateOTP(key string) (string, error) {
	hashedKey := ho.secured(key)
	if ho.cache.Exist(hashedKey) {
		return "", eris.New("OTP already requested; please wait in 2 minutes before retrying")
	}

	max := big.NewInt(1_000_000)
	nBig, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate random OTP")
	}
	otp := fmt.Sprintf("%06d", nBig.Int64())

	exp := time.Now().Add(otpTTL).Unix()
	claims := jwt.MapClaims{
		"key": hashedKey,
		"otp": otp,
		"exp": exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(ho.config.AppToken))
	if err != nil {
		return "", eris.Wrap(err, "failed to sign OTP JWT")
	}
	if err := ho.cache.Set(hashedKey, signed); err != nil {
		return "", eris.Wrap(err, "failed to store OTP token in cache")
	}

	if err := ho.cache.Set(hashedKey+":count", maxOTPTries); err != nil {
		return "", eris.Wrap(err, "failed to initialize OTP retry count")
	}

	ho.cache.log.Log(LogEntry{
		Category: CategoryOTP,
		Level:    LevelInfo,
		Message:  "OTP generated and stored",
		Fields: []zap.Field{
			zap.String("key", key),
			zap.String("hashedKey", hashedKey),
			zap.String("otp", otp),
			zap.Int64("expires_at", exp),
		},
	})

	return otp, nil
}

func (ho *HorizonOTP) VerifyOTP(key, value string) (bool, error) {
	hashedKey := ho.secured(key)

	rawCount, err := ho.cache.Get(hashedKey + ":count")
	if err != nil {
		return false, eris.Wrap(err, "failed to fetch OTP retry count")
	}
	if rawCount == nil {
		return false, eris.New("no retry count found (OTP expired or not generated)")
	}

	countFloat, ok := rawCount.(float64)
	if !ok {
		return false, eris.New("invalid type for retry count")
	}
	remaining := int(countFloat) - 1

	if remaining < 0 {
		_ = ho.cache.Delete(hashedKey)
		_ = ho.cache.Delete(hashedKey + ":count")
		return false, eris.New("too many incorrect attempts, please request a new OTP")
	}

	if err := ho.cache.Set(hashedKey+":count", remaining); err != nil {
		return false, eris.Wrap(err, "failed to update OTP retry count")
	}

	rawTok, err := ho.cache.Get(hashedKey)
	if err != nil {
		return false, eris.Wrap(err, "failed to fetch OTP token")
	}
	if rawTok == nil {
		return false, eris.New("OTP expired or not found")
	}
	tokenString, ok := rawTok.(string)
	if !ok {
		return false, eris.New("cached OTP token has invalid type")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(ho.config.AppToken), nil
	})
	if err != nil || !token.Valid {
		return false, eris.New("invalid or expired OTP token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, eris.New("could not parse JWT claims")
	}

	if claims["otp"] != value {
		return false, eris.New("OTP does not match")
	}

	_ = ho.cache.Delete(hashedKey)
	_ = ho.cache.Delete(hashedKey + ":count")

	return true, nil
}

func (ho *HorizonOTP) secured(key string) string {
	val := ho.security.Hash(key + ho.config.AppName)
	return string(val)
}
