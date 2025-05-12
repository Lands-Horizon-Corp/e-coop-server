package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /media
func (c *Controller) MediaList(ctx echo.Context) error {
	media, err := c.media.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MediaModels(media))
}

// GET /media/:media_id
func (c *Controller) MediaGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "media_id")
	if err != nil {
		return err
	}
	media, err := c.media.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MediaModel(media))
}

// POST /media
func (c *Controller) MediaCreate(ctx echo.Context) error {
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
		Status:     horizon.StorageStatusPending,
		Progress:   0,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := c.media.Manager.Create(initial); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	storage, err := c.storage.UploadFromHeader(file, func(progress, total int64, st *horizon.Storage) {

		_ = c.media.Manager.Update(&model.Media{
			ID:        initial.ID,
			Progress:  st.Progress,
			Status:    horizon.StorageStatusProgress,
			UpdatedAt: time.Now().UTC(),
		})
	})
	if err != nil {

		_ = c.media.Manager.Update(&model.Media{
			ID:        initial.ID,
			Status:    horizon.StorageStatusCorrupt,
			UpdatedAt: time.Now().UTC(),
		})
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	completed := &model.Media{
		ID:         initial.ID,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		URL:        storage.URL,
		BucketName: storage.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   100,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := c.media.Manager.Update(completed); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.MediaModel(completed))
}

// PUT /media/media_id
func (c *Controller) MediaUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "media_id")
	if err != nil {
		return err
	}
	req, err := c.model.MediaValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.Media{
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		StorageKey: req.StorageKey,
		URL:        req.URL,
		BucketName: req.BucketName,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := c.media.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.MediaModel(model))
}

// DELETE /media/media_id
func (c *Controller) MediaDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "media_id")
	if err != nil {
		return err
	}
	media, err := c.media.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if err := c.storage.DeleteFile(media.StorageKey); err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	if err := c.media.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
