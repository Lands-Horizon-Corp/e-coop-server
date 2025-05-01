package models

import "time"

type Feedback struct {
	ID           string     `gorm:"primaryKey;type:varchar(50);unique;default:uuid_generate_v4()"`
	Email        string     `gorm:"type:varchar(255)"`
	Description  string     `gorm:"type:text"`
	FeedbackType string     `gorm:"type:varchar(50);not null;default:'general'"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()"`
	DeletedAt    *time.Time `gorm:"index"`
}

// updated
// deleted
// created
