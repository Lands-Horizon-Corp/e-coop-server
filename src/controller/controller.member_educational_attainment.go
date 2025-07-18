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

	// Create a new educational attainment record for a member profile
	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Creates a new educational attainment record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create educational attainment record: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	// Update an existing educational attainment record by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/:member_educational_attainment_id",
		Method:   "PUT",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Updates an existing educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberEducationalAttainmentManager.GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Educational attainment record not found: " + err.Error()})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update educational attainment record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	// Delete an educational attainment record by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-educational-attainment/:member_educational_attainment_id",
		Method: "DELETE",
		Note:   "Deletes a member's educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		if err := c.model.MemberEducationalAttainmentManager.DeleteByID(context, *memberEducationalAttainmentID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete educational attainment record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete educational attainment records by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/member-educational-attainment/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple educational attainment records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			attainmentID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}
			if _, err := c.model.MemberEducationalAttainmentManager.GetByID(context, attainmentID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Educational attainment record with ID '%s' not found: %s", rawID, err.Error())})
			}
			if err := c.model.MemberEducationalAttainmentManager.DeleteByIDWithTx(context, tx, attainmentID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete educational attainment record with ID '%s': %s", rawID, err.Error())})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
