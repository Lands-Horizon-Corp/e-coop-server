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
	AccountType       string `json:"account_type"`
	jwt.RegisteredClaims
}

func (c UserOrganizatonClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserOrganizatonToken struct {
	model *model.Model
	Token horizon.TokenService[UserOrganizatonClaim]
}

func NewUserOrganizatonToken(provider *src.Provider, model *model.Model) (*UserOrganizatonToken, error) {
	context := context.Background()
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")
	isStaging := provider.Service.Environment.GetString("APP_ENV", "development") == "staging"

	token, err := provider.Service.Security.GenerateUUIDv5(context, appToken+"-user-organization")
	if err != nil {
		return nil, err
	}

	tokenService := horizon.NewTokenService[UserOrganizatonClaim](
		fmt.Sprintf("%s-%s", "X-SECURE-TOKEN-ORGANIZATION", appName),
		[]byte(token),
		isStaging,
	)
	return &UserOrganizatonToken{Token: tokenService, model: model}, nil
}

func (h *UserOrganizatonToken) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*model.UserOrganization, error) {
	claim, err := h.Token.GetToken(ctx, echoCtx)
	if err != nil {
		h.Token.CleanToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	id, err := uuid.Parse(claim.UserOrganizatonID)
	if err != nil {
		h.Token.CleanToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
	}
	userOrganization, err := h.model.UserOrganizationManager.GetByID(ctx, id)
	if err != nil {
		h.Token.CleanToken(ctx, echoCtx)
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return userOrganization, nil
}

func (h *UserOrganizatonToken) SetUserOrganization(ctx context.Context, echoCtx echo.Context, userOrganization *model.UserOrganization) error {
	h.Token.CleanToken(ctx, echoCtx)
	if err := h.Token.SetToken(ctx, echoCtx, UserOrganizatonClaim{
		UserOrganizatonID: userOrganization.ID.String(),
		UserID:            userOrganization.UserID.String(),
		BranchID:          userOrganization.BranchID.String(),
		OrganizationID:    userOrganization.OrganizationID.String(),
		AccountType:       userOrganization.UserType,
	}, 10*time.Hour); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token")
	}
	return nil
}
