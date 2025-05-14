package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-profile
func (c *Controller) MemberProfileList(ctx echo.Context) error {
	member_profile, err := c.memberProfile.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModels(member_profile))
}

// GET /member-profile/:member_profile_id
func (c *Controller) MemberProfileGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
	if err != nil {
		return err
	}
	member_profile, err := c.memberProfile.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModel(member_profile))
}

// POST /member-profile
func (c *Controller) MemberProfileCreate(ctx echo.Context) error {
	req, err := c.model.MemberProfileValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	memProfilemodel := &model.MemberProfile{
		CreatedAt:              time.Now().UTC(),
		CreatedByID:            user.UserID,
		UpdatedAt:              time.Now().UTC(),
		UpdatedByID:            user.UserID,
		BranchID:               *user.BranchID,
		OrganizationID:         user.OrganizationID,
		MediaID:                req.MediaID,
		SignatureMediaID:       req.SignatureMediaID,
		MemberCenterID:         req.MemberCenterID,
		MemberClassificationID: req.MemberClassificationID,
		MemberGenderID:         req.MemberGenderID,
		MemberGroupID:          req.MemberGroupID,
		MemberOccupationID:     req.MemberOccupationID,
		IsClosed:               req.IsClosed,
		IsMutualFundMember:     req.IsMutualFundMember,
		IsMicroFinanceMember:   req.IsMicroFinanceMember,
		FirstName:              req.FirstName,
		MiddleName:             req.MiddleName,
		LastName:               req.LastName,
		FullName:               req.FullName,
		Suffix:                 req.Suffix,
		Birthdate:              req.Birthdate,
		Status:                 req.Status,
		Description:            req.Description,
		Notes:                  req.Notes,
		ContactNumber:          req.ContactNumber,
		OldReferenceID:         req.OldReferenceID,
		Passbook:               req.Passbook,
		Occupation:             req.Occupation,
		BusinessAddress:        req.BusinessAddress,
		BusinessContactNumber:  req.BusinessContactNumber,
		CivilStatus:            req.CivilStatus,
	}
	if err := c.memberProfile.Manager.Create(memProfilemodel); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.memberVerification.Manager.Create(&model.MemberVerification{
		CreatedAt:       time.Now().UTC(),
		CreatedByID:     user.UserID,
		UpdatedAt:       time.Now().UTC(),
		UpdatedByID:     user.UserID,
		BranchID:        *user.BranchID,
		OrganizationID:  user.OrganizationID,
		MemberProfileID: memProfilemodel.ID,
	})
	return ctx.JSON(http.StatusCreated, c.model.MemberProfileModel(memProfilemodel))
}

func (c *Controller) MemberProfileUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
	}

	// Get existing member profile
	existing, err := c.memberProfile.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
	}

	// Validate request payload
	req, err := c.model.MemberProfileValidate(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Validate allowed status values
	validStatus := map[string]bool{
		"pending":     true,
		"for review":  true,
		"verified":    true,
		"not allowed": true,
	}
	if !validStatus[req.Status] {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid status value",
		})
	}

	// Get current user
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Prepare updated member profile model
	memProfileModel := &model.MemberProfile{
		UpdatedAt:              time.Now().UTC(),
		UpdatedByID:            user.UserID,
		BranchID:               *user.BranchID,
		OrganizationID:         user.OrganizationID,
		MediaID:                req.MediaID,
		SignatureMediaID:       req.SignatureMediaID,
		MemberCenterID:         req.MemberCenterID,
		MemberClassificationID: req.MemberClassificationID,
		MemberGenderID:         req.MemberGenderID,
		MemberGroupID:          req.MemberGroupID,
		MemberOccupationID:     req.MemberOccupationID,
		IsClosed:               req.IsClosed,
		IsMutualFundMember:     req.IsMutualFundMember,
		IsMicroFinanceMember:   req.IsMicroFinanceMember,
		FirstName:              req.FirstName,
		MiddleName:             req.MiddleName,
		LastName:               req.LastName,
		Suffix:                 req.Suffix,
		Birthdate:              req.Birthdate,
		Description:            req.Description,
		Notes:                  req.Notes,
		ContactNumber:          req.ContactNumber,
		OldReferenceID:         req.OldReferenceID,
		Passbook:               req.Passbook,
		Occupation:             req.Occupation,
		BusinessAddress:        req.BusinessAddress,
		BusinessContactNumber:  req.BusinessContactNumber,
		CivilStatus:            req.CivilStatus,
	}

	// Construct full name
	memProfileModel.FullName = fmt.Sprintf("%s %s %s %s", req.FirstName, req.MiddleName, req.LastName, req.Suffix)

	// Status update logic
	if existing.Status != req.Status {
		if user.UserType != "owner" && user.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "Only organization owners or employees can modify member status",
			})
		}

		verified, err := c.memberVerification.GetByMemberProfileID(existing.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Member verification record not found",
			})
		}

		verified.Status = req.Status
		verified.UpdatedAt = time.Now().UTC()
		verified.UpdatedByID = user.UserID
		verified.VerifiedByUserID = user.UserID

		if err := c.memberVerification.Manager.Update(verified); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update member verification",
			})
		}

		memProfileModel.Status = req.Status

		c.provider.Notification(ctx, "Member Status Update",
			fmt.Sprintf("Member %s status changed from %s to %s",
				existing.FullName, existing.Status, req.Status),
			"info")
	} else {
		memProfileModel.Status = existing.Status
	}

	// Save updated member profile
	if err := c.memberProfile.Manager.UpdateByID(*id, memProfileModel); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update member profile: " + err.Error(),
		})
	}

	// Log activity
	c.provider.UserFootstep(ctx, "member-profile",
		fmt.Sprintf("Updated member profile for %s", memProfileModel.FullName),
		memProfileModel)

	// Return updated profile
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModel(memProfileModel))
}

// DELETE /member-profile/member_profile_id
func (c *Controller) MemberProfileDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
	if err != nil {
		return err
	}
	if err := c.memberProfile.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-profile/branch/:branch_id
func (c *Controller) MemberProfileListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_profile, err := c.memberProfile.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModels(member_profile))
}

// GET member-profile/organization/:organization_id
func (c *Controller) MemberProfileListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_profile, err := c.memberProfile.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModels(member_profile))
}

// GET member_profile/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberProfileListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_profile, err := c.memberProfile.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberProfileModels(member_profile))
}
