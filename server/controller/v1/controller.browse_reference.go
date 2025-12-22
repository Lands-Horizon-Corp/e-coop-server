package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (c *Controller) browseReferenceController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference/:browse_reference_id",
		Method:       "PUT",
		ResponseType: core.BrowseReferenceResponse{},
		RequestType:  core.BrowseReferenceRequest{},
		Note:         "Updates an existing browse reference with nested interest rates.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		browseReferenceID, err := handlers.EngineUUIDParam(ctx, "browse_reference_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse reference ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update browse references"})
		}

		request, err := c.core.BrowseReferenceManager().Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		browseReference, err := c.core.BrowseReferenceManager().GetByID(context, *browseReferenceID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse reference not found"})
		}

		if browseReference.OrganizationID != userOrg.OrganizationID || browseReference.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this browse reference"})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		browseReference.Name = request.Name
		browseReference.Description = request.Description
		browseReference.InterestRate = request.InterestRate
		browseReference.MinimumBalance = request.MinimumBalance
		browseReference.Charges = request.Charges
		browseReference.AccountID = request.AccountID
		browseReference.MemberTypeID = request.MemberTypeID
		browseReference.InterestType = request.InterestType
		browseReference.DefaultMinimumBalance = request.DefaultMinimumBalance
		browseReference.DefaultInterestRate = request.DefaultInterestRate
		browseReference.UpdatedAt = time.Now().UTC()
		browseReference.UpdatedByID = userOrg.UserID

		if request.InterestRatesByYearDeleted != nil {
			for _, deletedID := range request.InterestRatesByYearDeleted {
				interestRateByYear, err := c.core.InterestRateByYearManager().GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find interest rate by year for deletion: " + endTx(err).Error()})
				}
				if interestRateByYear.BrowseReferenceID != browseReference.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete interest rate by year that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
				}
				interestRateByYear.DeletedByID = &userOrg.UserID
				if err := c.core.InterestRateByYearManager().DeleteWithTx(context, tx, interestRateByYear.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete interest rate by year: " + endTx(err).Error()})
				}
			}
		}

		if request.InterestRatesByDateDeleted != nil {
			for _, deletedID := range request.InterestRatesByDateDeleted {
				interestRateByDate, err := c.core.InterestRateByDateManager().GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find interest rate by date for deletion: " + endTx(err).Error()})
				}
				if interestRateByDate.BrowseReferenceID != browseReference.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete interest rate by date that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
				}
				interestRateByDate.DeletedByID = &userOrg.UserID
				if err := c.core.InterestRateByDateManager().DeleteWithTx(context, tx, interestRateByDate.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete interest rate by date: " + endTx(err).Error()})
				}
			}
		}

		if request.InterestRatesByAmountDeleted != nil {
			for _, deletedID := range request.InterestRatesByAmountDeleted {
				interestRateByAmount, err := c.core.InterestRateByAmountManager().GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find interest rate by amount for deletion: " + endTx(err).Error()})
				}
				if interestRateByAmount.BrowseReferenceID != browseReference.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete interest rate by amount that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
				}
				interestRateByAmount.DeletedByID = &userOrg.UserID
				if err := c.core.InterestRateByAmountManager().DeleteWithTx(context, tx, interestRateByAmount.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete interest rate by amount: " + endTx(err).Error()})
				}
			}
		}

		if request.InterestRatesByYear != nil {
			for _, rateReq := range request.InterestRatesByYear {
				if rateReq.ID != nil {
					existingRecord, err := c.core.InterestRateByYearManager().GetByID(context, *rateReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing interest rate by year: " + endTx(err).Error()})
					}
					if existingRecord.BrowseReferenceID != browseReference.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update interest rate by year that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.FromYear = rateReq.FromYear
					existingRecord.ToYear = rateReq.ToYear
					existingRecord.InterestRate = rateReq.InterestRate

					if err := c.core.InterestRateByYearManager().UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update interest rate by year: " + endTx(err).Error()})
					}
				} else {
					rateByYear := &core.InterestRateByYear{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						BrowseReferenceID: browseReference.ID,
						FromYear:          rateReq.FromYear,
						ToYear:            rateReq.ToYear,
						InterestRate:      rateReq.InterestRate,
					}

					if err := c.core.InterestRateByYearManager().CreateWithTx(context, tx, rateByYear); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by year: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.InterestRatesByDate != nil {
			for _, rateReq := range request.InterestRatesByDate {
				if rateReq.ID != nil {
					existingRecord, err := c.core.InterestRateByDateManager().GetByID(context, *rateReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing interest rate by date: " + endTx(err).Error()})
					}
					if existingRecord.BrowseReferenceID != browseReference.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update interest rate by date that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.FromDate = rateReq.FromDate
					existingRecord.ToDate = rateReq.ToDate
					existingRecord.InterestRate = rateReq.InterestRate

					if err := c.core.InterestRateByDateManager().UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update interest rate by date: " + endTx(err).Error()})
					}
				} else {
					rateByDate := &core.InterestRateByDate{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						BrowseReferenceID: browseReference.ID,
						FromDate:          rateReq.FromDate,
						ToDate:            rateReq.ToDate,
						InterestRate:      rateReq.InterestRate,
					}

					if err := c.core.InterestRateByDateManager().CreateWithTx(context, tx, rateByDate); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by date: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.InterestRatesByAmount != nil {
			for _, rateReq := range request.InterestRatesByAmount {
				if rateReq.ID != nil {
					existingRecord, err := c.core.InterestRateByAmountManager().GetByID(context, *rateReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing interest rate by amount: " + endTx(err).Error()})
					}
					if existingRecord.BrowseReferenceID != browseReference.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update interest rate by amount that doesn't belong to this browse reference: " + endTx(eris.New("invalid browse reference")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.FromAmount = rateReq.FromAmount
					existingRecord.ToAmount = rateReq.ToAmount
					existingRecord.InterestRate = rateReq.InterestRate

					if err := c.core.InterestRateByAmountManager().UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update interest rate by amount: " + endTx(err).Error()})
					}
				} else {
					rateByAmount := &core.InterestRateByAmount{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						BrowseReferenceID: browseReference.ID,
						FromAmount:        rateReq.FromAmount,
						ToAmount:          rateReq.ToAmount,
						InterestRate:      rateReq.InterestRate,
					}

					if err := c.core.InterestRateByAmountManager().CreateWithTx(context, tx, rateByAmount); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by amount: " + endTx(err).Error()})
					}
				}
			}
		}

		if err := c.core.BrowseReferenceManager().UpdateByIDWithTx(context, tx, browseReference.ID, browseReference); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update browse reference: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/browse-reference/:browse_reference_id): " + err.Error(),
				Module:      "BrowseReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		updatedBrowseReference, err := c.core.BrowseReferenceManager().GetByID(context, browseReference.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated browse reference: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Browse reference updated successfully",
			Module:      "BrowseReference",
		})

		return ctx.JSON(http.StatusOK, c.core.BrowseReferenceManager().ToModel(updatedBrowseReference))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference/:browse_reference_id",
		Method:       "GET",
		ResponseType: core.BrowseReferenceResponse{},
		Note:         "Retrieves a specific browse reference by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		browseReferenceID, err := handlers.EngineUUIDParam(ctx, "browse_reference_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse reference ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		browseReference, err := c.core.BrowseReferenceManager().GetByID(context, *browseReferenceID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse reference not found"})
		}

		if browseReference.OrganizationID != userOrg.OrganizationID || browseReference.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this browse reference"})
		}

		return ctx.JSON(http.StatusOK, c.core.BrowseReferenceManager().ToModel(browseReference))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference",
		Method:       "GET",
		ResponseType: core.BrowseReferenceResponse{},
		Note:         "Retrieves all browse references for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		browseReferences, err := c.core.BrowseReferenceCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve browse references: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.BrowseReferenceManager().ToModels(browseReferences))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference",
		Method:       "POST",
		ResponseType: core.BrowseReferenceResponse{},
		RequestType:  core.BrowseReferenceRequest{},
		Note:         "Creates a new browse reference with nested interest rates.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		request, err := c.core.BrowseReferenceManager().Validate(ctx)
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

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		browseReference := &core.BrowseReference{
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
			CreatedByID:           userOrg.UserID,
			UpdatedByID:           userOrg.UserID,
			OrganizationID:        userOrg.OrganizationID,
			BranchID:              *userOrg.BranchID,
			Name:                  request.Name,
			Description:           request.Description,
			InterestRate:          request.InterestRate,
			MinimumBalance:        request.MinimumBalance,
			Charges:               request.Charges,
			AccountID:             request.AccountID,
			MemberTypeID:          request.MemberTypeID,
			InterestType:          request.InterestType,
			DefaultMinimumBalance: request.DefaultMinimumBalance,
			DefaultInterestRate:   request.DefaultInterestRate,
		}

		if err := c.core.BrowseReferenceManager().CreateWithTx(context, tx, browseReference); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create browse reference: " + endTx(err).Error()})
		}

		if request.InterestRatesByYear != nil {
			for _, rateReq := range request.InterestRatesByYear {
				rateByYear := &core.InterestRateByYear{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					BrowseReferenceID: browseReference.ID,
					FromYear:          rateReq.FromYear,
					ToYear:            rateReq.ToYear,
					InterestRate:      rateReq.InterestRate,
				}

				if err := c.core.InterestRateByYearManager().CreateWithTx(context, tx, rateByYear); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by year: " + endTx(err).Error()})
				}
			}
		}

		if request.InterestRatesByDate != nil {
			for _, rateReq := range request.InterestRatesByDate {
				rateByDate := &core.InterestRateByDate{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					BrowseReferenceID: browseReference.ID,
					FromDate:          rateReq.FromDate,
					ToDate:            rateReq.ToDate,
					InterestRate:      rateReq.InterestRate,
				}

				if err := c.core.InterestRateByDateManager().CreateWithTx(context, tx, rateByDate); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by date: " + endTx(err).Error()})
				}
			}
		}

		if request.InterestRatesByAmount != nil {
			for _, rateReq := range request.InterestRatesByAmount {
				rateByAmount := &core.InterestRateByAmount{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					BrowseReferenceID: browseReference.ID,
					FromAmount:        rateReq.FromAmount,
					ToAmount:          rateReq.ToAmount,
					InterestRate:      rateReq.InterestRate,
				}

				if err := c.core.InterestRateByAmountManager().CreateWithTx(context, tx, rateByAmount); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create interest rate by amount: " + endTx(err).Error()})
				}
			}
		}

		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/browse-reference): " + err.Error(),
				Module:      "BrowseReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		createdBrowseReference, err := c.core.BrowseReferenceManager().GetByID(context, browseReference.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created browse reference: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Browse reference created successfully",
			Module:      "BrowseReference",
		})

		return ctx.JSON(http.StatusCreated, c.core.BrowseReferenceManager().ToModel(createdBrowseReference))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/browse-reference/:browse_reference_id",
		Method: "DELETE",
		Note:   "Deletes a browse reference and all related interest rates.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		browseReferenceID, err := handlers.EngineUUIDParam(ctx, "browse_reference_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid browse reference ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete browse references"})
		}

		browseReference, err := c.core.BrowseReferenceManager().GetByID(context, *browseReferenceID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Browse reference not found"})
		}

		if browseReference.OrganizationID != userOrg.OrganizationID || browseReference.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this browse reference"})
		}

		browseReference.DeletedByID = &userOrg.UserID
		if err := c.core.BrowseReferenceManager().Delete(context, browseReference.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete browse reference: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Browse reference deleted successfully",
			Module:      "BrowseReference",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Browse reference deleted successfully"})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference/by-member-type/:member_type_id",
		Method:       "GET",
		ResponseType: core.BrowseReferenceResponse{},
		Note:         "Retrieves browse references for a specific member type.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := handlers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member type ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		browseReferences, err := c.core.BrowseReferenceByMemberType(context, *memberTypeID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve browse references: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.BrowseReferenceManager().ToModels(browseReferences))
	})

	// GET /api/v1/browse-reference/account/:account_id/member-type/:member_type_id
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/browse-reference/account/:account_id/member-type/:member_type_id",
		Method:       "GET",
		ResponseType: core.BrowseReferenceResponse{},
		Note:         "Retrieves browse references by account and member type for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid account ID",
			})
		}
		memberTypeID, err := handlers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid member type ID",
			})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "User authentication failed or organization not found",
			})
		}
		browseReference, err := c.core.BrowseReferenceManager().FindOne(
			context,
			&core.BrowseReference{
				AccountID:      accountID,
				MemberTypeID:   memberTypeID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			},
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve browse references: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, c.core.BrowseReferenceManager().ToModel(browseReference))
	})

}
