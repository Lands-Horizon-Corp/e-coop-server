package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FootstepEvent struct {
	Description string
	Activity    string
	Module      string
}

func Footstep(ctx echo.Context, service *horizon.HorizonService, data FootstepEvent) {

	go func() {

		context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := CurrentUser(context, service, ctx)
		if err != nil || user == nil {
			return
		}

		userID := user.ID
		userOrganization, _ := CurrentUserOrganization(context, service, ctx)

		var userType types.UserOrganizationType
		var organizationID, branchID *uuid.UUID
		if userOrganization != nil {
			userType = userOrganization.UserType
			organizationID = &userOrganization.OrganizationID
			branchID = userOrganization.BranchID
		}

		claim, _ := CurrentUserCSRF(context, service, ctx)
		var latitude, longitude *float64
		var ipAddress, userAgent, referer, location, acceptLanguage string

		if claim != (UserCSRF{}) {
			latitude = &claim.Latitude
			longitude = &claim.Longitude
			ipAddress = claim.IPAddress
			userAgent = claim.UserAgent
			referer = claim.Referer
			location = claim.Location
			acceptLanguage = claim.AcceptLanguage
		}

		if err := core.FootstepManager(service).Create(context, &types.Footstep{
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
