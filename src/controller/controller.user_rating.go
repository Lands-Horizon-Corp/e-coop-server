package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserRatingController() {
	req := c.provider.Service.Request

	// Get all user ratings made by the specified user (rater)
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/user-rater/:user_id",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings given by the specified user (rater).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userRating, err := c.model.GetUserRater(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRating))
	})

	// Get all user ratings received by the specified user (ratee)
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/user-ratee/:user_id",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings received by the specified user (ratee).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userRating, err := c.model.GetUserRatee(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRating))
	})

	// Get a specific user rating by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/:user_rating_id",
		Method:   "GET",
		Response: "TUserRating",
		Note:     "Returns a specific user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := horizon.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid rating ID")
		}
		userRating, err := c.model.UserRatingManager.GetByID(context, *userRatingId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModel(userRating))
	})

	// Get all user ratings for the current user's branch
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating/branch",
		Method:   "GET",
		Response: "TUserRating[]",
		Note:     "Returns all user ratings in the current user's active branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		userRatig, err := c.model.GetOrganizationBranchRatings(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModels(userRatig))
	})

	// Create a new user rating in the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/user-rating",
		Method:   "POST",
		Response: "TUserRating",
		Request:  "TUserRating",
		Note:     "Creates a new user rating in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.UserRatingManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		userRating := &model.UserRating{
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

		if err := c.model.UserRatingManager.Create(context, userRating); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.UserRatingManager.ToModel(userRating))
	})

	// Delete a user rating by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/user-rating/:user_rating_id",
		Method: "DELETE",
		Note:   "Deletes a user rating by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userRatingId, err := horizon.EngineUUIDParam(ctx, "user_rating_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid rating ID")
		}
		if err := c.model.UserRatingManager.DeleteByID(context, *userRatingId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
