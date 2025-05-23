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
	"github.com/stretchr/testify/assert"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

// go test -v ./services/horizon_test/horizon.token_test.go
type TestClaim struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (c TestClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

func setupTokenService() horizon.TokenService[TestClaim] {
	env := horizon.NewEnvironmentService("../../.env")
	return horizon.NewTokenService[TestClaim](
		env.GetString("APP_NAME", "horizon-test"),
		[]byte(env.GetString("APP_TOKEN", base64.StdEncoding.EncodeToString([]byte("test-secret")))),
	)
}

func TestGenerateAndVerifyToken(t *testing.T) {
	service := setupTokenService()
	ctx := context.Background()

	claims := TestClaim{
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token, err := service.GenerateToken(ctx, claims, time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	verifiedClaims, err := service.VerifyToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, claims.Username, verifiedClaims.Username)
}

func TestSetAndGetToken(t *testing.T) {
	service := setupTokenService()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.Background()

	claims := TestClaim{
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	err := service.SetToken(ctx, c, claims, time.Hour)
	assert.NoError(t, err)

	cookie := rec.Result().Cookies()
	assert.NotEmpty(t, cookie)

	req.AddCookie(cookie[0])
	token, err := service.GetToken(ctx, c)
	assert.NoError(t, err)
	assert.Equal(t, claims.Username, token.Username)
}

func TestCleanToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.Background()

	svc := setupTokenService()

	// Set a dummy cookie to simulate an existing token
	http.SetCookie(rec, &http.Cookie{
		Name:     svc.(*horizon.HorizonTokenService[TestClaim]).Name,
		Value:    "dummy",
		Path:     "/",
		HttpOnly: true,
	})

	// Attach the request to the recorder's cookies
	req.AddCookie(&http.Cookie{
		Name:  svc.(*horizon.HorizonTokenService[TestClaim]).Name,
		Value: "dummy",
	})

	// Call CleanToken to clear the cookie
	svc.CleanToken(ctx, c)

	// Check if cookie was cleared (Expires set in the past)
	cleared := false
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == svc.(*horizon.HorizonTokenService[TestClaim]).Name {
			if cookie.Value == "" && cookie.Expires.Before(time.Now()) {
				cleared = true
			}
		}
	}
	assert.True(t, cleared, "Expected token cookie to be cleared")
}
