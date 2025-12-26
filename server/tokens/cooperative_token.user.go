package tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/labstack/echo/v4"
)

type CloudflareHeaders struct {
	Country      string
	ConnectingIP string
	CFRay        string
}

func GetCloudflareHeaders(c echo.Context) CloudflareHeaders {
	return CloudflareHeaders{
		Country:      c.Request().Header.Get("CF-IPCountry"),
		ConnectingIP: c.Request().Header.Get("CF-Connecting-IP"),
		CFRay:        c.Request().Header.Get("CF-Ray"),
	}
}

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
}

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

func (m *UserCSRFResponse) UserCSRFModel(data *UserCSRF) *UserCSRFResponse {
	if data == nil {
		return nil
	}
	return query.ToModel(data, func(data *UserCSRF) *UserCSRFResponse {
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

func (m *UserCSRFResponse) UserCSRFModels(data []*UserCSRF) []*UserCSRFResponse {
	return query.ToModels(data, m.UserCSRFModel)
}

func (m UserCSRF) GetID() string {
	return m.UserID
}

type UserToken struct {
	core                  *core.Core
	userOrganizationToken *UserOrganizationToken

	CSRF horizon.AuthService[UserCSRF]
}

func NewUserToken(provider *server.Provider, core *core.Core, userOrganizationToken *UserOrganizationToken) (*UserToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")

	csrfService := horizon.NewAuthServiceImpl[UserCSRF](
		provider.Service.Cache,
		"user-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER", appName),
		true,
	)

	return &UserToken{
		CSRF:                  csrfService,
		core:                  core,
		userOrganizationToken: userOrganizationToken,
	}, nil
}

func (h *UserToken) ClearCurrentCSRF(ctx context.Context, echoCtx echo.Context) {
	h.CSRF.ClearCSRF(ctx, echoCtx)
	h.userOrganizationToken.ClearCurrentToken(ctx, echoCtx)

}

func (h *UserToken) CurrentUser(ctx context.Context, echoCtx echo.Context) (*core.User, error) {
	claim, err := h.CSRF.GetCSRF(ctx, echoCtx)
	if err != nil {
		h.ClearCurrentCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	if claim.UserID == "" || claim.Email == "" || claim.ContactNumber == "" || claim.Password == "" {
		h.ClearCurrentCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user information")
	}
	user, err := h.core.UserManager().GetByID(ctx, handlers.ParseUUID(&claim.UserID))
	if err != nil || user == nil {
		h.ClearCurrentCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user not found")
	}
	if user.Email != claim.Email || user.ContactNumber != claim.ContactNumber || user.Password != claim.Password {
		h.ClearCurrentCSRF(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user information mismatch")
	}
	return user, nil
}

func (h *UserToken) SetUser(ctx context.Context, echoCtx echo.Context, user *core.User) error {
	h.ClearCurrentCSRF(ctx, echoCtx)
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
		h.ClearCurrentCSRF(ctx, echoCtx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}
