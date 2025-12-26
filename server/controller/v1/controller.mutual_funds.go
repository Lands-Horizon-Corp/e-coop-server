package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) mutualFundsController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund",
		Method:       "GET",
		Note:         "Returns all mutual funds for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := c.core.MutualFundCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual funds found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundManager().ToModels(mutualFunds))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund/search",
		Method:       "GET",
		Note:         "Returns a paginated list of mutual funds for the current user's organization and branch.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := c.core.MutualFundManager().NormalPagination(context, ctx, &core.MutualFund{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch mutual funds for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, mutualFunds)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund/member/:member_id",
		Method:       "GET",
		Note:         "Returns all mutual funds for a specific member profile.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberID, err := handlers.EngineUUIDParam(ctx, "member_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := c.core.MutualFundByMember(context, *memberID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual funds found for the specified member"})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundManager().ToModels(mutualFunds))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund/:mutual_fund_id",
		Method:       "GET",
		Note:         "Returns a single mutual fund by its ID.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByIDRaw(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		return ctx.JSON(http.StatusOK, mutualFund)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund",
		Method:       "POST",
		Note:         "Creates a new mutual fund for the current user's organization and branch.",
		RequestType:  core.MutualFundRequest{},
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MutualFundManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), user org error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), user not assigned to branch.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		mutualFund, err := c.core.CreateMutualFundValue(context, req, userOrg)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if err := c.core.MutualFundManager().CreateWithTx(context, tx, mutualFund); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), db error: " + err.Error(),
				Module:      "MutualFund",
			})

			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund: " + endTx(err).Error()})
		}
		for _, additionalMember := range mutualFund.AdditionalMembers {
			if err := c.core.MutualFundAdditionalMembersManager().CreateWithTx(context, tx, additionalMember); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund additional member creation failed: " + err.Error(),
					Module:      "MutualFund",
				})

				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create additional member: " + endTx(err).Error()})
			}
		}
		for _, mutualFundTable := range mutualFund.MutualFundTables {
			if err := c.core.MutualFundTableManager().CreateWithTx(context, tx, mutualFundTable); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund table creation failed: " + err.Error(),
					Module:      "MutualFund",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund table: " + endTx(err).Error()})
			}
		}
		mutualFundView, err := c.event.GenerateMutualFundEntries(context, userOrg, mutualFund)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund view: " + endTx(err).Error()})
		}
		for _, entry := range mutualFundView {
			if err := c.core.MutualFundEntryManager().CreateWithTx(context, tx, entry); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund entry creation failed: " + endTx(err).Error(),
					Module:      "MutualFund",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund entry: " + err.Error()})
			}
		}

		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + endTx(err).Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created mutual fund (/mutual-fund): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.JSON(http.StatusCreated, c.core.MutualFundManager().ToModel(mutualFund))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund/:mutual_fund_id",
		Method:       "PUT",
		Note:         "Updates an existing mutual fund by its ID.",
		RequestType:  core.MutualFundRequest{},
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), invalid mutual fund ID.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}

		req, err := c.core.MutualFundManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), user org error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), mutual fund not found.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}

		mutualFund.MemberProfileID = req.MemberProfileID
		mutualFund.Name = req.Name
		mutualFund.Description = req.Description
		mutualFund.DateOfDeath = req.DateOfDeath
		mutualFund.ExtensionOnly = req.ExtensionOnly
		mutualFund.Amount = req.Amount
		mutualFund.ComputationType = req.ComputationType
		mutualFund.UpdatedAt = time.Now().UTC()
		mutualFund.UpdatedByID = userOrg.UserID
		mutualFund.MemberTypeID = req.MemberTypeID

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		if err := c.core.MutualFundManager().UpdateByIDWithTx(context, tx, mutualFund.ID, mutualFund); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), db error: " + err.Error(),
				Module:      "MutualFund",
			})
			endTx(err)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund: " + err.Error()})
		}

		mfIds := make([]any, len(req.MutualFundAdditionalMembersDeleteIDs))
		for i, id := range req.MutualFundAdditionalMembersDeleteIDs {
			mfIds[i] = id
		}
		if len(req.MutualFundAdditionalMembersDeleteIDs) > 0 {
			if err := c.core.MutualFundAdditionalMembersManager().BulkDeleteWithTx(context, tx, mfIds); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to delete additional members: " + err.Error(),
					Module:      "MutualFund",
				})
				endTx(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete additional members: " + err.Error()})
			}
		}

		mftIds := make([]any, len(req.MutualFundTableDeleteIDs))
		for i, id := range req.MutualFundTableDeleteIDs {
			mftIds[i] = id
		}
		if len(req.MutualFundTableDeleteIDs) > 0 {
			if err := c.core.MutualFundTableManager().BulkDeleteWithTx(context, tx, mftIds); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to delete mutual fund tables: " + err.Error(),
					Module:      "MutualFund",
				})
				endTx(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete mutual fund tables: " + err.Error()})
			}
		}

		for _, additionalMember := range req.MutualFundAdditionalMembers {
			if additionalMember.ID != nil {
				existingMember, err := c.core.MutualFundAdditionalMembersManager().GetByID(context, *additionalMember.ID)
				if err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Additional member not found for update: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Additional member not found: " + err.Error()})
				}
				existingMember.MemberTypeID = additionalMember.MemberTypeID
				existingMember.NumberOfMembers = additionalMember.NumberOfMembers
				existingMember.Ratio = additionalMember.Ratio
				existingMember.UpdatedAt = time.Now().UTC()
				existingMember.UpdatedByID = userOrg.UserID
				if err := c.core.MutualFundAdditionalMembersManager().UpdateByIDWithTx(context, tx, existingMember.ID, existingMember); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Additional member update failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update additional member: " + err.Error()})
				}
			} else {
				additionalMemberData := &core.MutualFundAdditionalMembers{
					MutualFundID:    mutualFund.ID,
					MemberTypeID:    additionalMember.MemberTypeID,
					NumberOfMembers: additionalMember.NumberOfMembers,
					Ratio:           additionalMember.Ratio,
					CreatedAt:       time.Now().UTC(),
					CreatedByID:     userOrg.UserID,
					UpdatedAt:       time.Now().UTC(),
					UpdatedByID:     userOrg.UserID,
					BranchID:        *userOrg.BranchID,
					OrganizationID:  userOrg.OrganizationID,
				}
				if err := c.core.MutualFundAdditionalMembersManager().CreateWithTx(context, tx, additionalMemberData); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Additional member creation failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create additional member: " + err.Error()})
				}
			}
		}

		for _, mutualFundTable := range req.MutualFundTables {
			if mutualFundTable.ID != nil {
				existingTable, err := c.core.MutualFundTableManager().GetByID(context, *mutualFundTable.ID)
				if err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Mutual fund table not found for update: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund table not found: " + err.Error()})
				}
				existingTable.MonthFrom = mutualFundTable.MonthFrom
				existingTable.MonthTo = mutualFundTable.MonthTo
				existingTable.Amount = mutualFundTable.Amount
				existingTable.UpdatedAt = time.Now().UTC()
				existingTable.UpdatedByID = userOrg.UserID
				if err := c.core.MutualFundTableManager().UpdateByIDWithTx(context, tx, existingTable.ID, existingTable); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Mutual fund table update failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund table: " + err.Error()})
				}
			} else {
				mutualFundTableData := &core.MutualFundTable{
					MutualFundID:   mutualFund.ID,
					MonthFrom:      mutualFundTable.MonthFrom,
					MonthTo:        mutualFundTable.MonthTo,
					Amount:         mutualFundTable.Amount,
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
					BranchID:       *userOrg.BranchID,
					OrganizationID: userOrg.OrganizationID,
				}
				if err := c.core.MutualFundTableManager().CreateWithTx(context, tx, mutualFundTableData); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Mutual fund table creation failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund table: " + err.Error()})
				}
			}
		}
		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		mutualFundUpdated, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		mutualFundView, err := c.event.GenerateMutualFundEntries(context, userOrg, mutualFundUpdated)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund view: " + err.Error()})
		}
		for _, entry := range mutualFundView {
			if err := c.core.MutualFundEntryManager().CreateWithTx(context, tx, entry); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund entry creation failed: " + err.Error(),
					Module:      "MutualFund",
				})
				endTx(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund entry: " + err.Error()})
			}
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated mutual fund (/mutual-fund/:mutual_fund_id): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.JSON(http.StatusOK, c.core.MutualFundManager().ToModel(mutualFund))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/mutual-fund/:mutual_fund_id",
		Method: "DELETE",
		Note:   "Deletes the specified mutual fund by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), invalid mutual fund ID.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), not found.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		if err := c.core.MutualFundManager().Delete(context, *mutualFundID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), db error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete mutual fund: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted mutual fund (/mutual-fund/:mutual_fund_id): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/mutual-fund/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple mutual funds by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual funds (/mutual-fund/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual funds (/mutual-fund/bulk-delete) | no IDs provided",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No mutual fund IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.MutualFundManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual funds (/mutual-fund/bulk-delete) | error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete mutual funds: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted mutual funds (/mutual-fund/bulk-delete)",
			Module:      "MutualFund",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/mutual-fund/view",
		Method:       "POST",
		Note:         "Retrieves a summarized view of mutual funds including total amount and entries.",
		ResponseType: core.MutualFundView{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		reqData, err := c.core.MutualFundManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "Invalid mutual fund data: " + err.Error()},
			)
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), user org error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusUnauthorized,
				map[string]string{"error": "User organization not found or authentication failed"},
			)
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), user not assigned to branch.",
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "User is not assigned to a branch"},
			)
		}
		mutualFund, err := c.core.CreateMutualFundValue(context, reqData, userOrg)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), build error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)
		}
		mutualFundEntries, err := c.event.GenerateMutualFundEntries(context, userOrg, mutualFund)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), entry generation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to retrieve mutual fund view: " + err.Error()},
			)
		}
		total := 0.0
		for _, entry := range mutualFundEntries {
			total += entry.Amount
		}
		return ctx.JSON(http.StatusOK, core.MutualFundView{
			TotalAmount:       total,
			MutualFundEntries: c.core.MutualFundEntryManager().ToModels(mutualFundEntries),
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/mutual-fund/:mutual_fund_id/print",
		Note:   "Prints mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print mutual fund entries"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		if mutualFund == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		if mutualFund.PrintedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund has already been printed"})
		}
		now := time.Now().UTC()
		mutualFund.PrintedByUserID = &userOrg.UserID
		mutualFund.PrintedDate = &now
		if err := c.core.MutualFundManager().UpdateByID(context, mutualFund.ID, mutualFund); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund as printed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundManager().ToModel(mutualFund))
	})

	req.RegisterWebRoute(handlers.Route{
		Method: "PUT",
		Route:  "/api/v1/mutual-fund/:mutual_fund_id/print-undo",
		Note:   "Undoes the print status of mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print status of mutual fund entries"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		if mutualFund == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		if mutualFund.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund has not been printed yet"})
		}
		if mutualFund.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot undo print status - mutual fund has already been posted"})
		}
		mutualFund.PrintedByUserID = nil
		mutualFund.PrintedDate = nil
		if err := c.core.MutualFundManager().UpdateByID(context, mutualFund.ID, mutualFund); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print status of mutual fund: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MutualFundManager().ToModel(mutualFund))
	})

	req.RegisterWebRoute(handlers.Route{
		Method:      "PUT",
		Route:       "/api/v1/mutual-fund/:mutual_fund_id/post",
		RequestType: core.MutualFundViewPostRequest{},
		Note:        "Posts mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		var req core.MutualFundViewPostRequest
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post mutual fund entries"})
		}
		mutualFund, err := c.core.MutualFundManager().GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		if mutualFund.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund must be printed before posting"})
		}
		if mutualFund.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund has already been posted"})
		}
		if err := c.event.GenerateMutualFundEntriesPost(
			context, userOrg, mutualFundID, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post mutual fund entries: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Method:       "GET",
		Route:        "/api/v1/mutual-fund/:mutual_fund_id/view",
		ResponseType: core.MutualFundView{},
		Note:         "Returns mutual fund entries for a specific mutual fund ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := handlers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		entries, err := c.core.MutualFundEntryManager().Find(context, &core.MutualFundEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			MutualFundID:   *mutualFundID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund entries: " + err.Error()})
		}
		totalAmount := 0.0

		for _, entry := range entries {
			totalAmount = c.provider.Service.Decimal.Add(totalAmount, entry.Amount)
		}
		return ctx.JSON(http.StatusOK, core.MutualFundView{
			MutualFundEntries: c.core.MutualFundEntryManager().ToModels(entries),
			TotalAmount:       totalAmount,
		})
	})

}
