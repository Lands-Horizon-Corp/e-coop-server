package funds

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

func MutualFundsController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund",
		Method:       "GET",
		Note:         "Returns all mutual funds for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := core.MutualFundCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual funds found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundManager(service).ToModels(mutualFunds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund/search",
		Method:       "GET",
		Note:         "Returns a paginated list of mutual funds for the current user's organization and branch.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := core.MutualFundManager(service).NormalPagination(context, ctx, &types.MutualFund{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch mutual funds for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, mutualFunds)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund/member/:member_id",
		Method:       "GET",
		Note:         "Returns all mutual funds for a specific member profile.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberID, err := helpers.EngineUUIDParam(ctx, "member_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		mutualFunds, err := core.MutualFundByMember(context, service, *memberID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No mutual funds found for the specified member"})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundManager(service).ToModels(mutualFunds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund/:mutual_fund_id",
		Method:       "GET",
		Note:         "Returns a single mutual fund by its ID.",
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByIDRaw(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		return ctx.JSON(http.StatusOK, mutualFund)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund",
		Method:       "POST",
		Note:         "Creates a new mutual fund for the current user's organization and branch.",
		RequestType:  core.MutualFundRequest{},
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MutualFundManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), user org error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), user not assigned to branch.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		mutualFund, err := core.CreateMutualFundValue(context, service, req, userOrg)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)
		}
		tx, endTx := service.Database.StartTransaction(context)
		if err := core.MutualFundManager(service).CreateWithTx(context, tx, mutualFund); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Mutual fund creation failed (/mutual-fund), db error: " + err.Error(),
				Module:      "MutualFund",
			})

			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund: " + endTx(err).Error()})
		}
		for _, additionalMember := range mutualFund.AdditionalMembers {
			if err := core.MutualFundAdditionalMembersManager(service).CreateWithTx(context, tx, additionalMember); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund additional member creation failed: " + err.Error(),
					Module:      "MutualFund",
				})

				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create additional member: " + endTx(err).Error()})
			}
		}
		for _, mutualFundTable := range mutualFund.MutualFundTables {
			if err := core.MutualFundTableManager(service).CreateWithTx(context, tx, mutualFundTable); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund table creation failed: " + err.Error(),
					Module:      "MutualFund",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund table: " + endTx(err).Error()})
			}
		}
		mutualFundView, err := event.GenerateMutualFundEntries(context, service, userOrg, mutualFund)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund view: " + endTx(err).Error()})
		}
		for _, entry := range mutualFundView {
			if err := core.MutualFundEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
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
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created mutual fund (/mutual-fund): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.JSON(http.StatusCreated, core.MutualFundManager(service).ToModel(mutualFund))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund/:mutual_fund_id",
		Method:       "PUT",
		Note:         "Updates an existing mutual fund by its ID.",
		RequestType:  core.MutualFundRequest{},
		ResponseType: core.MutualFundResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), invalid mutual fund ID.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}

		req, err := core.MutualFundManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Mutual fund update failed (/mutual-fund/:mutual_fund_id), user org error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		tx, endTx := service.Database.StartTransaction(context)

		if err := core.MutualFundManager(service).UpdateByIDWithTx(context, tx, mutualFund.ID, mutualFund); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
			if err := core.MutualFundAdditionalMembersManager(service).BulkDeleteWithTx(context, tx, mfIds); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
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
			if err := core.MutualFundTableManager(service).BulkDeleteWithTx(context, tx, mftIds); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
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
				existingMember, err := core.MutualFundAdditionalMembersManager(service).GetByID(context, *additionalMember.ID)
				if err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
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
				if err := core.MutualFundAdditionalMembersManager(service).UpdateByIDWithTx(context, tx, existingMember.ID, existingMember); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Additional member update failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update additional member: " + err.Error()})
				}
			} else {
				additionalMemberData := &types.MutualFundAdditionalMembers{
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
				if err := core.MutualFundAdditionalMembersManager(service).CreateWithTx(context, tx, additionalMemberData); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
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
				existingTable, err := core.MutualFundTableManager(service).GetByID(context, *mutualFundTable.ID)
				if err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
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
				if err := core.MutualFundTableManager(service).UpdateByIDWithTx(context, tx, existingTable.ID, existingTable); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Mutual fund table update failed: " + err.Error(),
						Module:      "MutualFund",
					})
					endTx(err)
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund table: " + err.Error()})
				}
			} else {
				mutualFundTableData := &types.MutualFundTable{
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
				if err := core.MutualFundTableManager(service).CreateWithTx(context, tx, mutualFundTableData); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
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
		mutualFundUpdated, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		mutualFundView, err := event.GenerateMutualFundEntries(context, service, userOrg, mutualFundUpdated)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund view: " + err.Error()})
		}
		for _, entry := range mutualFundView {
			if err := core.MutualFundEntryManager(service).CreateWithTx(context, tx, entry); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Mutual fund entry creation failed: " + err.Error(),
					Module:      "MutualFund",
				})
				endTx(err)
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create mutual fund entry: " + err.Error()})
			}
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated mutual fund (/mutual-fund/:mutual_fund_id): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.JSON(http.StatusOK, core.MutualFundManager(service).ToModel(mutualFund))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/mutual-fund/:mutual_fund_id",
		Method: "DELETE",
		Note:   "Deletes the specified mutual fund by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), invalid mutual fund ID.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), not found.",
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}
		if err := core.MutualFundManager(service).Delete(context, *mutualFundID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Mutual fund delete failed (/mutual-fund/:mutual_fund_id), db error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete mutual fund: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted mutual fund (/mutual-fund/:mutual_fund_id): " + mutualFund.Name,
			Module:      "MutualFund",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/mutual-fund/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple mutual funds by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual funds (/mutual-fund/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.MutualFundManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete mutual funds (/mutual-fund/bulk-delete) | error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete mutual funds: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted mutual funds (/mutual-fund/bulk-delete)",
			Module:      "MutualFund",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/mutual-fund/view",
		Method:       "POST",
		Note:         "Retrieves a summarized view of mutual funds including total amount and entries.",
		ResponseType: core.MutualFundView{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		reqData, err := core.MutualFundManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), validation error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "Invalid mutual fund data: " + err.Error()},
			)
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), user not assigned to branch.",
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "User is not assigned to a branch"},
			)
		}
		mutualFund, err := core.CreateMutualFundValue(context, service, reqData, userOrg)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "view-error",
				Description: "Mutual fund view failed (/mutual-fund/view), build error: " + err.Error(),
				Module:      "MutualFund",
			})
			return ctx.JSON(
				http.StatusBadRequest,
				map[string]string{"error": err.Error()},
			)
		}
		mutualFundEntries, err := event.GenerateMutualFundEntries(context, service, userOrg, mutualFund)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
			MutualFundEntries: core.MutualFundEntryManager(service).ToModels(mutualFundEntries),
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Method: "PUT",
		Route:  "/api/v1/mutual-fund/:mutual_fund_id/print",
		Note:   "Prints mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print mutual fund entries"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
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
		if err := core.MutualFundManager(service).UpdateByID(context, mutualFund.ID, mutualFund); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update mutual fund as printed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundManager(service).ToModel(mutualFund))
	})

	req.RegisterWebRoute(horizon.Route{
		Method: "PUT",
		Route:  "/api/v1/mutual-fund/:mutual_fund_id/print-undo",
		Note:   "Undoes the print status of mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print status of mutual fund entries"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
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
		if err := core.MutualFundManager(service).UpdateByID(context, mutualFund.ID, mutualFund); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to undo print status of mutual fund: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MutualFundManager(service).ToModel(mutualFund))
	})

	req.RegisterWebRoute(horizon.Route{
		Method:      "PUT",
		Route:       "/api/v1/mutual-fund/:mutual_fund_id/post",
		RequestType: core.MutualFundViewPostRequest{},
		Note:        "Posts mutual fund entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		var req core.MutualFundViewPostRequest
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to post mutual fund entries"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund: " + err.Error()})
		}
		if mutualFund.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund must be printed before posting"})
		}
		if mutualFund.PostedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Mutual fund has already been posted"})
		}
		if err := event.GenerateMutualFundEntriesPost(
			context, service, userOrg, mutualFundID, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post mutual fund entries: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Method:       "GET",
		Route:        "/api/v1/mutual-fund/:mutual_fund_id/view",
		ResponseType: core.MutualFundView{},
		Note:         "Returns mutual fund entries for a specific mutual fund ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mutualFundID, err := helpers.EngineUUIDParam(ctx, "mutual_fund_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid mutual fund ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		mutualFund, err := core.MutualFundManager(service).GetByID(context, *mutualFundID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Mutual fund not found"})
		}

		entries, err := core.MutualFundEntryManager(service).Find(context, &types.MutualFundEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			MutualFundID:   mutualFund.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve mutual fund entries: " + err.Error()})
		}
		totalAmount := decimal.Zero

		for _, entry := range entries {
			totalAmount = totalAmount.Add(decimal.NewFromFloat(entry.Amount))
		}
		totalAmountFloat := totalAmount.InexactFloat64()
		return ctx.JSON(http.StatusOK, core.MutualFundView{
			MutualFundEntries: core.MutualFundEntryManager(service).ToModels(entries),
			TotalAmount:       totalAmountFloat,
			MutualFund:        core.MutualFundManager(service).ToModel(mutualFund),
		})
	})

}
