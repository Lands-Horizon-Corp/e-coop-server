package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
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
		user, err := e.userToken.CurrentUser(context, ctx)
		if err != nil || user == nil {
			return
		}

		userId := user.ID
		userOrganization, _ := e.userOrganizationToken.CurrentUserOrganization(context, ctx)

		var userType string
		var orgId, branchId *uuid.UUID
		if userOrganization != nil {
			userType = userOrganization.UserType
			orgId = &userOrganization.OrganizationID
			branchId = userOrganization.BranchID
		}

		claim, _ := e.userToken.CSRF.GetCSRF(context, ctx)
		var latitude, longitude *float64
		var ipAddress, userAgent, referer, location, acceptLanguage string

		if claim != (cooperative_tokens.UserCSRF{}) {
			latitude = &claim.Latitude
			longitude = &claim.Longitude
			ipAddress = claim.IPAddress
			userAgent = claim.UserAgent
			referer = claim.Referer
			location = claim.Location
			acceptLanguage = claim.AcceptLanguage
		}

		if err := e.model.FootstepManager.Create(context, &model.Footstep{
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
