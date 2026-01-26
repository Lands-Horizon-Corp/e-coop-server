package account

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func AccountClassificationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-classification/search",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch.",
		ResponseType: types.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := core.AccountClassificationManager(service).NormalPagination(context, ctx, &types.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, classifications)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-classification",
		Method:       "GET",
		Note:         "Retrieve all account classifications for the current branch (raw).",
		ResponseType: types.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classifications, err := core.AccountClassificationManager(service).Find(context, &types.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account classifications (raw): " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AccountClassificationManager(service).ToModels(classifications))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "GET",
		Note:         "Get an account classification by ID.",
		ResponseType: types.AccountClassificationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := core.AccountClassificationManager(service).GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AccountClassificationManager(service).ToModel(classification))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-classification",
		Method:       "POST",
		Note:         "Create a new account classification for the current branch.",
		ResponseType: types.AccountClassificationResponse{},
		RequestType:  types.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AccountClassificationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for account classification (/account-classification)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		accountClassification := &types.AccountClassification{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := core.AccountClassificationManager(service).Create(context, accountClassification); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Failed to create account classification (/account-classification): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account classification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account classification (/account-classification): " + accountClassification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusCreated, core.AccountClassificationManager(service).ToModel(accountClassification))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-classification/:account_classification_id",
		Method:       "PUT",
		Note:         "Update an account classification by ID.",
		ResponseType: types.AccountClassificationResponse{},
		RequestType:  types.AccountClassificationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AccountClassificationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): validation error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account classification validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := helpers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := core.AccountClassificationManager(service).GetByID(context, *classificationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.AccountClassificationManager(service).UpdateByID(context, classification.ID, classification); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account classification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, core.AccountClassificationManager(service).ToModel(classification))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/account-classification/:account_classification_id",
		Method: "DELETE",
		Note:   "Delete an account classification by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): user org error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for account classification (/account-classification/:account_classification_id)",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		classificationID, err := helpers.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): invalid UUID: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account classification ID: " + err.Error()})
		}
		classification, err := core.AccountClassificationManager(service).GetByID(context, *classificationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): not found: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account classification not found: " + err.Error()})
		}
		if err := core.AccountClassificationManager(service).Delete(context, classification.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to delete account classification (/account-classification/:account_classification_id): db error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account classification: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account classification (/account-classification/:account_classification_id): " + classification.Name,
			Module:      "AccountClassification",
		})
		return ctx.JSON(http.StatusOK, core.AccountClassificationManager(service).ToModel(classification))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/account-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple account classifications by IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | no IDs provided",
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.AccountClassificationManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete account classifications (/account-classification/bulk-delete) | error: " + err.Error(),
				Module:      "AccountClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete account classifications: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted account classifications (/account-classification/bulk-delete)",
			Module:      "AccountClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
