package cooperative_tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// UserOrganizationClaim defines the JWT claims for a user organization.
type UserOrganizationClaim struct {
	UserOrganizationID string                          `json:"user_organization_id"`
	UserID             string                          `json:"user_id"`
	BranchID           string                          `json:"branch_id"`
	OrganizationID     string                          `json:"organization_id"`
	UserType           model_core.UserOrganizationType `json:"user_type"`
	jwt.RegisteredClaims
}

// GetRegisteredClaims returns the JWT registered claims from UserOrganizationClaim.
func (c UserOrganizationClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

// UserOrganizationCSRF holds user organization info for CSRF protection.
type UserOrganizationCSRF struct {
	UserOrganizationID string                          `json:"user_organization_id"`
	UserID             string                          `json:"user_id"`
	BranchID           string                          `json:"branch_id"`
	OrganizationID     string                          `json:"organization_id"`
	UserType           model_core.UserOrganizationType `json:"user_type"`
	Language           string                          `json:"language"`
	Location           string                          `json:"location"`
	UserAgent          string                          `json:"user_agent"`
	IPAddress          string                          `json:"ip_address"`
	DeviceType         string                          `json:"device_type"`
	Longitude          float64                         `json:"longitude"`
	Latitude           float64                         `json:"latitude"`
	Referer            string                          `json:"referer"`
	AcceptLanguage     string                          `json:"accept_language"`
}

// UserOrganizationCSRFResponse is the response model for CSRF user organization info.
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

// UserOrganizationCSRFModel maps a UserOrganizationCSRF to a UserOrganizationCSRFResponse.
func (m *UserOrganizationCSRFResponse) UserOrganizationCSRFModel(data *UserOrganizationCSRF) *UserOrganizationCSRFResponse {
	if data == nil {
		return nil
	}
	return horizon_services.ToModel(data, func(data *UserOrganizationCSRF) *UserOrganizationCSRFResponse {
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

// UserOrganizationCSRFModels maps a slice of UserOrganizationCSRF to a slice of UserOrganizationCSRFResponse.
func (m *UserOrganizationCSRFResponse) UserOrganizationCSRFModels(data []*UserOrganizationCSRF) []*UserOrganizationCSRFResponse {
	return horizon_services.ToModels(data, m.UserOrganizationCSRFModel)
}

// GetID returns the user organization ID from the UserOrganizationCSRF struct.
func (m UserOrganizationCSRF) GetID() string {
	return m.UserOrganizationID
}

// UserOrganizationToken handles user organization token and CSRF logic.
type UserOrganizationToken struct {
	model_core *model_core.ModelCore
	provider   *src.Provider

	CSRF horizon.AuthService[UserOrganizationCSRF]
}

// NewUserOrganizationToken initializes a new UserOrganizationToken.
func NewUserOrganizationToken(provider *src.Provider, model_core *model_core.ModelCore) (*UserOrganizationToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")

	csrfService := horizon.NewHorizonAuthService[UserOrganizationCSRF](
		provider.Service.Cache,
		"user-organization-csrf",
		fmt.Sprintf("%s-%s", "X-SECURE-CSRF-USER-ORGANIZATION", appName),
		true,
	)

	return &UserOrganizationToken{
		CSRF:       csrfService,
		model_core: model_core,
		provider:   provider,
	}, nil
}

func (h *UserOrganizationToken) ClearCurrentToken(ctx context.Context, echoCtx echo.Context) {
	h.CSRF.ClearCSRF(ctx, echoCtx)
}

// CurrentUserOrganization retrieves the current user organization from the CSRF token, validating the information.
func (h *UserOrganizationToken) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*model_core.UserOrganization, error) {
	// Try Bearer token as fallback first
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		userOrganization, err := h.model_core.UserOrganizationManager.FindOne(ctx, &model_core.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})

		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}
		return userOrganization, nil
	}

	claim, err := h.CSRF.GetCSRF(ctx, echoCtx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	if claim.UserOrganizationID == "" || claim.UserID == "" || claim.BranchID == "" || claim.OrganizationID == "" {
		h.ClearCurrentToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: missing essential user organization information")
	}

	userOrganization, err := h.model_core.UserOrganizationManager.GetByID(ctx, handlers.ParseUUID(&claim.UserOrganizationID))
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

// SetUserOrganization sets the CSRF token for the provided user organization.
func (h *UserOrganizationToken) SetUserOrganization(ctx context.Context, echoCtx echo.Context, userOrganization *model_core.UserOrganization) error {
	h.ClearCurrentToken(ctx, echoCtx)
	if userOrganization == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization cannot be nil")
	}
	if userOrganization.UserID.String() == "" || userOrganization.BranchID.String() == "" || userOrganization.OrganizationID.String() == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "UserOrganization must have UserID, BranchID, and OrganizationID")
	}

	longitude := handlers.ParseCoordinate(echoCtx.Request().Header.Get("X-Longitude"))
	latitude := handlers.ParseCoordinate(echoCtx.Request().Header.Get("X-Latitude"))
	location := echoCtx.Request().Header.Get("Location")

	claim := UserOrganizationCSRF{
		UserOrganizationID: userOrganization.ID.String(),
		UserID:             userOrganization.UserID.String(),
		BranchID:           userOrganization.BranchID.String(),
		OrganizationID:     userOrganization.OrganizationID.String(),
		UserType:           userOrganization.UserType,
		Language:           echoCtx.Request().Header.Get("Accept-Language"),
		Location:           location,
		UserAgent:          echoCtx.Request().Header.Get("X-User-Agent"),
		IPAddress:          echoCtx.RealIP(),
		DeviceType:         echoCtx.Request().Header.Get("X-Device-Type"),
		Longitude:          longitude,
		Latitude:           latitude,
		Referer:            echoCtx.Request().Referer(),
		AcceptLanguage:     echoCtx.Request().Header.Get("Accept-Language"),
	}

	if err := h.CSRF.SetCSRF(ctx, echoCtx, claim, 144*time.Hour); err != nil {
		h.ClearCurrentToken(ctx, echoCtx)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set authentication token")
	}
	return nil
}
