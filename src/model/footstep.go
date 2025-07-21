package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Footstep struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID *uuid.UUID     `gorm:"type:uuid;index:idx_branch_org_footstep"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID     `gorm:"type:uuid;index:idx_branch_org_footstep"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID  *uuid.UUID `gorm:"type:uuid"`
		User    *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Description    string    `gorm:"type:text;not null"`
		Activity       string    `gorm:"type:text;not null"`
		UserType       string    `gorm:"type:varchar(11);unsigned" json:"user_type"`
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
		OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       *uuid.UUID            `json:"branch_id,omitempty"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		UserID  *uuid.UUID     `json:"user_id,omitempty"`
		User    *UserResponse  `json:"user,omitempty"`
		MediaID *uuid.UUID     `json:"media_id,omitempty"`
		Media   *MediaResponse `json:"media,omitempty"`

		Description    string   `json:"description"`
		Activity       string   `json:"activity"`
		UserType       string   `json:"user_type"`
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
)

func (m *Model) Footstep() {
	m.Migration = append(m.Migration, &Footstep{})
	m.FootstepManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Footstep, FootstepResponse, any]{
		Preloads: []string{
			"User",
			"User.Media",
			"Branch",
			"Branch.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Media",
		},
		Service: m.provider.Service,
		Resource: func(data *Footstep) *FootstepResponse {
			if data == nil {
				return nil
			}
			return &FootstepResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),

				UserID:  data.UserID,
				User:    m.UserManager.ToModel(data.User),
				MediaID: data.MediaID,
				Media:   m.MediaManager.ToModel(data.Media),

				Description:    data.Description,
				Activity:       data.Activity,
				UserType:       data.UserType,
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
		},
		Created: func(data *Footstep) []string {
			return []string{
				"footstep.create",
				fmt.Sprintf("footstep.create.%s", data.ID),
				fmt.Sprintf("footstep.create.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Footstep) []string {
			return []string{
				"footstep.update",
				fmt.Sprintf("footstep.update.%s", data.ID),
				fmt.Sprintf("footstep.update.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Footstep) []string {
			return []string{
				"footstep.delete",
				fmt.Sprintf("footstep.delete.%s", data.ID),
				fmt.Sprintf("footstep.delete.branch.%s", data.BranchID),
				fmt.Sprintf("footstep.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) GetFootstepByUser(context context.Context, userId uuid.UUID) ([]*Footstep, error) {
	return m.FootstepManager.Find(context, &Footstep{
		UserID: &userId,
	})
}

func (m *Model) GetFootstepByBranch(context context.Context, organizationId uuid.UUID, branchId uuid.UUID) ([]*Footstep, error) {
	return m.FootstepManager.Find(context, &Footstep{
		OrganizationID: &organizationId,
		BranchID:       &branchId,
	})
}

func (m *Model) GetFootstepByUserOrganization(context context.Context, userId uuid.UUID, organizationId uuid.UUID, branchId uuid.UUID) ([]*Footstep, error) {
	return m.FootstepManager.Find(context, &Footstep{
		UserID:         &userId,
		OrganizationID: &organizationId,
		BranchID:       &branchId,
	})
}
