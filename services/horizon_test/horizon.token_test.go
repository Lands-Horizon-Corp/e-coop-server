package horizon_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test ./services/horizon_test/horizon.token_test.go
// A minimal claim struct for testing
type TestClaim struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (c TestClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

func setupService() *horizon.HorizonTokenService[TestClaim] {
	env := horizon.NewEnvironmentService("../../.env")
	return &horizon.HorizonTokenService[TestClaim]{
		Name:   env.GetString("APP_NAME", ""),
		Secret: []byte(env.GetString("APP_TOKEN", "")),
	}
}

func TestGenerateAndVerifyToken(t *testing.T) {
	ctx := context.Background()
	svc := setupService()

	// Prepare a claim with ID and Issuer = svc.name
	claim := TestClaim{
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:     "user-123",
			Issuer: svc.Name,
		},
	}

	tokenB64, err := svc.GenerateToken(ctx, claim, 2*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenB64)
	raw, err := base64.StdEncoding.DecodeString(tokenB64)
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
	out, err := svc.VerifyToken(ctx, tokenB64)
	assert.NoError(t, err)
	assert.Equal(t, "alice", out.Username)
	assert.Equal(t, "user-123", out.RegisteredClaims.ID)
}

func TestSetGetAndCleanToken(t *testing.T) {
	ctx := context.Background()
	svc := setupService()

	// echo context with a recorder
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 1) SetToken
	claim := TestClaim{
		Username: "bob",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:     "user-456",
			Issuer: svc.Name,
		},
	}
	err := svc.SetToken(ctx, c, claim, time.Minute)
	assert.NoError(t, err)

	// Cookie was set in the response header
	cookies := rec.Result().Cookies()
	assert.Len(t, cookies, 1)
	setC := cookies[0]
	assert.Equal(t, svc.Name, setC.Name)
	assert.NotEmpty(t, setC.Value)
	assert.WithinDuration(t, time.Now().Add(time.Minute), setC.Expires, time.Second)

	// 2) Simulate a follow‑up request: attach that cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(setC)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// GetToken should retrieve our claim
	outClaim, err := svc.GetToken(ctx, c2)
	assert.NoError(t, err)
	assert.Equal(t, "bob", outClaim.Username)
	assert.Equal(t, "user-456", outClaim.RegisteredClaims.ID)

	// 3) CleanToken should clear it
	svc.CleanToken(ctx, c2)
	cleanCookies := rec2.Result().Cookies()
	// Echo may append multiple Set-Cookie; find ours
	var found *http.Cookie
	for _, ck := range cleanCookies {
		if ck.Name == svc.Name {
			found = ck
			break
		}
	}
	if assert.NotNil(t, found, "expected CleanToken to set a cookie for %q", svc.Name) {
		assert.Equal(t, "", found.Value)
		// Expires in the past
		assert.True(t, found.Expires.Before(time.Now()))
	}
}

func TestVerifyToken_BadBase64(t *testing.T) {
	svc := setupService()
	_, err := svc.VerifyToken(context.Background(), "not-base64!!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base64 token")
}

func TestVerifyToken_BadSignature(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")
	name := env.GetString("APP_NAME", "")

	ctx := context.Background()
	svcGood := setupService()
	svcBad := &horizon.HorizonTokenService[TestClaim]{Name: name, Secret: []byte("wrong-key")}

	claim := TestClaim{
		Username: "eve",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:     "u789",
			Issuer: svcGood.Name,
		},
	}
	token, err := svcGood.GenerateToken(ctx, claim, time.Hour)
	assert.NoError(t, err)

	_, err = svcBad.VerifyToken(ctx, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token parse failed")
}
