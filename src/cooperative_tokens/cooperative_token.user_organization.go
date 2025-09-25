package cooperative_tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserOrganizationClaim struct {
	UserOrganizationID string                     `json:"user_organization_id"`
	UserID             string                     `json:"user_id"`
	BranchID           string                     `json:"branch_id"`
	OrganizationID     string                     `json:"organization_id"`
	UserType           model.UserOrganizationType `json:"user_type"`
	jwt.RegisteredClaims
}

func (c UserOrganizationClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserOrganizationToken struct {
	model    *model.Model
	Token    horizon.TokenService[UserOrganizationClaim]
	provider *src.Provider
}

func NewUserOrganizationToken(provider *src.Provider, model *model.Model) (*UserOrganizationToken, error) {
	context := context.Background()
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")
	token, err := provider.Service.Security.GenerateUUIDv5(context, appToken+"-user-organization")
	if err != nil {
		return nil, err
	}

	tokenService := horizon.NewTokenService[UserOrganizationClaim](
		fmt.Sprintf("%s-%s", "X-SECURE-TOKEN-ORGANIZATION", appName),
		[]byte(token),
		true,
	)
	return &UserOrganizationToken{Token: tokenService, model: model, provider: provider}, nil
}

func (h *UserOrganizationToken) ClearCurrentToken(context context.Context, ctx echo.Context) {
	claim, err := h.Token.GetToken(context, ctx)
	if err == nil {
		id, err := uuid.Parse(claim.UserOrganizationID)
		if err != nil {
			h.Token.CleanToken(context, ctx)
			return
		}
		userOrg, err := h.model.UserOrganizationManager.GetByID(context, id)
		if err != nil {
			h.Token.CleanToken(context, ctx)
			return
		}
		userOrg.Status = model.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now().UTC()
		if err := h.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			h.Token.CleanToken(context, ctx)
			return
		}
		fmt.Println("wth is going on here -------------------")
		if err := h.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return
		}
	}
	h.Token.CleanToken(context, ctx)
}

func (h *UserOrganizationToken) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*model.UserOrganization, error) {
	// Try JWT token first
	claim, err := h.Token.GetToken(ctx, echoCtx)
	if err == nil {
		id, err := uuid.Parse(claim.UserOrganizationID)
		if err != nil {
			h.ClearCurrentToken(ctx, echoCtx)
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
		}
		userOrganization, err := h.model.UserOrganizationManager.GetByID(ctx, id)
		if err != nil {
			h.ClearCurrentToken(ctx, echoCtx)
			return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return userOrganization, nil
	}

	// Try Bearer token as fallback
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		userOrganization, err := h.model.UserOrganizationManager.FindOne(ctx, &model.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})

		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}
		return userOrganization, nil
	}

	// No valid authentication found - don't clear tokens here!
	return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
}

func (h *UserOrganizationToken) SetUserOrganization(ctx context.Context, echoCtx echo.Context, userOrganization *model.UserOrganization) error {
	h.ClearCurrentToken(ctx, echoCtx)
	if err := h.Token.SetToken(ctx, echoCtx, UserOrganizationClaim{
		UserOrganizationID: userOrganization.ID.String(),
		UserID:             userOrganization.UserID.String(),
		BranchID:           userOrganization.BranchID.String(),
		OrganizationID:     userOrganization.OrganizationID.String(),
		UserType:           userOrganization.UserType,
	}, 144*time.Hour); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}
