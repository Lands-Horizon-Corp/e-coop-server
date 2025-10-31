package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FootstepEvent struct {
	Description string
	Activity    string
	Module      string
}

func (e *Event) Footstep(context context.Context, ctx echo.Context, data FootstepEvent) {

	go func() {
		user, err := e.user_token.CurrentUser(context, ctx)
		if err != nil || user == nil {
			return
		}

		userID := user.ID
		userOrganization, _ := e.user_organization_token.CurrentUserOrganization(context, ctx)

		var userType modelcore.UserOrganizationType
		var organizationID, branchID *uuid.UUID
		if userOrganization != nil {
			userType = userOrganization.UserType
			organizationID = &userOrganization.OrganizationID
			branchID = userOrganization.BranchID
		}

		claim, _ := e.user_token.CSRF.GetCSRF(context, ctx)
		var latitude, longitude *float64
		var ipAddress, userAgent, referer, location, acceptLanguage string

		if claim != (tokens.UserCSRF{}) {
			latitude = &claim.Latitude
			longitude = &claim.Longitude
			ipAddress = claim.IPAddress
			userAgent = claim.UserAgent
			referer = claim.Referer
			location = claim.Location
			acceptLanguage = claim.AcceptLanguage
		}

		if err := e.modelcore.FootstepManager.Create(context, &modelcore.Footstep{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserID:         &userID,
			Description:    data.Description,
			Activity:       data.Activity,
			UserType:       userType,
			Module:         data.Module,
			Latitude:       latitude,
			Longitude:      longitude,
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
