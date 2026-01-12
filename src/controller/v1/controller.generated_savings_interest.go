package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/shopspring/decimal"

	"github.com/labstack/echo/v4"
)

func generateSavingsInterest(service *horizon.HorizonService) {

	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest/search",
		Method:       "GET",
		Note:         "Returns all generated savings interest for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.GeneratedSavingsInterestResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedSavingsInterests, err := core.GeneratedSavingsInterestManager(service).NormalPagination(context, ctx, &core.GeneratedSavingsInterest{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated savings interest found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedSavingsInterests)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-savings-interest/:generated_savings_interest_id",
		Method:       "GET",
		Note:         "Returns a single generated savings interest by its ID.",
		ResponseType: core.GeneratedSavingsInterestResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		generatedSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByIDRaw(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated savings interest not found"})
		}
		return ctx.JSON(http.StatusOK, generatedSavingsInterest)
	})

	req.RegisterWebRoute(horizon.Route{
		Method:       "GET",
		Route:        "/api/v1/generated-savings-interest/:generated_savings_interest_id/view",
		ResponseType: core.GeneratedSavingsInterestViewResponse{},
		Note:         "Returns generated savings interest entries for a specific generated savings interest ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		entries, err := core.GeneratedSavingsInterestEntryManager(service).Find(context, &core.GeneratedSavingsInterestEntry{
			GeneratedSavingsInterestID: *generatedSavingsInterestID,
			OrganizationID:             userOrg.OrganizationID,
			BranchID:                   *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest entries: " + err.Error()})
		}
		totalTax := decimal.Zero
		totalInterest := decimal.Zero

		for _, entry := range entries {
			totalTax = totalTax.Add(decimal.NewFromFloat(entry.InterestTax))
			totalInterest = totalInterest.Add(decimal.NewFromFloat(entry.InterestAmount))
		}

		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestViewResponse{
			Entries:       core.GeneratedSavingsInterestEntryManager(service).ToModels(entries),
			TotalTax:      totalTax.InexactFloat64(),
			TotalInterest: totalInterest.InexactFloat64(),
		})
	})
	req.RegisterWebRoute(horizon.Route{
		Method:       "POST",
		Route:        "/api/v1/generated-savings-interest/view",
		ResponseType: core.GeneratedSavingsInterestViewResponse{},
		RequestType:  core.GeneratedSavingsInterestRequest{},
		Note:         "Generates savings interest for all applicable accounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := core.GeneratedSavingsInterestManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create browse references"})
		}
		branch, err := core.BranchManager(service).GetByID(context, *userOrg.BranchID, "BranchSetting")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branch information: " + err.Error()})
		}
		if branch.BranchSetting == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Branch settings not found"})
		}
		annualDivisor := branch.BranchSetting.AnnualDivisor
		entries, err := GenerateSavingsInterestEntries(context, userOrg, core.GeneratedSavingsInterest{
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
		totalTax := decimal.Zero
		totalInterest := decimal.Zero

		for _, entry := range entries {
			totalTax = totalTax.Add(decimal.NewFromFloat(entry.InterestTax))
			totalInterest = totalInterest.Add(decimal.NewFromFloat(entry.InterestAmount))
		}

		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestViewResponse{
			Entries:       core.GeneratedSavingsInterestEntryManager(service).ToModels(entries),
			TotalTax:      totalTax.InexactFloat64(),
			TotalInterest: totalInterest.InexactFloat64(),
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Method:       "POST",
		Route:        "/api/v1/generated-savings-interest",
		ResponseType: core.GeneratedSavingsInterestEntry{},
		RequestType:  core.GeneratedSavingsInterestRequest{},
		Note:         "Generates savings interest for all applicable accounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := core.GeneratedSavingsInterestManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create browse references"})
		}
		branch, err := core.BranchManager(service).GetByID(context, *userOrg.BranchID, "BranchSetting")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branch information: " + err.Error()})
		}
		if branch.BranchSetting == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Branch settings not found"})
		}
		tx, endTx := service.Database.StartTransaction(context)
		totalTax := decimal.Zero
		totalInterest := decimal.Zero
		generatedSavingsInterest := &core.GeneratedSavingsInterest{
			CreatedAt:                       time.Now().UTC(),
			CreatedByID:                     userOrg.UserID,
			UpdatedAt:                       time.Now().UTC(),
			UpdatedByID:                     userOrg.UserID,
			OrganizationID:                  userOrg.OrganizationID,
			BranchID:                        *userOrg.BranchID,
			DocumentNo:                      request.DocumentNo,
			LastComputationDate:             request.LastComputationDate,
			NewComputationDate:              request.NewComputationDate,
			AccountID:                       request.AccountID,
			MemberTypeID:                    request.MemberTypeID,
			SavingsComputationType:          request.SavingsComputationType,
			IncludeClosedAccount:            request.IncludeClosedAccount,
			IncludeExistingComputedInterest: request.IncludeExistingComputedInterest,
			InterestTaxRate:                 request.InterestTaxRate,
			TotalInterest:                   totalInterest.InexactFloat64(),
			TotalTax:                        totalTax.InexactFloat64(),
		}
		if err := core.GeneratedSavingsInterestManager(service).CreateWithTx(context, tx, generatedSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest: " + err.Error()})
		}
		annualDivisor := branch.BranchSetting.AnnualDivisor
		entries, err := GenerateSavingsInterestEntries(context, userOrg, core.GeneratedSavingsInterest{
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
			totalTax = totalTax.Add(decimal.NewFromFloat(entry.InterestTax))
			totalInterest = totalInterest.Add(decimal.NewFromFloat(entry.InterestAmount))

			entry.GeneratedSavingsInterestID = generatedSavingsInterest.ID
			entry.OrganizationID = userOrg.OrganizationID
			entry.BranchID = *userOrg.BranchID
			entry.CreatedAt = time.Now().UTC()
			entry.CreatedByID = userOrg.UserID
			entry.UpdatedAt = time.Now().UTC()
			entry.UpdatedByID = userOrg.UserID
			if err := core.GeneratedSavingsInterestEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated savings interest entry: " + err.Error()})
			}
		}
		generatedSavingsInterest.TotalTax = totalTax.InexactFloat64()
		generatedSavingsInterest.TotalInterest = totalInterest.InexactFloat64()
		if err := core.GeneratedSavingsInterestManager(service).UpdateByIDWithTx(context, tx, generatedSavingsInterest.ID, generatedSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest: " + err.Error()})
		}

		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestEntryManager(service).ToModels(entries))
	})

	req.RegisterWebRoute(horizon.Route{
		Method: "PUT",
		Route:  "/api/v1/generated-savings-interest/:generated_savings_interest_id/print",
		Note:   "Prints generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post generated savings interest entries"})
		}
		generateSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(context, *generatedSavingsInterestID)
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
		if err := core.GeneratedSavingsInterestManager(service).UpdateByID(context, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated savings interest as printed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestManager(service).ToModel(generateSavingsInterest))
	})
	req.RegisterWebRoute(horizon.Route{
		Method: "PUT",
		Route:  "/api/v1/generated-savings-interest/:generated_savings_interest_id/print-undo",
		Note:   "Undoes the print status of generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print status of generated savings interest entries"})
		}
		generateSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(context, *generatedSavingsInterestID)
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
		if err := core.GeneratedSavingsInterestManager(service).UpdateByID(context, generateSavingsInterest.ID, generateSavingsInterest); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print status of generated savings interest: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedSavingsInterestManager(service).ToModel(generateSavingsInterest))
	})

	req.RegisterWebRoute(horizon.Route{
		Method:      "PUT",
		Route:       "/api/v1/generated-savings-interest/:generated_savings_interest_id/post",
		RequestType: core.GenerateSavingsInterestPostRequest{},
		Note:        "Posts generated savings interest entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedSavingsInterestID, err := helpers.EngineUUIDParam(ctx, "generated_savings_interest_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated savings interest ID"})
		}
		var req core.GenerateSavingsInterestPostRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post request payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post generated savings interest entries"})
		}
		generateSavingsInterest, err := core.GeneratedSavingsInterestManager(service).GetByID(context, *generatedSavingsInterestID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated savings interest: " + err.Error()})
		}
		if generateSavingsInterest.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest must be printed before posting"})
		}
		if generateSavingsInterest.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Generated savings interest has already been posted"})
		}
		if err := GenerateSavingsInterestEntriesPost(
			context, userOrg, generatedSavingsInterestID, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post generated savings interest entries: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
