package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type (
	RoleTemplate struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_org_branch,unique"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_org_branch,unique"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string         `gorm:"type:varchar(255);not null"`
		Description string         `gorm:"type:text"`
		Permissions pq.StringArray `gorm:"type:varchar[];default:'{}'"`
	}

	RoleTemplateRequest struct {
		OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
		BranchID       uuid.UUID `json:"branch_id" validate:"required"`
		Name           string    `json:"name" validate:"required,min=1,max=255"`
		Description    string    `json:"description,omitempty"`
		Permissions    []string  `json:"permissions,omitempty"`
	}

	RoleTemplateResponse struct {
		ID             uuid.UUID             `json:"id"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`
		Name           string                `json:"name"`
		Description    string                `json:"description,omitempty"`
		Permissions    []string              `json:"permissions"`
		CreatedAt      string                `json:"created_at"`
		UpdatedAt      string                `json:"updated_at"`
	}

	RoleTemplateCollection struct {
		validator       *validator.Validate
		organizationCol *OrganizationCollection
		branchCol       *BranchCollection
	}
)

func NewRoleTemplateCollection(
	organizationCol *OrganizationCollection,
	branchCol *BranchCollection,
) *RoleTemplateCollection {
	return &RoleTemplateCollection{
		validator:       validator.New(),
		organizationCol: organizationCol,
		branchCol:       branchCol,
	}
}

func (rtc *RoleTemplateCollection) ValidateCreate(c echo.Context) (*RoleTemplateRequest, error) {
	req := new(RoleTemplateRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := rtc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (rtc *RoleTemplateCollection) ToModel(rt *RoleTemplate) *RoleTemplateResponse {
	if rt == nil {
		return nil
	}
	return &RoleTemplateResponse{
		ID:             rt.ID,
		OrganizationID: rt.OrganizationID,
		Organization:   rtc.organizationCol.ToModel(rt.Organization),
		BranchID:       rt.BranchID,
		Branch:         rtc.branchCol.ToModel(rt.Branch),
		Name:           rt.Name,
		Description:    rt.Description,
		Permissions:    rt.Permissions,
		CreatedAt:      rt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      rt.UpdatedAt.Format(time.RFC3339),
	}
}

func (rtc *RoleTemplateCollection) ToModels(data []*RoleTemplate) []*RoleTemplateResponse {
	if len(data) == 0 {
		return []*RoleTemplateResponse{}
	}
	out := make([]*RoleTemplateResponse, 0, len(data))
	for _, rt := range data {
		if m := rtc.ToModel(rt); m != nil {
			out = append(out, m)
		}
	}
	return out
}
