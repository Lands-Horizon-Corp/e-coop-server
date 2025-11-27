package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"

	"github.com/labstack/echo/v4"
)

func (c *Controller) generateSavingsInterest() {

	req := c.provider.Service.Request
	req.RegisterRoute(handlers.Route{
		Method:       "POST",
		Route:        "/api/v1/generate-savings-interest/view",
		ResponseType: core.GeneratedSavingsInterestEntry{},
		RequestType:  core.GeneratedSavingsInterestRequest{},
		Note:         "Generates savings interest for all applicable accounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := c.core.GeneratedSavingsInterestManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create browse references"})
		}
		data, err := c.event.GenerateSavingsInterestEntries(context, userOrg, core.GeneratedSavingsInterest{
			LastComputationDate:             request.LastComputationDate,
			NewComputationDate:              request.NewComputationDate,
			AccountID:                       request.AccountID,
			MemberTypeID:                    request.MemberTypeID,
			SavingsComputationType:          request.SavingsComputationType,
			IncludeClosedAccount:            request.IncludeClosedAccount,
			IncludeExistingComputedInterest: request.IncludeExistingComputedInterest,
			InterestTaxRate:                 request.InterestTaxRate,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate savings interest entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestEntryManager.ToModels(data))
	})

	// PUT /api/v1/generate-savings-interest/:generated_savings_interest_id/print
	req.RegisterRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/generate-savings-interest/:generated_savings_interest_id/print",
		Note:   "Prints generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post generated savings interest entries"})
		}
		generateSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByID(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest: " + err.Error()})
		}
		if generateSavingsInterest == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest not found"})
		}
		if generateSavingsInterest.PrintedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest has already been printed"})
		}
		now := time.Now().UTC()
		generateSavingsInterest.PrintedByUserID = &userOrg.UserID
		generateSavingsInterest.PrintedDate = &now
		if err := c.core.GeneratedSavingsInterestManager.UpdateByID(context, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest as printed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestManager.ToModel(generateSavingsInterest))
	})
	// PUT /api/v1/generate-savings-interest/:generated_savings_interest_id/print-undo
	req.RegisterRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/generate-savings-interest/:generated_savings_interest_id/print-undo",
		Note:   "Undoes the print status of generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print status of generated savings interest entries"})
		}
		generateSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByID(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest: " + err.Error()})
		}
		if generateSavingsInterest == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest not found"})
		}
		if generateSavingsInterest.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest has not been printed yet"})
		}
		if generateSavingsInterest.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot undo print status - generated savings interest has already been posted"})
		}
		generateSavingsInterest.PrintedByUserID = nil
		generateSavingsInterest.PrintedDate = nil
		if err := c.core.GeneratedSavingsInterestManager.UpdateByID(context, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print status of generated savings interest: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestManager.ToModel(generateSavingsInterest))
	})

	// PUT /api/v1/generate-savings-interest/:generated_savings_interest_id/post
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
		var req core.GenerateSavingsInterestPostRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post request payload: " + err.Error()})
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
		generateSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByID(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest: " + err.Error()})
		}
		if generateSavingsInterest.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest must be printed before posting"})
		}
		if generateSavingsInterest.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest has already been posted"})
		}
		if err := c.event.GenerateSavingsInterestEntriesPost(
			context, userOrg, generatedSavingsInterestID, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post generated savings interest entries: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
