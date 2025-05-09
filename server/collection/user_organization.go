package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	UserOrganization struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_org_branch,unique"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_org_branch,unique"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserType               string    `gorm:"type:varchar(50);not null"`
		UserID                 uuid.UUID `gorm:"type:uuid;not null"`
		User                   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
		Description            string    `gorm:"type:text" json:"description,omitempty"`
		ApplicationDescription string    `gorm:"type:text" json:"application_description,omitempty"`
		ApplicationStatus      string    `gorm:"type:varchar(50);not null;default:'pending'" json:"application_status"`

		DeveloperSecretKey string `gorm:"type:varchar(255);not null;unique" json:"developer_secret_key"`
	}

	UserOrganizationRequest struct {
		UserType               string    `json:"user_type" validate:"required,oneof=employee owner member"`
		OrganizationID         uuid.UUID `json:"organization_id" validate:"required"`
		BranchID               uuid.UUID `json:"branch_id" validate:"required"`
		UserID                 uuid.UUID `json:"user_id" validate:"required"`
		Description            string    `json:"description,omitempty"`
		ApplicationDescription string    `json:"application_description,omitempty"`
		ApplicationStatus      string    `json:"application_status" validate:"required,oneof=pending reported accepted ban"`
	}

	UserOrganizationResponse struct {
		ID                     uuid.UUID             `json:"id"`
		UserType               string                `json:"user_type"`
		OrganizationID         uuid.UUID             `json:"organization_id"`
		Organization           *OrganizationResponse `json:"organization,omitempty"`
		BranchID               uuid.UUID             `json:"branch_id"`
		Branch                 *BranchResponse       `json:"branch,omitempty"`
		UserID                 uuid.UUID             `json:"user_id"`
		User                   *UserResponse         `json:"user,omitempty"`
		Description            string                `json:"description,omitempty"`
		ApplicationDescription string                `json:"application_description,omitempty"`
		ApplicationStatus      string                `json:"application_status"`
		DeveloperSecretKey     string                `json:"developer_secret_key"`
		CreatedAt              string                `json:"created_at"`
		UpdatedAt              string                `json:"updated_at"`

		// ROles : []
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
		ID:                     uo.ID,
		UserType:               uo.UserType,
		OrganizationID:         uo.OrganizationID,
		Organization:           uoc.organizationCol.ToModel(uo.Organization),
		BranchID:               uo.BranchID,
		Branch:                 uoc.branchCol.ToModel(uo.Branch),
		UserID:                 uo.UserID,
		User:                   uoc.userCol.ToModel(uo.User),
		Description:            uo.Description,
		ApplicationDescription: uo.ApplicationDescription,
		ApplicationStatus:      uo.ApplicationStatus,
		CreatedAt:              uo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              uo.UpdatedAt.Format(time.RFC3339),
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
