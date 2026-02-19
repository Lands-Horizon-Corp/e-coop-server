package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func FeedController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/search",
		Method:       "GET",
		Note:         "Returns a paginated list of feeds.",
		ResponseType: types.FeedResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication failed"})
		}
		feeds, err := core.FeedManager(service).NormalPagination(context, ctx, &types.Feed{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch feeds: " + err.Error()})
		}
		for i := range feeds.Data {
			for _, like := range feeds.Data[i].UserLikes {
				if like.UserID == userOrg.UserID {
					feeds.Data[i].IsLiked = true
					break
				}
			}
		}
		return ctx.JSON(http.StatusOK, feeds)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/:feed_id",
		Method:       "GET",
		Note:         "Returns a single feed by its ID.",
		ResponseType: types.FeedResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedID, err := helpers.EngineUUIDParam(ctx, "feed_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feed ID"})
		}
		feed, err := core.FeedManager(service).GetByIDRaw(context, *feedID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feed not found"})
		}
		return ctx.JSON(http.StatusOK, feed)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed",
		Method:       "POST",
		Note:         "Creates a new feed post.",
		RequestType:  types.FeedRequest{},
		ResponseType: types.FeedResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.FeedManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		tx, endTx := service.Database.StartTransaction(context)
		feed := &types.Feed{
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}
		if err := core.FeedManager(service).CreateWithTx(context, tx, feed); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity: "create-error", Module: "Feed",
				Description: "Feed creation failed: " + err.Error(),
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if len(req.MediaIDs) > 0 {
			var feedMedias []*types.FeedMedia
			for _, mID := range req.MediaIDs {
				if err := core.FeedMediaManager(service).CreateWithTx(context, tx, &types.FeedMedia{
					FeedID:         feed.ID,
					MediaID:        *mID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					CreatedByID:    userOrg.UserID,
					UpdatedByID:    userOrg.UserID,
				}); err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity: "create-error", Module: "Feed",
						Description: "Failed to associate media with feed: " + endTx(err).Error(),
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
				}
			}
			if err := tx.Create(&feedMedias).Error; err != nil {
				return err
			}
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity: "create-error", Module: "Feed",
				Description: "Transaction commit failed: " + err.Error(),
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity: "create-success", Module: "Feed",
			Description: fmt.Sprintf("Created feed: %s", feed.ID),
		})
		return ctx.JSON(http.StatusCreated, core.FeedManager(service).ToModel(feed))
	})

	// PUT: Update Feed
	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/:feed_id",
		Method:       "PUT",
		Note:         "Updates an existing feed description.",
		RequestType:  types.FeedRequest{},
		ResponseType: types.FeedResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedID, err := helpers.EngineUUIDParam(ctx, "feed_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}

		req, err := core.FeedManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		feed, err := core.FeedManager(service).GetByID(context, *feedID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feed not found"})
		}

		feed.Description = req.Description
		feed.UpdatedAt = time.Now().UTC()
		feed.UpdatedByID = userOrg.UserID

		if err := core.FeedManager(service).UpdateByID(context, feed.ID, feed); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity: "update-success", Module: "Feed",
			Description: "Updated feed: " + feed.ID.String(),
		})
		return ctx.JSON(http.StatusOK, core.FeedManager(service).ToModel(feed))
	})

	// DELETE: Single Feed
	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/feed/:feed_id",
		Method: "DELETE",
		Note:   "Deletes a feed post.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedID, err := helpers.EngineUUIDParam(ctx, "feed_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}

		if err := core.FeedManager(service).Delete(context, *feedID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity: "delete-success", Module: "Feed",
			Description: "Deleted feed: " + feedID.String(),
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE: Bulk Delete Feeds
	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/feed/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple feeds.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid body"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}

		if err := core.FeedManager(service).BulkDelete(context, ids); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity: "bulk-delete-success", Module: "Feed",
			Description: "Bulk deleted feeds",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/:feed_id/like",
		Method:       "PUT",
		Note:         "Toggles a like on a feed post. If already liked, it will unlike.",
		ResponseType: map[string]any{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedID, err := helpers.EngineUUIDParam(ctx, "feed_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feed ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		feed, err := core.FeedManager(service).GetByID(context, *feedID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feed not found"})
		}
		var existingLike types.FeedLike
		db := service.Database.Client().WithContext(context)
		result := db.Where("feed_id = ? AND user_id = ?", feed.ID, userOrg.UserID).First(&existingLike)

		if result.Error == nil {
			if err := core.FeedLikeManager(service).Delete(context, existingLike.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unlike"})
			}
			return ctx.JSON(http.StatusOK, map[string]string{"message": "Unliked", "status": "unliked"})
		}
		newLike := &types.FeedLike{
			FeedID:         feed.ID,
			UserID:         userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CreatedByID:    userOrg.UserID,
			UpdatedByID:    userOrg.UserID,
		}

		if err := core.FeedLikeManager(service).Create(context, newLike); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to like feed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Liked", "status": "liked"})
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/:feed_id/comment",
		Method:       "POST",
		Note:         "Adds a comment to a specific feed post.",
		RequestType:  types.FeedCommentRequest{},
		ResponseType: types.FeedCommentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		feedID, err := helpers.EngineUUIDParam(ctx, "feed_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid feed ID"})
		}

		req, err := core.FeedCommentManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		feed, err := core.FeedManager(service).GetByID(context, *feedID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Feed not found"})
		}
		comment := &types.FeedComment{
			FeedID:         feed.ID,
			UserID:         userOrg.UserID,
			Comment:        req.Comment,
			MediaID:        req.MediaID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CreatedByID:    userOrg.UserID,
			UpdatedByID:    userOrg.UserID,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		if err := core.FeedCommentManager(service).Create(context, comment); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post comment"})
		}
		return ctx.JSON(http.StatusCreated, core.FeedCommentManager(service).ToModel(comment))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/feed/comment/:comment_id",
		Method: "DELETE",
		Note:   "Deletes a comment. Users can only delete their own comments.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		commentID, err := helpers.EngineUUIDParam(ctx, "comment_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		comment, err := core.FeedCommentManager(service).GetByID(context, *commentID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Comment not found"})
		}
		if comment.UserID != userOrg.UserID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You can only delete your own comments"})
		}
		if err := core.FeedCommentManager(service).Delete(context, *commentID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comment"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
