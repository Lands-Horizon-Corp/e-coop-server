package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"

	"github.com/labstack/echo/v4"
)

func (c *Controller) generateSavingsInterest() {

	req := c.provider.Service.Request
	// req.RegisterRoute(handlers.Route{
	// 	Method:       "POST",
	// 	Route:        "/api/v1/generate-savings-interest",
	// 	ResponseType: core.GeneratedSavingsInterestEntry{},
	// 	RequestType:  core.GeneratedSavingsInterestRequest{},
	// 	Note:         "Generates savings interest for all applicable accounts.",
	// }, func(ctx echo.Context) error {
	// 	context := ctx.Request().Context()
	// 	request, err := c.core.BrowseReferenceManager.Validate(ctx)
	// 	if err != nil {
	// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
	// 	}

	// 	userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
	// 	if err != nil {
	// 		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
	// 	}
	// 	if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
	// 		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create browse references"})
	// 	}

	// 	data:= &core.GeneratedSavingsInterest{

	// 	}

	// })

	// PUT /api/v1/generate-savings-interest/print
	// PUT /api/v1/generate-savings-interest/print-undo

	// PUT /api/v1/generate-savings-interest/post
	req.RegisterRoute(handlers.Route{
		Method:      "PUT",
		Route:       "/api/v1/generate-savings-interest/:generated_savings_interest_id/post",
		RequestType: core.GenerateSavingsInterestPostRequest{},
		Note:        "Posts generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		var req core.UserSettingsChangeProfileRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid change profile payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post generated savings interest entries"})
		}

		if err := c.event.GenerateSavingsInterestPost(context, userOrg, *generatedSavingsInterestID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post generated savings interest entries: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
