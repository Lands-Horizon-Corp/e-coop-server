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

func (c *Controller) MemberTypeReferenceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/member-type/:member_type_id/search",
		Method:   "GET",
		Response: "TMemberTypeReference[]",
		Note:     "Get all member type references by member_type_id for the current branch",
	}, func(ctx echo.Context) error {
		fmt.Println("DEBUG: Handler entered") // 1
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			fmt.Println("DEBUG: memberTypeID error:", err) // 2
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		if memberTypeID == nil {
			fmt.Println("DEBUG: memberTypeID is nil") // 3
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			fmt.Println("DEBUG: userOrganizationToken error:", err) // 4
			return ctx.NoContent(http.StatusNoContent)
		}
		if user.BranchID == nil {
			fmt.Println("DEBUG: user.BranchID is nil") // 5
			return c.BadRequest(ctx, "Branch ID is required")
		}
		fmt.Println("DEBUG: About to call Find with org:", user.OrganizationID, "branch:", *user.BranchID, "memberTypeID:", *memberTypeID) // 6
		refs, err := c.model.MemberTypeReferenceManager.Find(context, &model.MemberTypeReference{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			MemberTypeID:   *memberTypeID,
		})
		if err != nil {
			fmt.Println("DEBUG: Find error:", err) // 7
			return c.NotFound(ctx, "MemberTypeReference")
		}
		fmt.Println("DEBUG: Success, returning refs") // 8
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.Pagination(context, ctx, refs))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "GET",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}
		ref, err := c.model.MemberTypeReferenceManager.GetByIDRaw(context, *id)
		if err != nil {
			return c.NotFound(ctx, "MemberTypeReference")
		}
		return ctx.JSON(http.StatusOK, ref)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference",
		Method:   "POST",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		ref := &model.MemberTypeReference{
			AccountID:                  req.AccountID,
			MemberTypeID:               req.MemberTypeID,
			MaintainingBalance:         req.MaintainingBalance,
			Description:                req.Description,
			InterestRate:               req.InterestRate,
			MinimumBalance:             req.MinimumBalance,
			Charges:                    req.Charges,
			ActiveMemberMinimumBalance: req.ActiveMemberMinimumBalance,
			ActiveMemberRatio:          req.ActiveMemberRatio,
			OtherInterestOnSavingComputationMinimumBalance: req.OtherInterestOnSavingComputationMinimumBalance,
			OtherInterestOnSavingComputationInterestRate:   req.OtherInterestOnSavingComputationInterestRate,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberTypeReferenceManager.Create(context, ref); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "PUT",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}

		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		ref, err := c.model.MemberTypeReferenceManager.GetByID(context, *id)
		if err != nil {
			return c.NotFound(ctx, "MemberTypeReference")
		}
		ref.AccountID = req.AccountID
		ref.MemberTypeID = req.MemberTypeID
		ref.MaintainingBalance = req.MaintainingBalance
		ref.Description = req.Description
		ref.InterestRate = req.InterestRate
		ref.MinimumBalance = req.MinimumBalance
		ref.Charges = req.Charges
		ref.ActiveMemberMinimumBalance = req.ActiveMemberMinimumBalance
		ref.ActiveMemberRatio = req.ActiveMemberRatio
		ref.OtherInterestOnSavingComputationMinimumBalance = req.OtherInterestOnSavingComputationMinimumBalance
		ref.OtherInterestOnSavingComputationInterestRate = req.OtherInterestOnSavingComputationInterestRate
		ref.UpdatedAt = time.Now().UTC()
		ref.UpdatedByID = user.UserID
		if err := c.model.MemberTypeReferenceManager.UpdateFields(context, ref.ID, ref); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-type-reference/:member_type_reference_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}
		if err := c.model.MemberTypeReferenceManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-type-reference/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member type reference records",
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
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.MemberTypeReferenceManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberTypeReference with ID %s", rawID))
			}
			if err := c.model.MemberTypeReferenceManager.DeleteByIDWithTx(context, tx, id); err != nil {
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
