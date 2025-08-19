package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) AccountClassificationController() {
	req := c.provider.Service.Request

	// GET endpoints (no footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/search",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch.",
		ResponseType: model.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := c.model.AccountClassificationManager.Find(context, &model.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.Pagination(context, ctx, classifications))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch (raw).",
		ResponseType: model.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := c.model.AccountClassificationManager.Find(context, &model.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.Filtered(context, ctx, classifications))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "GET",
		Note:         "Get an account classification by ID.",
		ResponseType: model.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.model.AccountClassificationManager.GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(classification))
	})

	// POST - Create (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification",
		Method:       "POST",
		Note:         "Create a new account classification for the current branch.",
		ResponseType: model.AccountClassificationResponse{},
		RequestType:  model.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for account classification (/account-classification)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountClassification := &model.AccountClassification{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.model.AccountClassificationManager.Create(context, accountClassification); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account classification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account classification (/account-classification): " + accountClassification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusCreated, c.model.AccountClassificationManager.ToModel(accountClassification))
	})

	// PUT - Update (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "PUT",
		Note:         "Update an account classification by ID.",
		ResponseType: model.AccountClassificationResponse{},
		RequestType:  model.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.model.AccountClassificationManager.GetByID(context, *classificationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		if err := c.model.AccountClassificationManager.UpdateFields(context, classification.ID, classification); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account classification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(classification))
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := handlers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := c.model.AccountClassificationManager.GetByID(context, *classificationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): not found: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		if err := c.model.AccountClassificationManager.DeleteByID(context, classification.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account classification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(classification))
	})

	// BULK DELETE (with footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/account-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account classifications by IDs.",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): invalid request body: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): no IDs provided",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): begin tx error: " + tx.Error.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): invalid UUID: " + rawID + " - " + err.Error(),
					Module:      "AccountClassification",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid UUID: " + rawID + " - " + err.Error()})
			}
			if _, err := c.model.AccountClassificationManager.GetByID(context, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): not found: " + rawID + " - " + err.Error(),
					Module:      "AccountClassification",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification with ID " + rawID + " not found: " + err.Error()})
			}
			if err := c.model.AccountClassificationManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): delete error: " + rawID + " - " + err.Error(),
					Module:      "AccountClassification",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account classification with ID " + rawID + ": " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete): commit error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity: "bulk-delete-success",
			Description: "Bulk deleted account classifications (/account-classification/bulk-delete): IDs=" + func() string {
				b := ""
				for _, id := range reqBody.IDs {
					b += id + ","
				}
				return b
			}(),
			Module: "AccountClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
