package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

func (h *Handler) MediaList(c echo.Context) error {
	media, err := h.repository.MediaList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.MediaModels(media))
}

func (h *Handler) MediaGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	media, err := h.repository.MediaGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.MediaModel(media))
}

func (h *Handler) MediaCreate(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing file")
	}
	// 2. Insert placeholder record with Status = pending
	initial := &model.Media{
		FileName:   file.Filename,
		FileSize:   0,
		FileType:   file.Header.Get("Content-Type"),
		StorageKey: "",
		URL:        "",
		BucketName: "",
		Status:     horizon.StorageStatusPending,
		Progress:   0,
	}
	if err := h.repository.MediaCreate(initial); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	storage, err := h.storage.UploadFromHeader(file, func(progress, total int64, st *horizon.Storage) {
		update := &model.Media{
			ID:       initial.ID,
			Progress: st.Progress,
			Status:   horizon.StorageStatusProgress,
		}
		_ = h.repository.MediaUpdate(update)
	})
	if err != nil {
		_ = h.repository.MediaUpdate(&model.Media{
			ID:     initial.ID,
			Status: horizon.StorageStatusCorrupt,
		})
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	completed := &model.Media{
		ID:         initial.ID,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		URL:        storage.URL,
		BucketName: storage.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   storage.Progress,
	}

	if err := h.repository.MediaUpdate(completed); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.MediaModel(completed))
}

func (h *Handler) MediaUpdate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	req, err := h.model.MediaValidate(c)
	if err != nil {
		return err
	}
	model := &model.Media{
		ID:         id,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		StorageKey: req.StorageKey,
		URL:        req.URL,
		BucketName: req.BucketName,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := h.repository.MediaUpdate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.MediaModel(model))

}

func (h *Handler) MediaDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	media, err := h.repository.MediaGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if err := h.storage.DeleteFile(media.StorageKey); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	model := &model.Media{ID: id}
	if err := h.repository.MediaDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
