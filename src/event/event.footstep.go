package event

import (
	"context"
	"fmt"
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

func (e *Event) Footstep(ctx context.Context, echoCtx echo.Context, data FootstepEvent) {
	go func() {
		fmt.Println("[Footstep] Logging event:", data.Activity, data.Module, data.Description) // <-- Add this line

		userOrganization, _ := e.userOrganizationToken.Token.GetToken(ctx, echoCtx)

		user, err := e.userToken.CurrentUser(ctx, echoCtx)
		if err != nil {
			fmt.Println("Failed to get current user:", err)
			return
		}

		userId := user.ID

		var orgId, branchId *uuid.UUID
		var userType string
		if userOrganization != nil {
			if parsedOrgId, err := uuid.Parse(userOrganization.OrganizationID); err == nil {
				orgId = &parsedOrgId
			}
			if parsedBranchId, err := uuid.Parse(userOrganization.BranchID); err == nil {
				branchId = &parsedBranchId
			}
			userType = userOrganization.UserType
		}

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
			UserType:       userType,
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
			fmt.Println("Failed to save footstep:", err)
			return
		}
		fmt.Println("[Footstep] Event saved successfully!") // <-- Add this line
	}()
}
