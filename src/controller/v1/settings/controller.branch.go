package settings

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func BranchController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch",
		Method:       "GET",
		Note:         "Returns all branches if unauthenticated; otherwise, returns branches filtered by the user's organization from cache.",
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil || userOrg == nil {
			branches, err := core.BranchManager(service).List(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve branches: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := core.GetBranchesByOrganization(context, service, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.BranchManager(service).ToModels(branches))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch/kyc",
		Method:       "GET",
		Note:         "Returns all branches belonging to the specified organization in members portal.",
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		org, ok := event.GetOrganization(service, ctx)
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization"})
		}
		branches, err := core.GetBranchesByOrganization(context, service, org.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.BranchManager(service).ToModels(branches))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "GET",
		Note:         "Returns all branches belonging to the specified organization.",
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		branches, err := core.GetBranchesByOrganization(context, service, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.BranchManager(service).ToModels(branches))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "POST",
		Note:         "Creates a new branch for the given organization. If the user already has a branch, a new user organization is created; otherwise, the user's current user organization is updated with the new branch.",
		RequestType:  types.BranchRequest{},
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.BranchManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate branch data for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Invalid organization ID for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}

		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: "User authentication required for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		userOrganization, err := core.UserOrganizationManager(service).FindOne(context, &types.UserOrganization{
			UserID:         user.ID,
			OrganizationID: *organizationID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("User organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found " + err.Error()})
		}
		if userOrganization.UserType != types.UserOrganizationTypeOwner {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: "Only organization owners can create branches for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners can create branches "})
		}

		organization, err := core.OrganizationManager(service).GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found " + err.Error()})
		}

		branchCount, err := core.GetBranchesByOrganizationCount(context, service, organization.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed branch count for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not count organization branches: " + err.Error()})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: "Branch limit reached for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for the current subscription plan "})
		}

		branch := &types.Branch{
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             user.ID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             user.ID,
			OrganizationID:          userOrganization.OrganizationID,
			MediaID:                 req.MediaID,
			Type:                    req.Type,
			Name:                    req.Name,
			Email:                   req.Email,
			Description:             req.Description,
			CurrencyID:              req.CurrencyID,
			ContactNumber:           req.ContactNumber,
			Address:                 req.Address,
			Province:                req.Province,
			City:                    req.City,
			Region:                  req.Region,
			Barangay:                req.Barangay,
			PostalCode:              req.PostalCode,
			Latitude:                req.Latitude,
			Longitude:               req.Longitude,
			IsMainBranch:            req.IsMainBranch,
			TaxIdentificationNumber: req.TaxIdentificationNumber,
		}

		tx, endTx := service.Database.StartTransaction(context)

		if err := core.BranchManager(service).CreateWithTx(context, tx, branch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + endTx(err).Error()})
		}

		branchSetting := &types.BranchSetting{
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			BranchID:   branch.ID,
			CurrencyID: *req.CurrencyID,
		}

		if err := core.BranchSettingManager(service).CreateWithTx(context, tx, branchSetting); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch settings for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + endTx(err).Error()})
		}

		if userOrganization.BranchID == nil {
			userOrganization.BranchID = &branch.ID
			userOrganization.UpdatedAt = time.Now().UTC()
			userOrganization.UpdatedByID = user.ID

			if err := core.UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to update user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + endTx(err).Error()})
			}
		} else {
			developerKey, err := service.Security.GenerateUUIDv5(user.ID.String())
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to generate developer key for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + endTx(err).Error()})
			}

			newUserOrg := &types.UserOrganization{
				CreatedAt:          time.Now().UTC(),
				CreatedByID:        user.ID,
				UpdatedAt:          time.Now().UTC(),
				UpdatedByID:        user.ID,
				OrganizationID:     userOrganization.OrganizationID,
				BranchID:           &branch.ID,
				UserID:             user.ID,
				UserType:           types.UserOrganizationTypeOwner,
				ApplicationStatus:  "accepted",
				DeveloperSecretKey: developerKey + uuid.NewString() + "-horizon",
				PermissionName:     string(types.UserOrganizationTypeOwner),
				Permissions:        []string{},
			}

			if err := core.UserOrganizationManager(service).CreateWithTx(context, tx, newUserOrg); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to create new user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + endTx(err).Error()})
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to commit transaction for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Notification(ctx, service, event.NotificationEvent{
			Title:       fmt.Sprintf("Create: %s", branch.Name),
			Description: fmt.Sprintf("Created a new branch: %s", branch.Name),
		})

		event.OrganizationDirectNotification(ctx, service, userOrganization.OrganizationID, event.NotificationEvent{
			Description:      fmt.Sprintf("New branch '%s' has been created by %s %s", branch.Name, *user.FirstName, *user.LastName),
			Title:            "New Branch Created",
			NotificationType: types.NotificationInfo,
		})

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, core.BranchManager(service).ToModel(branch))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch/:branch_id",
		Method:       "PUT",
		Note:         "Updates branch information for the specified branch. Only allowed for the owner of the branch.",
		RequestType:  types.BranchRequest{},
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.BranchManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch data for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "User authentication required for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		branchID, err := helpers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid branch id for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}

		userOrg, err := core.UserOrganizationManager(service).FindOne(context, &types.UserOrganization{
			UserID:   user.ID,
			BranchID: branchID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("User organization not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization for this branch not found: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "Only the branch owner can update branch for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can update branch information "})
		}

		branch, err := core.BranchManager(service).GetByID(context, *branchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		branch.UpdatedAt = time.Now().UTC()
		branch.UpdatedByID = user.ID
		branch.MediaID = req.MediaID
		branch.Type = req.Type
		branch.Name = req.Name
		branch.Email = req.Email
		branch.Description = req.Description
		branch.CurrencyID = req.CurrencyID
		branch.ContactNumber = req.ContactNumber
		branch.Address = req.Address
		branch.Province = req.Province
		branch.City = req.City
		branch.Region = req.Region
		branch.Barangay = req.Barangay
		branch.PostalCode = req.PostalCode
		branch.Latitude = req.Latitude
		branch.Longitude = req.Longitude
		branch.IsMainBranch = req.IsMainBranch
		branch.TaxIdentificationNumber = req.TaxIdentificationNumber

		if err := core.BranchManager(service).UpdateByID(context, branch.ID, branch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		event.Notification(ctx, service, event.NotificationEvent{
			Title:       fmt.Sprintf("Update: %s", branch.Name),
			Description: fmt.Sprintf("Updated branch: %s", branch.Name),
		})

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, core.BranchManager(service).ToModel(branch))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/branch/:branch_id",
		Method: "DELETE",
		Note:   "Deletes the specified branch if the user is the owner and there are less than 3 members in the branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchID, err := helpers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid branch ID for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: "User authentication required for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required "})
		}
		branch, err := core.BranchManager(service).GetByID(context, *branchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Branch not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		userOrganization, err := core.UserOrganizationManager(service).FindOne(context, &types.UserOrganization{
			UserID:         user.ID,
			BranchID:       branchID,
			OrganizationID: branch.OrganizationID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("User organization not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if userOrganization.UserType != types.UserOrganizationTypeOwner {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Only the branch owner can delete this branch for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can delete this branch"})
		}
		count, err := core.CountUserOrganizationPerBranch(context, service, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Could not check branch membership for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not check branch membership: " + err.Error()})
		}
		if count > 2 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Cannot delete branch with more than 2 members for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete branch with more than 2 members"})
		}
		tx, endTx := service.Database.StartTransaction(context)

		if err := core.BranchManager(service).DeleteWithTx(context, tx, branch.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete branch for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + endTx(err).Error()})
		}
		if err := core.UserOrganizationManager(service).DeleteWithTx(context, tx, userOrganization.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete user organization for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to commit transaction for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		event.Notification(ctx, service, event.NotificationEvent{
			Title:       fmt.Sprintf("Delete: %s", branch.Name),
			Description: fmt.Sprintf("Deleted branch: %s", branch.Name),
		})
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch/:branch_id",
		Method:       "GET",
		Note:         "Returns a single branch by its ID.",
		ResponseType: types.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchID, err := helpers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID"})
		}
		branch, err := core.BranchManager(service).GetByIDRaw(context, *branchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found"})
		}

		return ctx.JSON(http.StatusOK, branch)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch-settings",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  types.BranchSettingRequest{},
		ResponseType: types.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var settingsReq types.BranchSettingRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if err := service.Validator.Struct(settingsReq); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil || userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}

		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "Insufficient permissions to update branch settings for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to update branch settings"})
		}

		var branchSetting *types.BranchSetting
		branchSetting, err = core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{
			BranchID: *userOrg.BranchID,
		})
		branchSetting.UpdatedAt = time.Now().UTC()
		branchSetting.WithdrawAllowUserInput = settingsReq.WithdrawAllowUserInput
		branchSetting.WithdrawPrefix = settingsReq.WithdrawPrefix
		branchSetting.WithdrawORStart = settingsReq.WithdrawORStart
		branchSetting.WithdrawORCurrent = settingsReq.WithdrawORCurrent
		branchSetting.WithdrawOREnd = settingsReq.WithdrawOREnd
		branchSetting.WithdrawORIteration = settingsReq.WithdrawORIteration
		branchSetting.WithdrawUseDateOR = settingsReq.WithdrawUseDateOR
		branchSetting.WithdrawPadding = settingsReq.WithdrawPadding
		branchSetting.WithdrawCommonOR = settingsReq.WithdrawCommonOR

		branchSetting.DepositORStart = settingsReq.DepositORStart
		branchSetting.DepositORCurrent = settingsReq.DepositORCurrent
		branchSetting.DepositOREnd = settingsReq.DepositOREnd
		branchSetting.DepositORIteration = settingsReq.DepositORIteration
		branchSetting.DepositUseDateOR = settingsReq.DepositUseDateOR
		branchSetting.DepositPadding = settingsReq.DepositPadding
		branchSetting.DepositCommonOR = settingsReq.DepositCommonOR

		branchSetting.CashCheckVoucherAllowUserInput = settingsReq.CashCheckVoucherAllowUserInput
		branchSetting.CashCheckVoucherORUnique = settingsReq.CashCheckVoucherORUnique
		branchSetting.CashCheckVoucherPrefix = settingsReq.CashCheckVoucherPrefix
		branchSetting.CashCheckVoucherORStart = settingsReq.CashCheckVoucherORStart
		branchSetting.CashCheckVoucherORCurrent = settingsReq.CashCheckVoucherORCurrent
		branchSetting.CashCheckVoucherPadding = settingsReq.CashCheckVoucherPadding

		branchSetting.JournalVoucherAllowUserInput = settingsReq.JournalVoucherAllowUserInput
		branchSetting.JournalVoucherORUnique = settingsReq.JournalVoucherORUnique
		branchSetting.JournalVoucherPrefix = settingsReq.JournalVoucherPrefix
		branchSetting.JournalVoucherORStart = settingsReq.JournalVoucherORStart
		branchSetting.JournalVoucherORCurrent = settingsReq.JournalVoucherORCurrent
		branchSetting.JournalVoucherPadding = settingsReq.JournalVoucherPadding

		branchSetting.AdjustmentVoucherAllowUserInput = settingsReq.AdjustmentVoucherAllowUserInput
		branchSetting.AdjustmentVoucherORUnique = settingsReq.AdjustmentVoucherORUnique
		branchSetting.AdjustmentVoucherPrefix = settingsReq.AdjustmentVoucherPrefix
		branchSetting.AdjustmentVoucherORStart = settingsReq.AdjustmentVoucherORStart
		branchSetting.AdjustmentVoucherORCurrent = settingsReq.AdjustmentVoucherORCurrent
		branchSetting.AdjustmentVoucherPadding = settingsReq.AdjustmentVoucherPadding

		branchSetting.LoanVoucherAllowUserInput = settingsReq.LoanVoucherAllowUserInput
		branchSetting.LoanVoucherORUnique = settingsReq.LoanVoucherORUnique
		branchSetting.LoanVoucherPrefix = settingsReq.LoanVoucherPrefix
		branchSetting.LoanVoucherORStart = settingsReq.LoanVoucherORStart
		branchSetting.LoanVoucherORCurrent = settingsReq.LoanVoucherORCurrent
		branchSetting.LoanVoucherPadding = settingsReq.LoanVoucherPadding

		branchSetting.CheckVoucherGeneral = settingsReq.CheckVoucherGeneral
		branchSetting.CheckVoucherGeneralAllowUserInput = settingsReq.CheckVoucherGeneralAllowUserInput
		branchSetting.CheckVoucherGeneralORUnique = settingsReq.CheckVoucherGeneralORUnique
		branchSetting.CheckVoucherGeneralPrefix = settingsReq.CheckVoucherGeneralPrefix
		branchSetting.CheckVoucherGeneralORStart = settingsReq.CheckVoucherGeneralORStart
		branchSetting.CheckVoucherGeneralORCurrent = settingsReq.CheckVoucherGeneralORCurrent
		branchSetting.CheckVoucherGeneralPadding = settingsReq.CheckVoucherGeneralPadding
		branchSetting.DefaultMemberTypeID = settingsReq.DefaultMemberTypeID
		branchSetting.DefaultMemberGenderID = settingsReq.DefaultMemberGenderID
		branchSetting.LoanAppliedEqualToBalance = settingsReq.LoanAppliedEqualToBalance
		branchSetting.AnnualDivisor = settingsReq.AnnualDivisor
		branchSetting.TaxInterest = settingsReq.TaxInterest

		if err := core.BranchSettingManager(service).UpdateByID(context, branchSetting.ID, branchSetting); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings for branch ID: %s", userOrg.BranchID),
			Module:      "branch",
		})

		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Title:       "Branch Settings Updated",
			Description: "Branch settings have been successfully updated",
		})
		newBranchSettings, err := core.BranchSettingManager(service).GetByIDRaw(context, branchSetting.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get latest branch settings: " + err.Error()})

		}
		return ctx.JSON(http.StatusOK, newBranchSettings)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/branch-settings/currency",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  types.BranchSettingsCurrencyRequest{},
		ResponseType: types.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var settingsReq types.BranchSettingsCurrencyRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil || userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings/currency",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}
		if err := service.Validator.Struct(settingsReq); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		branchSetting, err := core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{
			BranchID: *userOrg.BranchID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch settings not found for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch settings not found: " + err.Error()})
		}

		branchSetting.CurrencyID = settingsReq.CurrencyID
		branchSetting.CompassionFundAccountID = settingsReq.CompassionFundAccountID
		branchSetting.AccountWalletID = &settingsReq.AccountWalletID
		branchSetting.PaidUpSharedCapitalAccountID = &settingsReq.PaidUpSharedCapitalAccountID
		branchSetting.CashOnHandAccountID = &settingsReq.CashOnHandAccountID
		branchSetting.UpdatedAt = time.Now().UTC()

		if err := core.BranchSettingManager(service).UpdateByID(context, branchSetting.ID, branchSetting); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings currency: " + err.Error()})
		}

		for _, id := range settingsReq.UnbalancedAccountDeleteIDs {
			if err := core.UnbalancedAccountManager(service).Delete(context, id); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to delete unbalanced account: " + err.Error(),
					Module:      "UnbalancedAccount",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete unbalanced account: " + err.Error()})
			}
		}

		for _, accountReq := range settingsReq.UnbalancedAccount {
			if accountReq.ID != nil {
				existingAccount, err := core.UnbalancedAccountManager(service).GetByID(context, *accountReq.ID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get unbalanced account: " + err.Error()})
				}

				existingAccount.Name = accountReq.Name
				existingAccount.Description = accountReq.Description
				existingAccount.CurrencyID = accountReq.CurrencyID
				existingAccount.AccountForShortageID = accountReq.AccountForShortageID
				existingAccount.AccountForOverageID = accountReq.AccountForOverageID
				existingAccount.CashOnHandAccountID = accountReq.CashOnHandAccountID
				existingAccount.MemberProfileIDForShortage = accountReq.MemberProfileIDForShortage
				existingAccount.MemberProfileIDForOverage = accountReq.MemberProfileIDForOverage

				existingAccount.UpdatedAt = time.Now().UTC()
				existingAccount.UpdatedByID = userOrg.UserID
				if err := core.UnbalancedAccountManager(service).UpdateByID(context, existingAccount.ID, existingAccount); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme account: " + err.Error()})
				}
			} else {
				newUnbalancedAccount := &types.UnbalancedAccount{
					BranchSettingsID: branchSetting.ID,
					CreatedAt:        time.Now().UTC(),
					CreatedByID:      userOrg.UserID,
					UpdatedAt:        time.Now().UTC(),
					UpdatedByID:      userOrg.UserID,

					Name:                 accountReq.Name,
					Description:          accountReq.Description,
					CurrencyID:           accountReq.CurrencyID,
					AccountForShortageID: accountReq.AccountForShortageID,
					AccountForOverageID:  accountReq.AccountForOverageID,

					CashOnHandAccountID:        accountReq.CashOnHandAccountID,
					MemberProfileIDForShortage: accountReq.MemberProfileIDForShortage,
					MemberProfileIDForOverage:  accountReq.MemberProfileIDForOverage,
				}
				if err := core.UnbalancedAccountManager(service).Create(context, newUnbalancedAccount); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create unbalanced account: " + err.Error()})
				}
			}
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings currency for branch settings ID: %s", branchSetting.ID),
			Module:      "branch",
		})

		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Title:            "Branch Settings Updated",
			Description:      "Branch settings have been successfully updated",
			NotificationType: types.NotificationAlert,
		})
		newBranchSettings, err := core.BranchSettingManager(service).GetByIDRaw(context, branchSetting.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get latest branch settings: " + err.Error()})

		}
		return ctx.JSON(http.StatusOK, newBranchSettings)
	})
}
