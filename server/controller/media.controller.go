package controller

import (
	"fmt"
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
}

func NewMediaController(
	repository *repository.MediaRepository,
	collection *collection.MediaCollection,
	storage *horizon.HorizonStorage,
) (*MediaController, error) {
	return &MediaController{
		repository: repository,
		collection: collection,
		storage:    storage,
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
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	media, err := fc.storage.UploadFromHeader(file, func(progress int64, total int64, storage *horizon.Storage) {
		fmt.Println(total)
		fmt.Println(progress)
		fmt.Println(storage)
		fmt.Println("--------")
	})
	model := &collection.Media{
		FileName:   media.FileName,
		FileSize:   media.FileSize,
		FileType:   media.FileType,
		StorageKey: media.StorageKey,
		URL:        media.URL,
		BucketName: media.BucketName,
	}
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if err := fc.repository.Create(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, fc.collection.ToModel(model))

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
