package model

import (
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Media struct {
		ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		FileName   string `gorm:"type:varchar(2048);unsigned" json:"file_name"`
		FileSize   int64  `gorm:"unsigned" json:"file_size"`
		FileType   string `gorm:"type:varchar(50);unsigned" json:"file_type"`
		StorageKey string `gorm:"type:varchar(2048)" json:"storage_key"`
		URL        string `gorm:"type:varchar(2048);unsigned" json:"url"`
		Key        string `gorm:"type:varchar(2048)" json:"key"`
		BucketName string `gorm:"type:varchar(2048)" json:"bucket_name"`
		Status     string `gorm:"type:varchar(50);default:'pending'" json:"status"`
		Progress   int64  `gorm:"unsigned" json:"progress"`
	}

	MediaResponse struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
		FileName    string    `json:"file_name"`
		FileSize    int64     `json:"file_size"`
		FileType    string    `json:"file_type"`
		StorageKey  string    `json:"storage_key"`
		URL         string    `json:"url"`
		Key         string    `json:"key"`
		DownloadURL string    `json:"download_url"`
		BucketName  string    `json:"bucket_name"`
		Status      string    `json:"status"`
		Progress    int64     `json:"progress"`
	}

	MediaRequest struct {
		ID       *uuid.UUID `json:"id,omitempty"`
		FileName string     `json:"file_name" validate:"required,max=255"`
	}
)

func (m *Model) Media() {
	m.Migration = append(m.Migration, &Media{})
	m.MediaManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Media, MediaResponse, MediaRequest]{
		Preloads: nil,
		Service:  m.provider.Service,
		Resource: func(data *Media) *MediaResponse {
			if data == nil {
				return nil
			}
			return &MediaResponse{
				ID:         data.ID,
				CreatedAt:  data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:  data.UpdatedAt.Format(time.RFC3339),
				FileName:   data.FileName,
				FileSize:   data.FileSize,
				FileType:   data.FileType,
				StorageKey: data.StorageKey,
				URL:        data.URL,
				Key:        data.Key,
				BucketName: data.BucketName,
				Status:     data.Status,
				Progress:   data.Progress,
			}
		},
		Created: func(data *Media) []string {
			return []string{
				"media.create",
				"media.create." + data.ID.String(),
			}
		},
		Updated: func(data *Media) []string {
			return []string{
				"media.update",
				"media.update." + data.ID.String(),
			}
		},
		Deleted: func(data *Media) []string {
			return []string{
				"media.delete",
				"media.delete." + data.ID.String(),
			}
		},
	})
}
