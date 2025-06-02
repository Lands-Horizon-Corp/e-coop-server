package controller

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) FootstepController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/me",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Getting your own footstep (the logged in user)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/branch",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Get footstep on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/user-organization/:user_organization_id",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Getting Footstep of users that is (member or employee or owner) on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByUserOrganization(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/:footstep_id",
		Method:   "GET",
		Response: "TFootstep",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		footstepId, err := horizon.EngineUUIDParam(ctx, "footstep_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid footstep ID")
		}
		footstep, err := c.model.FootstepManager.GetByIDRaw(context, *footstepId)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

}

func (c *Controller) NotificationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/notification/me",
		Method:   "GET",
		Response: "TNotification[]",
		Note:     "Getting your own notifications (the logged in user)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		notification, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.NotificationManager.ToModels(notification))
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/notification/view",
		Method:  "PUT",
		Request: "string[] ids",
		Note:    "Apply View on multiple notification records",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()

		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}

		tx := c.provider.Service.Database.Client().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		for _, rawID := range reqBody.IDs {
			notificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			notification, err := c.model.NotificationManager.GetByID(context, notificationID)
			if err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Notification with ID %s not found", rawID))
			}

			if notification.IsViewed {
				continue
			}

			notification.IsViewed = true
			if err := c.model.NotificationManager.UpdateByID(context, notification.ID, notification); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		notifications, err := c.model.GetNotificationByUser(context, user.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.NotificationManager.ToModels(notifications))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/notification/:notification_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		notificationId, err := horizon.EngineUUIDParam(ctx, "notification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid notification ID")
		}
		if err := c.model.NotificationManager.DeleteByID(context, *notificationId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
