package core

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
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
		FileType   string `gorm:"type:varchar(255);unsigned" json:"file_type"`
		StorageKey string `gorm:"type:varchar(2048)" json:"storage_key"`

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

func (m *Core) media() {
	m.Migration = append(m.Migration, &Media{})
	m.MediaManager = *registry.NewRegistry(registry.RegistryParams[Media, MediaResponse, MediaRequest]{
		Preloads: nil,
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Media) *MediaResponse {
			context := context.Background()
			if data == nil {
				return nil
			}
			temporary, err := m.provider.Service.Storage.GeneratePresignedURL(
				context, &horizon.Storage{
					StorageKey: data.StorageKey,
					BucketName: data.BucketName,
					FileName:   data.FileName,
				}, time.Hour)
			if err != nil {
				temporary = ""
			}
			return &MediaResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
				FileName:    data.FileName,
				FileSize:    data.FileSize,
				FileType:    data.FileType,
				StorageKey:  data.StorageKey,
				Key:         data.Key,
				BucketName:  data.BucketName,
				Status:      data.Status,
				Progress:    data.Progress,
				DownloadURL: temporary,
			}
		},
		Created: func(data *Media) registry.Topics {
			return []string{
				"media.create",
				"media.create." + data.ID.String(),
			}
		},
		Updated: func(data *Media) registry.Topics {
			return []string{
				"media.update",
				"media.update." + data.ID.String(),
			}
		},
		Deleted: func(data *Media) registry.Topics {
			return []string{
				"media.delete",
				"media.delete." + data.ID.String(),
			}
		},
	})
}

func (m *Core) MediaDelete(context context.Context, mediaID uuid.UUID) error {
	if mediaID == uuid.Nil {
		return nil
	}
	media, err := m.MediaManager.GetByID(context, mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return nil
	}
	if err := m.MediaManager.Delete(context, media.ID); err != nil {
		return err
	}
	if err := m.provider.Service.Storage.DeleteFile(context, &horizon.Storage{
		FileName:   media.FileName,
		FileSize:   media.FileSize,
		FileType:   media.FileType,
		StorageKey: media.StorageKey,
		BucketName: media.BucketName,
		Status:     "delete",
	}); err != nil {
		return err
	}

	return nil

}
