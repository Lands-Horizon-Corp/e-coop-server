package user

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

func FeedCommentController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/:feed_id/comments",
		Method:       "GET",
		Note:         "Returns all comments for a feed.",
		ResponseType: []types.FeedCommentResponse{},
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
		comments, err := core.FeedCommentManager(service).Find(
			context,
			&types.FeedComment{
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				FeedID:         *feedID,
			},
		)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No comments found"})
		}
		return ctx.JSON(http.StatusOK, core.FeedCommentManager(service).ToModels(comments))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/feed/comment/:comment_id",
		Method:       "PUT",
		Note:         "Updates a comment.",
		RequestType:  types.FeedCommentRequest{},
		ResponseType: types.FeedCommentResponse{},
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()

		commentID, err := helpers.EngineUUIDParam(ctx, "comment_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
		}

		req, err := core.FeedCommentManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot edit this comment"})
		}

		comment.Comment = req.Comment
		comment.UpdatedAt = time.Now().UTC()
		comment.UpdatedByID = userOrg.UserID

		if err := core.FeedCommentManager(service).UpdateByID(context, comment.ID, comment); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.FeedCommentManager(service).ToModel(comment))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/feed/comment/:comment_id",
		Method: "DELETE",
		Note:   "Deletes a comment.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot delete this comment"})
		}

		if err := core.FeedCommentManager(service).Delete(context, comment.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
