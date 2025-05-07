package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type MediaController struct {
	repository *repository.MediaRepository
	collection *collection.MediaCollection
	storage    *horizon.HorizonStorage
	broadcast  *horizon.HorizonBroadcast
}

func NewMediaController(
	repository *repository.MediaRepository,
	collection *collection.MediaCollection,
	storage *horizon.HorizonStorage,
	broadcast *horizon.HorizonBroadcast,
) (*MediaController, error) {
	return &MediaController{
		repository: repository,
		collection: collection,
		storage:    storage,
		broadcast:  broadcast,
	}, nil
}

func (fc *MediaController) List(c echo.Context) error {
	medias, err := fc.repository.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	responses := fc.collection.ToModels(medias)
	return c.JSON(http.StatusOK, responses)
}

func (fc *MediaController) Get(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	media, err := fc.repository.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	resp := fc.collection.ToModel(media)
	return c.JSON(http.StatusOK, resp)
}

func (fc *MediaController) Create(c echo.Context) error {
	// 1. Bind multipart file
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing file")
	}

	// 2. Insert placeholder record with Status = pending
	initial := &collection.Media{
		FileName:   file.Filename,
		FileSize:   0,
		FileType:   file.Header.Get("Content-Type"),
		StorageKey: "",
		URL:        "",
		BucketName: "",
		Status:     horizon.StorageStatusPending,
		Progress:   0,
	}
	if err := fc.repository.Create(initial); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 3. Kick off upload with a callback that updates the same DB record
	storage, err := fc.storage.UploadFromHeader(file, func(progress, total int64, st *horizon.Storage) {
		// update record progress and status
		update := &collection.Media{
			ID:       initial.ID, // ensure we update the same record
			Progress: st.Progress,
			Status:   horizon.StorageStatusProgress,
		}
		_ = fc.repository.Update(update) // ignore error here, but you could log it
	})
	if err != nil {
		// mark record as corrupt on error
		_ = fc.repository.Update(&collection.Media{
			ID:     initial.ID,
			Status: horizon.StorageStatusCorrupt,
		})
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 4. Final update: set completed metadata
	completed := &collection.Media{
		ID:         initial.ID,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		URL:        storage.URL,
		BucketName: storage.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   storage.Progress,
	}
	if err := fc.repository.Update(completed); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 5. Fetch and return the up-to-date model
	respModel := fc.collection.ToModel(completed)
	return c.JSON(http.StatusCreated, respModel)
}

func (fc *MediaController) Update(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	req, err := fc.collection.ValidateCreate(c)
	if err != nil {
		return err
	}
	model := &collection.Media{
		ID:         id,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		FileType:   req.FileType,
		StorageKey: req.StorageKey,
		URL:        req.URL,
		BucketName: req.BucketName,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := fc.repository.Update(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, fc.collection.ToModel(model))
}

func (fc *MediaController) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid media ID"})
	}
	media, err := fc.repository.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if err := fc.storage.DeleteFile(media.StorageKey); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	model := &collection.Media{ID: id}
	if err := fc.repository.Delete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (fc *MediaController) APIRoutes(service *echo.Echo) {
	service.GET("/media", fc.List)
	service.GET("/media/:id", fc.Get)
	service.POST("/media", fc.Create)
	service.PUT("/media/:id", fc.Update)
	service.DELETE("/media/:id", fc.Delete)
}
