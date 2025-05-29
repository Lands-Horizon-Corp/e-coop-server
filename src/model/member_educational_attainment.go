package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberEducationalAttainment struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name                  string `gorm:"type:varchar(255);not null"`
		SchoolName            string `gorm:"type:varchar(255)"`
		SchoolYear            int    `gorm:"type:int"`
		ProgramCourse         string `gorm:"type:varchar(255)"`
		EducationalAttainment string `gorm:"type:varchar(255)"` // Bachelors Degree, Phd, Masteral
		Description           string `gorm:"type:text"`
	}

	MemberEducationalAttainmentResponse struct {
		ID                    uuid.UUID              `json:"id"`
		CreatedAt             string                 `json:"created_at"`
		CreatedByID           uuid.UUID              `json:"created_by_id"`
		CreatedBy             *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt             string                 `json:"updated_at"`
		UpdatedByID           uuid.UUID              `json:"updated_by_id"`
		UpdatedBy             *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID        uuid.UUID              `json:"organization_id"`
		Organization          *OrganizationResponse  `json:"organization,omitempty"`
		BranchID              uuid.UUID              `json:"branch_id"`
		Branch                *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID       uuid.UUID              `json:"member_profile_id"`
		MemberProfile         *MemberProfileResponse `json:"member_profile,omitempty"`
		Name                  string                 `json:"name"`
		SchoolName            string                 `json:"school_name"`
		SchoolYear            int                    `json:"school_year"`
		ProgramCourse         string                 `json:"program_course"`
		EducationalAttainment string                 `json:"educational_attainment"`
		Description           string                 `json:"description"`
	}

	MemberEducationalAttainmentRequest struct {
		MemberProfileID       uuid.UUID `json:"member_profile_id" validate:"required"`
		Name                  string    `json:"name" validate:"required,min=1,max=255"`
		SchoolName            string    `json:"school_name,omitempty"`
		SchoolYear            int       `json:"school_year,omitempty"`
		ProgramCourse         string    `json:"program_course,omitempty"`
		EducationalAttainment string    `json:"educational_attainment,omitempty"`
		Description           string    `json:"description,omitempty"`
	}
)

func (m *Model) MemberEducationalAttainment() {
	m.Migration = append(m.Migration, &MemberEducationalAttainment{})
	m.MemberEducationalAttainmentManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberEducationalAttainment, MemberEducationalAttainmentResponse, MemberEducationalAttainmentRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberEducationalAttainment) *MemberEducationalAttainmentResponse {
			if data == nil {
				return nil
			}
			return &MemberEducationalAttainmentResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          m.OrganizationManager.ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                m.BranchManager.ToModel(data.Branch),
				MemberProfileID:       data.MemberProfileID,
				MemberProfile:         m.MemberProfileManager.ToModel(data.MemberProfile),
				Name:                  data.Name,
				SchoolName:            data.SchoolName,
				SchoolYear:            data.SchoolYear,
				ProgramCourse:         data.ProgramCourse,
				EducationalAttainment: data.EducationalAttainment,
				Description:           data.Description,
			}
		},
		Created: func(data *MemberEducationalAttainment) []string {
			return []string{
				"member_educational_attainment.create",
				fmt.Sprintf("member_educational_attainment.create.%s", data.ID),
			}
		},
		Updated: func(data *MemberEducationalAttainment) []string {
			return []string{
				"member_educational_attainment.update",
				fmt.Sprintf("member_educational_attainment.update.%s", data.ID),
			}
		},
		Deleted: func(data *MemberEducationalAttainment) []string {
			return []string{
				"member_educational_attainment.delete",
				fmt.Sprintf("member_educational_attainment.delete.%s", data.ID),
			}
		},
	})
}
