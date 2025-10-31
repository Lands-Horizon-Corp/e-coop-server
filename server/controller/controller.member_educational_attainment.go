package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberEducationalAttainmentController() {
	req := c.provider.Service.Request

	// Create a new educational attainment record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-educational-attainment/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType:  modelcore.MemberEducationalAttainmentRequest{},
		ResponseType: modelcore.MemberEducationalAttainmentResponse{},
		Note:         "Creates a new educational attainment record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.modelcore.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &modelcore.MemberEducationalAttainment{
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

		if err := c.modelcore.MemberEducationalAttainmentManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create educational attainment record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created educational attainment (/member-educational-attainment/member-profile/:member_profile_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberEducationalAttainmentManager.ToModel(value))
	})

	// Update an existing educational attainment record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-educational-attainment/:member_educational_attainment_id",
		Method:       "PUT",
		RequestType:  modelcore.MemberEducationalAttainmentRequest{},
		ResponseType: modelcore.MemberEducationalAttainmentResponse{},
		Note:         "Updates an existing educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := handlers.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), invalid member_educational_attainment_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		req, err := c.modelcore.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), validation error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), user org error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.modelcore.MemberEducationalAttainmentManager.GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), record not found: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
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

		if err := c.modelcore.MemberEducationalAttainmentManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update educational attainment record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated educational attainment (/member-educational-attainment/:member_educational_attainment_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.MemberEducationalAttainmentManager.ToModel(value))
	})

	// Delete an educational attainment record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-educational-attainment/:member_educational_attainment_id",
		Method: "DELETE",
		Note:   "Deletes a member's educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := handlers.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), invalid member_educational_attainment_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		value, err := c.modelcore.MemberEducationalAttainmentManager.GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), record not found: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Educational attainment record not found: " + err.Error()})
		}
		if err := c.modelcore.MemberEducationalAttainmentManager.DeleteByID(context, *memberEducationalAttainmentID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete educational attainment record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted educational attainment (/member-educational-attainment/:member_educational_attainment_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete educational attainment records by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-educational-attainment/bulk-delete",
		Method:      "DELETE",
		RequestType: modelcore.IDSRequest{},
		Note:        "Deletes multiple educational attainment records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), invalid request body.",
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), no IDs provided.",
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			attainmentID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberEducationalAttainment",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}
			value, err := c.modelcore.MemberEducationalAttainmentManager.GetByID(context, attainmentID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), not found: " + rawID,
					Module:      "MemberEducationalAttainment",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Educational attainment record with ID '%s' not found: %s", rawID, err.Error())})
			}
			names += value.SchoolName + ","
			if err := c.modelcore.MemberEducationalAttainmentManager.DeleteByIDWithTx(context, tx, attainmentID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), db error: " + err.Error(),
					Module:      "MemberEducationalAttainment",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete educational attainment record with ID '%s': %s", rawID, err.Error())})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted educational attainments (/member-educational-attainment/bulk-delete): " + names,
			Module:      "MemberEducationalAttainment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
