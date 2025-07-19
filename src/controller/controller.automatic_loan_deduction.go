package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// AutomaticLoanDeductionController registers routes for managing automatic loan deductions.
func (c *Controller) AutomaticLoanDeductionController() {
	req := c.provider.Service.Request

	// GET /automatic-loan-deduction/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(horizon.Route{
		Route:    "/automatic-loan-deduction/computation-sheet/:computation_sheet_id",
		Method:   "GET",
		Response: "AutomaticLoanDeduction[]",
		Note:     "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := horizon.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		// Find all for this computation sheet, org, and branch
		alds, err := c.model.AutomaticLoanDeductionManager.Find(context, &model.AutomaticLoanDeduction{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.AutomaticLoanDeductionManager.ToModels(alds))
	})

	// GET /automatic-loan-deduction/computation-sheet/:computation_sheet_id/search
	req.RegisterRoute(horizon.Route{
		Route:    "/automatic-loan-deduction/computation-sheet/:computation_sheet_id/search",
		Method:   "GET",
		Response: "AutomaticLoanDeduction[]",
		Note:     "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := horizon.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		// Find all for this computation sheet, org, and branch
		alds, err := c.model.AutomaticLoanDeductionManager.Find(context, &model.AutomaticLoanDeduction{
			OrganizationID:     user.OrganizationID,
			BranchID:           *user.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.model.AutomaticLoanDeductionManager.Pagination(context, ctx, alds))
	})

	// POST /automatic-loan-deduction
	req.RegisterRoute(horizon.Route{
		Route:    "/automatic-loan-deduction",
		Method:   "POST",
		Request:  "AutomaticLoanDeduction",
		Response: "AutomaticLoanDeduction",
		Note:     "Creates a new automatic loan deduction for the current user's org/branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AutomaticLoanDeductionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user not assigned to branch.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		ald := &model.AutomaticLoanDeduction{
			AccountID:          req.AccountID,
			ComputationSheetID: req.ComputationSheetID,
			LinkAccountID:      req.LinkAccountID,
			ChargesPercentage1: req.ChargesPercentage1,
			ChargesPercentage2: req.ChargesPercentage2,
			ChargesAmount:      req.ChargesAmount,
			ChargesDivisor:     req.ChargesDivisor,
			MinAmount:          req.MinAmount,
			MaxAmount:          req.MaxAmount,
			Anum:               req.Anum,
			AddOn:              req.AddOn,
			AoRest:             req.AoRest,
			ExcludeRenewal:     req.ExcludeRenewal,
			Ct:                 req.Ct,
			Name:               req.Name,
			Description:        req.Description,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.UserID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		}

		if err := c.model.AutomaticLoanDeductionManager.Create(context, ald); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created automatic loan deduction (/automatic-loan-deduction): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusCreated, c.model.AutomaticLoanDeductionManager.ToModel(ald))
	})

	// PUT /automatic-loan-deduction/:automatic_loan_deduction_id
	req.RegisterRoute(horizon.Route{
		Route:    "/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method:   "PUT",
		Request:  "AutomaticLoanDeduction",
		Response: "AutomaticLoanDeduction",
		Note:     "Updates an existing automatic loan deduction by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}

		req, err := c.model.AutomaticLoanDeductionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		ald, err := c.model.AutomaticLoanDeductionManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		ald.AccountID = req.AccountID
		ald.ComputationSheetID = req.ComputationSheetID
		ald.LinkAccountID = req.LinkAccountID
		ald.ChargesPercentage1 = req.ChargesPercentage1
		ald.ChargesPercentage2 = req.ChargesPercentage2
		ald.ChargesAmount = req.ChargesAmount
		ald.ChargesDivisor = req.ChargesDivisor
		ald.MinAmount = req.MinAmount
		ald.MaxAmount = req.MaxAmount
		ald.Anum = req.Anum
		ald.AddOn = req.AddOn
		ald.AoRest = req.AoRest
		ald.ExcludeRenewal = req.ExcludeRenewal
		ald.Ct = req.Ct
		ald.Name = req.Name
		ald.Description = req.Description
		ald.UpdatedAt = time.Now().UTC()
		ald.UpdatedByID = user.UserID

		if err := c.model.AutomaticLoanDeductionManager.UpdateFields(context, ald.ID, ald); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusOK, c.model.AutomaticLoanDeductionManager.ToModel(ald))
	})

	// DELETE /automatic-loan-deduction/:automatic_loan_deduction_id
	req.RegisterRoute(horizon.Route{
		Route:  "/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method: "DELETE",
		Note:   "Deletes the specified automatic loan deduction by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}
		ald, err := c.model.AutomaticLoanDeductionManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		if err := c.model.AutomaticLoanDeductionManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /automatic-loan-deduction/bulk-delete
	req.RegisterRoute(horizon.Route{
		Route:   "/automatic-loan-deduction/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple automatic loan deductions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), invalid request body.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), no IDs provided.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), invalid UUID: " + rawID,
					Module:      "AutomaticLoanDeduction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			ald, err := c.model.AutomaticLoanDeductionManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), not found: " + rawID,
					Module:      "AutomaticLoanDeduction",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Automatic loan deduction not found with ID: %s", rawID)})
			}
			names += ald.Name + ","
			if err := c.model.AutomaticLoanDeductionManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), db error: " + err.Error(),
					Module:      "AutomaticLoanDeduction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete automatic loan deduction: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/automatic-loan-deduction/bulk-delete), commit error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted automatic loan deductions (/automatic-loan-deduction/bulk-delete): " + names,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
