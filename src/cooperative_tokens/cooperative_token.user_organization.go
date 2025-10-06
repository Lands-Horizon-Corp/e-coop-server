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
	"go.uber.org/zap"
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
	h.provider.Service.Logger.Info("AUTH: ClearCurrentToken called")

	claim, err := h.Token.GetToken(context, ctx)
	if err == nil {
		h.provider.Service.Logger.Info("AUTH: Found token to clear",
			zap.String("user_organization_id", claim.UserOrganizationID),
			zap.String("user_id", claim.UserID),
		)

		id, err := uuid.Parse(claim.UserOrganizationID)
		if err != nil {
			h.provider.Service.Logger.Error("AUTH: Invalid UUID when clearing token",
				zap.String("user_organization_id", claim.UserOrganizationID),
				zap.Error(err),
			)
			h.Token.CleanToken(context, ctx)
			return
		}
		userOrg, err := h.model.UserOrganizationManager.GetByID(context, id)
		if err != nil {
			h.provider.Service.Logger.Error("AUTH: User organization not found when clearing token",
				zap.String("user_organization_id", id.String()),
				zap.Error(err),
			)
			h.Token.CleanToken(context, ctx)
			return
		}

		h.provider.Service.Logger.Info("AUTH: Setting user organization to offline",
			zap.String("user_organization_id", userOrg.ID.String()),
			zap.String("previous_status", string(userOrg.Status)),
		)

		userOrg.Status = model.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now().UTC()
		if err := h.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			h.provider.Service.Logger.Error("AUTH: Failed to update user organization status",
				zap.String("user_organization_id", userOrg.ID.String()),
				zap.Error(err),
			)
			h.Token.CleanToken(context, ctx)
			return
		}
		if err := h.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_organization.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_organization.status.organization.%s", userOrg.OrganizationID),
		}, nil); err != nil {
			h.provider.Service.Logger.Error("AUTH: Failed to dispatch status update",
				zap.String("user_organization_id", userOrg.ID.String()),
				zap.Error(err),
			)
			return
		}

		h.provider.Service.Logger.Info("AUTH: Successfully cleared token and updated status")
	} else {
		h.provider.Service.Logger.Info("AUTH: No token found to clear",
			zap.Error(err),
		)
	}

	h.provider.Service.Logger.Info("AUTH: Cleaning token cookie")
	h.Token.CleanToken(context, ctx)
}

func (h *UserOrganizationToken) CurrentUserOrganization(ctx context.Context, echoCtx echo.Context) (*model.UserOrganization, error) {
	h.provider.Service.Logger.Info("AUTH: Starting CurrentUserOrganization",
		zap.String("method", echoCtx.Request().Method),
		zap.String("path", echoCtx.Request().URL.Path),
		zap.String("remote_addr", echoCtx.Request().RemoteAddr),
	)

	// Try JWT token first
	claim, err := h.Token.GetToken(ctx, echoCtx)
	if err == nil {
		h.provider.Service.Logger.Info("AUTH: JWT token found, validating claim",
			zap.String("user_organization_id", claim.UserOrganizationID),
			zap.String("user_id", claim.UserID),
			zap.String("user_type", string(claim.UserType)),
		)

		// We have a JWT token, so we should clear it if there are any issues
		id, err := uuid.Parse(claim.UserOrganizationID)
		if err != nil {
			h.provider.Service.Logger.Error("AUTH: Invalid UUID in JWT token, clearing token",
				zap.String("user_organization_id", claim.UserOrganizationID),
				zap.Error(err),
			)
			h.ClearCurrentToken(ctx, echoCtx)
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid user ID in token")
		}

		userOrganization, err := h.model.UserOrganizationManager.GetByID(ctx, id)
		if err != nil {
			h.provider.Service.Logger.Error("AUTH: User organization not found in database, clearing token",
				zap.String("user_organization_id", id.String()),
				zap.Error(err),
			)
			h.ClearCurrentToken(ctx, echoCtx)
			return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		h.provider.Service.Logger.Info("AUTH: JWT authentication successful",
			zap.String("user_organization_id", userOrganization.ID.String()),
			zap.String("user_id", userOrganization.UserID.String()),
			zap.String("status", string(userOrganization.Status)),
		)
		return userOrganization, nil
	}

	h.provider.Service.Logger.Info("AUTH: No valid JWT token found, trying Bearer token",
		zap.Error(err),
	)

	// Try Bearer token as fallback
	authHeader := echoCtx.Request().Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		bearerToken := authHeader[7:]
		h.provider.Service.Logger.Info("AUTH: Bearer token found, validating",
			zap.String("token_prefix", bearerToken[:min(8, len(bearerToken))]),
		)

		userOrganization, err := h.model.UserOrganizationManager.FindOne(ctx, &model.UserOrganization{
			DeveloperSecretKey: bearerToken,
		})

		if err != nil {
			h.provider.Service.Logger.Error("AUTH: Invalid bearer token",
				zap.String("token_prefix", bearerToken[:min(8, len(bearerToken))]),
				zap.Error(err),
			)
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}

		h.provider.Service.Logger.Info("AUTH: Bearer token authentication successful",
			zap.String("user_organization_id", userOrganization.ID.String()),
			zap.String("user_id", userOrganization.UserID.String()),
		)
		return userOrganization, nil
	}

	h.provider.Service.Logger.Info("AUTH: No authentication provided",
		zap.Bool("has_auth_header", len(authHeader) > 0),
		zap.String("auth_header_prefix", authHeader[:min(10, len(authHeader))]),
	)
	return nil, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
}

func (h *UserOrganizationToken) SetUserOrganization(ctx context.Context, echoCtx echo.Context, userOrganization *model.UserOrganization) error {
	h.provider.Service.Logger.Info("AUTH: Setting user organization token",
		zap.String("user_organization_id", userOrganization.ID.String()),
		zap.String("user_id", userOrganization.UserID.String()),
		zap.String("user_type", string(userOrganization.UserType)),
	)

	h.ClearCurrentToken(ctx, echoCtx)

	if err := h.Token.SetToken(ctx, echoCtx, UserOrganizationClaim{
		UserOrganizationID: userOrganization.ID.String(),
		UserID:             userOrganization.UserID.String(),
		BranchID:           userOrganization.BranchID.String(),
		OrganizationID:     userOrganization.OrganizationID.String(),
		UserType:           userOrganization.UserType,
	}, 144*time.Hour); err != nil {
		h.provider.Service.Logger.Error("AUTH: Failed to set authentication token",
			zap.String("user_organization_id", userOrganization.ID.String()),
			zap.Error(err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set authentication token: "+err.Error())
	}

	h.provider.Service.Logger.Info("AUTH: Successfully set user organization token")
	return nil
}

// Helper function for safe string truncation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
