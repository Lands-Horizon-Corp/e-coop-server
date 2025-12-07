package v1

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// MemberProfileArchiveController registers routes for managing member profile archive.
func (c *Controller) memberProfileArchiveController() {
	req := c.provider.Service.WebRequest

	// GET /api/v1/member-profile-archive/member-profile/:member_profile_id/search: Get all media for a specific member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive/member-profile/:member_profile_id",
		Method:       "GET",
		Note:         "Get all member profile archive for a specific member profile.",
		ResponseType: core.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), invalid member profile ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		// Verify member profile belongs to user's organization
		memberProfile, err := c.core.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), member profile not found.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
		}
		// Search for all member profile archive for the specified member profile
		memberProfileArchiveList, err := c.core.MemberProfileArchiveManager.FindRaw(context, &core.MemberProfileArchive{
			BranchID:        userOrg.BranchID,
			OrganizationID:  &userOrg.OrganizationID,
			MemberProfileID: &memberProfile.ID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search member profile archive: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "member-profile-search-success",
			Description: "Member profile archive member profile search successful (/member-profile-archive/member-profile/:member_profile_id/search), found " + strconv.Itoa(len(memberProfileArchiveList)) + " media items.",
			Module:      "MemberProfileArchive",
		})

		return ctx.JSON(http.StatusOK, memberProfileArchiveList)
	})

	// POST /api/v1/member-profile-archive: Create a new member profile archive
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive",
		Method:       "POST",
		Note:         "Creates a new member profile archive for the current user's organization and branch.",
		RequestType:  core.MemberProfileArchiveRequest{},
		ResponseType: core.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.core.MemberProfileArchiveManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), validation error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), user not assigned to branch.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileArchive := &core.MemberProfileArchive{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			Category:       req.Category,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		}

		if err := c.core.MemberProfileArchiveManager.Create(context, memberProfileArchive); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile archive: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Member profile archive created successfully (/member-profile-archive), ID: " + memberProfileArchive.ID.String(),
			Module:      "MemberProfileArchive",
		})

		result, err := c.core.MemberProfileArchiveManager.GetByID(context, memberProfileArchive.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created member profile archive: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	// PUT /api/v1/member-profile-archive/:member_profile_archive_id: Update a member profile archive
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method:       "PUT",
		Note:         "Update a member profile archive by ID.",
		RequestType:  core.MemberProfileArchiveRequest{},
		ResponseType: core.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := handlers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), invalid member profile archive ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		req, err := c.core.MemberProfileArchiveManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), validation error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileArchive, err := c.core.MemberProfileArchiveManager.GetByID(context, *memberProfileArchiveID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), member profile archive not found.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile archive not found"})
		}
		memberProfileArchive.MediaID = req.MediaID
		memberProfileArchive.Name = req.Name
		memberProfileArchive.Description = req.Description
		memberProfileArchive.Category = req.Category
		memberProfileArchive.UpdatedAt = time.Now().UTC()
		memberProfileArchive.UpdatedByID = userOrg.UserID

		if err := c.core.MemberProfileArchiveManager.UpdateByID(context, memberProfileArchive.ID, memberProfileArchive); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile archive: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Member profile archive updated successfully (/member-profile-archive/:member_profile_archive_id), ID: " + memberProfileArchiveID.String(),
			Module:      "MemberProfileArchive",
		})

		result, err := c.core.MemberProfileArchiveManager.GetByID(context, *memberProfileArchiveID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated member profile archive: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// DELETE /api/v1/member-profile-archive/:member_profile_archive_id: Delete a member profile archive
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method: "DELETE",
		Note:   "Delete a member profile archive by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := handlers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), invalid member profile archive ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		_, err = c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileArchive, err := c.core.MemberProfileArchiveManager.GetByID(context, *memberProfileArchiveID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), not found.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile archive not found"})
		}
		if err := c.core.MediaDelete(context, *memberProfileArchive.MediaID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := c.core.MemberProfileArchiveManager.Delete(context, memberProfileArchive.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile archive: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Member profile archive deleted successfully (/member-profile-archive/:member_profile_archive_id), ID: " + memberProfileArchiveID.String(),
			Module:      "MemberProfileArchive",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Member profile archive deleted successfully"})
	})
	// GET /api/v1/member-profile-archive/:member_profile_archive_id: Get a specific member profile archive by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method:       "GET",
		Note:         "Get a specific member profile archive by ID.",
		ResponseType: core.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := handlers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		memberProfileArchive, err := c.core.MemberProfileArchiveManager.GetByIDRaw(context, *memberProfileArchiveID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile archive not found"})
		}

		return ctx.JSON(http.StatusOK, memberProfileArchive)
	})

	// POST /api/v1/member-profile-archive/bulk/member-profile/:member_profile_id: Bulk create member profile archive for a specific member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive/bulk/member-profile/:member_profile_id",
		Method:       "POST",
		Note:         "Bulk create member profile archive for a specific member profile.",
		RequestType:  core.MemberProfileArchiveBulkRequest{},
		ResponseType: core.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		var req core.MemberProfileArchiveBulkRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*core.MemberProfileArchive
		for _, mediaID := range req.IDs {
			media, err := c.core.MediaManager.GetByID(context, mediaID)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found: " + mediaID.String()})
			}
			memberProfileArchive := &core.MemberProfileArchive{
				MediaID:         &mediaID,
				CreatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       time.Now().UTC(),
				UpdatedByID:     userOrg.UserID,
				BranchID:        userOrg.BranchID,
				OrganizationID:  &userOrg.OrganizationID,
				MemberProfileID: memberProfileID,
				Name:            media.FileName,
				Description:     media.FileName + " at " + time.Now().Format(time.RFC3339),
				Category:        req.Category,
			}

			if err := c.core.MemberProfileArchiveManager.Create(context, memberProfileArchive); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile archive: " + err.Error()})
			}

			createdMedia = append(createdMedia, memberProfileArchive)
		}
		return ctx.JSON(http.StatusCreated, c.core.MemberProfileArchiveManager.ToModels(createdMedia))
	})

	// GET api/v1/member-profile-archive/member-profile/:member_profile_id/category
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-archive/member-profile/:member_profile_id/category",
		Method:       "GET",
		Note:         "Get distinct categories of member profile archive for a specific member profile.",
		ResponseType: core.MemberProfileArchiveCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileArchives, err := c.core.MemberProfileArchiveManager.Find(context, &core.MemberProfileArchive{
			MemberProfileID: memberProfileID,
			OrganizationID:  &userOrg.OrganizationID,
			BranchID:        userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve categories: " + err.Error()})
		}

		// Build counts per category, treating nil/empty as "Uncategorized"
		// Normalize by trimming, collapsing spaces and lower-casing for case-insensitive grouping
		normalize := func(s string) string {
			trimmed := strings.Join(strings.Fields(strings.TrimSpace(s)), " ") // collapse multiple spaces
			return strings.ToLower(trimmed)
		}
		titleize := func(s string) string {
			// simple title case for display: capitalize first letter of each word
			words := strings.Fields(s)
			for i, w := range words {
				if len(w) == 0 {
					continue
				}
				words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
			return strings.Join(words, " ")
		}

		counts := make(map[string]int)            // normalized -> count
		displayVariant := make(map[string]string) // normalized -> first-seen trimmed display

		for _, a := range memberProfileArchives {
			if a == nil {
				continue
			}
			var raw string
			switch v := any(a.Category).(type) {
			case *string:
				if v != nil {
					raw = *v
				}
			case string:
				raw = v
			}
			trimmed := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
			if trimmed == "" {
				trimmed = "Uncategorized"
			}
			norm := normalize(trimmed)
			if _, ok := displayVariant[norm]; !ok {
				displayVariant[norm] = trimmed
			}
			counts[norm]++
		}

		// Default categories commonly used in cooperative bank member profiles
		defaultCategories := []string{
			"Identity Documents",
			"Passports",
			"Driver's License",
			"KYC Documents",
			"Proof of Address",
			"Financial Documents",
			"Loans & Mortgages",
			"Tax Documents",
			"Insurance",
			"Agreements & Contracts",
			"Certificates (Birth/Marriage)",
			"Employment Records",
			"Medical Records",
			"Photos & Signatures",
			"Correspondence",
			"Legal Documents",
			"Education / Qualifications",
			"Miscellaneous",
			"Uncategorized",
		}

		// prepare normalized map for defaults so we present canonical display names
		normDefault := make(map[string]string)
		for _, d := range defaultCategories {
			normDefault[normalize(d)] = d
		}

		// Prepare result with defaults (preserve ordering) and include any additional categories found
		result := make([]core.MemberProfileArchiveCategoryResponse, 0, len(defaultCategories))
		seen := make(map[string]bool)
		for _, name := range defaultCategories {
			n := normalize(name)
			result = append(result, core.MemberProfileArchiveCategoryResponse{
				Name:  normDefault[n],
				Count: int64(counts[n]),
			})
			seen[n] = true
		}
		// add any categories present in counts but not in defaults
		for norm, cnt := range counts {
			if seen[norm] {
				continue
			}
			display := displayVariant[norm]
			if display == "" {
				// fallback: make a readable display from normalized key
				display = titleize(strings.ReplaceAll(norm, "_", " "))
			} else {
				// use cleaned display (collapse spaces) and titleize for consistent casing
				display = titleize(display)
			}
			result = append(result, core.MemberProfileArchiveCategoryResponse{
				Name:  display,
				Count: int64(cnt),
			})
		}

		return ctx.JSON(http.StatusOK, result)
	})
}
