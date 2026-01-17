package charges

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func ChargesRateSchemeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme",
		Method:       "GET",
		Note:         "Returns a paginated list of charges rate schemes for the current user's organization and branch.",
		ResponseType: types.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		chargesRateSchemes, err := core.ChargesRateSchemeCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch charges rate schemes for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.ChargesRateSchemeManager(service).ToModels(chargesRateSchemes))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme/currency/:currency_id",
		Method:       "GET",
		Note:         "Returns a list of charges rate schemes for a specific currency.",
		ResponseType: types.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		chargesRateSchemes, err := core.ChargesRateSchemeManager(service).Find(context, &types.ChargesRateScheme{
			CurrencyID:     *currencyID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch charges rate schemes by currency ID: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.ChargesRateSchemeManager(service).ToModels(chargesRateSchemes))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "GET",
		Note:         "Returns a single charges rate scheme by its ID.",
		ResponseType: types.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		chargesRateScheme, err := core.ChargesRateSchemeManager(service).GetByIDRaw(context, *chargesRateSchemeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}
		return ctx.JSON(http.StatusOK, chargesRateScheme)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme",
		Method:       "POST",
		Note:         "Creates a new charges rate scheme for the current user's organization and branch.",
		RequestType:  types.ChargesRateSchemeRequest{},
		ResponseType: types.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.ChargesRateSchemeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), validation error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), user org error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), user not assigned to branch.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateScheme := &types.ChargesRateScheme{
			MemberTypeID:          req.MemberTypeID,
			ModeOfPayment:         req.ModeOfPayment,
			Name:                  req.Name,
			Description:           req.Description,
			Icon:                  req.Icon,
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
			ByTermHeader1:         req.ByTermHeader1,
			ByTermHeader2:         req.ByTermHeader2,
			ByTermHeader3:         req.ByTermHeader3,
			ByTermHeader4:         req.ByTermHeader4,
			ByTermHeader5:         req.ByTermHeader5,
			ByTermHeader6:         req.ByTermHeader6,
			ByTermHeader7:         req.ByTermHeader7,
			ByTermHeader8:         req.ByTermHeader8,
			ByTermHeader9:         req.ByTermHeader9,
			ByTermHeader10:        req.ByTermHeader10,
			ByTermHeader11:        req.ByTermHeader11,
			ByTermHeader12:        req.ByTermHeader12,
			ByTermHeader13:        req.ByTermHeader13,
			ByTermHeader14:        req.ByTermHeader14,
			ByTermHeader15:        req.ByTermHeader15,
			ByTermHeader16:        req.ByTermHeader16,
			ByTermHeader17:        req.ByTermHeader17,
			ByTermHeader18:        req.ByTermHeader18,
			ByTermHeader19:        req.ByTermHeader19,
			ByTermHeader20:        req.ByTermHeader20,
			ByTermHeader21:        req.ByTermHeader21,
			ByTermHeader22:        req.ByTermHeader22,
			CreatedAt:             time.Now().UTC(),
			CreatedByID:           userOrg.UserID,
			UpdatedAt:             time.Now().UTC(),
			UpdatedByID:           userOrg.UserID,
			BranchID:              *userOrg.BranchID,
			OrganizationID:        userOrg.OrganizationID,
			CurrencyID:            req.CurrencyID,
			Type:                  req.Type,
		}

		if err := core.ChargesRateSchemeManager(service).Create(context, chargesRateScheme); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme creation failed (/charges-rate-scheme), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme: " + err.Error()})
		}

		if len(req.AccountIDs) > 0 {
			for _, accountID := range req.AccountIDs {
				chargesRateSchemeAccount := &types.ChargesRateSchemeAccount{
					ChargesRateSchemeID: chargesRateScheme.ID,
					AccountID:           accountID,
					CreatedAt:           time.Now().UTC(),
					CreatedByID:         userOrg.UserID,
					UpdatedAt:           time.Now().UTC(),
					UpdatedByID:         userOrg.UserID,
					BranchID:            *userOrg.BranchID,
					OrganizationID:      userOrg.OrganizationID,
				}
				if err := core.ChargesRateSchemeAccountManager(service).Create(context, chargesRateSchemeAccount); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Charges rate scheme account creation failed (/charges-rate-scheme), db error: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme account: " + err.Error()})
				}
			}
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate scheme (/charges-rate-scheme): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})
		return ctx.JSON(http.StatusCreated, core.ChargesRateSchemeManager(service).ToModel(chargesRateScheme))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme by its ID.",
		RequestType:  types.ChargesRateSchemeRequest{},
		ResponseType: types.ChargesRateSchemeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), invalid charges rate scheme ID.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}

		req, err := core.ChargesRateSchemeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), validation error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), user org error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateScheme, err := core.ChargesRateSchemeManager(service).GetByID(context, *chargesRateSchemeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), charges rate scheme not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}

		tx, endTx := service.Database.StartTransaction(context)

		chargesRateScheme.MemberTypeID = req.MemberTypeID
		chargesRateScheme.ModeOfPayment = req.ModeOfPayment
		chargesRateScheme.Name = req.Name
		chargesRateScheme.Description = req.Description
		chargesRateScheme.Icon = req.Icon
		chargesRateScheme.Type = req.Type
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
		chargesRateScheme.UpdatedByID = userOrg.UserID
		chargesRateScheme.CurrencyID = req.CurrencyID

		if err := core.ChargesRateSchemeManager(service).UpdateByIDWithTx(context, tx, chargesRateScheme.ID, chargesRateScheme); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme update failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme: " + endTx(err).Error()})
		}

		if req.ChargesRateSchemeAccountsDeleted != nil {
			for _, id := range req.ChargesRateSchemeAccountsDeleted {
				if err := core.ChargesRateSchemeAccountManager(service).DeleteWithTx(context, tx, id); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate scheme account: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme account: " + endTx(err).Error()})
				}
			}
		}

		if req.ChargesRateByRangeOrMinimumAmountsDeleted != nil {
			for _, id := range req.ChargesRateByRangeOrMinimumAmountsDeleted {
				if err := core.ChargesRateByRangeOrMinimumAmountManager(service).DeleteWithTx(context, tx, id); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate by range or minimum amount: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by range or minimum amount: " + endTx(err).Error()})
				}
			}
		}

		if req.ChargesRateSchemeModeOfPaymentsDeleted != nil {
			for _, id := range req.ChargesRateSchemeModeOfPaymentsDeleted {
				if err := core.ChargesRateSchemeModeOfPaymentManager(service).DeleteWithTx(context, tx, id); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate scheme mode of payment: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme mode of payment: " + endTx(err).Error()})
				}
			}
		}

		if req.ChargesRateByTermsDeleted != nil {
			for _, id := range req.ChargesRateByTermsDeleted {
				if err := core.ChargesRateByTermManager(service).DeleteWithTx(context, tx, id); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Failed to delete charges rate by term: " + err.Error(),
						Module:      "ChargesRateScheme",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by term: " + endTx(err).Error()})
				}
			}
		}

		if req.ChargesRateSchemeAccounts != nil {
			for _, accountReq := range req.ChargesRateSchemeAccounts {
				if accountReq.ID != nil {
					existingAccount, err := core.ChargesRateSchemeAccountManager(service).GetByID(context, *accountReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate scheme account: " + endTx(err).Error()})
					}
					existingAccount.AccountID = accountReq.AccountID
					existingAccount.UpdatedAt = time.Now().UTC()
					existingAccount.UpdatedByID = userOrg.UserID
					if err := core.ChargesRateSchemeAccountManager(service).UpdateByIDWithTx(context, tx, existingAccount.ID, existingAccount); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme account: " + endTx(err).Error()})
					}
				} else {
					newAccount := &types.ChargesRateSchemeAccount{
						ChargesRateSchemeID: chargesRateScheme.ID,
						AccountID:           accountReq.AccountID,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         userOrg.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         userOrg.UserID,
						BranchID:            *userOrg.BranchID,
						OrganizationID:      userOrg.OrganizationID,
					}
					if err := core.ChargesRateSchemeAccountManager(service).CreateWithTx(context, tx, newAccount); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme account: " + endTx(err).Error()})
					}
				}
			}
		}

		if req.ChargesRateByRangeOrMinimumAmounts != nil {
			for _, rangeReq := range req.ChargesRateByRangeOrMinimumAmounts {
				if rangeReq.ID != nil {
					existingRange, err := core.ChargesRateByRangeOrMinimumAmountManager(service).GetByID(context, *rangeReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate by range or minimum amount: " + endTx(err).Error()})
					}
					existingRange.From = rangeReq.From
					existingRange.To = rangeReq.To
					existingRange.Charge = rangeReq.Charge
					existingRange.Amount = rangeReq.Amount
					existingRange.MinimumAmount = rangeReq.MinimumAmount
					existingRange.UpdatedAt = time.Now().UTC()
					existingRange.UpdatedByID = userOrg.UserID
					if err := core.ChargesRateByRangeOrMinimumAmountManager(service).UpdateByIDWithTx(context, tx, existingRange.ID, existingRange); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by range or minimum amount: " + endTx(err).Error()})
					}
				} else {
					newRange := &types.ChargesRateByRangeOrMinimumAmount{
						ChargesRateSchemeID: chargesRateScheme.ID,
						From:                rangeReq.From,
						To:                  rangeReq.To,
						Charge:              rangeReq.Charge,
						Amount:              rangeReq.Amount,
						MinimumAmount:       rangeReq.MinimumAmount,
						CreatedAt:           time.Now().UTC(),
						CreatedByID:         userOrg.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         userOrg.UserID,
						BranchID:            *userOrg.BranchID,
						OrganizationID:      userOrg.OrganizationID,
					}
					if err := core.ChargesRateByRangeOrMinimumAmountManager(service).CreateWithTx(context, tx, newRange); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by range or minimum amount: " + endTx(err).Error()})
					}
				}
			}
		}

		if req.ChargesRateSchemeModeOfPayments != nil {
			for _, modeReq := range req.ChargesRateSchemeModeOfPayments {
				if modeReq.ID != nil {
					existingMode, err := core.ChargesRateSchemeModeOfPaymentManager(service).GetByID(context, *modeReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate scheme mode of payment: " + endTx(err).Error()})
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
					existingMode.UpdatedByID = userOrg.UserID
					if err := core.ChargesRateSchemeModeOfPaymentManager(service).UpdateByIDWithTx(context, tx, existingMode.ID, existingMode); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme mode of payment: " + endTx(err).Error()})
					}
				} else {
					newMode := &types.ChargesRateSchemeModeOfPayment{
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
						CreatedByID:         userOrg.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         userOrg.UserID,
						BranchID:            *userOrg.BranchID,
						OrganizationID:      userOrg.OrganizationID,
					}
					if err := core.ChargesRateSchemeModeOfPaymentManager(service).CreateWithTx(context, tx, newMode); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme mode of payment: " + endTx(err).Error()})
					}
				}
			}
		}

		if req.ChargesRateByTerms != nil {
			for _, termReq := range req.ChargesRateByTerms {
				if termReq.ID != nil {
					existingTerm, err := core.ChargesRateByTermManager(service).GetByID(context, *termReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get charges rate by term: " + endTx(err).Error()})
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
					existingTerm.UpdatedByID = userOrg.UserID
					if err := core.ChargesRateByTermManager(service).UpdateByIDWithTx(context, tx, existingTerm.ID, existingTerm); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by term: " + endTx(err).Error()})
					}
				} else {
					newTerm := &types.ChargesRateByTerm{
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
						CreatedByID:         userOrg.UserID,
						UpdatedAt:           time.Now().UTC(),
						UpdatedByID:         userOrg.UserID,
						BranchID:            *userOrg.BranchID,
						OrganizationID:      userOrg.OrganizationID,
					}
					if err := core.ChargesRateByTermManager(service).CreateWithTx(context, tx, newTerm); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by term: " + endTx(err).Error()})
					}
				}
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to commit charges rate scheme update transaction: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit charges rate scheme update: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme (/charges-rate-scheme/:charges_rate_scheme_id): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})

		newRateScheme, err := core.ChargesRateSchemeManager(service).GetByIDRaw(context, chargesRateScheme.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated charges rate scheme: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newRateScheme)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/charges-rate-scheme/:charges_rate_scheme_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate scheme by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), invalid charges rate scheme ID.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		chargesRateScheme, err := core.ChargesRateSchemeManager(service).GetByID(context, *chargesRateSchemeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), not found.",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme not found"})
		}
		if err := core.ChargesRateSchemeManager(service).Delete(context, *chargesRateSchemeID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme delete failed (/charges-rate-scheme/:charges_rate_scheme_id), db error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate scheme (/charges-rate-scheme/:charges_rate_scheme_id): " + chargesRateScheme.Name,
			Module:      "ChargesRateScheme",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/charges-rate-scheme/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate schemes by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete) | no IDs provided",
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate scheme IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.ChargesRateSchemeManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme/bulk-delete) | error: " + err.Error(),
				Module:      "ChargesRateScheme",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete charges rate schemes: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate schemes (/charges-rate-scheme/bulk-delete)",
			Module:      "ChargesRateScheme",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
