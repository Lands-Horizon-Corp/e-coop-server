package types

import (
	"time"

	"github.com/google/uuid"
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_educational_attainment"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_educational_attainment"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		SchoolName            string `gorm:"type:varchar(255)"`
		SchoolYear            int    `gorm:"type:int"`
		ProgramCourse         string `gorm:"type:varchar(255)"`
		EducationalAttainment string `gorm:"type:varchar(255)"`
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
		SchoolName            string    `json:"school_name,omitempty" validate:"required,min=1,max=255"`
		SchoolYear            int       `json:"school_year,omitempty"`
		ProgramCourse         string    `json:"program_course,omitempty"`
		EducationalAttainment string    `json:"educational_attainment,omitempty"`
		Description           string    `json:"description,omitempty"`
	}
)
