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

	// GET /api/v1/generated-savings-interest: List all generated savings interest for the current user's branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest",
		Method:       "GET",
		Note:         "Returns all generated savings interest for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.GeneratedSavingsInterestResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedSavingsInterests, err := c.core.GeneratedSavingsInterestManager.Find(context, &core.GeneratedSavingsInterest{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated savings interest found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestManager.ToModels(generatedSavingsInterests))
	})

	// GET /api/v1/generated-savings-interest/:generated_savings_interest_id: Get specific generated savings interest by ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/generated-savings-interest/:generated_savings_interest_id",
		Method:       "GET",
		Note:         "Returns a single generated savings interest by its ID.",
		ResponseType: core.GeneratedSavingsInterestResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := handlers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		generatedSavingsInterest, err := c.core.GeneratedSavingsInterestManager.GetByIDRaw(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest not found"})
		}
		return ctx.JSON(http.StatusOK, generatedSavingsInterest)
	})

	// GET /api/v1/generated-savings-interest/:genereated_savings_interest/view
	req.RegisterWebRoute(handlers.Route{
		Method:       "GET",
		Route:        "/api/v1/generated-savings-interest/:generated_savings_interest_id/view",
		ResponseType: core.GeneratedSavingsInterestViewResponse{},
		Note:         "Returns generated savings interest entries for a specific generated savings interest ID.",
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
		entries, err := c.core.GeneratedSavingsInterestEntryManager.Find(context, &core.GeneratedSavingsInterestEntry{
			GeneratedSavingsInterestID: *generatedSavingsInterestID,
			OrganizationID:             userOrg.OrganizationID,
			BranchID:                   *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest entries: " + err.Error()})
		}
		totalTax, totalInterest := 0.0, 0.0

		for _, entry := range entries {
			totalTax = c.provider.Service.Decimal.Add(totalTax, entry.InterestTax)
			totalInterest = c.provider.Service.Decimal.Add(totalInterest, entry.InterestAmount)

		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestViewResponse{
			Entries:       c.core.GeneratedSavingsInterestEntryManager.ToModels(entries),
			TotalTax:      totalTax,      // You might want to calculate this value
			TotalInterest: totalInterest, // You might want to calculate this value
		})
	})
	req.RegisterWebRoute(handlers.Route{
		Method:       "POST",
		Route:        "/api/v1/generated-savings-interest/view",
		ResponseType: core.GeneratedSavingsInterestViewResponse{},
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
		branch, err := c.core.BranchManager.GetByID(context, *userOrg.BranchID, "BranchSetting")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branch information: " + err.Error()})
		}
		if branch.BranchSetting == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Branch settings not found"})
		}
		annualDivisor := branch.BranchSetting.AnnualDivisor
		entries, err := c.event.GenerateSavingsInterestEntries(context, userOrg, core.GeneratedSavingsInterest{
			LastComputationDate:             request.LastComputationDate,
			NewComputationDate:              request.NewComputationDate,
			AccountID:                       request.AccountID,
			MemberTypeID:                    request.MemberTypeID,
			SavingsComputationType:          request.SavingsComputationType,
			IncludeClosedAccount:            request.IncludeClosedAccount,
			IncludeExistingComputedInterest: request.IncludeExistingComputedInterest,
			InterestTaxRate:                 request.InterestTaxRate,
		}, annualDivisor)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate savings interest entries: " + err.Error()})
		}
		totalTax, totalInterest := 0.0, 0.0

		for _, entry := range entries {
			totalTax = c.provider.Service.Decimal.Add(totalTax, entry.InterestTax)
			totalInterest = c.provider.Service.Decimal.Add(totalInterest, entry.InterestAmount)

		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestViewResponse{
			Entries:       c.core.GeneratedSavingsInterestEntryManager.ToModels(entries),
			TotalTax:      totalTax,      // You might want to calculate this value
			TotalInterest: totalInterest, // You might want to calculate this value
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Method:       "POST",
		Route:        "/api/v1/generated-savings-interest",
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
		branch, err := c.core.BranchManager.GetByID(context, *userOrg.BranchID, "BranchSetting")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branch information: " + err.Error()})
		}
		if branch.BranchSetting == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Branch settings not found"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		totalTax, totalInterest := 0.0, 0.0
		generatedSavingsInterest := &core.GeneratedSavingsInterest{
			CreatedAt:                       time.Now().UTC(),
			CreatedByID:                     userOrg.UserID,
			UpdatedAt:                       time.Now().UTC(),
			UpdatedByID:                     userOrg.UserID,
			OrganizationID:                  userOrg.OrganizationID,
			BranchID:                        *userOrg.BranchID,
			LastComputationDate:             request.LastComputationDate,
			NewComputationDate:              request.NewComputationDate,
			AccountID:                       request.AccountID,
			MemberTypeID:                    request.MemberTypeID,
			SavingsComputationType:          request.SavingsComputationType,
			IncludeClosedAccount:            request.IncludeClosedAccount,
			IncludeExistingComputedInterest: request.IncludeExistingComputedInterest,
			InterestTaxRate:                 request.InterestTaxRate,
			TotalInterest:                   totalInterest,
			TotalTax:                        totalTax,
		}
		if err := c.core.GeneratedSavingsInterestManager.CreateWithTx(context, tx, generatedSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest: " + err.Error()})
		}
		annualDivisor := branch.BranchSetting.AnnualDivisor
		entries, err := c.event.GenerateSavingsInterestEntries(context, userOrg, core.GeneratedSavingsInterest{
			LastComputationDate:             request.LastComputationDate,
			NewComputationDate:              request.NewComputationDate,
			AccountID:                       request.AccountID,
			MemberTypeID:                    request.MemberTypeID,
			SavingsComputationType:          request.SavingsComputationType,
			IncludeClosedAccount:            request.IncludeClosedAccount,
			IncludeExistingComputedInterest: request.IncludeExistingComputedInterest,
			InterestTaxRate:                 request.InterestTaxRate,
		}, annualDivisor)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate savings interest entries: " + err.Error()})
		}

		for _, entry := range entries {
			totalTax = c.provider.Service.Decimal.Add(totalTax, entry.InterestTax)
			totalInterest = c.provider.Service.Decimal.Add(totalInterest, entry.InterestAmount)

			entry.GeneratedSavingsInterestID = generatedSavingsInterest.ID
			entry.OrganizationID = userOrg.OrganizationID
			entry.BranchID = *userOrg.BranchID
			entry.CreatedAt = time.Now().UTC()
			entry.CreatedByID = userOrg.UserID
			entry.UpdatedAt = time.Now().UTC()
			entry.UpdatedByID = userOrg.UserID
			if err := c.core.GeneratedSavingsInterestEntryManager.CreateWithTx(context, tx, entry); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest entry: " + err.Error()})
			}
		}
		generatedSavingsInterest.TotalTax = totalTax
		generatedSavingsInterest.TotalInterest = totalInterest
		if err := c.core.GeneratedSavingsInterestManager.UpdateByIDWithTx(context, tx, generatedSavingsInterest.ID, generatedSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest: " + err.Error()})
		}

		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedSavingsInterestEntryManager.ToModels(entries))
	})

	// PUT /api/v1/generated-savings-interest/:generated_savings_interest_id/print
	req.RegisterWebRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/generated-savings-interest/:generated_savings_interest_id/print",
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
	// PUT /api/v1/generated-savings-interest/:generated_savings_interest_id/print-undo
	req.RegisterWebRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/generated-savings-interest/:generated_savings_interest_id/print-undo",
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

	// PUT /api/v1/generated-savings-interest/:generated_savings_interest_id/post
	req.RegisterWebRoute(handlers.Route{
		Method:      "PUT",
		Route:       "/api/v1/generated-savings-interest/:generated_savings_interest_id/post",
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
