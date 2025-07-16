package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
)

type FootstepEvent struct {
	Description string
	Activity    string
	Module      string
}

// Records a footstep event for a user, scoped to their organization and branch.
// This is only triggered if the user has valid CSRF and organization tokens.
func (e *Event) Footstep(ctx context.Context, echoCtx echo.Context, data FootstepEvent) {
	go func() {
		userOrganization, err := e.userOrganizationToken.Token.GetToken(ctx, echoCtx)
		if err != nil || userOrganization == nil {
			return
		}
		user, err := e.userToken.CSRF.GetCSRF(ctx, echoCtx)
		if err != nil {
			return
		}
		userId, err := uuid.Parse(userOrganization.UserID)
		if err != nil {
			return
		}
		orgId, err := uuid.Parse(userOrganization.OrganizationID)
		if err != nil {
			return
		}
		branchId, err := uuid.Parse(userOrganization.BranchID)
		if err != nil {
			return
		}
		if err := e.model.FootstepManager.Create(ctx, &model.Footstep{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userId,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userId,
			OrganizationID: orgId,
			BranchID:       branchId,
			UserID:         &userId,
			Description:    data.Description,
			Activity:       data.Activity,
			AccountType:    userOrganization.AccountType,
			Module:         data.Module,
			Latitude:       &user.Latitude,
			Longitude:      &user.Longitude,
			Timestamp:      time.Now().UTC(),
			IPAddress:      user.IPAddress,
			UserAgent:      user.UserAgent,
			Referer:        user.Referer,
			Location:       user.Location,
			AcceptLanguage: user.AcceptLanguage,
		}); err != nil {
			return
		}
	}()
}
