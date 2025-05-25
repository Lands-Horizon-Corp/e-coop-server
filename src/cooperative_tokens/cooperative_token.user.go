package cooperative_tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/model"
)

type CloudflareHeaders struct {
	country      string
	connectingIP string
	cfRay        string
}

func GetCloudflareHeaders(c echo.Context) CloudflareHeaders {
	return CloudflareHeaders{
		country:      c.Request().Header.Get("CF-IPCountry"),
		connectingIP: c.Request().Header.Get("CF-Connecting-IP"),
		cfRay:        c.Request().Header.Get("CF-Ray"),
	}
}

type UserClaim struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Password      string `json:"password"`
	Username      string `json:"username"`

	Timezone       string
	Language       string
	Location       string
	UserAgent      string
	IPAddress      string
	DeviceID       string
	DeviceType     string
	Longitude      float64
	Latitude       float64
	Referer        string
	AcceptLanguage string
	// Cloudflare specific headers (if you're behind Cloudflare)
	CFCountry      string
	CFConnectingIP string
	CFRay          string
	jwt.RegisteredClaims
}

type UserCSRF struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Password      string `json:"password"`
	Username      string `json:"username"`

	Timezone       string
	Language       string
	Location       string
	UserAgent      string
	IPAddress      string
	DeviceID       string
	DeviceType     string
	Longitude      float64
	Latitude       float64
	Referer        string
	AcceptLanguage string
	// Cloudflare specific headers (if you're behind Cloudflare)
	CFCountry      string
	CFConnectingIP string
	CFRay          string
}

func (m UserCSRF) GetID() string {
	return m.UserID
}

func (c UserClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserToken struct {
	model *model.Model

	Token horizon.TokenService[UserClaim]
	CSRF  horizon.AuthService[UserCSRF]
}

func NewUserToken(provider *src.Provider, model *model.Model) (*UserToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")

	token, err := provider.Service.Security.GenerateUUIDv5(context.Background(), appToken+"-user")
	if err != nil {
		return nil, err
	}

	tokenService := horizon.NewTokenService[UserClaim](
		fmt.Sprintf("%s-%s", "X-SECURE-TOKEN-USER", appName),
		[]byte(token),
	)

	csrfService := horizon.NewHorizonAuthService[UserCSRF](
		provider.Service.Cache,
		"user-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER", appName),
	)

	return &UserToken{
		Token: tokenService,
		CSRF:  csrfService,
		model: model,
	}, nil
}

// Key generates a key for the CSRF token
func (h *UserToken) CurrentUser(context context.Context, ctx echo.Context) (*model.User, error) {
	claim, err := h.CSRF.GetCSRF(context, ctx)
	if err != nil {
		h.CSRF.ClearCSRF(context, ctx)
		return nil, echo.NewHTTPError(401, "Unauthorized")
	}
	if claim.UserID == "" || claim.Email == "" || claim.ContactNumber == "" || claim.Password == "" {
		h.CSRF.ClearCSRF(context, ctx)
		return nil, echo.NewHTTPError(401, "Unauthorized [important user important information]")
	}
	user, err := h.model.UserManager.GetByID(context, horizon.ParseUUID(&claim.UserID))
	if err != nil || user == nil {
		h.CSRF.ClearCSRF(context, ctx)
		return nil, echo.NewHTTPError(401, "Unauthorized [user not found]")
	}
	if user.Email != claim.Email || user.ContactNumber != claim.ContactNumber || user.Password != claim.Password {
		h.CSRF.ClearCSRF(context, ctx)
		return nil, echo.NewHTTPError(401, "Unauthorized [user information mismatch]")
	}
	return user, nil
}

func (h *UserToken) SetUser(context context.Context, ctx echo.Context, user *model.User) error {
	h.CSRF.ClearCSRF(context, ctx)
	if user == nil {
		h.CSRF.ClearCSRF(context, ctx)
		return echo.NewHTTPError(http.StatusBadRequest, "user cannot be nil")
	}
	if user.Email == "" || user.ContactNumber == "" || user.Password == "" || user.UserName == "" {
		h.CSRF.ClearCSRF(context, ctx)
		return echo.NewHTTPError(http.StatusBadRequest, "user must have ID, Email, ContactNumber, Password, and Username")
	}

	longitude := horizon.ParseCoordinate(ctx.Request().Header.Get("X-Longitude"))
	latitude := horizon.ParseCoordinate(ctx.Request().Header.Get("X-Latitude"))
	cfHeaders := GetCloudflareHeaders(ctx)

	claim := UserCSRF{
		UserID:        user.ID.String(),
		Email:         user.Email,
		ContactNumber: user.ContactNumber,
		Password:      user.Password,
		Username:      user.UserName,

		Timezone:       ctx.Request().Header.Get("Timezone"),
		Language:       ctx.Request().Header.Get("Accept-Language"),
		Location:       ctx.Request().Header.Get("Location"),
		UserAgent:      ctx.Request().UserAgent(),
		IPAddress:      ctx.RealIP(),
		DeviceID:       ctx.Request().Header.Get("X-Device-ID"),
		DeviceType:     ctx.Request().Header.Get("X-Device-Type"),
		Longitude:      longitude,
		Latitude:       latitude,
		Referer:        ctx.Request().Referer(),
		AcceptLanguage: ctx.Request().Header.Get("Accept-Language"),
		CFCountry:      cfHeaders.country,
		CFConnectingIP: cfHeaders.connectingIP,
		CFRay:          cfHeaders.cfRay,
	}
	if err := h.CSRF.SetCSRF(context, ctx, claim, 8*time.Hour); err != nil {
		h.CSRF.ClearCSRF(context, ctx)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}
