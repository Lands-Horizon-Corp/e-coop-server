package event

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

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

func user(service *horizon.HorizonService) *horizon.AuthImpl[UserCSRF] {
	return horizon.NewAuthImpl[UserCSRF](
		service.Cache,
		"user-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER", service.Config.AppName),
		true,
	)
}

func ClearCurrentCSRF(ctx context.Context, service *horizon.HorizonService, echoCtx echo.Context) {
	user(service).ClearCSRF(ctx, echoCtx)
	ClearCurrentToken(ctx, service, echoCtx)

}

func CurrentUser(ctx context.Context, service *horizon.HorizonService, echoCtx echo.Context) (*types.User, error) {
	claim, err := user(service).GetCSRF(ctx, echoCtx)
	if err != nil {
		ClearCurrentCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	if claim.UserID == "" || claim.Email == "" || claim.ContactNumber == "" || claim.Password == "" {
		ClearCurrentCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user information")
	}
	user, err := core.UserManager(service).GetByID(ctx, helpers.ParseUUID(&claim.UserID))
	if err != nil || user == nil {
		ClearCurrentCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user not found")
	}
	if user.Email != claim.Email || user.ContactNumber != claim.ContactNumber || user.Password != claim.Password {
		ClearCurrentCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user information mismatch")
	}
	return user, nil
}

func CurrentUserCSRF(context context.Context, service *horizon.HorizonService, ctx echo.Context) (UserCSRF, error) {
	return user(service).GetCSRF(context, ctx)
}

func LogoutOtherDevices(context context.Context, service *horizon.HorizonService, ctx echo.Context) error {
	return user(service).LogoutOtherDevices(context, ctx)
}

func LoggedInUsers(context context.Context, service *horizon.HorizonService, ctx echo.Context) ([]UserCSRF, error) {
	return user(service).GetLoggedInUsers(context, ctx)
}

func SetUser(ctx context.Context, service *horizon.HorizonService, echoCtx echo.Context, data *types.User) error {
	ClearCurrentCSRF(ctx, service, echoCtx)
	if data == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User cannot be nil")
	}
	if data.Email == "" || data.ContactNumber == "" || data.Password == "" || data.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User must have ID, Email, ContactNumber, Password, and Username")
	}
	if err := user(service).SetCSRF(ctx, echoCtx, UserCSRF{
		UserID:         data.ID.String(),
		Email:          data.Email,
		ContactNumber:  data.ContactNumber,
		Password:       data.Password,
		Username:       data.Username,
		Language:       echoCtx.Request().Header.Get("Accept-Language"),
		Location:       echoCtx.Request().Header.Get("Location"),
		UserAgent:      echoCtx.Request().Header.Get("X-User-Agent"),
		IPAddress:      echoCtx.RealIP(),
		DeviceType:     echoCtx.Request().Header.Get("X-Device-Type"),
		Longitude:      helpers.ParseCoordinate(echoCtx.Request().Header.Get("X-Longitude")),
		Latitude:       helpers.ParseCoordinate(echoCtx.Request().Header.Get("X-Latitude")),
		Referer:        echoCtx.Request().Referer(),
		AcceptLanguage: echoCtx.Request().Header.Get("Accept-Language"),
	}, 144*time.Hour); err != nil {
		ClearCurrentCSRF(ctx, service, echoCtx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}
