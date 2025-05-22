package controller

import (
	"context"
	"net/http"
	"time"

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
		media, err := c.media.Manager.ListRaw(context.Background())
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

		media, err := c.media.Manager.GetByIDRaw(context, *mediaId)
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
		if err := c.media.Manager.Create(context, initial); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		storage, err := c.provider.Service.Storage.UploadFromHeader(context, file, func(progress, total int64, storage *horizon.Storage) {
			_ = c.media.Manager.Update(context, &model.Media{
				ID:        initial.ID,
				Progress:  progress,
				Status:    "progress",
				UpdatedAt: time.Now().UTC(),
			})
		})
		if err != nil {
			_ = c.media.Manager.Update(context, &model.Media{
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
		if err := c.media.Manager.Update(context, completed); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.media.Manager.ToModel(completed))
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
		req, err := c.media.Manager.Validate(ctx)
		if err != nil {
			return err
		}
		model := &model.Media{
			FileName:  req.FileName,
			UpdatedAt: time.Now().UTC(),
		}

		if err := c.media.Manager.UpdateByID(context, *mediaId, model); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.media.Manager.ToModel(model))

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
		media, err := c.media.Manager.GetByID(context, *mediaId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
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
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		if err := c.media.Manager.DeleteByID(context, *mediaId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
