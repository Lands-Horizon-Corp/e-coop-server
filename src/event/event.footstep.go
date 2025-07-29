package event

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
)

type FootstepEvent struct {
	Description string
	Activity    string
	Module      string
}

func (e *Event) Footstep(context context.Context, ctx echo.Context, data FootstepEvent) {
	fmt.Println("[Footstep] Logging event:", data.Activity, data.Module, data.Description)

	go func() {
		// fmt.Println("[Footstep] Logging event:", data.Activity, data.Module, data.Description)

		// user, err := e.userToken.CurrentUser(context, ctx)
		// if err != nil {
		// 	fmt.Println("Failed to get current user:", err)
		// 	return
		// }

		// userOrganization, _ := e.userOrganizationToken.CurrentUserOrganization(context, ctx)

		// userId := user.ID

		// var userType string
		// if userOrganization != nil {
		// 	userType = userOrganization.UserType
		// }

		// claim, _ := e.userToken.CSRF.GetCSRF(context, ctx)
		// latitude := claim.Latitude
		// longitude := claim.Longitude
		// ipAddress := claim.IPAddress
		// userAgent := claim.UserAgent
		// referer := claim.Referer
		// location := claim.Location
		// acceptLanguage := claim.AcceptLanguage

		// if err := e.model.FootstepManager.Create(context, &model.Footstep{
		// 	CreatedAt:      time.Now().UTC(),
		// 	CreatedByID:    userId,
		// 	UpdatedAt:      time.Now().UTC(),
		// 	UpdatedByID:    userId,
		// 	OrganizationID: &userOrganization.OrganizationID,
		// 	BranchID:       userOrganization.BranchID,
		// 	UserID:         &userId,
		// 	Description:    data.Description,
		// 	Activity:       data.Activity,
		// 	UserType:       userType,
		// 	Module:         data.Module,
		// 	Latitude:       &latitude,
		// 	Longitude:      &longitude,
		// 	Timestamp:      time.Now().UTC(),
		// 	IPAddress:      ipAddress,
		// 	UserAgent:      userAgent,
		// 	Referer:        referer,
		// 	Location:       location,
		// 	AcceptLanguage: acceptLanguage,
		// }); err != nil {
		// 	fmt.Println("Failed to save footstep:", err)
		// 	return
		// }
		// fmt.Println("[Footstep] Event saved successfully!") // <-- Add this line
	}()
}
