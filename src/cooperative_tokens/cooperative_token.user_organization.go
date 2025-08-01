package cooperative_tokens

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/model"
)

type UserOrganizatonClaim struct {
	UserOrganizatonID string `json:"user_organization_id"`
	UserID            string `json:"user_id"`
	BranchID          string `json:"branch_id"`
	OrganizationID    string `json:"organization_id"`
	UserType          string `json:"user_type"`
	jwt.RegisteredClaims
}

func (c UserOrganizatonClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserOrganizatonToken struct {
	model    *model.Model
	Token    horizon.TokenService[UserOrganizatonClaim]
	provider *src.Provider
}

func NewUserOrganizatonToken(provider *src.Provider, model *model.Model) (*UserOrganizatonToken, error) {
	context := context.Background()
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")
	token, err := provider.Service.Security.GenerateUUIDv5(context, appToken+"-user-organization")
	if err != nil {
		return nil, err
	}

	tokenService := horizon.NewTokenService[UserOrganizatonClaim](
		fmt.Sprintf("%s-%s", "X-SECURE-TOKEN-ORGANIZATION", appName),
		[]byte(token),
		true,
	)
	return &UserOrganizatonToken{Token: tokenService, model: model, provider: provider}, nil
}

func (h *UserOrganizatonToken) ClearCurrentToken(context context.Context, ctx echo.Context) {
	claim, err := h.Token.GetToken(context, ctx)
	if err == nil {
		id, err := uuid.Parse(claim.UserOrganizatonID)
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
		userOrg.LastOnlineAt = time.Now()
		if err := h.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			h.Token.CleanToken(context, ctx)
			return
		}
		if err := h.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			return
		}
	}
	h.Token.CleanToken(context, ctx)
}

func (h *UserOrganizatonToken) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*model.UserOrganization, error) {
	// Try JWT token first
	claim, err := h.Token.GetToken(ctx, echoCtx)
	if err == nil {
		id, err := uuid.Parse(claim.UserOrganizatonID)
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
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		userOrganization, err := h.model.UserOrganizationManager.FindOne(ctx, &model.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})

		if err != nil {
			return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return userOrganization, nil
	}

	h.ClearCurrentToken(ctx, echoCtx)
	return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
}

func (h *UserOrganizatonToken) SetUserOrganization(ctx context.Context, echoCtx echo.Context, userOrganization *model.UserOrganization) error {
	h.ClearCurrentToken(ctx, echoCtx)
	if err := h.Token.SetToken(ctx, echoCtx, UserOrganizatonClaim{
		UserOrganizatonID: userOrganization.ID.String(),
		UserID:            userOrganization.UserID.String(),
		BranchID:          userOrganization.BranchID.String(),
		OrganizationID:    userOrganization.OrganizationID.String(),
		UserType:          userOrganization.UserType,
	}, 144*time.Hour); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}
