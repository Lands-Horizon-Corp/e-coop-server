package event

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	core_admin "github.com/Lands-Horizon-Corp/e-coop-server/src/db/admin"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

//
// =======================
// CSRF MODELS
// =======================
//

type AdminCSRF struct {
	AdminID        string  `json:"admin_id"`
	Email          string  `json:"email"`
	Username       string  `json:"username"`
	Password       string  `json:"password"`
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

type AdminCSRFResponse struct {
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

func (m *AdminCSRFResponse) AdminCSRFModel(data *AdminCSRF) *AdminCSRFResponse {
	if data == nil {
		return nil
	}
	return query.ToModel(data, func(data *AdminCSRF) *AdminCSRFResponse {
		return &AdminCSRFResponse{
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

func (m *AdminCSRFResponse) AdminCSRFModels(data []*AdminCSRF) []*AdminCSRFResponse {
	return query.ToModels(data, m.AdminCSRFModel)
}

func (m AdminCSRF) GetID() string {
	return m.AdminID
}

func admin(service *horizon.HorizonService) *horizon.AuthImpl[AdminCSRF] {
	return horizon.NewAuthImpl[AdminCSRF](service.Cache, "admin-csrf", fmt.Sprintf("%s-%s", "X-SECURE-CSRF-ADMIN", service.Config.AppName), true)
}

func ClearCurrentAdminCSRF(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
) {
	admin(service).ClearCSRF(ctx, echoCtx)
	ClearCurrentToken(ctx, service, echoCtx)
}

func CurrentAdmin(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
) (*types.Admin, error) {
	claim, err := admin(service).GetCSRF(ctx, echoCtx)
	if err != nil {
		ClearCurrentAdminCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	if claim.AdminID == "" || claim.Email == "" || claim.Username == "" || claim.Password == "" {
		ClearCurrentAdminCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential admin information")
	}
	adminUser, err := core_admin.AdminManager(service).GetByID(ctx, helpers.ParseUUID(&claim.AdminID))
	if err != nil || adminUser == nil {
		ClearCurrentAdminCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: admin not found")
	}

	if adminUser.Email != claim.Email || adminUser.Username != claim.Username || adminUser.Password != claim.Password {
		ClearCurrentAdminCSRF(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: admin information mismatch")
	}

	return adminUser, nil
}

func CurrentAdminCSRF(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
) (AdminCSRF, error) {
	return admin(service).GetCSRF(ctx, echoCtx)
}

func LogoutOtherAdminDevices(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
) error {
	return admin(service).LogoutOtherDevices(ctx, echoCtx)
}

func LoggedInAdmins(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
) ([]AdminCSRF, error) {
	return admin(service).GetLoggedInUsers(ctx, echoCtx)
}
func SetAdmin(
	ctx context.Context,
	service *horizon.HorizonService,
	echoCtx echo.Context,
	data *types.Admin,
) error {
	ClearCurrentAdminCSRF(ctx, service, echoCtx)
	if data == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Admin cannot be nil")
	}
	if data.Email == "" || data.Username == "" || data.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Admin must have ID, Email, Username, and Password")
	}
	if err := admin(service).SetCSRF(ctx, echoCtx, AdminCSRF{
		AdminID:        data.ID.String(),
		Email:          data.Email,
		Username:       data.Username,
		Password:       data.Password,
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

		ClearCurrentAdminCSRF(ctx, service, echoCtx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set admin authentication token")
	}

	return nil
}
