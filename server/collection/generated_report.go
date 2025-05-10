package collection

import (
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneratedReport struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID      *uuid.UUID `gorm:"type:uuid"`
		User        *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID     *uuid.UUID `gorm:"type:uuid;not null"`
		Media       *Media     `gorm:"foreignKey:MediaID"`
		Name        string     `gorm:"type:varchar(255);not null"`
		Description string     `gorm:"type:text;not null"`
		Status      string     `gorm:"type:varchar(50);not null"`
		Progress    float64    `gorm:"not null"`
	}
	GeneratedReportResponse struct {
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

		UserID      *uuid.UUID     `json:"user_id"`
		User        *UserResponse  `json:"user"`
		MediaID     *uuid.UUID     `json:"media_id"`
		Media       *MediaResponse `json:"media,omitempty"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Status      string         `json:"status"`
		Progress    float64        `json:"progress"`
	}

	GeneratedReportRequest struct {
		Name        string `json:"firstName" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1"`
	}

	GeneratedReportCollection struct {
		validator *validator.Validate
		branchCol *BranchCollection
		orgCol    *OrganizationCollection
		mediaCol  *MediaCollection
		userCol   *UserCollection
	}
)

func NewGeneratedReportCollection(
	branchCol *BranchCollection,
	orgCol *OrganizationCollection,
	mediaCol *MediaCollection,
	userCol *UserCollection,
) *GeneratedReportCollection {
	return &GeneratedReportCollection{
		validator: validator.New(),
		branchCol: branchCol,
		orgCol:    orgCol,
		mediaCol:  mediaCol,
		userCol:   userCol,
	}
}

func (c *GeneratedReportCollection) ToModel(m *GeneratedReport) *GeneratedReportResponse {
	if m == nil {
		return nil
	}

	return &GeneratedReportResponse{
		ID:             m.ID,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		CreatedByID:    m.CreatedByID,
		CreatedBy:      c.userCol.ToModel(m.CreatedBy),
		UpdatedAt:      m.UpdatedAt.Format(time.RFC3339),
		UpdatedByID:    m.UpdatedByID,
		UpdatedBy:      c.userCol.ToModel(m.UpdatedBy),
		OrganizationID: m.OrganizationID,
		Organization:   c.orgCol.ToModel(m.Organization),
		BranchID:       m.BranchID,
		Branch:         c.branchCol.ToModel(m.Branch),
		UserID:         m.UserID,
		User:           c.userCol.ToModel(m.User),
		MediaID:        m.MediaID,
		Media:          c.mediaCol.ToModel(m.Media),
		Name:           m.Name,
		Description:    m.Description,
		Status:         m.Status,
		Progress:       m.Progress,
	}
}

func (c *GeneratedReportCollection) ToModels(data []*GeneratedReport) []*GeneratedReportResponse {
	if data == nil {
		return []*GeneratedReportResponse{}
	}

	responses := make([]*GeneratedReportResponse, 0, len(data))
	for _, report := range data {
		if response := c.ToModel(report); response != nil {
			responses = append(responses, response)
		}
	}
	return responses
}
