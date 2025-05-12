package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Footstep struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_footstep"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_footstep"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID  *uuid.UUID `gorm:"type:uuid"`
		User    *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Description    string    `gorm:"type:text;not null"`
		Activity       string    `gorm:"type:text;not null"`
		AccountType    string    `gorm:"type:varchar(11);unsigned" json:"account_type"`
		Module         string    `gorm:"type:varchar(255);unsigned" json:"module"`
		Latitude       *float64  `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
		Longitude      *float64  `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
		Timestamp      time.Time `gorm:"type:timestamp" json:"timestamp"`
		IsDeleted      bool      `gorm:"default:false" json:"is_deleted"`
		IPAddress      string    `gorm:"type:varchar(45)" json:"ip_address"`
		UserAgent      string    `gorm:"type:varchar(1000)" json:"user_agent"`
		Referer        string    `gorm:"type:varchar(1000)" json:"referer"`
		Location       string    `gorm:"type:varchar(255)" json:"location"`
		AcceptLanguage string    `gorm:"type:varchar(255)" json:"accept_language"`
	}

	FootstepResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		UserID  *uuid.UUID     `json:"user_id,omitempty"`
		User    *UserResponse  `json:"user,omitempty"`
		MediaID *uuid.UUID     `json:"media_id,omitempty"`
		Media   *MediaResponse `json:"media,omitempty"`

		Description    string   `json:"description"`
		Activity       string   `json:"activity"`
		AccountType    string   `json:"account_type"`
		Module         string   `json:"module"`
		Latitude       *float64 `json:"latitude,omitempty"`
		Longitude      *float64 `json:"longitude,omitempty"`
		Timestamp      string   `json:"timestamp"`
		IsDeleted      bool     `json:"is_deleted"`
		IPAddress      string   `json:"ip_address"`
		UserAgent      string   `json:"user_agent"`
		Referer        string   `json:"referer"`
		Location       string   `json:"location"`
		AcceptLanguage string   `json:"accept_language"`
	}

	FootstepCollection struct {
		Manager CollectionManager[Footstep]
	}
)

func (m *Model) FootstepModel(data *Footstep) *FootstepResponse {
	return ToModel(data, func(data *Footstep) *FootstepResponse {
		return &FootstepResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.UpdatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),
			UserID:         data.UserID,
			User:           m.UserModel(data.User),
			MediaID:        data.MediaID,
			Media:          m.MediaModel(data.Media),
			AccountType:    data.AccountType,
			Module:         data.Module,
			Latitude:       data.Latitude,
			Longitude:      data.Longitude,
			Timestamp:      data.Timestamp.Format(time.RFC3339),
			IsDeleted:      data.IsDeleted,
			IPAddress:      data.IPAddress,
			UserAgent:      data.UserAgent,
			Referer:        data.Referer,
			Location:       data.Location,
			AcceptLanguage: data.AcceptLanguage,
		}
	})
}

func (m *Model) FootstepModels(data []*Footstep) []*FootstepResponse {
	return ToModels(data, m.FootstepModel)
}
