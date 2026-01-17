package user

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func UserRatingController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user-rating/user-rater/:user_id",
		Method:       "GET",
		ResponseType: core.UserRatingResponse{},
		Note:         "Returns all user ratings given by the specified user (rater).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userRating, err := core.GetUserRater(context, service, *userID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings given by user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.UserRatingManager(service).ToModels(userRating))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user-rating/user-ratee/:user_id",
		Method:       "GET",
		ResponseType: core.UserRatingResponse{},
		Note:         "Returns all user ratings received by the specified user (ratee).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userRating, err := core.GetUserRatee(context, service, *userID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings received by user: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.UserRatingManager(service).ToModels(userRating))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user-rating/:user_rating_id",
		Method:       "GET",
		ResponseType: core.UserRatingResponse{},
		Note:         "Returns a specific user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingID, err := helpers.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_rating_id: " + err.Error()})
		}
		userRating, err := core.UserRatingManager(service).GetByID(context, *userRatingID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user rating: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.UserRatingManager(service).ToModel(userRating))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user-rating/branch",
		Method:       "GET",
		ResponseType: core.UserRatingResponse{},
		Note:         "Returns all user ratings in the current user's active branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userRating, err := core.UserRatingCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ratings for branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.UserRatingManager(service).ToModels(userRating))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/user-rating",
		Method:       "POST",
		ResponseType: core.UserRatingResponse{},
		RequestType:  core.UserRatingRequest{},
		Note:         "Creates a new user rating in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.UserRatingManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: validation error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: get user org error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		userRating := &types.UserRating{
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

		if err := core.UserRatingManager(service).Create(context, userRating); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user rating failed: create error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user rating: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created user rating for ratee " + req.RateeUserID.String() + " by rater " + req.RaterUserID.String(),
			Module:      "UserRating",
		})

		return ctx.JSON(http.StatusOK, core.UserRatingManager(service).ToModel(userRating))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/user-rating/:user_rating_id",
		Method: "DELETE",
		Note:   "Deletes a user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingID, err := helpers.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user rating failed: invalid user_rating_id: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_rating_id: " + err.Error()})
		}
		if err := core.UserRatingManager(service).Delete(context, *userRatingID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user rating failed: delete error: " + err.Error(),
				Module:      "UserRating",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user rating: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted user rating with ID " + userRatingID.String(),
			Module:      "UserRating",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
