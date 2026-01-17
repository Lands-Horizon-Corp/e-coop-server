package types

import (
	"time"

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
