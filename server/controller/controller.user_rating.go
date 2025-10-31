package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) userRatingController(
	req := c.provider.Service.Request

	// Returns all user ratings given by the specified user (rater)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-rating/user-rater/:user_id",
		Method:       "GET",
		ResponseType: modelcore.UserRatingResponse{},
		Note:         "Returns all user ratings given by the specified user (rater).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userRating, err := c.modelcore.GetUserRater(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings given by user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.UserRatingManager.Filtered(context, ctx, userRating))
	})

	// Returns all user ratings received by the specified user (ratee)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-rating/user-ratee/:user_id",
		Method:       "GET",
		ResponseType: modelcore.UserRatingResponse{},
		Note:         "Returns all user ratings received by the specified user (ratee).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userRating, err := c.modelcore.GetUserRatee(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings received by user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.UserRatingManager.Filtered(context, ctx, userRating))
	})

	// Returns a specific user rating by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-rating/:user_rating_id",
		Method:       "GET",
		ResponseType: modelcore.UserRatingResponse{},
		Note:         "Returns a specific user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := handlers.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_rating_id: " + err.Error()})
		}
		userRating, err := c.modelcore.UserRatingManager.GetByID(context, *userRatingId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user rating: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.UserRatingManager.ToModel(userRating))
	})

	// Returns all user ratings in the current user's active branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-rating/branch",
		Method:       "GET",
		ResponseType: modelcore.UserRatingResponse{},
		Note:         "Returns all user ratings in the current user's active branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userRating, err := c.modelcore.UserRatingCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings for branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.UserRatingManager.Filtered(context, ctx, userRating))
	})

	// Creates a new user rating in the current user's branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-rating",
		Method:       "POST",
		ResponseType: modelcore.UserRatingResponse{},
		RequestType:  modelcore.UserRatingRequest{},
		Note:         "Creates a new user rating in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.UserRatingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: validation error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: get user org error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		userRating := &modelcore.UserRating{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			RateeUserID:    req.RateeUserID,
			RaterUserID:    req.RaterUserID,
			Rate:           req.Rate,
			Remark:         req.Remark,
		}

		if err := c.modelcore.UserRatingManager.Create(context, userRating); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: create error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user rating: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created user rating for ratee " + req.RateeUserID.String() + " by rater " + req.RaterUserID.String(),
			Module:      "UserRating",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.UserRatingManager.ToModel(userRating))
	})

	// Deletes a user rating by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-rating/:user_rating_id",
		Method: "DELETE",
		Note:   "Deletes a user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := handlers.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user rating failed: invalid user_rating_id: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_rating_id: " + err.Error()})
		}
		if err := c.modelcore.UserRatingManager.DeleteByID(context, *userRatingId); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user rating failed: delete error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user rating: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted user rating with ID " + userRatingId.String(),
			Module:      "UserRating",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
