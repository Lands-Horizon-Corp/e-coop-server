package collection

import (
	"time"

	"github.com/google/uuid"
)

type Feedback struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email        string     `gorm:"type:varchar(255)"`
	Description  string     `gorm:"type:text"`
	FeedbackType string     `gorm:"type:varchar(50);not null;default:'general'"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()"`
	DeletedAt    *time.Time `gorm:"index"`
}
