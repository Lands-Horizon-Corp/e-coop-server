package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// MediaController manages endpoints for media records.
func (c *Controller) MediaController() {

	req := c.provider.Service.Request

	// GET /media: List all media records. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/media",
		Method:       "GET",
		Note:         "Returns all media records in the system.",
		ResponseType: model.MediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		media, err := c.model.MediaManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve media records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MediaManager.Filtered(context, ctx, media))
	})

	// GET /media/:media_id: Get a specific media record by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/media/:media_id",
		Method:       "GET",
		Note:         "Returns a specific media record by its ID.",
		ResponseType: model.MediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaId, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}

		media, err := c.model.MediaManager.GetByIDRaw(context, *mediaId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		return ctx.JSON(http.StatusOK, media)
	})

	// POST /media: Upload a new media file. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/media",
		Method:       "POST",
		ResponseType: model.MediaResponse{},
		Note:         "Uploads a file and creates a new media record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		file, err := ctx.FormFile("file")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), missing file in upload.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing file in upload"})
		}
		initial := &model.Media{
			FileName:   file.Filename,
			FileSize:   0,
			FileType:   file.Header.Get("Content-Type"),
			StorageKey: "",
			URL:        "",
			BucketName: "",
			Status:     "pending",
			Progress:   0,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		if err := c.model.MediaManager.Create(context, initial); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create media record: " + err.Error()})
		}
		storage, err := c.provider.Service.Storage.UploadFromHeader(context, file, func(progress, total int64, storage *horizon.Storage) {
			_ = c.model.MediaManager.Update(context, &model.Media{
				ID:        initial.ID,
				Progress:  progress,
				Status:    "progress",
				UpdatedAt: time.Now().UTC(),
			})
		})
		if err != nil {
			_ = c.model.MediaManager.Update(context, &model.Media{
				ID:        initial.ID,
				Status:    "error",
				UpdatedAt: time.Now().UTC(),
			})
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), file upload failed: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "File upload failed: " + err.Error()})
		}
		completed := &model.Media{
			FileName:   storage.FileName,
			FileType:   storage.FileType,
			FileSize:   storage.FileSize,
			StorageKey: storage.StorageKey,
			URL:        storage.URL,
			BucketName: storage.BucketName,
			Status:     "completed",
			Progress:   100,
			CreatedAt:  initial.CreatedAt,
			UpdatedAt:  time.Now().UTC(),
			ID:         initial.ID,
		}
		if err := c.model.MediaManager.Update(context, completed); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Media upload failed (/media), update after upload error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update media record after upload: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Uploaded and created media (/media): " + completed.FileName,
			Module:      "Media",
		})
		return ctx.JSON(http.StatusCreated, c.model.MediaManager.ToModel(completed))
	})

	// PUT /media/:media_id: Update media file's name. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/media/:media_id",
		Method:       "PUT",
		RequestType:  model.MediaRequest{},
		ResponseType: model.MediaResponse{},
		Note:         "Updates the file name of a media record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaId, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), invalid media ID.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}
		req, err := c.model.MediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), validation error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media data: " + err.Error()})
		}
		media, err := c.model.MediaManager.GetByID(context, *mediaId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), record not found.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		media.FileName = req.FileName
		media.UpdatedAt = time.Now().UTC()
		if err := c.model.MediaManager.UpdateFields(context, *mediaId, media); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Media update failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update media record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated media (/media/:media_id): " + media.FileName,
			Module:      "Media",
		})
		return ctx.JSON(http.StatusOK, c.model.MediaManager.ToModel(media))
	})

	// DELETE /media/:media_id: Delete a media record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/media/:media_id",
		Method: "DELETE",
		Note:   "Deletes a specific media record by its ID and associated file.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaId, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), invalid media ID.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media ID"})
		}
		media, err := c.model.MediaManager.GetByID(context, *mediaId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), record not found.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media record not found"})
		}
		if err := c.model.MediaDelete(context, media.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted media (/media/:media_id): " + media.FileName,
			Module:      "Media",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /media/bulk-delete: Bulk delete media records by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/media/bulk-delete",
		Method:      "DELETE",
		RequestType: model.IDSRequest{},
		Note:        "Deletes multiple media records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete), invalid request body.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete), no IDs provided.",
				Module:      "Media",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			mediaID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Media bulk delete failed (/media/bulk-delete), invalid UUID: " + rawID,
					Module:      "Media",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			media, err := c.model.MediaManager.GetByID(context, mediaID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Media bulk delete failed (/media/bulk-delete), not found: " + rawID,
					Module:      "Media",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Media record not found with ID: %s", rawID)})
			}
			names += media.FileName + ","
			if err := c.provider.Service.Storage.DeleteFile(context, &horizon.Storage{
				FileName:   media.FileName,
				FileSize:   media.FileSize,
				FileType:   media.FileType,
				StorageKey: media.StorageKey,
				URL:        media.URL,
				BucketName: media.BucketName,
				Status:     "delete",
			}); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Media bulk delete failed (/media/bulk-delete), storage delete error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete file from storage: " + err.Error()})
			}
			if err := c.model.MediaManager.DeleteByIDWithTx(context, tx, mediaID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Media bulk delete failed (/media/bulk-delete), db error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Media bulk delete failed (/media/bulk-delete), commit error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted media (/media/bulk-delete): " + names,
			Module:      "Media",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
