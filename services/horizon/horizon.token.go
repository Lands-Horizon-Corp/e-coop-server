package horizon

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// TokenService manages JWT token lifecycle
type TokenService[T jwt.Claims] interface {
	// GetToken extracts and validates token from request context
	GetToken(ctx context.Context, c echo.Context) (*T, error)

	// CleanToken removes token from response context
	CleanToken(ctx context.Context, c echo.Context)

	// VerifyToken validates a token string and returns claims
	VerifyToken(ctx context.Context, value string) (*T, error)

	// SetToken creates and sets a new token in response context
	SetToken(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error

	// GenerateToken creates a new signed token with claims
	GenerateToken(ctx context.Context, claims T, expiry time.Duration) (string, error)
}

type HorizonTokenService[T jwt.Claims] struct {
	Name   string
	Secret []byte
}

func NewTokenService[T jwt.Claims](name string, secret []byte) TokenService[T] {
	return &HorizonTokenService[T]{
		Name:   name,
		Secret: secret,
	}
}

// CleanToken implements TokenService.
func (h *HorizonTokenService[T]) CleanToken(ctx context.Context, c echo.Context) {
	cookie := &http.Cookie{
		Name:     h.Name,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// GetToken implements TokenService.
func (h *HorizonTokenService[T]) GetToken(ctx context.Context, c echo.Context) (*T, error) {
	cookie, err := c.Cookie(h.Name)
	if err != nil {
		h.CleanToken(ctx, c)
		return nil, eris.New("authentication token not found")
	}
	rawToken := cookie.Value
	if rawToken == "" {
		h.CleanToken(ctx, c)
		return nil, eris.New("authentication token is empty")
	}

	claim, err := h.VerifyToken(ctx, rawToken)
	if err != nil {
		h.CleanToken(ctx, c)
		return nil, eris.Wrap(err, "invalid or expired authentication token")
	}

	return claim, nil
}

// SetToken implements TokenService.
func (h *HorizonTokenService[T]) SetToken(ctx context.Context, c echo.Context, claim T, expiry time.Duration) error {
	tok, err := h.GenerateToken(ctx, claim, expiry)
	if err != nil {
		return eris.Wrap(err, "GenerateToken failed")
	}
	cookie := &http.Cookie{
		Name:     h.Name,
		Value:    tok,
		Path:     "/",
		Expires:  time.Now().Add(expiry),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

// GenerateToken implements TokenService.
func (h *HorizonTokenService[T]) GenerateToken(ctx context.Context, claims T, expiry time.Duration) (string, error) {
	now := time.Now()

	// Check if the claims implement GetRegisteredClaims via pointer receiver
	if rcGetter, ok := any(&claims).(interface{ GetRegisteredClaims() *jwt.RegisteredClaims }); ok {
		rc := rcGetter.GetRegisteredClaims()
		if rc.NotBefore == nil {
			rc.NotBefore = jwt.NewNumericDate(now)
		}
		if rc.IssuedAt == nil {
			rc.IssuedAt = jwt.NewNumericDate(now)
		}
		if rc.ExpiresAt == nil {
			rc.ExpiresAt = jwt.NewNumericDate(now.Add(expiry))
		}
		if rc.Subject == "" && rc.ID != "" {
			rc.Subject = rc.ID
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.Secret)
	if err != nil {
		return "", eris.Wrap(err, "signing token failed")
	}
	return base64.StdEncoding.EncodeToString([]byte(signed)), nil
}
func (h *HorizonTokenService[T]) VerifyToken(ctx context.Context, value string) (*T, error) {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, eris.Wrap(err, "invalid base64 token")
	}
	var claim T
	token, err := jwt.ParseWithClaims(string(raw), any(&claim).(jwt.Claims), func(tkn *jwt.Token) (any, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, eris.Wrap(jwt.ErrSignatureInvalid, "unexpected signing method")
		}
		return h.Secret, nil
	})
	if err != nil {
		return nil, eris.Wrap(err, "token parse failed")
	}

	if !token.Valid {
		return nil, eris.New("invalid token")
	}

	return &claim, nil
}
