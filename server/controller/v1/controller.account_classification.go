package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) accountClassificationController() {
	req := c.provider.Service.WebRequest

	// GET endpoints (no footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/search",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch.",
		ResponseType: core.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := c.core.AccountClassificationManager.PaginationWithFields(context, ctx, &core.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, classifications)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch (raw).",
		ResponseType: core.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := c.core.AccountClassificationManager.Find(context, &core.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountClassificationManager.ToModels(classifications))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "GET",
		Note:         "Get an account classification by ID.",
		ResponseType: core.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.core.AccountClassificationManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountClassificationManager.ToModel(classification))
	})

	// POST - Create (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification",
		Method:       "POST",
		Note:         "Create a new account classification for the current branch.",
		ResponseType: core.AccountClassificationResponse{},
		RequestType:  core.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for account classification (/account-classification)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountClassification := &core.AccountClassification{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.core.AccountClassificationManager.Create(context, accountClassification); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account classification: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account classification (/account-classification): " + accountClassification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusCreated, c.core.AccountClassificationManager.ToModel(accountClassification))
	})

	// PUT - Update (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "PUT",
		Note:         "Update an account classification by ID.",
		ResponseType: core.AccountClassificationResponse{},
		RequestType:  core.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.core.AccountClassificationManager.GetByID(context, *classificationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): not found: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		classification.UpdatedByID = userOrg.UserID
		classification.UpdatedAt = time.Now().UTC()
		classification.Name = req.Name
		classification.Description = req.Description
		classification.BranchID = *userOrg.BranchID
		classification.OrganizationID = userOrg.OrganizationID
		if err := c.core.AccountClassificationManager.UpdateByID(context, classification.ID, classification); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account classification: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountClassificationManager.ToModel(classification))
	})

	// DELETE (single) - with footstep
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/account-classification/:account_classification_id",
		Method: "DELETE",
		Note:   "Delete an account classification by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.core.AccountClassificationManager.GetByID(context, *classificationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): not found: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		if err := c.core.AccountClassificationManager.Delete(context, classification.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account classification: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountClassificationManager.ToModel(classification))
	})

	// BULK DELETE (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/account-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account classifications by IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | no IDs provided",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}

		if err := c.core.AccountClassificationManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete account classifications: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted account classifications (/account-classification/bulk-delete)",
			Module:      "AccountClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
