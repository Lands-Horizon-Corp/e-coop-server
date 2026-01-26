package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberIncomeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-income/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: types.MemberIncomeResponse{},
		RequestType:  types.MemberIncomeRequest{},
		Note:         "Creates a new income record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberIncomeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &types.MemberIncome{
			MemberProfileID: *memberProfileID,
			MediaID:         req.MediaID,
			Name:            req.Name,
			Source:          req.Source,
			Amount:          req.Amount,
			ReleaseDate:     req.ReleaseDate,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}

		if err := core.MemberIncomeManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member income: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member income (/member-income/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberIncome",
		})

		return ctx.JSON(http.StatusOK, core.MemberIncomeManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-income/:member_income_id",
		Method:       "PUT",
		ResponseType: types.MemberIncomeResponse{},
		RequestType:  types.MemberIncomeRequest{},
		Note:         "Updates an existing income record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := helpers.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), invalid member_income_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_income_id: " + err.Error()})
		}
		req, err := core.MemberIncomeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), validation error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), user org error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberIncomeManager(service).GetByID(context, *memberIncomeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), record not found: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member income not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.MediaID = req.MediaID
		value.Name = req.Name
		value.Source = req.Source
		value.Amount = req.Amount
		value.ReleaseDate = req.ReleaseDate

		if err := core.MemberIncomeManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member income: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member income (/member-income/:member_income_id): " + value.Name,
			Module:      "MemberIncome",
		})
		return ctx.JSON(http.StatusOK, core.MemberIncomeManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-income/:member_income_id",
		Method: "DELETE",
		Note:   "Deletes a member's income record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := helpers.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), invalid member_income_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_income_id: " + err.Error()})
		}
		value, err := core.MemberIncomeManager(service).GetByID(context, *memberIncomeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), record not found: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member income not found: " + err.Error()})
		}
		if err := core.MemberIncomeManager(service).Delete(context, *memberIncomeID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member income: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member income (/member-income/:member_income_id): " + value.Name,
			Module:      "MemberIncome",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
