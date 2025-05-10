package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	Media struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		FileName   string                `gorm:"type:varchar(2048);unsigned" json:"file_name"`
		FileSize   int64                 `gorm:"unsigned" json:"file_size"`
		FileType   string                `gorm:"type:varchar(50);unsigned" json:"file_type"`
		StorageKey string                `gorm:"type:varchar(2048)" json:"storage_key"`
		URL        string                `gorm:"type:varchar(2048);unsigned" json:"url"`
		Key        string                `gorm:"type:varchar(2048)" json:"key"`
		BucketName string                `gorm:"type:varchar(2048)" json:"bucket_name"`
		Status     horizon.StorageStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
		Progress   int64                 `gorm:"unsigned" json:"progress"`
	}

	MediaResponse struct {
		ID          uuid.UUID             `json:"id"`
		CreatedAt   string                `json:"created_at"`
		UpdatedAt   string                `json:"updated_at"`
		FileName    string                `json:"file_name"`
		FileSize    int64                 `json:"file_size"`
		FileType    string                `json:"file_type"`
		StorageKey  string                `json:"storage_key"`
		URL         string                `json:"url"`
		Key         string                `json:"key"`
		DownloadURL string                `json:"download_url"`
		BucketName  string                `json:"bucket_name"`
		Status      horizon.StorageStatus `json:"status"`
		Progress    int64                 `json:"progress"`
	}

	MediaRequest struct {
		ID         *uint  `json:"id"`
		FileName   string `json:"file_name" validate:"required,max=255"`
		FileSize   int64  `json:"file_size" validate:"required,min=1"`
		FileType   string `json:"file_type" validate:"required,max=50"`
		StorageKey string `json:"storage_key" validate:"required,max=255"`
		URL        string `json:"url" validate:"required,url,max=255"`
		Key        string `json:"key,omitempty" validate:"max=255"`
		BucketName string `json:"bucket_name,omitempty" validate:"max=255"`
		Progress   int64  `json:"status"`
	}
)

func (m *Model) MediaValidate(ctx echo.Context) (*MediaRequest, error) {
	return Validate[MediaRequest](ctx, m.validator)
}

func (m *Model) MediaModel(data *Media) *MediaResponse {
	temporaryURL, err := m.storage.GeneratePresignedURL(data.StorageKey)
	if err != nil {
		temporaryURL = ""
	}
	return ToModel(data, func(data *Media) *MediaResponse {
		return &MediaResponse{
			ID:          data.ID,
			CreatedAt:   data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
			FileName:    data.FileName,
			FileSize:    data.FileSize,
			FileType:    data.FileType,
			StorageKey:  data.StorageKey,
			URL:         data.URL,
			Key:         data.Key,
			BucketName:  data.BucketName,
			DownloadURL: temporaryURL,
			Status:      data.Status,
			Progress:    data.Progress,
		}
	})
}

func (m *Model) MediaModels(data []*Media) []*MediaResponse {
	return ToModels(data, m.MediaModel)
}

// func NewMediaCollection(
// 	storage *horizon.HorizonStorage,

// ) (*MediaCollection, error) {
// 	return &MediaCollection{
// 		validator: validator.New(),
// 		storage:   storage,
// 	}, nil
// }

// func (fc *MediaCollection) ValidateCreate(c echo.Context) (*MediaRequest, error) {
// 	u := new(MediaRequest)
// 	if err := c.Bind(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := fc.validator.Struct(u); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return u, nil
// }

// func (m *MediaCollection) ToModel(media *Media) *MediaResponse {
// 	if media == nil {
// 		return nil
// 	}
// 	temporaryURL, err := m.storage.GeneratePresignedURL(media.StorageKey)
// 	if err != nil {
// 		temporaryURL = ""
// 	}
// 	return &MediaResponse{
// 		ID:        media.ID,
// 		CreatedAt: media.CreatedAt.Format(time.RFC3339),
// 		UpdatedAt: media.UpdatedAt.Format(time.RFC3339),

// 		FileName:    media.FileName,
// 		FileSize:    media.FileSize,
// 		FileType:    media.FileType,
// 		StorageKey:  media.StorageKey,
// 		URL:         media.URL,
// 		Key:         media.Key,
// 		BucketName:  media.BucketName,
// 		DownloadURL: temporaryURL,
// 		Status:      media.Status,
// 		Progress:    media.Progress,
// 	}
// }

// func (m *MediaCollection) ToModels(mediaList []*Media) []*MediaResponse {
// 	if mediaList == nil {
// 		return make([]*MediaResponse, 0)
// 	}
// 	var mediaResponses []*MediaResponse
// 	for _, media := range mediaList {
// 		model := m.ToModel(media)
// 		if model != nil {
// 			mediaResponses = append(mediaResponses, m.ToModel(media))
// 		}
// 	}
// 	if len(mediaResponses) <= 0 {
// 		return make([]*MediaResponse, 0)
// 	}
// 	return mediaResponses
// }
