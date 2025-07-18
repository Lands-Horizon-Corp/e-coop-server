package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberEducationalAttainmentController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Create a new educational attainment record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberEducationalAttainment{
			MemberProfileID:       *memberProfileID,
			SchoolName:            req.SchoolName,
			SchoolYear:            req.SchoolYear,
			ProgramCourse:         req.ProgramCourse,
			EducationalAttainment: req.EducationalAttainment,
			Description:           req.Description,
			CreatedAt:             time.Now().UTC(),
			CreatedByID:           user.UserID,
			UpdatedAt:             time.Now().UTC(),
			UpdatedByID:           user.UserID,
			BranchID:              *user.BranchID,
			OrganizationID:        user.OrganizationID,
		}

		if err := c.model.MemberEducationalAttainmentManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/:member_educational_attainment_id",
		Method:   "PUT",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Update an existing educational attainment record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberEducationalAttainmentManager.GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			return c.NotFound(ctx, "MemberEducationalAttainment")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.SchoolName = req.SchoolName
		value.SchoolYear = req.SchoolYear
		value.ProgramCourse = req.ProgramCourse
		value.EducationalAttainment = req.EducationalAttainment
		value.Description = req.Description
		if err := c.model.MemberEducationalAttainmentManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-educational-attainment/:member_educational_attainment_id",
		Method: "DELETE",
		Note:   "Delete a member's educational attainment record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		if err := c.model.MemberEducationalAttainmentManager.DeleteByID(context, *memberEducationalAttainmentID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-educational-attainment/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member educational attainment records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			bankID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.MemberEducationalAttainmentManager.GetByID(context, bankID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberEducationalAttainment with ID %s", rawID))
			}
			if err := c.model.MemberEducationalAttainmentManager.DeleteByIDWithTx(context, tx, bankID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
