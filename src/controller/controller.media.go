package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MediaController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/media",
		Method:   "GET",
		Response: "TMedia[]",
	}, func(ctx echo.Context) error {
		media, err := c.model.MediaManager.ListRaw(context.Background())
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, media)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/media/:media_id",
		Method:   "GET",
		Response: "TMedia",
	}, func(ctx echo.Context) error {
		context := context.Background()
		mediaId, err := horizon.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return err
		}

		media, err := c.model.MediaManager.GetByIDRaw(context, *mediaId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, media)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/media",
		Method:   "POST",
		Request:  "File - multipart/form-data",
		Response: "TMedia",
		Note:     "this route is used for uploading files",
	}, func(ctx echo.Context) error {
		context := context.Background()
		file, err := ctx.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "missing file")
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		completed := &model.Media{
			FileName:   file.Filename,
			FileType:   file.Header.Get("Content-Type"),
			ID:         initial.ID,
			FileSize:   storage.FileSize,
			StorageKey: storage.StorageKey,
			URL:        storage.URL,
			BucketName: storage.BucketName,
			Status:     "completed",
			Progress:   100,
			UpdatedAt:  time.Now().UTC(),
		}
		if err := c.model.MediaManager.Update(context, completed); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.MediaManager.ToModel(completed))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/media/:media_id",
		Method:   "PUT",
		Request:  "TMedia",
		Response: "TMedia",
		Note:     "This only change file name",
	}, func(ctx echo.Context) error {
		context := context.Background()
		mediaId, err := horizon.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return err
		}
		req, err := c.model.MediaManager.Validate(ctx)
		if err != nil {
			return err
		}
		model := &model.Media{
			FileName:  req.FileName,
			UpdatedAt: time.Now().UTC(),
		}

		if err := c.model.MediaManager.UpdateByID(context, *mediaId, model); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.MediaManager.ToModel(model))

	})

	req.RegisterRoute(horizon.Route{
		Route:  "/media/:media_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := context.Background()
		mediaId, err := horizon.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return err
		}
		media, err := c.model.MediaManager.GetByID(context, *mediaId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err := c.model.MediaDelete(context, media.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/media/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple media records",
	}, func(ctx echo.Context) error {
		context := context.Background()
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
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		for _, rawID := range reqBody.IDs {
			mediaID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			media, err := c.model.MediaManager.GetByID(context, mediaID)
			if err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Media with ID %s", rawID))
			}
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
				return err
			}
			if err := c.model.MediaManager.DeleteByIDWithTx(context, tx, mediaID); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
