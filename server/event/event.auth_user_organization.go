package event

import (
	"context"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

type UserOrganizationCSRF struct {
	UserOrganizationID string                    `json:"user_organization_id"`
	UserID             string                    `json:"user_id"`
	BranchID           string                    `json:"branch_id"`
	OrganizationID     string                    `json:"organization_id"`
	UserType           core.UserOrganizationType `json:"user_type"`
	Language           string                    `json:"language"`
	Location           string                    `json:"location"`
	UserAgent          string                    `json:"user_agent"`
	IPAddress          string                    `json:"ip_address"`
	DeviceType         string                    `json:"device_type"`
	Longitude          float64                   `json:"longitude"`
	Latitude           float64                   `json:"latitude"`
	Referer            string                    `json:"referer"`
	AcceptLanguage     string                    `json:"accept_language"`
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

func (h *Event) ClearCurrentToken(ctx context.Context, echoCtx echo.Context) {
	h.userOrgCSRF.ClearCSRF(ctx, echoCtx)
}

func (h *Event) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*core.UserOrganization, error) {
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		userOrganization, err := h.core.UserOrganizationManager().FindOne(ctx, &core.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})

		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}
		return userOrganization, nil
	}

	claim, err := h.userOrgCSRF.GetCSRF(ctx, echoCtx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	if claim.UserOrganizationID == "" || claim.UserID == "" || claim.BranchID == "" || claim.OrganizationID == "" {
		h.ClearCurrentToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user organization information")
	}

	userOrganization, err := h.core.UserOrganizationManager().GetByID(
		ctx, handlers.ParseUUID(&claim.UserOrganizationID),
	)
	if err != nil || userOrganization == nil {
		h.ClearCurrentToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user organization not found")
	}

	if userOrganization.UserID.String() != claim.UserID ||
		userOrganization.BranchID.String() != claim.BranchID ||
		userOrganization.OrganizationID.String() != claim.OrganizationID ||
		userOrganization.UserType != claim.UserType {
		h.ClearCurrentToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: user organization information mismatch")
	}

	return userOrganization, nil
}

func (h *Event) SetUserOrganization(context context.Context, ctx echo.Context, userOrganization *core.UserOrganization) error {
	h.ClearCurrentToken(context, ctx)
	if userOrganization == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization cannot be nil")
	}
	if userOrganization.UserID.String() == "" || userOrganization.BranchID.String() == "" || userOrganization.OrganizationID.String() == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization must have UserID, BranchID, and OrganizationID")
	}
	if err := h.userOrgCSRF.SetCSRF(context, ctx, UserOrganizationCSRF{
		UserOrganizationID: userOrganization.ID.String(),
		UserID:             userOrganization.UserID.String(),
		BranchID:           userOrganization.BranchID.String(),
		OrganizationID:     userOrganization.OrganizationID.String(),
		UserType:           userOrganization.UserType,
		Language:           ctx.Request().Header.Get("Accept-Language"),
		Location:           ctx.Request().Header.Get("Location"),
		UserAgent:          ctx.Request().Header.Get("X-User-Agent"),
		IPAddress:          handlers.GetClientIP(ctx),
		DeviceType:         ctx.Request().Header.Get("X-Device-Type"),
		Longitude:          handlers.ParseCoordinate(ctx.Request().Header.Get("X-Longitude")),
		Latitude:           handlers.ParseCoordinate(ctx.Request().Header.Get("X-Latitude")),
		Referer:            ctx.Request().Referer(),
		AcceptLanguage:     ctx.Request().Header.Get("Accept-Language"),
	}, 144*time.Hour); err != nil {
		h.ClearCurrentToken(context, ctx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}

func (h *Event) GetOrganization(ctx echo.Context) (*core.Organization, bool) {
	orgID := ctx.Request().Header.Get("X-Organization-ID")
	if orgID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "organization ID not provided",
		})
		return nil, false
	}

	org, err := h.core.OrganizationManager().GetByID(ctx.Request().Context(), orgID)
	if err != nil || org == nil {
		ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "organization not found",
		})
		return nil, false
	}

	return org, true
}
