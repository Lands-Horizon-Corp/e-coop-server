package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
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
		ResponseType: modelCore.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		chargesRateSchemes, err := c.modelCore.ChargesRateSchemeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch charges rate schemes for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.ChargesRateSchemeManager.ToModels(chargesRateSchemes))
	})

	// GET	 /api/v1/charges-rate-scheme/currency/:currency_id: Get charges rate schemes by currency ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme/currency/:currency_id",
		Method:       "GET",
		Note:         "Returns a list of charges rate schemes for a specific currency.",
		ResponseType: modelCore.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		chargesRateSchemes, err := c.modelCore.ChargesRateSchemeManager.Find(context, &modelCore.ChargesRateScheme{
			CurrencyID:     *currencyID,
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch charges rate schemes by currency ID: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.ChargesRateSchemeManager.ToModels(chargesRateSchemes))
	})

	// GET /charges-rate-scheme/:charges_rate_scheme_id: Get specific charges rate scheme by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "GET",
		Note:         "Returns a single charges rate scheme by its ID.",
		ResponseType: modelCore.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		chargesRateScheme, err := c.modelCore.ChargesRateSchemeManager.GetByIDRaw(context, *chargesRateSchemeID)
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
		RequestType:  modelCore.ChargesRateSchemeRequest{},
		ResponseType: modelCore.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.ChargesRateSchemeManager.Validate(ctx)
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

		chargesRateScheme := &modelCore.ChargesRateScheme{
			MemberTypeID:  req.MemberTypeID,
			ModeOfPayment: req.ModeOfPayment,
			Name:          req.Name,
			Description:   req.Description,
			Icon:          req.Icon,
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
			CurrencyID:     req.CurrencyID,
			Type:           req.Type,
		}

		if err := c.modelCore.ChargesRateSchemeManager.Create(context, chargesRateScheme); err != nil {
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
				chargesRateSchemeAccount := &modelCore.ChargesRateSchemeAccount{
					ChargesRateSchemeID: chargesRateScheme.ID,
					AccountID:           accountID,
					CreatedAt:           time.Now().UTC(),
					CreatedByID:         user.UserID,
					UpdatedAt:           time.Now().UTC(),
					UpdatedByID:         user.UserID,
					BranchID:            *user.BranchID,
					OrganizationID:      user.OrganizationID,
				}
				if err := c.modelCore.ChargesRateSchemeAccountManager.Create(context, chargesRateSchemeAccount); err != nil {
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
		return ctx.JSON(http.StatusCreated, c.modelCore.ChargesRateSchemeManager.ToModel(chargesRateScheme))
	})

	// PUT /charges-rate-scheme/:charges_rate_scheme_id: Update charges rate scheme by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme by its ID.",
		RequestType:  modelCore.ChargesRateSchemeRequest{},
		ResponseType: modelCore.ChargesRateSchemeResponse{},
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

		req, err := c.modelCore.ChargesRateSchemeManager.Validate(ctx)
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
		chargesRateScheme, err := c.modelCore.ChargesRateSchemeManager.GetByID(context, *chargesRateSchemeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), charges rate scheme not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}

		// Start database transaction
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		chargesRateScheme.MemberTypeID = req.MemberTypeID
		chargesRateScheme.ModeOfPayment = req.ModeOfPayment
		chargesRateScheme.Name = req.Name
		chargesRateScheme.Description = req.Description
		chargesRateScheme.Icon = req.Icon
		chargesRateScheme.Type = req.Type
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
		chargesRateScheme.CurrencyID = req.CurrencyID

		if err := c.modelCore.ChargesRateSchemeManager.UpdateFieldsWithTx(context, tx, chargesRateScheme.ID, chargesRateScheme); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme: " + err.Error()})
		}

		// Handle deletions first
		if req.ChargesRateSchemeAccountsDeleted != nil {
			for _, id := range req.ChargesRateSchemeAccountsDeleted {
				if err := c.modelCore.ChargesRateSchemeAccountManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate scheme account: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme account: " + err.Error()})
				}
			}
		}

		if req.ChargesRateByRangeOrMinimumAmountsDeleted != nil {
			for _, id := range req.ChargesRateByRangeOrMinimumAmountsDeleted {
				if err := c.modelCore.ChargesRateByRangeOrMinimumAmountManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate by range or minimum amount: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by range or minimum amount: " + err.Error()})
				}
			}
		}

		if req.ChargesRateSchemeModeOfPaymentsDeleted != nil {
			for _, id := range req.ChargesRateSchemeModeOfPaymentsDeleted {
				if err := c.modelCore.ChargesRateSchemeModeOfPaymentManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate scheme mode of payment: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme mode of payment: " + err.Error()})
				}
			}
		}

		if req.ChargesRateByTermsDeleted != nil {
			for _, id := range req.ChargesRateByTermsDeleted {
				if err := c.modelCore.ChargesRateByTermManager.DeleteByIDWithTx(context, tx, id); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate by term: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by term: " + err.Error()})
				}
			}
		}

		// Handle ChargesRateSchemeAccounts creation/update
		if req.ChargesRateSchemeAccounts != nil {
			for _, accountReq := range req.ChargesRateSchemeAccounts {
				if accountReq.ID != nil {
					// Update existing record
					existingAccount, err := c.modelCore.ChargesRateSchemeAccountManager.GetByID(context, *accountReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate scheme account: " + err.Error()})
					}
					existingAccount.AccountID = accountReq.AccountID
					existingAccount.UpdatedAt = time.Now().UTC()
					existingAccount.UpdatedByID = user.UserID
					if err := c.modelCore.ChargesRateSchemeAccountManager.UpdateFieldsWithTx(context, tx, existingAccount.ID, existingAccount); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme account: " + err.Error()})
					}
				} else {
					// Create new record
					newAccount := &modelCore.ChargesRateSchemeAccount{
						ChargesRateSchemeID: chargesRateScheme.ID,
						AccountID:           accountReq.AccountID,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         user.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         user.UserID,
						BranchID:            *user.BranchID,
						OrganizationID:      user.OrganizationID,
					}
					if err := c.modelCore.ChargesRateSchemeAccountManager.CreateWithTx(context, tx, newAccount); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme account: " + err.Error()})
					}
				}
			}
		}

		// Handle ChargesRateByRangeOrMinimumAmounts creation/update
		if req.ChargesRateByRangeOrMinimumAmounts != nil {
			for _, rangeReq := range req.ChargesRateByRangeOrMinimumAmounts {
				if rangeReq.ID != nil {
					// Update existing record
					existingRange, err := c.modelCore.ChargesRateByRangeOrMinimumAmountManager.GetByID(context, *rangeReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate by range or minimum amount: " + err.Error()})
					}
					existingRange.From = rangeReq.From
					existingRange.To = rangeReq.To
					existingRange.Charge = rangeReq.Charge
					existingRange.Amount = rangeReq.Amount
					existingRange.MinimumAmount = rangeReq.MinimumAmount
					existingRange.UpdatedAt = time.Now().UTC()
					existingRange.UpdatedByID = user.UserID
					if err := c.modelCore.ChargesRateByRangeOrMinimumAmountManager.UpdateFieldsWithTx(context, tx, existingRange.ID, existingRange); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by range or minimum amount: " + err.Error()})
					}
				} else {
					// Create new record
					newRange := &modelCore.ChargesRateByRangeOrMinimumAmount{
						ChargesRateSchemeID: chargesRateScheme.ID,
						From:                rangeReq.From,
						To:                  rangeReq.To,
						Charge:              rangeReq.Charge,
						Amount:              rangeReq.Amount,
						MinimumAmount:       rangeReq.MinimumAmount,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         user.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         user.UserID,
						BranchID:            *user.BranchID,
						OrganizationID:      user.OrganizationID,
					}
					if err := c.modelCore.ChargesRateByRangeOrMinimumAmountManager.CreateWithTx(context, tx, newRange); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by range or minimum amount: " + err.Error()})
					}
				}
			}
		}

		// Handle ChargesRateSchemeModeOfPayments creation/update
		if req.ChargesRateSchemeModeOfPayments != nil {
			for _, modeReq := range req.ChargesRateSchemeModeOfPayments {
				if modeReq.ID != nil {
					// Update existing record
					existingMode, err := c.modelCore.ChargesRateSchemeModeOfPaymentManager.GetByID(context, *modeReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate scheme mode of payment: " + err.Error()})
					}
					existingMode.From = modeReq.From
					existingMode.To = modeReq.To
					existingMode.Column1 = modeReq.Column1
					existingMode.Column2 = modeReq.Column2
					existingMode.Column3 = modeReq.Column3
					existingMode.Column4 = modeReq.Column4
					existingMode.Column5 = modeReq.Column5
					existingMode.Column6 = modeReq.Column6
					existingMode.Column7 = modeReq.Column7
					existingMode.Column8 = modeReq.Column8
					existingMode.Column9 = modeReq.Column9
					existingMode.Column10 = modeReq.Column10
					existingMode.Column11 = modeReq.Column11
					existingMode.Column12 = modeReq.Column12
					existingMode.Column13 = modeReq.Column13
					existingMode.Column14 = modeReq.Column14
					existingMode.Column15 = modeReq.Column15
					existingMode.Column16 = modeReq.Column16
					existingMode.Column17 = modeReq.Column17
					existingMode.Column18 = modeReq.Column18
					existingMode.Column19 = modeReq.Column19
					existingMode.Column20 = modeReq.Column20
					existingMode.Column21 = modeReq.Column21
					existingMode.Column22 = modeReq.Column22
					existingMode.UpdatedAt = time.Now().UTC()
					existingMode.UpdatedByID = user.UserID
					if err := c.modelCore.ChargesRateSchemeModeOfPaymentManager.UpdateFieldsWithTx(context, tx, existingMode.ID, existingMode); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme mode of payment: " + err.Error()})
					}
				} else {
					// Create new record
					newMode := &modelCore.ChargesRateSchemeModeOfPayment{
						ChargesRateSchemeID: chargesRateScheme.ID,
						From:                modeReq.From,
						To:                  modeReq.To,
						Column1:             modeReq.Column1,
						Column2:             modeReq.Column2,
						Column3:             modeReq.Column3,
						Column4:             modeReq.Column4,
						Column5:             modeReq.Column5,
						Column6:             modeReq.Column6,
						Column7:             modeReq.Column7,
						Column8:             modeReq.Column8,
						Column9:             modeReq.Column9,
						Column10:            modeReq.Column10,
						Column11:            modeReq.Column11,
						Column12:            modeReq.Column12,
						Column13:            modeReq.Column13,
						Column14:            modeReq.Column14,
						Column15:            modeReq.Column15,
						Column16:            modeReq.Column16,
						Column17:            modeReq.Column17,
						Column18:            modeReq.Column18,
						Column19:            modeReq.Column19,
						Column20:            modeReq.Column20,
						Column21:            modeReq.Column21,
						Column22:            modeReq.Column22,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         user.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         user.UserID,
						BranchID:            *user.BranchID,
						OrganizationID:      user.OrganizationID,
					}
					if err := c.modelCore.ChargesRateSchemeModeOfPaymentManager.CreateWithTx(context, tx, newMode); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme mode of payment: " + err.Error()})
					}
				}
			}
		}

		// Handle ChargesRateByTerms creation/update
		if req.ChargesRateByTerms != nil {
			for _, termReq := range req.ChargesRateByTerms {
				if termReq.ID != nil {
					// Update existing record
					existingTerm, err := c.modelCore.ChargesRateByTermManager.GetByID(context, *termReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate by term: " + err.Error()})
					}
					existingTerm.Name = termReq.Name
					existingTerm.Description = termReq.Description
					existingTerm.ModeOfPayment = termReq.ModeOfPayment
					existingTerm.Rate1 = termReq.Rate1
					existingTerm.Rate2 = termReq.Rate2
					existingTerm.Rate3 = termReq.Rate3
					existingTerm.Rate4 = termReq.Rate4
					existingTerm.Rate5 = termReq.Rate5
					existingTerm.Rate6 = termReq.Rate6
					existingTerm.Rate7 = termReq.Rate7
					existingTerm.Rate8 = termReq.Rate8
					existingTerm.Rate9 = termReq.Rate9
					existingTerm.Rate10 = termReq.Rate10
					existingTerm.Rate11 = termReq.Rate11
					existingTerm.Rate12 = termReq.Rate12
					existingTerm.Rate13 = termReq.Rate13
					existingTerm.Rate14 = termReq.Rate14
					existingTerm.Rate15 = termReq.Rate15
					existingTerm.Rate16 = termReq.Rate16
					existingTerm.Rate17 = termReq.Rate17
					existingTerm.Rate18 = termReq.Rate18
					existingTerm.Rate19 = termReq.Rate19
					existingTerm.Rate20 = termReq.Rate20
					existingTerm.Rate21 = termReq.Rate21
					existingTerm.Rate22 = termReq.Rate22
					existingTerm.UpdatedAt = time.Now().UTC()
					existingTerm.UpdatedByID = user.UserID
					if err := c.modelCore.ChargesRateByTermManager.UpdateFieldsWithTx(context, tx, existingTerm.ID, existingTerm); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by term: " + err.Error()})
					}
				} else {
					// Create new record
					newTerm := &modelCore.ChargesRateByTerm{
						ChargesRateSchemeID: chargesRateScheme.ID,
						Name:                termReq.Name,
						Description:         termReq.Description,
						ModeOfPayment:       termReq.ModeOfPayment,
						Rate1:               termReq.Rate1,
						Rate2:               termReq.Rate2,
						Rate3:               termReq.Rate3,
						Rate4:               termReq.Rate4,
						Rate5:               termReq.Rate5,
						Rate6:               termReq.Rate6,
						Rate7:               termReq.Rate7,
						Rate8:               termReq.Rate8,
						Rate9:               termReq.Rate9,
						Rate10:              termReq.Rate10,
						Rate11:              termReq.Rate11,
						Rate12:              termReq.Rate12,
						Rate13:              termReq.Rate13,
						Rate14:              termReq.Rate14,
						Rate15:              termReq.Rate15,
						Rate16:              termReq.Rate16,
						Rate17:              termReq.Rate17,
						Rate18:              termReq.Rate18,
						Rate19:              termReq.Rate19,
						Rate20:              termReq.Rate20,
						Rate21:              termReq.Rate21,
						Rate22:              termReq.Rate22,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         user.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         user.UserID,
						BranchID:            *user.BranchID,
						OrganizationID:      user.OrganizationID,
					}
					if err := c.modelCore.ChargesRateByTermManager.CreateWithTx(context, tx, newTerm); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by term: " + err.Error()})
					}
				}
			}
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to commit charges rate scheme update transaction: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit charges rate scheme update: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme (/charges-rate-scheme/:charges_rate_scheme_id): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})

		newRateScheme, err := c.modelCore.ChargesRateSchemeManager.GetByIDRaw(context, chargesRateScheme.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated charges rate scheme: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newRateScheme)
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
		chargesRateScheme, err := c.modelCore.ChargesRateSchemeManager.GetByID(context, *chargesRateSchemeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}
		if err := c.modelCore.ChargesRateSchemeManager.DeleteByID(context, *chargesRateSchemeID); err != nil {
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
		RequestType: modelCore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelCore.IDSRequest
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
			chargesRateScheme, err := c.modelCore.ChargesRateSchemeManager.GetByID(context, chargesRateSchemeID)
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
			if err := c.modelCore.ChargesRateSchemeManager.DeleteByIDWithTx(context, tx, chargesRateSchemeID); err != nil {
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
