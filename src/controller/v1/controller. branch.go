package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BranchController registers routes related to branch management.
func (c *Controller) BranchController() {
	req := c.provider.Service.Request

	// GET /branch: List all branches or filter by user's organization from JWT if available.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch",
		Method:       "GET",
		Note:         "Returns all branches if unauthenticated; otherwise, returns branches filtered by the user's organization from JWT.",
		ResponseType: model_core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg == nil {
			branches, err := c.model_core.BranchManager.List(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve branches: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := c.model_core.GetBranchesByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model_core.BranchManager.Filtered(context, ctx, branches))
	})

	// GET /branch/organization/:organization_id: List branches by organization ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "GET",
		Note:         "Returns all branches belonging to the specified organization.",
		ResponseType: model_core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		branches, err := c.model_core.GetBranchesByOrganization(context, *orgId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.BranchManager.Filtered(context, ctx, branches))
	})

	// POST /branch/organization/:organization_id: Create a branch for an organization.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "POST",
		Note:         "Creates a new branch for the given organization. If the user already has a branch, a new user organization is created; otherwise, the user's current user organization is updated with the new branch.",
		Private:      true,
		RequestType:  model_core.BranchRequest{},
		ResponseType: model_core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model_core.BranchManager.Validate(ctx)
		if err != nil {
			// Footstep for create branch error
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate branch data for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		organizationId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Invalid organization ID for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "User authentication required for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		userOrganization, err := c.model_core.UserOrganizationManager.FindOne(context, &model_core.UserOrganization{
			UserID:         user.ID,
			OrganizationID: *organizationId,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("User organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found " + err.Error()})
		}
		if userOrganization.UserType != model_core.UserOrganizationTypeOwner {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Only organization owners can create branches for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners can create branches "})
		}

		organization, err := c.model_core.OrganizationManager.GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found " + err.Error()})
		}

		branchCount, err := c.model_core.GetBranchesByOrganizationCount(context, organization.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed branch count for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not count organization branches: " + err.Error()})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Branch limit reached for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for the current subscription plan "})
		}

		branch := &model_core.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: userOrganization.OrganizationID,
			MediaID:        req.MediaID,
			Type:           req.Type,
			Name:           req.Name,
			Email:          req.Email,
			Description:    req.Description,
			CountryCode:    req.CountryCode,
			ContactNumber:  req.ContactNumber,
			Address:        req.Address,
			Province:       req.Province,
			City:           req.City,
			Region:         req.Region,
			Barangay:       req.Barangay,
			PostalCode:     req.PostalCode,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			IsMainBranch:   req.IsMainBranch,
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to start DB transaction for POST /branch/organization/:organization_id: %v", tx.Error),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		if err := c.model_core.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}
		currency, err := c.model_core.CurrencyFindByAlpha2(context, branch.CountryCode)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to find currency for branch country code for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find currency for branch country code: " + err.Error()})
		}
		// Create default branch settings for the new branch
		branchSetting := &model_core.BranchSetting{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			BranchID:  branch.ID,

			// Withdraw Settings
			WithdrawAllowUserInput: true,
			WithdrawPrefix:         "WD",
			WithdrawORStart:        1,
			WithdrawORCurrent:      1,
			WithdrawOREnd:          999999,
			WithdrawORIteration:    1,
			WithdrawORUnique:       true,
			WithdrawUseDateOR:      false,

			// Deposit Settings
			DepositAllowUserInput: true,
			DepositPrefix:         "DP",
			DepositORStart:        1,
			DepositORCurrent:      1,
			DepositOREnd:          999999,
			DepositORIteration:    1,
			DepositORUnique:       true,
			DepositUseDateOR:      false,

			// Loan Settings
			LoanAllowUserInput: true,
			LoanPrefix:         "LN",
			LoanORStart:        1,
			LoanORCurrent:      1,
			LoanOREnd:          999999,
			LoanORIteration:    1,
			LoanORUnique:       true,
			LoanUseDateOR:      false,

			// Check Voucher Settings
			CheckVoucherAllowUserInput: true,
			CheckVoucherPrefix:         "CV",
			CheckVoucherORStart:        1,
			CheckVoucherORCurrent:      1,
			CheckVoucherOREnd:          999999,
			CheckVoucherORIteration:    1,
			CheckVoucherORUnique:       true,
			CheckVoucherUseDateOR:      false,

			// Default Member Type - can be set later
			DefaultMemberTypeID:       nil,
			LoanAppliedEqualToBalance: true,
			CurrencyID:                currency.ID,
		}

		if err := c.model_core.BranchSettingManager.CreateWithTx(context, tx, branchSetting); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch settings for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + err.Error()})
		}

		if userOrganization.BranchID == nil {
			// Assign branch to existing user organization
			userOrganization.BranchID = &branch.ID
			userOrganization.UpdatedAt = time.Now().UTC()
			userOrganization.UpdatedByID = user.ID

			if err := c.model_core.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to update user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + err.Error()})
			}
		} else {
			// Create new user organization for this branch
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to generate developer key for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
			}

			newUserOrg := &model_core.UserOrganization{
				CreatedAt:                time.Now().UTC(),
				CreatedByID:              user.ID,
				UpdatedAt:                time.Now().UTC(),
				UpdatedByID:              user.ID,
				OrganizationID:           userOrganization.OrganizationID,
				BranchID:                 &branch.ID,
				UserID:                   user.ID,
				UserType:                 model_core.UserOrganizationTypeOwner,
				ApplicationStatus:        "accepted",
				DeveloperSecretKey:       developerKey + uuid.NewString() + "-horizon",
				PermissionName:           string(model_core.UserOrganizationTypeOwner),
				Permissions:              []string{},
				UserSettingStartOR:       0,
				UserSettingEndOR:         1000,
				UserSettingUsedOR:        0,
				UserSettingStartVoucher:  0,
				UserSettingEndVoucher:    5,
				UserSettingUsedVoucher:   0,
				UserSettingNumberPadding: 7,
			}

			if err := c.model_core.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to create new user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to commit transaction for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		// Event notification
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Create: %s", branch.Name),
			Description: fmt.Sprintf("Created a new branch: %s", branch.Name),
		})

		// Footstep for create branch success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.model_core.BranchManager.ToModel(branch))
	})

	// PUT /branch/:branch_id: Update an existing branch (only by owner).
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/:branch_id",
		Method:       "PUT",
		Note:         "Updates branch information for the specified branch. Only allowed for the owner of the branch.",
		Private:      true,
		RequestType:  model_core.BranchRequest{},
		ResponseType: model_core.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model_core.BranchManager.Validate(ctx)
		if err != nil {
			// Footstep for update error
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch data for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User authentication required for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required " + err.Error()})
		}

		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid branch id for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}

		userOrg, err := c.model_core.UserOrganizationManager.FindOne(context, &model_core.UserOrganization{
			UserID:   user.ID,
			BranchID: branchId,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("User organization not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization for this branch not found: " + err.Error()})
		}
		if userOrg.UserType != model_core.UserOrganizationTypeOwner {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "Only the branch owner can update branch for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can update branch information "})
		}

		branch, err := c.model_core.BranchManager.GetByID(context, *branchId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		// Update branch fields
		branch.UpdatedAt = time.Now().UTC()
		branch.UpdatedByID = user.ID
		branch.MediaID = req.MediaID
		branch.Type = req.Type
		branch.Name = req.Name
		branch.Email = req.Email
		branch.Description = req.Description
		branch.CountryCode = req.CountryCode
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

		if err := c.model_core.BranchManager.UpdateFields(context, branch.ID, branch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Update: %s", branch.Name),
			Description: fmt.Sprintf("Updated branch: %s", branch.Name),
		})

		// Footstep for update branch success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.model_core.BranchManager.ToModel(branch))
	})

	// DELETE /branch/:branch_id: Delete a branch (owner only, if fewer than 3 members).
	req.RegisterRoute(handlers.Route{
		Route:   "/api/v1/branch/:branch_id",
		Method:  "DELETE",
		Note:    "Deletes the specified branch if the user is the owner and there are less than 3 members in the branch.",
		Private: true,
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid branch ID for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "User authentication required for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required "})
		}
		branch, err := c.model_core.BranchManager.GetByID(context, *branchId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Branch not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		userOrganization, err := c.model_core.UserOrganizationManager.FindOne(context, &model_core.UserOrganization{
			UserID:         user.ID,
			BranchID:       branchId,
			OrganizationID: branch.OrganizationID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("User organization not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if userOrganization.UserType != model_core.UserOrganizationTypeOwner {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Only the branch owner can delete this branch for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can delete this branch"})
		}
		count, err := c.model_core.CountUserOrganizationPerBranch(context, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Could not check branch membership for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not check branch membership: " + err.Error()})
		}
		if count > 2 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Cannot delete branch with more than 2 members for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete branch with more than 2 members"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to start DB transaction for DELETE /branch/:branch_id: %v", tx.Error),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		if err := c.model_core.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete branch for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
		}
		if err := c.model_core.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete user organization for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to commit transaction for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Delete: %s", branch.Name),
			Description: fmt.Sprintf("Deleted branch: %s", branch.Name),
		})
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch-settings",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  model_core.BranchSettingRequest{},
		ResponseType: model_core.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the branch settings request
		var settingsReq model_core.BranchSettingRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(settingsReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// Get user's current branch
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}

		// Check if user has permission to update branch settings
		if userOrg.UserType != model_core.UserOrganizationTypeOwner && userOrg.UserType != model_core.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "Insufficient permissions to update branch settings for PUT /branch-settings",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions to update branch settings"})
		}

		// Get existing branch settings or create new one
		var branchSetting *model_core.BranchSetting
		branchSetting, err = c.model_core.BranchSettingManager.FindOne(context, &model_core.BranchSetting{
			BranchID: *userOrg.BranchID,
		})

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to start DB transaction for PUT /branch-settings: %v", tx.Error),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		if err != nil {
			// Create new branch settings if they don't exist
			branchSetting = &model_core.BranchSetting{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				BranchID:  *userOrg.BranchID,

				// Withdraw Settings
				WithdrawAllowUserInput: settingsReq.WithdrawAllowUserInput,
				WithdrawPrefix:         settingsReq.WithdrawPrefix,
				WithdrawORStart:        settingsReq.WithdrawORStart,
				WithdrawORCurrent:      settingsReq.WithdrawORCurrent,
				WithdrawOREnd:          settingsReq.WithdrawOREnd,
				WithdrawORIteration:    settingsReq.WithdrawORIteration,
				WithdrawORUnique:       settingsReq.WithdrawORUnique,
				WithdrawUseDateOR:      settingsReq.WithdrawUseDateOR,

				// Deposit Settings
				DepositAllowUserInput: settingsReq.DepositAllowUserInput,
				DepositPrefix:         settingsReq.DepositPrefix,
				DepositORStart:        settingsReq.DepositORStart,
				DepositORCurrent:      settingsReq.DepositORCurrent,
				DepositOREnd:          settingsReq.DepositOREnd,
				DepositORIteration:    settingsReq.DepositORIteration,
				DepositORUnique:       settingsReq.DepositORUnique,
				DepositUseDateOR:      settingsReq.DepositUseDateOR,

				// Loan Settings
				LoanAllowUserInput: settingsReq.LoanAllowUserInput,
				LoanPrefix:         settingsReq.LoanPrefix,
				LoanORStart:        settingsReq.LoanORStart,
				LoanORCurrent:      settingsReq.LoanORCurrent,
				LoanOREnd:          settingsReq.LoanOREnd,
				LoanORIteration:    settingsReq.LoanORIteration,
				LoanORUnique:       settingsReq.LoanORUnique,
				LoanUseDateOR:      settingsReq.LoanUseDateOR,

				// Check Voucher Settings
				CheckVoucherAllowUserInput: settingsReq.CheckVoucherAllowUserInput,
				CheckVoucherPrefix:         settingsReq.CheckVoucherPrefix,
				CheckVoucherORStart:        settingsReq.CheckVoucherORStart,
				CheckVoucherORCurrent:      settingsReq.CheckVoucherORCurrent,
				CheckVoucherOREnd:          settingsReq.CheckVoucherOREnd,
				CheckVoucherORIteration:    settingsReq.CheckVoucherORIteration,
				CheckVoucherORUnique:       settingsReq.CheckVoucherORUnique,
				CheckVoucherUseDateOR:      settingsReq.CheckVoucherUseDateOR,

				// Default Member Type
				DefaultMemberTypeID: settingsReq.DefaultMemberTypeID,
			}

			if err := c.model_core.BranchSettingManager.CreateWithTx(context, tx, branchSetting); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update error",
					Description: fmt.Sprintf("Failed to create branch settings for PUT /branch-settings: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + err.Error()})
			}
		} else {
			// Update existing branch settings
			branchSetting.UpdatedAt = time.Now().UTC()

			// Withdraw Settings
			branchSetting.WithdrawAllowUserInput = settingsReq.WithdrawAllowUserInput
			branchSetting.WithdrawPrefix = settingsReq.WithdrawPrefix
			branchSetting.WithdrawORStart = settingsReq.WithdrawORStart
			branchSetting.WithdrawORCurrent = settingsReq.WithdrawORCurrent
			branchSetting.WithdrawOREnd = settingsReq.WithdrawOREnd
			branchSetting.WithdrawORIteration = settingsReq.WithdrawORIteration
			branchSetting.WithdrawORUnique = settingsReq.WithdrawORUnique
			branchSetting.WithdrawUseDateOR = settingsReq.WithdrawUseDateOR

			// Deposit Settings
			branchSetting.DepositAllowUserInput = settingsReq.DepositAllowUserInput
			branchSetting.DepositPrefix = settingsReq.DepositPrefix
			branchSetting.DepositORStart = settingsReq.DepositORStart
			branchSetting.DepositORCurrent = settingsReq.DepositORCurrent
			branchSetting.DepositOREnd = settingsReq.DepositOREnd
			branchSetting.DepositORIteration = settingsReq.DepositORIteration
			branchSetting.DepositORUnique = settingsReq.DepositORUnique
			branchSetting.DepositUseDateOR = settingsReq.DepositUseDateOR

			// Loan Settings
			branchSetting.LoanAllowUserInput = settingsReq.LoanAllowUserInput
			branchSetting.LoanPrefix = settingsReq.LoanPrefix
			branchSetting.LoanORStart = settingsReq.LoanORStart
			branchSetting.LoanORCurrent = settingsReq.LoanORCurrent
			branchSetting.LoanOREnd = settingsReq.LoanOREnd
			branchSetting.LoanORIteration = settingsReq.LoanORIteration
			branchSetting.LoanORUnique = settingsReq.LoanORUnique
			branchSetting.LoanUseDateOR = settingsReq.LoanUseDateOR

			// Check Voucher Settings
			branchSetting.CheckVoucherAllowUserInput = settingsReq.CheckVoucherAllowUserInput
			branchSetting.CheckVoucherPrefix = settingsReq.CheckVoucherPrefix
			branchSetting.CheckVoucherORStart = settingsReq.CheckVoucherORStart
			branchSetting.CheckVoucherORCurrent = settingsReq.CheckVoucherORCurrent
			branchSetting.CheckVoucherOREnd = settingsReq.CheckVoucherOREnd
			branchSetting.CheckVoucherORIteration = settingsReq.CheckVoucherORIteration
			branchSetting.CheckVoucherORUnique = settingsReq.CheckVoucherORUnique
			branchSetting.CheckVoucherUseDateOR = settingsReq.CheckVoucherUseDateOR

			// Default Member Type
			branchSetting.DefaultMemberTypeID = settingsReq.DefaultMemberTypeID
			branchSetting.LoanAppliedEqualToBalance = settingsReq.LoanAppliedEqualToBalance

			if err := c.model_core.BranchSettingManager.UpdateFieldsWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update error",
					Description: fmt.Sprintf("Failed to update branch settings for PUT /branch-settings: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to commit transaction for PUT /branch-settings: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		// Log success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings for branch ID: %s", userOrg.BranchID),
			Module:      "branch",
		})

		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       "Branch Settings Updated",
			Description: "Branch settings have been successfully updated",
		})

		return ctx.JSON(http.StatusOK, c.model_core.BranchSettingManager.ToModel(branchSetting))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch-settings/currency",
		Method:       "PUT",
		Note:         "Updates branch settings for the current user's branch.",
		RequestType:  model_core.BranchSettingsCurrencyRequest{},
		ResponseType: model_core.BranchSettingResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate the branch settings currency request
		var settingsReq model_core.BranchSettingsCurrencyRequest
		if err := ctx.Bind(&settingsReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to bind branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User not assigned to a branch for PUT /branch-settings/currency",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not assigned to a branch"})
		}
		if err := c.provider.Service.Validator.Struct(settingsReq); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		branchSetting, err := c.model_core.BranchSettingManager.FindOne(context, &model_core.BranchSetting{
			BranchID: *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch settings not found for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch settings not found: " + err.Error()})
		}
		branchSetting.CurrencyID = settingsReq.CurrencyID
		branchSetting.PaidUpSharedCapitalAccountID = &settingsReq.PaidUpSharedCapitalAccountID
		branchSetting.CashOnHandAccountID = &settingsReq.CashOnHandAccountID

		branchSetting.UpdatedAt = time.Now().UTC()

		if err := c.model_core.BranchSettingManager.UpdateFields(context, branchSetting.ID, branchSetting); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch settings currency for PUT /branch-settings/currency: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings currency: " + err.Error()})
		}

		// Log success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch settings currency for branch settings ID: %s", branchSetting.ID),
			Module:      "branch",
		})

		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       "Branch Settings Currency Updated",
			Description: "Branch settings currency have been successfully updated",
		})

		return ctx.JSON(http.StatusOK, c.model_core.BranchSettingManager.ToModel(branchSetting))
	})
}
