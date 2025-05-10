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
	UserOrganization struct {
		ID                     uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt              time.Time      `gorm:"not null;default:now()"`
		CreatedByID            uuid.UUID      `gorm:"type:uuid"`
		CreatedBy              *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt              time.Time      `gorm:"not null;default:now()"`
		UpdatedByID            uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy              *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt              gorm.DeletedAt `gorm:"index"`
		DeletedByID            *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy              *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_org_branch"`
		Organization           *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID               uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_org_branch"`
		Branch                 *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`
		UserID                 uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_org_branch"`
		User                   *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
		UserType               string         `gorm:"type:varchar(50);not null"`
		Description            string         `gorm:"type:text" json:"description,omitempty"`
		ApplicationDescription string         `gorm:"type:text" json:"application_description,omitempty"`
		ApplicationStatus      string         `gorm:"type:varchar(50);not null;default:'pending'" json:"application_status"`
		DeveloperSecretKey     string         `gorm:"type:varchar(255);not null;unique" json:"developer_secret_key"`
		PermissionName         string         `gorm:"type:varchar(255);not null" json:"permission_name"`
		PermissionDescription  string         `gorm:"type:varchar(255);not null" json:"permission_description"`
		Permissions            pq.StringArray `gorm:"type:varchar[];default:'{}'"`
	}

	UserOrganizationRequest struct {
		UserType               string         `json:"user_type" validate:"required,oneof=employee owner member"`
		OrganizationID         uuid.UUID      `json:"organization_id" validate:"required"`
		BranchID               uuid.UUID      `json:"branch_id" validate:"required"`
		UserID                 uuid.UUID      `json:"user_id" validate:"required"`
		Description            string         `json:"description,omitempty"`
		ApplicationDescription string         `json:"application_description,omitempty"`
		ApplicationStatus      string         `json:"application_status" validate:"required,oneof=pending reported accepted ban"`
		PermissionName         string         `json:"permission_name" validate:"required"`
		PermissionDescription  string         `json:"permission_description" validate:"required"`
		Permissions            pq.StringArray `json:"permissions,omitempty" validate:"dive,required"`
	}

	UserOrganizationResponse struct {
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

		UserType               string        `json:"user_type"`
		UserID                 uuid.UUID     `json:"user_id"`
		User                   *UserResponse `json:"user,omitempty"`
		Description            string        `json:"description,omitempty"`
		ApplicationDescription string        `json:"application_description,omitempty"`
		ApplicationStatus      string        `json:"application_status"`
		DeveloperSecretKey     string        `json:"developer_secret_key"`
		PermissionName         string        `json:"permission_name"`
		PermissionDescription  string        `json:"permission_description"`
		Permissions            []string      `json:"permissions"`
	}

	UserOrganizationCollection struct {
		validator       *validator.Validate
		organizationCol *OrganizationCollection
		branchCol       *BranchCollection
		userCol         *UserCollection
	}
)

func NewUserOrganizationCollection(
	organizationCol *OrganizationCollection,
	branchCol *BranchCollection,
	userCol *UserCollection,
) (*UserOrganizationCollection, error) {
	return &UserOrganizationCollection{
		validator:       validator.New(),
		organizationCol: organizationCol,
		branchCol:       branchCol,
		userCol:         userCol,
	}, nil
}

func (uoc *UserOrganizationCollection) ValidateCreate(c echo.Context) (*UserOrganizationRequest, error) {
	req := new(UserOrganizationRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := uoc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (uoc *UserOrganizationCollection) ToModel(uo *UserOrganization) *UserOrganizationResponse {
	if uo == nil {
		return nil
	}
	return &UserOrganizationResponse{
		ID:             uo.ID,
		CreatedAt:      uo.CreatedAt.Format(time.RFC3339),
		CreatedByID:    uo.CreatedByID,
		CreatedBy:      uoc.userCol.ToModel(uo.CreatedBy),
		UpdatedAt:      uo.UpdatedAt.Format(time.RFC3339),
		UpdatedByID:    uo.UpdatedByID,
		UpdatedBy:      uoc.userCol.ToModel(uo.UpdatedBy),
		OrganizationID: uo.OrganizationID,
		Organization:   uoc.organizationCol.ToModel(uo.Organization),
		BranchID:       uo.BranchID,
		Branch:         uoc.branchCol.ToModel(uo.Branch),

		UserType:               uo.UserType,
		UserID:                 uo.UserID,
		User:                   uoc.userCol.ToModel(uo.User),
		Description:            uo.Description,
		ApplicationDescription: uo.ApplicationDescription,
		ApplicationStatus:      uo.ApplicationStatus,
		DeveloperSecretKey:     uo.DeveloperSecretKey,
		PermissionName:         uo.PermissionName,
		PermissionDescription:  uo.PermissionDescription,
		Permissions:            uo.Permissions,
	}
}

// ToModels maps a slice of UserOrganization models to the response format
func (uoc *UserOrganizationCollection) ToModels(data []*UserOrganization) []*UserOrganizationResponse {
	if data == nil {
		return []*UserOrganizationResponse{}
	}
	var out []*UserOrganizationResponse
	for _, uo := range data {
		if m := uoc.ToModel(uo); m != nil {
			out = append(out, m)
		}
	}
	return out
}
