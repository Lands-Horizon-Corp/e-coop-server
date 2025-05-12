package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /user-rating
func (c *Controller) UserRatingList(ctx echo.Context) error {
	ratings, err := c.userRating.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/:rating_id
func (c *Controller) UserRatingGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "rating_id")
	if err != nil {
		return err
	}
	rating, err := c.userRating.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModel(rating))
}

// DELETE /user-rating/:rating_id
func (c *Controller) UserRatingDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "rating_id")
	if err != nil {
		return err
	}
	if err := c.userRating.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET /user-rating/user-ratee/:user_ratee_id
func (c *Controller) UserRatingListByRatee(ctx echo.Context) error {
	rateeId, err := horizon.EngineUUIDParam(ctx, "user_ratee_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByUserRatee(*rateeId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/user-rater/:user_rater_id
func (c *Controller) UserRatingListByRater(ctx echo.Context) error {
	raterId, err := horizon.EngineUUIDParam(ctx, "user_rater_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByUserRater(*raterId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/branch/:branch_id
func (c *Controller) UserRatingListByBranch(ctx echo.Context) error {
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByBranch(*branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id
func (c *Controller) UserRatingListByOrganization(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrganization(*orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id/branch/:branch_id
func (c *Controller) UserRatingListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrganizationBranch(*orgId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/branch/:branch_id/ratee/:ratee_user_id
func (c *Controller) UserRatingListByBranchRatee(ctx echo.Context) error {
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	rateeId, err := horizon.EngineUUIDParam(ctx, "ratee_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByBranchRatee(*branchId, *rateeId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/branch/:branch_id/rater/:rater_user_id
func (c *Controller) UserRatingListByBranchRater(ctx echo.Context) error {
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	raterId, err := horizon.EngineUUIDParam(ctx, "rater_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByBranchRater(*branchId, *raterId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id/ratee/:ratee_user_id
func (c *Controller) UserRatingListByOrganizationRatee(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	rateeId, err := horizon.EngineUUIDParam(ctx, "ratee_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrganizationRatee(*orgId, *rateeId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id/rater/:rater_user_id
func (c *Controller) UserRatingListByOrganizationRater(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	raterId, err := horizon.EngineUUIDParam(ctx, "rater_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrganizationRater(*orgId, *raterId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id/branch/:branch_id/ratee/:ratee_user_id
func (c *Controller) UserRatingListByOrgBranchRatee(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	rateeId, err := horizon.EngineUUIDParam(ctx, "ratee_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrgBranchRatee(*orgId, *branchId, *rateeId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}

// GET /user-rating/organization/:organization_id/branch/:branch_id/rater/:rater_user_id
func (c *Controller) UserRatingListByOrgBranchRater(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	raterId, err := horizon.EngineUUIDParam(ctx, "rater_user_id")
	if err != nil {
		return err
	}
	ratings, err := c.userRating.ListByOrgBranchRater(*orgId, *branchId, *raterId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserRatingModels(ratings))
}
