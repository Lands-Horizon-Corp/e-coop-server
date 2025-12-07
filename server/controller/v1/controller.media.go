package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/labstack/echo/v4"
)

// MediaController manages endpoints for media records.
func (c *Controller) mediaController() {

	req := c.provider.Service.Request

	// GET /media: List all media records. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/media",
		Method:       "GET",
		Note:         "Returns all media records in the system.",
		ResponseType: core.MediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		media, err := c.core.MediaManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve media records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MediaManager.ToModels(media))
	})

	// GET /media/:media_id: Get a specific media record by ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/media/:media_id",
		Method:       "GET",
		Note:         "Returns a specific media record by its ID.",
		ResponseType: core.MediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaID, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}

		media, err := c.core.MediaManager.GetByIDRaw(context, *mediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		return ctx.JSON(http.StatusOK, media)
	})

	// POST /media: Upload a new media file. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/media",
		Method:       "POST",
		ResponseType: core.MediaResponse{},
		Note:         "Uploads a file and creates a new media record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		file, err := ctx.FormFile("file")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), missing file in upload.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing file in upload"})
		}
		// Ensure filename has proper extension based on content type
		fileName := file.Filename
		contentType := file.Header.Get("Content-Type")

		// If filename doesn't have extension, add it based on content type
		if fileName != "" && !handlers.HasFileExtension(fileName) {
			if ext := handlers.GetExtensionFromContentType(contentType); ext != "" {
				fileName += ext
			}
		}
		initial := &core.Media{
			FileName:   fileName,
			FileSize:   0,
			FileType:   contentType,
			StorageKey: "",
			BucketName: "",
			Status:     "pending",
			Progress:   0,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		if err := c.core.MediaManager.Create(context, initial); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create media record: " + err.Error()})
		}
		storage, err := c.provider.Service.Storage.UploadFromHeader(context, file, func(progress, _ int64, _ *horizon.Storage) {
			_ = c.core.MediaManager.UpdateByID(context, initial.ID, &core.Media{
				Progress:  progress,
				Status:    "progress",
				UpdatedAt: time.Now().UTC(),
			})
		})
		if err != nil {
			_ = c.core.MediaManager.UpdateByID(context, initial.ID, &core.Media{
				ID:        initial.ID,
				Status:    "error",
				UpdatedAt: time.Now().UTC(),
			})
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), file upload failed: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "File upload failed: " + err.Error()})
		}
		completed := &core.Media{
			FileName:   storage.FileName,
			FileType:   storage.FileType,
			FileSize:   storage.FileSize,
			StorageKey: storage.StorageKey,
			BucketName: storage.BucketName,
			Status:     "completed",
			Progress:   100,
			CreatedAt:  initial.CreatedAt,
			UpdatedAt:  time.Now().UTC(),
			ID:         initial.ID,
		}
		if err := c.core.MediaManager.UpdateByID(context, completed.ID, completed); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), update after upload error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update media record after upload: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Uploaded and created media (/media): " + completed.FileName,
			Module:      "Media",
		})
		return ctx.JSON(http.StatusCreated, c.core.MediaManager.ToModel(completed))
	})

	// PUT /media/:media_id: Update media file's name. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/media/:media_id",
		Method:       "PUT",
		RequestType:  core.MediaRequest{},
		ResponseType: core.MediaResponse{},
		Note:         "Updates the file name of a media record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaID, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), invalid media ID.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}
		req, err := c.core.MediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), validation error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media data: " + err.Error()})
		}
		media, err := c.core.MediaManager.GetByID(context, *mediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), record not found.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		media.FileName = req.FileName
		media.UpdatedAt = time.Now().UTC()
		if err := c.core.MediaManager.UpdateByID(context, *mediaID, media); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update media record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated media (/media/:media_id): " + media.FileName,
			Module:      "Media",
		})
		return ctx.JSON(http.StatusOK, c.core.MediaManager.ToModel(media))
	})

	// DELETE /media/:media_id: Delete a media record by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/media/:media_id",
		Method: "DELETE",
		Note:   "Deletes a specific media record by its ID and associated file.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaID, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), invalid media ID.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}
		media, err := c.core.MediaManager.GetByID(context, *mediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), record not found.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		if err := c.core.MediaDelete(context, media.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted media (/media/:media_id): " + media.FileName,
			Module:      "Media",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for media (moves storage + DB work into manager)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/media/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple media records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete) | no IDs provided",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate storage deletion and DB transaction to the manager.
		// Assumes MediaManager.BulkDelete(ctx context.Context, ids []string) error exists
		if err := c.core.MediaManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete) | error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete media records: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted media (/media/bulk-delete)",
			Module:      "Media",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
