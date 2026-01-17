package member_profile

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberProfileArchiveController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive/member-profile/:member_profile_id",
		Method:       "GET",
		Note:         "Get all member profile archive for a specific member profile.",
		ResponseType: types.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), invalid member profile ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), member profile not found.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
		}
		memberProfileArchiveList, err := core.MemberProfileArchiveManager(service).FindRaw(context, &types.MemberProfileArchive{
			BranchID:        userOrg.BranchID,
			OrganizationID:  &userOrg.OrganizationID,
			MemberProfileID: &memberProfile.ID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile archive member profile search failed (/member-profile-archive/member-profile/:member_profile_id/search), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search member profile archive: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "member-profile-search-success",
			Description: "Member profile archive member profile search successful (/member-profile-archive/member-profile/:member_profile_id/search), found " + strconv.Itoa(len(memberProfileArchiveList)) + " media items.",
			Module:      "MemberProfileArchive",
		})

		return ctx.JSON(http.StatusOK, memberProfileArchiveList)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive",
		Method:       "POST",
		Note:         "Creates a new member profile archive for the current user's organization and branch.",
		RequestType:  types.MemberProfileArchiveRequest{},
		ResponseType: types.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.MemberProfileArchiveManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), validation error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), user not assigned to branch.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileArchive := &types.MemberProfileArchive{
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

		if err := core.MemberProfileArchiveManager(service).Create(context, memberProfileArchive); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile archive creation failed (/member-profile-archive), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile archive: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Member profile archive created successfully (/member-profile-archive), ID: " + memberProfileArchive.ID.String(),
			Module:      "MemberProfileArchive",
		})

		result, err := core.MemberProfileArchiveManager(service).GetByID(context, memberProfileArchive.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created member profile archive: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method:       "PUT",
		Note:         "Update a member profile archive by ID.",
		RequestType:  types.MemberProfileArchiveRequest{},
		ResponseType: types.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := helpers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), invalid member profile archive ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		req, err := core.MemberProfileArchiveManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), validation error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileArchive, err := core.MemberProfileArchiveManager(service).GetByID(context, *memberProfileArchiveID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.MemberProfileArchiveManager(service).UpdateByID(context, memberProfileArchive.ID, memberProfileArchive); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile archive update failed (/member-profile-archive/:member_profile_archive_id), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile archive: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Member profile archive updated successfully (/member-profile-archive/:member_profile_archive_id), ID: " + memberProfileArchiveID.String(),
			Module:      "MemberProfileArchive",
		})

		result, err := core.MemberProfileArchiveManager(service).GetByID(context, *memberProfileArchiveID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated member profile archive: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method: "DELETE",
		Note:   "Delete a member profile archive by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := helpers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), invalid member profile archive ID.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		_, err = event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), user org error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileArchive, err := core.MemberProfileArchiveManager(service).GetByID(context, *memberProfileArchiveID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), not found.",
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile archive not found"})
		}
		if err := core.MediaDelete(context, service, *memberProfileArchive.MediaID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := core.MemberProfileArchiveManager(service).Delete(context, memberProfileArchive.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile archive delete failed (/member-profile-archive/:member_profile_archive_id), db error: " + err.Error(),
				Module:      "MemberProfileArchive",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile archive: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Member profile archive deleted successfully (/member-profile-archive/:member_profile_archive_id), ID: " + memberProfileArchiveID.String(),
			Module:      "MemberProfileArchive",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Member profile archive deleted successfully"})
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive/:member_profile_archive_id",
		Method:       "GET",
		Note:         "Get a specific member profile archive by ID.",
		ResponseType: types.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileArchiveID, err := helpers.EngineUUIDParam(ctx, "member_profile_archive_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile archive ID"})
		}

		memberProfileArchive, err := core.MemberProfileArchiveManager(service).GetByIDRaw(context, *memberProfileArchiveID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile archive not found"})
		}

		return ctx.JSON(http.StatusOK, memberProfileArchive)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive/bulk/member-profile/:member_profile_id",
		Method:       "POST",
		Note:         "Bulk create member profile archive for a specific member profile.",
		RequestType:  types.MemberProfileArchiveBulkRequest{},
		ResponseType: types.MemberProfileArchiveResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		var req types.MemberProfileArchiveBulkRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*types.MemberProfileArchive
		for _, mediaID := range req.IDs {
			media, err := core.MediaManager(service).GetByID(context, mediaID)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found: " + mediaID.String()})
			}
			memberProfileArchive := &types.MemberProfileArchive{
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

			if err := core.MemberProfileArchiveManager(service).Create(context, memberProfileArchive); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile archive: " + err.Error()})
			}

			createdMedia = append(createdMedia, memberProfileArchive)
		}
		return ctx.JSON(http.StatusCreated, core.MemberProfileArchiveManager(service).ToModels(createdMedia))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-archive/member-profile/:member_profile_id/category",
		Method:       "GET",
		Note:         "Get distinct categories of member profile archive for a specific member profile.",
		ResponseType: types.MemberProfileArchiveCategoryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileArchives, err := core.MemberProfileArchiveManager(service).Find(context, &types.MemberProfileArchive{
			MemberProfileID: memberProfileID,
			OrganizationID:  &userOrg.OrganizationID,
			BranchID:        userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve categories: " + err.Error()})
		}

		normalize := func(s string) string {
			trimmed := strings.Join(strings.Fields(strings.TrimSpace(s)), " ") // collapse multiple spaces
			return strings.ToLower(trimmed)
		}
		titleize := func(s string) string {
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

		normDefault := make(map[string]string)
		for _, d := range defaultCategories {
			normDefault[normalize(d)] = d
		}

		result := make([]types.MemberProfileArchiveCategoryResponse, 0, len(defaultCategories))
		seen := make(map[string]bool)
		for _, name := range defaultCategories {
			n := normalize(name)
			result = append(result, types.MemberProfileArchiveCategoryResponse{
				Name:  normDefault[n],
				Count: int64(counts[n]),
			})
			seen[n] = true
		}
		for norm, cnt := range counts {
			if seen[norm] {
				continue
			}
			display := displayVariant[norm]
			if display == "" {
				display = titleize(strings.ReplaceAll(norm, "_", " "))
			} else {
				display = titleize(display)
			}
			result = append(result, types.MemberProfileArchiveCategoryResponse{
				Name:  display,
				Count: int64(cnt),
			})
		}

		return ctx.JSON(http.StatusOK, result)
	})
}
