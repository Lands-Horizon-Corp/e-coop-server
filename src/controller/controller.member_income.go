package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberIncomeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-income/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberIncome",
		Response: "TMemberIncome",
		Note:     "Create a new income record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberIncome{
			MemberProfileID: *memberProfileID,

			MediaID:     req.MediaID,
			Name:        req.Name,
			Amount:      req.Amount,
			ReleaseDate: req.ReleaseDate,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberIncomeManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-income/:member_income_id",
		Method:   "PUT",
		Request:  "TMemberIncome",
		Response: "TMemberIncome",
		Note:     "Update an existing income record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := horizon.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member income ID")
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberIncomeManager.GetByID(context, *memberIncomeID)
		if err != nil {
			return c.NotFound(ctx, "MemberIncome")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MediaID = req.MediaID
		value.Name = req.Name
		value.Amount = req.Amount
		value.ReleaseDate = req.ReleaseDate
		if err := c.model.MemberIncomeManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-income/:member_income_id",
		Method: "DELETE",
		Note:   "Delete a member's income record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := horizon.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member income ID")
		}
		if err := c.model.MemberIncomeManager.DeleteByID(context, *memberIncomeID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
