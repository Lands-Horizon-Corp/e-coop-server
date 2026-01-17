package event

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

type UserOrganizationCSRF struct {
	UserOrganizationID string                     `json:"user_organization_id"`
	UserID             string                     `json:"user_id"`
	BranchID           string                     `json:"branch_id"`
	OrganizationID     string                     `json:"organization_id"`
	UserType           types.UserOrganizationType `json:"user_type"`
	Language           string                     `json:"language"`
	Location           string                     `json:"location"`
	UserAgent          string                     `json:"user_agent"`
	IPAddress          string                     `json:"ip_address"`
	DeviceType         string                     `json:"device_type"`
	Longitude          float64                    `json:"longitude"`
	Latitude           float64                    `json:"latitude"`
	Referer            string                     `json:"referer"`
	AcceptLanguage     string                     `json:"accept_language"`
}
type UserOrganizationCSRFResponse struct {
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

func (m *UserOrganizationCSRFResponse) UserOrganizationCSRFModel(data *UserOrganizationCSRF) *UserOrganizationCSRFResponse {
	if data == nil {
		return nil
	}
	return query.ToModel(data, func(data *UserOrganizationCSRF) *UserOrganizationCSRFResponse {
		return &UserOrganizationCSRFResponse{
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

func (m *UserOrganizationCSRFResponse) UserOrganizationCSRFModels(data []*UserOrganizationCSRF) []*UserOrganizationCSRFResponse {
	return query.ToModels(data, m.UserOrganizationCSRFModel)
}

func (m UserOrganizationCSRF) GetID() string {
	return m.UserOrganizationID
}

func userOrganization(service *horizon.HorizonService) *horizon.AuthImpl[UserOrganizationCSRF] {
	return horizon.NewAuthImpl[UserOrganizationCSRF](
		service.Cache,
		"user-organization-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER-ORGANIZATION", service.Config.AppName),
		true,
	)
}

func ClearCurrentToken(ctx context.Context, service *horizon.HorizonService, echoCtx echo.Context) {
	userOrganization(service).ClearCSRF(ctx, echoCtx)
}

func CurrentUserOrganization(
	ctx context.Context, service *horizon.HorizonService, echoCtx echo.Context) (*types.UserOrganization, error) {
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		userOrganization, err := core.UserOrganizationManager(service).FindOne(ctx, &types.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}
		return userOrganization, nil
	}

	claim, err := userOrganization(service).GetCSRF(ctx, echoCtx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	if claim.UserOrganizationID == "" || claim.UserID == "" || claim.BranchID == "" || claim.OrganizationID == "" {
		ClearCurrentToken(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user organization information")
	}

	parsedUUID := helpers.ParseUUID(&claim.UserOrganizationID)
	userOrganization, err := core.UserOrganizationManager(service).GetByID(ctx, parsedUUID)
	if err != nil || userOrganization == nil {
		ClearCurrentToken(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user organization not found")
	}
	if userOrganization.UserID.String() != claim.UserID ||
		userOrganization.BranchID.String() != claim.BranchID ||
		userOrganization.OrganizationID.String() != claim.OrganizationID ||
		userOrganization.UserType != claim.UserType {
		ClearCurrentToken(ctx, service, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user organization information mismatch")
	}

	return userOrganization, nil
}

func SetUserOrganization(context context.Context, service *horizon.HorizonService, ctx echo.Context, userOrg *types.UserOrganization) error {
	ClearCurrentToken(context, service, ctx)
	if userOrg == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization cannot be nil")
	}
	if userOrg.UserID.String() == "" || userOrg.BranchID.String() == "" || userOrg.OrganizationID.String() == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization must have UserID, BranchID, and OrganizationID")
	}
	if err := userOrganization(service).SetCSRF(context, ctx, UserOrganizationCSRF{
		UserOrganizationID: userOrg.ID.String(),
		UserID:             userOrg.UserID.String(),
		BranchID:           userOrg.BranchID.String(),
		OrganizationID:     userOrg.OrganizationID.String(),
		UserType:           userOrg.UserType,
		Language:           ctx.Request().Header.Get("Accept-Language"),
		Location:           ctx.Request().Header.Get("Location"),
		UserAgent:          ctx.Request().Header.Get("X-User-Agent"),
		IPAddress:          helpers.GetClientIP(ctx),
		DeviceType:         ctx.Request().Header.Get("X-Device-Type"),
		Longitude:          helpers.ParseCoordinate(ctx.Request().Header.Get("X-Longitude")),
		Latitude:           helpers.ParseCoordinate(ctx.Request().Header.Get("X-Latitude")),
		Referer:            ctx.Request().Referer(),
		AcceptLanguage:     ctx.Request().Header.Get("Accept-Language"),
	}, 144*time.Hour); err != nil {
		ClearCurrentToken(context, service, ctx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}

func GetOrganization(service *horizon.HorizonService, ctx echo.Context) (*types.Organization, bool) {
	orgID := ctx.Request().Header.Get("X-Organization-ID")
	if orgID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "organization ID not provided",
		})
		return nil, false
	}
	org, err := core.OrganizationManager(service).GetByID(ctx.Request().Context(), orgID)
	if err != nil || org == nil {
		if err := ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "organization not found",
		}); err != nil {
			return nil, false
		}
		return nil, false
	}
	return org, true
}
