package cooperative_tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/model"
)

// CloudflareHeaders holds specific Cloudflare-related HTTP headers.
type CloudflareHeaders struct {
	Country      string
	ConnectingIP string
	CFRay        string
}

// GetCloudflareHeaders extracts Cloudflare-specific headers from the request.
func GetCloudflareHeaders(c echo.Context) CloudflareHeaders {
	return CloudflareHeaders{
		Country:      c.Request().Header.Get("CF-IPCountry"),
		ConnectingIP: c.Request().Header.Get("CF-Connecting-IP"),
		CFRay:        c.Request().Header.Get("CF-Ray"),
	}
}

// UserClaim defines the JWT claims for a user.
type UserClaim struct {
	UserID         string  `json:"user_id"`
	Email          string  `json:"email"`
	ContactNumber  string  `json:"contact_number"`
	Password       string  `json:"password"`
	Username       string  `json:"username"`
	Language       string  `json:"language"`
	Location       string  `json:"location"`
	UserAgent      string  `json:"user_agent"`
	IPAddress      string  `json:"ip_address"`
	DeviceType     string  `json:"device_type"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	Referer        string  `json:"referer"`
	AcceptLanguage string  `json:"accept_language"`
	jwt.RegisteredClaims
}

// UserCSRF holds user info for CSRF protection.
type UserCSRF struct {
	UserID         string  `json:"user_id"`
	Email          string  `json:"email"`
	ContactNumber  string  `json:"contact_number"`
	Password       string  `json:"password"`
	Username       string  `json:"username"`
	Language       string  `json:"language"`
	Location       string  `json:"location"`
	UserAgent      string  `json:"user_agent"`
	IPAddress      string  `json:"ip_address"`
	DeviceType     string  `json:"device_type"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	Referer        string  `json:"referer"`
	AcceptLanguage string  `json:"accept_language"`
}

// UserCSRFResponse is the response model for CSRF user info.
type UserCSRFResponse struct {
	Language       string  `json:"language"`
	Location       string  `json:"location"`
	UserAgent      string  `json:"user_agent"`
	IPAddress      string  `json:"ip_address"`
	DeviceType     string  `json:"device_type"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	Referer        string  `json:"referer"`
	AcceptLanguage string  `json:"accept_language"`
}

// UserCSRFModel maps a UserCSRF to a UserCSRFResponse.
func (m *UserCSRFResponse) UserCSRFModel(data *UserCSRF) *UserCSRFResponse {
	if data == nil {
		return nil
	}
	return horizon_services.ToModel(data, func(data *UserCSRF) *UserCSRFResponse {
		return &UserCSRFResponse{
			Language:       data.Language,
			Location:       data.Location,
			UserAgent:      data.UserAgent,
			IPAddress:      data.IPAddress,
			DeviceType:     data.DeviceType,
			Longitude:      data.Longitude,
			Latitude:       data.Latitude,
			Referer:        data.Referer,
			AcceptLanguage: data.AcceptLanguage,
		}
	})
}

// UserCSRFModels maps a slice of UserCSRF to a slice of UserCSRFResponse.
func (m *UserCSRFResponse) UserCSRFModels(data []*UserCSRF) []*UserCSRFResponse {
	return horizon_services.ToModels(data, m.UserCSRFModel)
}

// GetID returns the user ID from the UserCSRF struct.
func (m UserCSRF) GetID() string {
	return m.UserID
}

// GetRegisteredClaims returns the JWT registered claims from UserClaim.
func (c UserClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

// UserToken handles user token and CSRF logic.
type UserToken struct {
	model *model.Model

	CSRF horizon.AuthService[UserCSRF]
}

// NewUserToken initializes a new UserToken.
func NewUserToken(provider *src.Provider, model *model.Model) (*UserToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")

	csrfService := horizon.NewHorizonAuthService[UserCSRF](
		provider.Service.Cache,
		"user-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER", appName),
		true,
	)

	return &UserToken{
		CSRF:  csrfService,
		model: model,
	}, nil
}

// CurrentUser retrieves the current user from the CSRF token, validating the information.
func (h *UserToken) CurrentUser(ctx context.Context, echoCtx echo.Context) (*model.User, error) {
	claim, err := h.CSRF.GetCSRF(ctx, echoCtx)
	if err != nil {
		h.CSRF.ClearCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	if claim.UserID == "" || claim.Email == "" || claim.ContactNumber == "" || claim.Password == "" {
		h.CSRF.ClearCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user information")
	}
	user, err := h.model.UserManager.GetByID(ctx, handlers.ParseUUID(&claim.UserID))
	if err != nil || user == nil {
		h.CSRF.ClearCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user not found")
	}
	if user.Email != claim.Email || user.ContactNumber != claim.ContactNumber || user.Password != claim.Password {
		h.CSRF.ClearCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user information mismatch")
	}
	return user, nil
}

// SetUser sets the CSRF token for the provided user.
func (h *UserToken) SetUser(ctx context.Context, echoCtx echo.Context, user *model.User) error {
	h.CSRF.ClearCSRF(ctx, echoCtx)
	if user == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User cannot be nil")
	}
	if user.Email == "" || user.ContactNumber == "" || user.Password == "" || user.UserName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User must have ID, Email, ContactNumber, Password, and Username")
	}

	longitude := handlers.ParseCoordinate(echoCtx.Request().Header.Get("X-Longitude"))
	latitude := handlers.ParseCoordinate(echoCtx.Request().Header.Get("X-Latitude"))
	location := echoCtx.Request().Header.Get("Location")

	claim := UserCSRF{
		UserID:         user.ID.String(),
		Email:          user.Email,
		ContactNumber:  user.ContactNumber,
		Password:       user.Password,
		Username:       user.UserName,
		Language:       echoCtx.Request().Header.Get("Accept-Language"),
		Location:       location,
		UserAgent:      echoCtx.Request().Header.Get("X-User-Agent"),
		IPAddress:      echoCtx.RealIP(),
		DeviceType:     echoCtx.Request().Header.Get("X-Device-Type"),
		Longitude:      longitude,
		Latitude:       latitude,
		Referer:        echoCtx.Request().Referer(),
		AcceptLanguage: echoCtx.Request().Header.Get("Accept-Language"),
	}
	if err := h.CSRF.SetCSRF(ctx, echoCtx, claim, 144*time.Hour); err != nil {
		h.CSRF.ClearCSRF(ctx, echoCtx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}
