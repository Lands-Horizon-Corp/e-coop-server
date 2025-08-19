package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ChargesRateSchemeController registers routes for managing charges rate schemes.
func (c *Controller) ChargesRateSchemeController() {
	req := c.provider.Service.Request

	// GET /charges-rate-scheme: Paginated list of charges rate schemes for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme",
		Method:       "GET",
		Note:         "Returns a paginated list of charges rate schemes for the current user's organization and branch.",
		ResponseType: model.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		chargesRateSchemes, err := c.model.ChargesRateSchemeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch charges rate schemes for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.ChargesRateSchemeManager.ToModels(chargesRateSchemes))
	})

	// GET /charges-rate-scheme/:charges_rate_scheme_id: Get specific charges rate scheme by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "GET",
		Note:         "Returns a single charges rate scheme by its ID.",
		ResponseType: model.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		chargesRateScheme, err := c.model.ChargesRateSchemeManager.GetByIDRaw(context, *chargesRateSchemeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}
		return ctx.JSON(http.StatusOK, chargesRateScheme)
	})

	// POST /charges-rate-scheme: Create a new charges rate scheme. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme",
		Method:       "POST",
		Note:         "Creates a new charges rate scheme for the current user's organization and branch.",
		RequestType:  model.ChargesRateSchemeRequest{},
		ResponseType: model.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.ChargesRateSchemeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), validation error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), user org error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), user not assigned to branch.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateScheme := &model.ChargesRateScheme{
			ChargesRateByTermHeaderID: req.ChargesRateByTermHeaderID,
			MemberTypeID:              req.MemberTypeID,
			ModeOfPayment:             req.ModeOfPayment,
			Name:                      req.Name,
			Description:               req.Description,
			Icon:                      req.Icon,
			// ModeOfPayment header fields
			ModeOfPaymentHeader1:  req.ModeOfPaymentHeader1,
			ModeOfPaymentHeader2:  req.ModeOfPaymentHeader2,
			ModeOfPaymentHeader3:  req.ModeOfPaymentHeader3,
			ModeOfPaymentHeader4:  req.ModeOfPaymentHeader4,
			ModeOfPaymentHeader5:  req.ModeOfPaymentHeader5,
			ModeOfPaymentHeader6:  req.ModeOfPaymentHeader6,
			ModeOfPaymentHeader7:  req.ModeOfPaymentHeader7,
			ModeOfPaymentHeader8:  req.ModeOfPaymentHeader8,
			ModeOfPaymentHeader9:  req.ModeOfPaymentHeader9,
			ModeOfPaymentHeader10: req.ModeOfPaymentHeader10,
			ModeOfPaymentHeader11: req.ModeOfPaymentHeader11,
			ModeOfPaymentHeader12: req.ModeOfPaymentHeader12,
			ModeOfPaymentHeader13: req.ModeOfPaymentHeader13,
			ModeOfPaymentHeader14: req.ModeOfPaymentHeader14,
			ModeOfPaymentHeader15: req.ModeOfPaymentHeader15,
			ModeOfPaymentHeader16: req.ModeOfPaymentHeader16,
			ModeOfPaymentHeader17: req.ModeOfPaymentHeader17,
			ModeOfPaymentHeader18: req.ModeOfPaymentHeader18,
			ModeOfPaymentHeader19: req.ModeOfPaymentHeader19,
			ModeOfPaymentHeader20: req.ModeOfPaymentHeader20,
			ModeOfPaymentHeader21: req.ModeOfPaymentHeader21,
			ModeOfPaymentHeader22: req.ModeOfPaymentHeader22,
			// ByTerm header fields
			ByTermHeader1:  req.ByTermHeader1,
			ByTermHeader2:  req.ByTermHeader2,
			ByTermHeader3:  req.ByTermHeader3,
			ByTermHeader4:  req.ByTermHeader4,
			ByTermHeader5:  req.ByTermHeader5,
			ByTermHeader6:  req.ByTermHeader6,
			ByTermHeader7:  req.ByTermHeader7,
			ByTermHeader8:  req.ByTermHeader8,
			ByTermHeader9:  req.ByTermHeader9,
			ByTermHeader10: req.ByTermHeader10,
			ByTermHeader11: req.ByTermHeader11,
			ByTermHeader12: req.ByTermHeader12,
			ByTermHeader13: req.ByTermHeader13,
			ByTermHeader14: req.ByTermHeader14,
			ByTermHeader15: req.ByTermHeader15,
			ByTermHeader16: req.ByTermHeader16,
			ByTermHeader17: req.ByTermHeader17,
			ByTermHeader18: req.ByTermHeader18,
			ByTermHeader19: req.ByTermHeader19,
			ByTermHeader20: req.ByTermHeader20,
			ByTermHeader21: req.ByTermHeader21,
			ByTermHeader22: req.ByTermHeader22,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.ChargesRateSchemeManager.Create(context, chargesRateScheme); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme: " + err.Error()})
		}

		// Create associated accounts if provided
		if len(req.AccountIDs) > 0 {
			for _, accountID := range req.AccountIDs {
				chargesRateSchemeAccount := &model.ChargesRateSchemeAccount{
					ChargesRateSchemeID: chargesRateScheme.ID,
					AccountID:           accountID,
					CreatedAt:           time.Now().UTC(),
					CreatedByID:         user.UserID,
					UpdatedAt:           time.Now().UTC(),
					UpdatedByID:         user.UserID,
					BranchID:            *user.BranchID,
					OrganizationID:      user.OrganizationID,
				}
				if err := c.model.ChargesRateSchemeAccountManager.Create(context, chargesRateSchemeAccount); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Charges rate scheme account creation failed (/charges-rate-scheme), db error: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme account: " + err.Error()})
				}
			}
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate scheme (/charges-rate-scheme): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})
		return ctx.JSON(http.StatusCreated, c.model.ChargesRateSchemeManager.ToModel(chargesRateScheme))
	})

	// PUT /charges-rate-scheme/:charges_rate_scheme_id: Update charges rate scheme by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme by its ID.",
		RequestType:  model.ChargesRateSchemeRequest{},
		ResponseType: model.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), invalid charges rate scheme ID.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}

		req, err := c.model.ChargesRateSchemeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), validation error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), user org error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateScheme, err := c.model.ChargesRateSchemeManager.GetByID(context, *chargesRateSchemeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), charges rate scheme not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}

		chargesRateScheme.ChargesRateByTermHeaderID = req.ChargesRateByTermHeaderID
		chargesRateScheme.MemberTypeID = req.MemberTypeID
		chargesRateScheme.ModeOfPayment = req.ModeOfPayment
		chargesRateScheme.Name = req.Name
		chargesRateScheme.Description = req.Description
		chargesRateScheme.Icon = req.Icon
		// ModeOfPayment header fields
		chargesRateScheme.ModeOfPaymentHeader1 = req.ModeOfPaymentHeader1
		chargesRateScheme.ModeOfPaymentHeader2 = req.ModeOfPaymentHeader2
		chargesRateScheme.ModeOfPaymentHeader3 = req.ModeOfPaymentHeader3
		chargesRateScheme.ModeOfPaymentHeader4 = req.ModeOfPaymentHeader4
		chargesRateScheme.ModeOfPaymentHeader5 = req.ModeOfPaymentHeader5
		chargesRateScheme.ModeOfPaymentHeader6 = req.ModeOfPaymentHeader6
		chargesRateScheme.ModeOfPaymentHeader7 = req.ModeOfPaymentHeader7
		chargesRateScheme.ModeOfPaymentHeader8 = req.ModeOfPaymentHeader8
		chargesRateScheme.ModeOfPaymentHeader9 = req.ModeOfPaymentHeader9
		chargesRateScheme.ModeOfPaymentHeader10 = req.ModeOfPaymentHeader10
		chargesRateScheme.ModeOfPaymentHeader11 = req.ModeOfPaymentHeader11
		chargesRateScheme.ModeOfPaymentHeader12 = req.ModeOfPaymentHeader12
		chargesRateScheme.ModeOfPaymentHeader13 = req.ModeOfPaymentHeader13
		chargesRateScheme.ModeOfPaymentHeader14 = req.ModeOfPaymentHeader14
		chargesRateScheme.ModeOfPaymentHeader15 = req.ModeOfPaymentHeader15
		chargesRateScheme.ModeOfPaymentHeader16 = req.ModeOfPaymentHeader16
		chargesRateScheme.ModeOfPaymentHeader17 = req.ModeOfPaymentHeader17
		chargesRateScheme.ModeOfPaymentHeader18 = req.ModeOfPaymentHeader18
		chargesRateScheme.ModeOfPaymentHeader19 = req.ModeOfPaymentHeader19
		chargesRateScheme.ModeOfPaymentHeader20 = req.ModeOfPaymentHeader20
		chargesRateScheme.ModeOfPaymentHeader21 = req.ModeOfPaymentHeader21
		chargesRateScheme.ModeOfPaymentHeader22 = req.ModeOfPaymentHeader22
		// ByTerm header fields
		chargesRateScheme.ByTermHeader1 = req.ByTermHeader1
		chargesRateScheme.ByTermHeader2 = req.ByTermHeader2
		chargesRateScheme.ByTermHeader3 = req.ByTermHeader3
		chargesRateScheme.ByTermHeader4 = req.ByTermHeader4
		chargesRateScheme.ByTermHeader5 = req.ByTermHeader5
		chargesRateScheme.ByTermHeader6 = req.ByTermHeader6
		chargesRateScheme.ByTermHeader7 = req.ByTermHeader7
		chargesRateScheme.ByTermHeader8 = req.ByTermHeader8
		chargesRateScheme.ByTermHeader9 = req.ByTermHeader9
		chargesRateScheme.ByTermHeader10 = req.ByTermHeader10
		chargesRateScheme.ByTermHeader11 = req.ByTermHeader11
		chargesRateScheme.ByTermHeader12 = req.ByTermHeader12
		chargesRateScheme.ByTermHeader13 = req.ByTermHeader13
		chargesRateScheme.ByTermHeader14 = req.ByTermHeader14
		chargesRateScheme.ByTermHeader15 = req.ByTermHeader15
		chargesRateScheme.ByTermHeader16 = req.ByTermHeader16
		chargesRateScheme.ByTermHeader17 = req.ByTermHeader17
		chargesRateScheme.ByTermHeader18 = req.ByTermHeader18
		chargesRateScheme.ByTermHeader19 = req.ByTermHeader19
		chargesRateScheme.ByTermHeader20 = req.ByTermHeader20
		chargesRateScheme.ByTermHeader21 = req.ByTermHeader21
		chargesRateScheme.ByTermHeader22 = req.ByTermHeader22
		chargesRateScheme.UpdatedAt = time.Now().UTC()
		chargesRateScheme.UpdatedByID = user.UserID
		if err := c.model.ChargesRateSchemeManager.UpdateFields(context, chargesRateScheme.ID, chargesRateScheme); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme: " + err.Error()})
		}

		// Handle account associations - delete existing and create new ones
		if req.AccountIDs != nil {
			// Delete existing associations
			existingAccounts, err := c.model.ChargesRateSchemeAccountManager.Find(context, &model.ChargesRateSchemeAccount{
				ChargesRateSchemeID: chargesRateScheme.ID,
			})
			if err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to fetch existing charges rate scheme accounts (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
					Module:      "ChargesRateScheme",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch existing accounts: " + err.Error()})
			}

			for _, existingAccount := range existingAccounts {
				if err := c.model.ChargesRateSchemeAccountManager.DeleteByID(context, existingAccount.ID); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete existing charges rate scheme account (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete existing account association: " + err.Error()})
				}
			}

			// Create new associations
			for _, accountID := range req.AccountIDs {
				chargesRateSchemeAccount := &model.ChargesRateSchemeAccount{
					ChargesRateSchemeID: chargesRateScheme.ID,
					AccountID:           accountID,
					CreatedAt:           time.Now().UTC(),
					CreatedByID:         user.UserID,
					UpdatedAt:           time.Now().UTC(),
					UpdatedByID:         user.UserID,
					BranchID:            *user.BranchID,
					OrganizationID:      user.OrganizationID,
				}
				if err := c.model.ChargesRateSchemeAccountManager.Create(context, chargesRateSchemeAccount); err != nil {
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Charges rate scheme account creation failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme account: " + err.Error()})
				}
			}
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme (/charges-rate-scheme/:charges_rate_scheme_id): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})
		return ctx.JSON(http.StatusOK, c.model.ChargesRateSchemeManager.ToModel(chargesRateScheme))
	})

	// DELETE /charges-rate-scheme/:charges_rate_scheme_id: Delete a charges rate scheme by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate scheme by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), invalid charges rate scheme ID.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		chargesRateScheme, err := c.model.ChargesRateSchemeManager.GetByID(context, *chargesRateSchemeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}
		if err := c.model.ChargesRateSchemeManager.DeleteByID(context, *chargesRateSchemeID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate scheme (/charges-rate-scheme/:charges_rate_scheme_id): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /charges-rate-scheme/bulk-delete: Bulk delete charges rate schemes by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/charges-rate-scheme/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate schemes by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), invalid request body.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), no IDs provided.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate scheme IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			chargesRateSchemeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), invalid UUID: " + rawID,
					Module:      "ChargesRateScheme",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			chargesRateScheme, err := c.model.ChargesRateSchemeManager.GetByID(context, chargesRateSchemeID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), not found: " + rawID,
					Module:      "ChargesRateScheme",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Charges rate scheme not found with ID: %s", rawID)})
			}
			names += chargesRateScheme.Name + ","
			if err := c.model.ChargesRateSchemeManager.DeleteByIDWithTx(context, tx, chargesRateSchemeID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), db error: " + err.Error(),
					Module:      "ChargesRateScheme",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete), commit error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate schemes (/charges-rate-scheme/bulk-delete): " + names,
			Module:      "ChargesRateScheme",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
