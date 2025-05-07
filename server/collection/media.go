package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

type (
	Media struct {
		ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time  `gorm:"not null;default:now()"`
		UpdatedAt time.Time  `gorm:"not null;default:now()"`
		DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`

		FileName   string                `gorm:"type:varchar(255);unsigned" json:"file_name"`
		FileSize   int64                 `gorm:"unsigned" json:"file_size"`
		FileType   string                `gorm:"type:varchar(50);unsigned" json:"file_type"`
		StorageKey string                `gorm:"type:varchar(255);unique;unsigned" json:"storage_key"`
		URL        string                `gorm:"type:varchar(255);unsigned" json:"url"`
		Key        string                `gorm:"type:varchar(255)" json:"key"`
		BucketName string                `gorm:"type:varchar(255)" json:"bucket_name"`
		Status     horizon.StorageStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	}

	MediaResponse struct {
		ID          uuid.UUID             `json:"id"`
		CreatedAt   string                `json:"createdAt"`
		UpdatedAt   string                `json:"updatedAt"`
		FileName    string                `json:"fileName"`
		FileSize    int64                 `json:"fileSize"`
		FileType    string                `json:"fileType"`
		StorageKey  string                `json:"storageKey"`
		URL         string                `json:"uRL"`
		Key         string                `json:"key"`
		DownloadURL string                `json:"downloadURL"`
		BucketName  string                `json:"bucketName"`
		Status      horizon.StorageStatus `json:"status"`
	}

	MediaRequest struct {
		ID         *uint  `json:"id"`
		FileName   string `json:"fileName" validate:"required,max=255"`
		FileSize   int64  `json:"fileSize" validate:"required,min=1"`
		FileType   string `json:"fileType" validate:"required,max=50"`
		StorageKey string `json:"storageKey" validate:"required,max=255"`
		URL        string `json:"url" validate:"required,url,max=255"`
		Key        string `json:"key,omitempty" validate:"max=255"`
		BucketName string `json:"bucketName,omitempty" validate:"max=255"`
	}
)

type MediaCollection struct {
	validator *validator.Validate
	storage   *horizon.HorizonStorage
}

func NewMediaCollection(
	storage *horizon.HorizonStorage,
) (*MediaCollection, error) {
	return &MediaCollection{
		validator: validator.New(),
		storage:   storage,
	}, nil
}

func (fc *MediaCollection) ValidateCreate(c echo.Context) (*MediaRequest, error) {
	u := new(MediaRequest)
	if err := c.Bind(u); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := fc.validator.Struct(u); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return u, nil
}

func (m *MediaCollection) ToModel(media *Media) *MediaResponse {
	if media == nil {
		return nil
	}
	temporaryURL, err := m.storage.GeneratePresignedURL(media.StorageKey)
	if err != nil {
		return nil
	}
	return &MediaResponse{
		ID:        media.ID,
		CreatedAt: media.CreatedAt.Format(time.RFC3339),
		UpdatedAt: media.UpdatedAt.Format(time.RFC3339),

		FileName:    media.FileName,
		FileSize:    media.FileSize,
		FileType:    media.FileType,
		StorageKey:  media.StorageKey,
		URL:         media.URL,
		Key:         media.Key,
		BucketName:  media.BucketName,
		DownloadURL: temporaryURL,
		Status:      media.Status,
	}
}

func (m *MediaCollection) ToModels(mediaList []*Media) []*MediaResponse {
	if mediaList == nil {
		return nil
	}
	var mediaResources []*MediaResponse
	for _, media := range mediaList {
		mediaResources = append(mediaResources, m.ToModel(media))
	}
	return mediaResources
}
