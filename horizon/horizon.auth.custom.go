package horizon

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

const NAME = "organization-branch"

type CustomClaim struct {
	UserOrganizationID string `json:"user_organization_id"`
	UserID             string `json:"user_id"`
	OrganizationID     string `json:"organization_id"`
	BranchID           string `json:"branch_id"`
	jwt.RegisteredClaims
}

type HorizonAuthCustom struct {
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
}

func NewHorizonAuthCustom(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
) (*HorizonAuthCustom, error) {
	return &HorizonAuthCustom{
		config:   config,
		log:      log,
		security: security,
	}, nil
}

func (ha *HorizonAuthCustom) GetCustomFromToken(c echo.Context) (*CustomClaim, error) {
	cookie, err := c.Cookie(NAME)
	if err != nil {
		_ = ha.CleanToken(c)
		return nil, eris.New("authentication token not found")
	}
	rawToken := cookie.Value
	if rawToken == "" {
		_ = ha.CleanToken(c)
		return nil, eris.New("authentication token is empty")
	}

	claim, err := ha.VerifyToken(rawToken)
	if err != nil {
		_ = ha.CleanToken(c)
		return nil, eris.Wrap(err, "invalid or expired authentication token")
	}
	return claim, nil
}

func (ha *HorizonAuthCustom) CleanToken(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     NAME,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

func (ha *HorizonAuthCustom) VerifyToken(encoded string) (*CustomClaim, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, eris.Wrap(err, "invalid base64 token")
	}
	tok, err := jwt.ParseWithClaims(string(raw), &CustomClaim{}, func(tkn *jwt.Token) (any, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, eris.Wrap(jwt.ErrSignatureInvalid, "unexpected signing method")
		}
		return []byte(ha.config.AppToken), nil
	})
	if err != nil {
		return nil, eris.Wrap(err, "token parse failed")
	}
	claims, ok := tok.Claims.(*CustomClaim)
	if !ok || !tok.Valid || claims.Issuer != NAME {
		return nil, eris.New("invalid token")
	}
	return claims, nil
}

func (ha *HorizonAuthCustom) SetToken(c echo.Context, claims CustomClaim) error {
	tok, err := ha.GenerateToken(claims)
	if err != nil {
		return eris.Wrap(err, "GenerateToken failed")
	}
	cookie := &http.Cookie{
		Name:     NAME,
		Value:    tok,
		Path:     "/",
		Expires:  time.Now().Add(AuthExpiration),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

func (ha *HorizonAuthCustom) GenerateToken(c CustomClaim) (string, error) {
	now := time.Now()
	// Default timings
	if c.NotBefore == nil {
		c.NotBefore = jwt.NewNumericDate(now)
	}
	if c.IssuedAt == nil {
		c.IssuedAt = jwt.NewNumericDate(now)
	}
	if c.ExpiresAt == nil {
		c.ExpiresAt = jwt.NewNumericDate(now.Add(AuthExpiration))
	}
	if c.Subject == "" {
		c.Subject = c.ID
	}
	c.Issuer = NAME
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	signed, err := t.SignedString([]byte(ha.config.AppToken))
	if err != nil {
		return "", eris.Wrap(err, "signing token failed")
	}
	return base64.StdEncoding.EncodeToString([]byte(signed)), nil
}
