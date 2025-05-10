package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"gorm.io/gorm"
)

type (
	PermissionTemplate struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string         `gorm:"type:varchar(255);not null"`
		Description string         `gorm:"type:text"`
		Permissions pq.StringArray `gorm:"type:varchar[];default:'{}'"`
	}

	PermissionTemplateRequest struct {
		OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
		BranchID       uuid.UUID `json:"branch_id" validate:"required"`
		Name           string    `json:"name" validate:"required,min=1,max=255"`
		Description    string    `json:"description,omitempty"`
		Permissions    []string  `json:"permissions,omitempty"`
	}

	PermissionTemplateResponse struct {
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

		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions"`
	}
)

// func NewPermissionTemplateCollection(
// 	organizationCol *OrganizationCollection,
// 	branchCol *BranchCollection,
// 	userCol *UserCollection,
// ) *PermissionTemplateCollection {
// 	return &PermissionTemplateCollection{
// 		validator:       validator.New(),
// 		organizationCol: organizationCol,
// 		branchCol:       branchCol,
// 		userCol:         userCol,
// 	}
// }

// func (rtc *PermissionTemplateCollection) ValidateCreate(c echo.Context) (*PermissionTemplateRequest, error) {
// 	req := new(PermissionTemplateRequest)
// 	if err := c.Bind(req); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := rtc.validator.Struct(req); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return req, nil
// }

// func (rtc *PermissionTemplateCollection) ToModel(rt *PermissionTemplate) *PermissionTemplateResponse {
// 	if rt == nil {
// 		return nil
// 	}
// 	return &PermissionTemplateResponse{
// 		ID:             rt.ID,
// 		CreatedAt:      rt.CreatedAt.Format(time.RFC3339),
// 		CreatedByID:    rt.CreatedByID,
// 		CreatedBy:      rtc.userCol.ToModel(rt.CreatedBy),
// 		UpdatedAt:      rt.UpdatedAt.Format(time.RFC3339),
// 		UpdatedByID:    rt.UpdatedByID,
// 		UpdatedBy:      rtc.userCol.ToModel(rt.UpdatedBy),
// 		OrganizationID: rt.OrganizationID,
// 		Organization:   rtc.organizationCol.ToModel(rt.Organization),
// 		BranchID:       rt.BranchID,
// 		Branch:         rtc.branchCol.ToModel(rt.Branch),

// 		Name:        rt.Name,
// 		Description: rt.Description,
// 		Permissions: rt.Permissions,
// 	}
// }

// func (rtc *PermissionTemplateCollection) ToModels(data []*PermissionTemplate) []*PermissionTemplateResponse {
// 	if len(data) == 0 {
// 		return []*PermissionTemplateResponse{}
// 	}
// 	out := make([]*PermissionTemplateResponse, 0, len(data))
// 	for _, rt := range data {
// 		if m := rtc.ToModel(rt); m != nil {
// 			out = append(out, m)
// 		}
// 	}
// 	return out
// }
