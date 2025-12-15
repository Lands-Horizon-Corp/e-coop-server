package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) branchController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch",
		Method:       "GET",
		Note:         "Returns all branches if unauthenticated; otherwise, returns branches filtered by the user's organization from JWT.",
		ResponseType: core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg == nil {
			branches, err := c.core.BranchManager.List(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve branches: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := c.core.GetBranchesByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.BranchManager.ToModels(branches))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "GET",
		Note:         "Returns all branches belonging to the specified organization.",
		ResponseType: core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		branches, err := c.core.GetBranchesByOrganization(context, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.BranchManager.ToModels(branches))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "POST",
		Note:         "Creates a new branch for the given organization. If the user already has a branch, a new user organization is created; otherwise, the user's current user organization is updated with the new branch.",
		Private:      true,
		RequestType:  core.BranchRequest{},
		ResponseType: core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.core.BranchManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate branch data for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Invalid organization ID for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "User authentication required for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		userOrganization, err := c.core.UserOrganizationManager.FindOne(context, &core.UserOrganization{
			UserID:         user.ID,
			OrganizationID: *organizationID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("User organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found " + err.Error()})
		}
		if userOrganization.UserType != core.UserOrganizationTypeOwner {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Only organization owners can create branches for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners can create branches "})
		}

		organization, err := c.core.OrganizationManager.GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found " + err.Error()})
		}

		branchCount, err := c.core.GetBranchesByOrganizationCount(context, organization.ID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed branch count for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not count organization branches: " + err.Error()})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Branch limit reached for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for the current subscription plan "})
		}

		branch := &core.Branch{
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

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		if err := c.core.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + endTx(err).Error()})
		}

		branchSetting := &core.BranchSetting{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			BranchID:  branch.ID,

			WithdrawAllowUserInput: true,
			WithdrawPrefix:         "WD",
			WithdrawORStart:        1,
			WithdrawORCurrent:      1,
			WithdrawOREnd:          999999,
			WithdrawORIteration:    1,
			WithdrawORUnique:       true,
			WithdrawUseDateOR:      false,

			DepositAllowUserInput: true,
			DepositPrefix:         "DP",
			DepositORStart:        1,
			DepositORCurrent:      1,
			DepositOREnd:          999999,
			DepositORIteration:    1,
			DepositORUnique:       true,
			DepositUseDateOR:      false,

			LoanAllowUserInput: true,
			LoanPrefix:         "LN",
			LoanORStart:        1,
			LoanORCurrent:      1,
			LoanOREnd:          999999,
			LoanORIteration:    1,
			LoanORUnique:       true,
			LoanUseDateOR:      false,

			CheckVoucherAllowUserInput: true,
			CheckVoucherPrefix:         "CV",
			CheckVoucherORStart:        1,
			CheckVoucherORCurrent:      1,
			CheckVoucherOREnd:          999999,
			CheckVoucherORIteration:    1,
			CheckVoucherORUnique:       true,
			CheckVoucherUseDateOR:      false,

			DefaultMemberTypeID:       nil,
			LoanAppliedEqualToBalance: true,
			CurrencyID:                *req.CurrencyID,
		}

		if err := c.core.BranchSettingManager.CreateWithTx(context, tx, branchSetting); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

			if err := c.core.UserOrganizationManager.UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to update user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + endTx(err).Error()})
			}
		} else {
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to generate developer key for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + endTx(err).Error()})
			}

			newUserOrg := &core.UserOrganization{
				CreatedAt:                time.Now().UTC(),
				CreatedByID:              user.ID,
				UpdatedAt:                time.Now().UTC(),
				UpdatedByID:              user.ID,
				OrganizationID:           userOrganization.OrganizationID,
				BranchID:                 &branch.ID,
				UserID:                   user.ID,
				UserType:                 core.UserOrganizationTypeOwner,
				ApplicationStatus:        "accepted",
				DeveloperSecretKey:       developerKey + uuid.NewString() + "-horizon",
				PermissionName:           string(core.UserOrganizationTypeOwner),
				Permissions:              []string{},
				UserSettingStartOR:       0,
				UserSettingEndOR:         1000,
				UserSettingUsedOR:        0,
				UserSettingStartVoucher:  0,
				UserSettingEndVoucher:    5,
				UserSettingUsedVoucher:   0,
				UserSettingNumberPadding: 7,
			}

			if err := c.core.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to create new user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + endTx(err).Error()})
			}
		}

		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to commit transaction for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Notification(ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Create: %s", branch.Name),
			Description: fmt.Sprintf("Created a new branch: %s", branch.Name),
		})

		c.event.OrganizationDirectNotification(ctx, userOrganization.OrganizationID, event.NotificationEvent{
			Description:      fmt.Sprintf("New branch '%s' has been created by %s %s", branch.Name, *user.FirstName, *user.LastName),
			Title:            "New Branch Created",
			NotificationType: core.NotificationInfo,
		})

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.core.BranchManager.ToModel(branch))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch/:branch_id",
		Method:       "PUT",
		Note:         "Updates branch information for the specified branch. Only allowed for the owner of the branch.",
		Private:      true,
		RequestType:  core.BranchRequest{},
		ResponseType: core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.core.BranchManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch data for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User authentication required for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid branch id for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}

		userOrg, err := c.core.UserOrganizationManager.FindOne(context, &core.UserOrganization{
			UserID:   user.ID,
			BranchID: branchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("User organization not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization for this branch not found: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "Only the branch owner can update branch for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can update branch information "})
		}

		branch, err := c.core.BranchManager.GetByID(context, *branchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.BranchManager.UpdateByID(context, branch.ID, branch); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		c.event.Notification(ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Update: %s", branch.Name),
			Description: fmt.Sprintf("Updated branch: %s", branch.Name),
		})

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.core.BranchManager.ToModel(branch))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:   "/api/v1/branch/:branch_id",
		Method:  "DELETE",
		Note:    "Deletes the specified branch if the user is the owner and there are less than 3 members in the branch.",
		Private: true,
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid branch ID for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "User authentication required for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required "})
		}
		branch, err := c.core.BranchManager.GetByID(context, *branchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Branch not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		userOrganization, err := c.core.UserOrganizationManager.FindOne(context, &core.UserOrganization{
			UserID:         user.ID,
			BranchID:       branchID,
			OrganizationID: branch.OrganizationID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("User organization not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if userOrganization.UserType != core.UserOrganizationTypeOwner {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Only the branch owner can delete this branch for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can delete this branch"})
		}
		count, err := c.core.CountUserOrganizationPerbranch(context, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Could not check branch membership for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not check branch membership: " + err.Error()})
		}
		if count > 2 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Cannot delete branch with more than 2 members for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete branch with more than 2 members"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		if err := c.core.BranchManager.DeleteWithTx(context, tx, branch.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete branch for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + endTx(err).Error()})
		}
		if err := c.core.UserOrganizationManager.DeleteWithTx(context, tx, userOrganization.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete user organization for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to commit transaction for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Notification(ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Delete: %s", branch.Name),
			Description: fmt.Sprintf("Deleted branch: %s", branch.Name),
		})
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch-settings",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  core.BranchSettingRequest{},
		ResponseType: core.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var settingsReq core.BranchSettingRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(settingsReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "Insufficient permissions to update branch settings for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to update branch settings"})
		}

		var branchSetting *core.BranchSetting
		branchSetting, err = c.core.BranchSettingManager.FindOne(context, &core.BranchSetting{
			BranchID: *userOrg.BranchID,
		})

		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		if err != nil {
			branchSetting = &core.BranchSetting{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				BranchID:  *userOrg.BranchID,

				WithdrawAllowUserInput: settingsReq.WithdrawAllowUserInput,
				WithdrawPrefix:         settingsReq.WithdrawPrefix,
				WithdrawORStart:        settingsReq.WithdrawORStart,
				WithdrawORCurrent:      settingsReq.WithdrawORCurrent,
				WithdrawOREnd:          settingsReq.WithdrawOREnd,
				WithdrawORIteration:    settingsReq.WithdrawORIteration,
				WithdrawORUnique:       settingsReq.WithdrawORUnique,
				WithdrawUseDateOR:      settingsReq.WithdrawUseDateOR,

				DepositAllowUserInput: settingsReq.DepositAllowUserInput,
				DepositPrefix:         settingsReq.DepositPrefix,
				DepositORStart:        settingsReq.DepositORStart,
				DepositORCurrent:      settingsReq.DepositORCurrent,
				DepositOREnd:          settingsReq.DepositOREnd,
				DepositORIteration:    settingsReq.DepositORIteration,
				DepositORUnique:       settingsReq.DepositORUnique,
				DepositUseDateOR:      settingsReq.DepositUseDateOR,

				LoanAllowUserInput: settingsReq.LoanAllowUserInput,
				LoanPrefix:         settingsReq.LoanPrefix,
				LoanORStart:        settingsReq.LoanORStart,
				LoanORCurrent:      settingsReq.LoanORCurrent,
				LoanOREnd:          settingsReq.LoanOREnd,
				LoanORIteration:    settingsReq.LoanORIteration,
				LoanORUnique:       settingsReq.LoanORUnique,
				LoanUseDateOR:      settingsReq.LoanUseDateOR,

				CheckVoucherAllowUserInput: settingsReq.CheckVoucherAllowUserInput,
				CheckVoucherPrefix:         settingsReq.CheckVoucherPrefix,
				CheckVoucherORStart:        settingsReq.CheckVoucherORStart,
				CheckVoucherORCurrent:      settingsReq.CheckVoucherORCurrent,
				CheckVoucherOREnd:          settingsReq.CheckVoucherOREnd,
				CheckVoucherORIteration:    settingsReq.CheckVoucherORIteration,
				CheckVoucherORUnique:       settingsReq.CheckVoucherORUnique,
				CheckVoucherUseDateOR:      settingsReq.CheckVoucherUseDateOR,
				AnnualDivisor:              settingsReq.AnnualDivisor,

				DefaultMemberTypeID: settingsReq.DefaultMemberTypeID,
				TaxInterest:         settingsReq.TaxInterest,
			}

			if err := c.core.BranchSettingManager.CreateWithTx(context, tx, branchSetting); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update error",
					Description: fmt.Sprintf("Failed to create branch settings for PUT /branch-settings: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + endTx(err).Error()})
			}
		} else {
			branchSetting.UpdatedAt = time.Now().UTC()

			branchSetting.WithdrawAllowUserInput = settingsReq.WithdrawAllowUserInput
			branchSetting.WithdrawPrefix = settingsReq.WithdrawPrefix
			branchSetting.WithdrawORStart = settingsReq.WithdrawORStart
			branchSetting.WithdrawORCurrent = settingsReq.WithdrawORCurrent
			branchSetting.WithdrawOREnd = settingsReq.WithdrawOREnd
			branchSetting.WithdrawORIteration = settingsReq.WithdrawORIteration
			branchSetting.WithdrawORUnique = settingsReq.WithdrawORUnique
			branchSetting.WithdrawUseDateOR = settingsReq.WithdrawUseDateOR

			branchSetting.DepositAllowUserInput = settingsReq.DepositAllowUserInput
			branchSetting.DepositPrefix = settingsReq.DepositPrefix
			branchSetting.DepositORStart = settingsReq.DepositORStart
			branchSetting.DepositORCurrent = settingsReq.DepositORCurrent
			branchSetting.DepositOREnd = settingsReq.DepositOREnd
			branchSetting.DepositORIteration = settingsReq.DepositORIteration
			branchSetting.DepositORUnique = settingsReq.DepositORUnique
			branchSetting.DepositUseDateOR = settingsReq.DepositUseDateOR

			branchSetting.LoanAllowUserInput = settingsReq.LoanAllowUserInput
			branchSetting.LoanPrefix = settingsReq.LoanPrefix
			branchSetting.LoanORStart = settingsReq.LoanORStart
			branchSetting.LoanORCurrent = settingsReq.LoanORCurrent
			branchSetting.LoanOREnd = settingsReq.LoanOREnd
			branchSetting.LoanORIteration = settingsReq.LoanORIteration
			branchSetting.LoanORUnique = settingsReq.LoanORUnique
			branchSetting.LoanUseDateOR = settingsReq.LoanUseDateOR

			branchSetting.CheckVoucherAllowUserInput = settingsReq.CheckVoucherAllowUserInput
			branchSetting.CheckVoucherPrefix = settingsReq.CheckVoucherPrefix
			branchSetting.CheckVoucherORStart = settingsReq.CheckVoucherORStart
			branchSetting.CheckVoucherORCurrent = settingsReq.CheckVoucherORCurrent
			branchSetting.CheckVoucherOREnd = settingsReq.CheckVoucherOREnd
			branchSetting.CheckVoucherORIteration = settingsReq.CheckVoucherORIteration
			branchSetting.CheckVoucherORUnique = settingsReq.CheckVoucherORUnique
			branchSetting.CheckVoucherUseDateOR = settingsReq.CheckVoucherUseDateOR

			branchSetting.DefaultMemberTypeID = settingsReq.DefaultMemberTypeID
			branchSetting.LoanAppliedEqualToBalance = settingsReq.LoanAppliedEqualToBalance
			branchSetting.AnnualDivisor = settingsReq.AnnualDivisor
			branchSetting.TaxInterest = settingsReq.TaxInterest

			if err := c.core.BranchSettingManager.UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update error",
					Description: fmt.Sprintf("Failed to update branch settings for PUT /branch-settings: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings: " + endTx(err).Error()})
			}
		}

		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to commit transaction for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings for branch ID: %s", userOrg.BranchID),
			Module:      "branch",
		})

		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Title:       "Branch Settings Updated",
			Description: "Branch settings have been successfully updated",
		})

		return ctx.JSON(http.StatusOK, c.core.BranchSettingManager.ToModel(branchSetting))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/branch-settings/currency",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  core.BranchSettingsCurrencyRequest{},
		ResponseType: core.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var settingsReq core.BranchSettingsCurrencyRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings/currency",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}
		if err := c.provider.Service.Validator.Struct(settingsReq); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		branchSetting, err := c.core.BranchSettingManager.FindOne(context, &core.BranchSetting{
			BranchID: *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch settings not found for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch settings not found: " + err.Error()})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		branchSetting.CurrencyID = settingsReq.CurrencyID
		branchSetting.PaidUpSharedCapitalAccountID = &settingsReq.PaidUpSharedCapitalAccountID
		branchSetting.CashOnHandAccountID = &settingsReq.CashOnHandAccountID
		branchSetting.UpdatedAt = time.Now().UTC()

		if err := c.core.BranchSettingManager.UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings currency: " + endTx(err).Error()})
		}

		for _, id := range settingsReq.UnbalancedAccountDeleteIDs {
			if err := c.core.UnbalancedAccountManager.DeleteWithTx(context, tx, id); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to delete unbalanced account: " + err.Error(),
					Module:      "UnbalancedAccount",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete unbalanced account: " + endTx(err).Error()})
			}
		}

		for _, accountReq := range settingsReq.UnbalancedAccount {
			if accountReq.ID != nil {
				existingAccount, err := c.core.UnbalancedAccountManager.GetByID(context, *accountReq.ID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get unbalanced account: " + endTx(err).Error()})
				}

				existingAccount.Name = accountReq.Name
				existingAccount.Description = accountReq.Description
				existingAccount.CurrencyID = accountReq.CurrencyID
				existingAccount.AccountForShortageID = accountReq.AccountForShortageID
				existingAccount.AccountForOverageID = accountReq.AccountForOverageID
				existingAccount.MemberProfileIDForShortage = accountReq.MemberProfileIDForShortage
				existingAccount.MemberProfileIDForOverage = accountReq.MemberProfileIDForOverage

				existingAccount.UpdatedAt = time.Now().UTC()
				existingAccount.UpdatedByID = userOrg.UserID
				if err := c.core.UnbalancedAccountManager.UpdateByIDWithTx(context, tx, existingAccount.ID, existingAccount); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme account: " + endTx(err).Error()})
				}
			} else {
				newUnbalancedAccount := &core.UnbalancedAccount{
					CreatedAt:   time.Now().UTC(),
					CreatedByID: userOrg.UserID,
					UpdatedAt:   time.Now().UTC(),
					UpdatedByID: userOrg.UserID,

					Name:                       accountReq.Name,
					Description:                accountReq.Description,
					CurrencyID:                 accountReq.CurrencyID,
					AccountForShortageID:       accountReq.AccountForShortageID,
					AccountForOverageID:        accountReq.AccountForOverageID,
					MemberProfileIDForShortage: accountReq.MemberProfileIDForShortage,
					MemberProfileIDForOverage:  accountReq.MemberProfileIDForOverage,
				}
				if err := c.core.UnbalancedAccountManager.CreateWithTx(context, tx, newUnbalancedAccount); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create unbalanced account: " + endTx(err).Error()})
				}
			}
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to commit unbalanced account update transaction: " + err.Error(),
				Module:      "UnbalancedAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit unbalanced account update: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings currency for branch settings ID: %s", branchSetting.ID),
			Module:      "branch",
		})

		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Title:            "Branch Settings Updated",
			Description:      "Branch settings have been successfully updated",
			NotificationType: core.NotificationAlert,
		})
		return ctx.JSON(http.StatusOK, c.core.BranchSettingManager.ToModel(branchSetting))
	})
}
