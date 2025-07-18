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
		userOrganization, _ := e.userOrganizationToken.Token.GetToken(ctx, echoCtx)

		user, err := e.userToken.CurrentUser(ctx, echoCtx)
		if err != nil {
			return
		}

		// Get userId from user struct
		userId := user.ID

		// Get orgId, branchId, accountType from userOrganization if present
		var orgId, branchId *uuid.UUID
		var accountType string
		if userOrganization != nil {
			if parsedOrgId, err := uuid.Parse(userOrganization.OrganizationID); err == nil {
				orgId = &parsedOrgId
			}
			if parsedBranchId, err := uuid.Parse(userOrganization.BranchID); err == nil {
				branchId = &parsedBranchId
			}
			accountType = userOrganization.AccountType
		}

		// Get geo and agent info from CSRF claim
		claim, _ := e.userToken.CSRF.GetCSRF(ctx, echoCtx)
		latitude := claim.Latitude
		longitude := claim.Longitude
		ipAddress := claim.IPAddress
		userAgent := claim.UserAgent
		referer := claim.Referer
		location := claim.Location
		acceptLanguage := claim.AcceptLanguage

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
			AccountType:    accountType,
			Module:         data.Module,
			Latitude:       &latitude,
			Longitude:      &longitude,
			Timestamp:      time.Now().UTC(),
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			Referer:        referer,
			Location:       location,
			AcceptLanguage: acceptLanguage,
		}); err != nil {
			return
		}
	}()
}
