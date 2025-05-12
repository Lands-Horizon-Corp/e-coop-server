package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
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
		UserType               string         `json:"user_type" validate:"required,oneof=employee member"`
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

		UserID                 uuid.UUID     `json:"user_id"`
		User                   *UserResponse `json:"user,omitempty"`
		UserType               string        `json:"user_type"`
		Description            string        `json:"description,omitempty"`
		ApplicationDescription string        `json:"application_description,omitempty"`
		ApplicationStatus      string        `json:"application_status"`
		DeveloperSecretKey     string        `json:"developer_secret_key"`
		PermissionName         string        `json:"permission_name"`
		PermissionDescription  string        `json:"permission_description"`
		Permissions            []string      `json:"permissions"`
	}
	UserOrganizationCollection struct {
		Manager CollectionManager[UserOrganization]
	}
)

func (m *Model) UserOrganizationValidate(ctx echo.Context) (*UserOrganizationRequest, error) {
	return Validate[UserOrganizationRequest](ctx, m.validator)
}

func (m *Model) UserOrganizationModel(data *UserOrganization) *UserOrganizationResponse {
	return ToModel(data, func(data *UserOrganization) *UserOrganizationResponse {
		return &UserOrganizationResponse{
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

			UserType:               data.UserType,
			UserID:                 data.UserID,
			User:                   m.UserModel(data.User),
			Description:            data.Description,
			ApplicationDescription: data.ApplicationDescription,
			ApplicationStatus:      data.ApplicationStatus,
			DeveloperSecretKey:     data.DeveloperSecretKey,
			PermissionName:         data.PermissionName,
			PermissionDescription:  data.PermissionDescription,
			Permissions:            data.Permissions,
		}
	})
}

func (m *Model) UserOrganizationModels(data []*UserOrganization) []*UserOrganizationResponse {
	return ToModels(data, m.UserOrganizationModel)
}

func NewUserOrganizationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*UserOrganizationCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.create",
				fmt.Sprintf("user_organization.create.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.update",
				fmt.Sprintf("user_organization.update.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.delete",
				fmt.Sprintf("user_organization.delete.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		[]string{"Branch", "User", "Organization"},
	)
	return &UserOrganizationCollection{
		Manager: manager,
	}, nil
}

// user_organization/user/:user_id
func (fc *UserOrganizationCollection) ListByUser(userID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		UserID: userID,
	})
}

// user_organization/branch/:branch_id
func (fc *UserOrganizationCollection) ListByBranch(branchID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		BranchID: branchID,
	})
}

// user_organization/organization/:organization_id
func (fc *UserOrganizationCollection) ListByOrganization(organizationID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		OrganizationID: organizationID,
	})
}

// user_organization/organization/:organization_id/branch/:branch_id
func (fc *UserOrganizationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// user_organization/user/:user_id/branch/:branch_id
func (fc *UserOrganizationCollection) ListByUserBranch(userID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		UserID:   userID,
		BranchID: branchID,
	})
}

// user_organization/user/:user_id/organization/:organization_id
func (fc *UserOrganizationCollection) ListByUserOrganization(userID uuid.UUID, organizationID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
}

// user_organization/user/:user_id/organization/:organization_id/branch/:branch_id
func (fc *UserOrganizationCollection) ByUserOrganizationBranch(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (*UserOrganization, error) {
	return fc.Manager.FindOne(&UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
