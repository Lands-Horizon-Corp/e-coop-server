package member_profile

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

func MemberEducationalAttainmentController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-educational-attainment/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType: types.MemberEducationalAttainmentRequest{},
		ResponseType: types.MemberEducationalAttainmentResponse{},
		Note:         "Creates a new educational attainment record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberEducationalAttainmentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &types.MemberEducationalAttainment{
			MemberProfileID:       *memberProfileID,
			SchoolName:            req.SchoolName,
			SchoolYear:            req.SchoolYear,
			ProgramCourse:         req.ProgramCourse,
			EducationalAttainment: req.EducationalAttainment,
			Description:           req.Description,
			CreatedAt:             time.Now().UTC(),
			CreatedByID:           userOrg.UserID,
			UpdatedAt:             time.Now().UTC(),
			UpdatedByID:           userOrg.UserID,
			BranchID:              *userOrg.BranchID,
			OrganizationID:        userOrg.OrganizationID,
		}

		if err := core.MemberEducationalAttainmentManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create educational attainment failed (/member-educational-attainment/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create educational attainment record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created educational attainment (/member-educational-attainment/member-profile/:member_profile_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})

		return ctx.JSON(http.StatusOK, core.MemberEducationalAttainmentManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-educational-attainment/:member_educational_attainment_id",
		Method:       "PUT",
		RequestType: types.MemberEducationalAttainmentRequest{},
		ResponseType: types.MemberEducationalAttainmentResponse{},
		Note:         "Updates an existing educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := helpers.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), invalid member_educational_attainment_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		req, err := core.MemberEducationalAttainmentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), validation error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), user org error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberEducationalAttainmentManager(service).GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), record not found: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Educational attainment record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.SchoolName = req.SchoolName
		value.SchoolYear = req.SchoolYear
		value.ProgramCourse = req.ProgramCourse
		value.EducationalAttainment = req.EducationalAttainment
		value.Description = req.Description

		if err := core.MemberEducationalAttainmentManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update educational attainment record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated educational attainment (/member-educational-attainment/:member_educational_attainment_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})
		return ctx.JSON(http.StatusOK, core.MemberEducationalAttainmentManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-educational-attainment/:member_educational_attainment_id",
		Method: "DELETE",
		Note:   "Deletes a member's educational attainment record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := helpers.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), invalid member_educational_attainment_id: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_educational_attainment_id: " + err.Error()})
		}
		value, err := core.MemberEducationalAttainmentManager(service).GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), record not found: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Educational attainment record not found: " + err.Error()})
		}
		if err := core.MemberEducationalAttainmentManager(service).Delete(context, *memberEducationalAttainmentID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete educational attainment failed (/member-educational-attainment/:member_educational_attainment_id), db error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete educational attainment record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted educational attainment (/member-educational-attainment/:member_educational_attainment_id): " + value.SchoolName,
			Module:      "MemberEducationalAttainment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-educational-attainment/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple educational attainment records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete) | no IDs provided",
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberEducationalAttainmentManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete educational attainment failed (/member-educational-attainment/bulk-delete) | error: " + err.Error(),
				Module:      "MemberEducationalAttainment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete educational attainment records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted educational attainments (/member-educational-attainment/bulk-delete)",
			Module:      "MemberEducationalAttainment",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
