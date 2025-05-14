package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	MemberType struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_type"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_type"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Prefix      string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text;not null"`
	}

	MemberTypeResponse struct {
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

		Name        string `json:"name"`
		Prefix      string `json:"prefix"`
		Description string `json:"description"`
	}

	MemberTypeRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		PRefix      string `json:"prefix,omitempty" validate:"max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberTypeCollection struct {
		Manager horizon_manager.CollectionManager[MemberType]
	}
)

func (m *Model) MemberTypeValidate(ctx echo.Context) (*MemberTypeRequest, error) {
	return horizon_manager.Validate[MemberTypeRequest](ctx, m.validator)
}

func (m *Model) MemberTypeModel(data *MemberType) *MemberTypeResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberType) *MemberTypeResponse {
		return &MemberTypeResponse{
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
			Name:           data.Name,
			Prefix:         data.Prefix,
			Description:    data.Description,
		}
	})
}

func NewMemberTypeCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberTypeCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberType) ([]string, any) {
			return []string{
				fmt.Sprintf("member_type.create.%s", data.ID),
				fmt.Sprintf("member_type.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_type.create.organization.%s", data.OrganizationID),
			}, model.MemberTypeModel(data)
		},
		func(data *MemberType) ([]string, any) {
			return []string{
				"member_type.update",
				fmt.Sprintf("member_type.update.%s", data.ID),
				fmt.Sprintf("member_type.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_type.update.organization.%s", data.OrganizationID),
			}, model.MemberTypeModel(data)
		},
		func(data *MemberType) ([]string, any) {
			return []string{
				"member_type.delete",
				fmt.Sprintf("member_type.delete.%s", data.ID),
				fmt.Sprintf("member_type.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_type.delete.organization.%s", data.OrganizationID),
			}, model.MemberTypeModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberTypeCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberTypeModels(data []*MemberType) []*MemberTypeResponse {
	return horizon_manager.ToModels(data, m.MemberTypeModel)
}

// member-type/branch/:branch_id
func (fc *MemberTypeCollection) ListByBranch(branchID uuid.UUID) ([]*MemberType, error) {
	return fc.Manager.Find(&MemberType{
		BranchID: branchID,
	})
}

// member-type/organization/:organization_id
func (fc *MemberTypeCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberType, error) {
	return fc.Manager.Find(&MemberType{
		OrganizationID: organizationID,
	})
}

// member-type/organization/:organization_id/branch/:branch_id
func (fc *MemberTypeCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberType, error) {
	return fc.Manager.Find(&MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (fc *MemberTypeCollection) Seeder(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberType, error) {
	now := time.Now()

	types := []*MemberType{
		{
			ID:          uuid.New(),
			Name:        "New",
			Prefix:      "NEW",
			Description: "Recently registered member, no activity yet.",
		},
		{
			ID:          uuid.New(),
			Name:        "Active",
			Prefix:      "ACT",
			Description: "Regularly engaged member with no issues.",
		},
		{
			ID:          uuid.New(),
			Name:        "Loyal",
			Prefix:      "LOY",
			Description: "Consistently active over a long period; high retention.",
		},
		{
			ID:          uuid.New(),
			Name:        "VIP",
			Prefix:      "VIP",
			Description: "Very high-value member with premium privileges.",
		},
		{
			ID:          uuid.New(),
			Name:        "Reported",
			Prefix:      "RPT",
			Description: "Flagged by community or system for review.",
		},
		{
			ID:          uuid.New(),
			Name:        "Suspended",
			Prefix:      "SUS",
			Description: "Temporarily barred from activities pending resolution.",
		},
		{
			ID:          uuid.New(),
			Name:        "Banned",
			Prefix:      "BAN",
			Description: "Permanently barred due to policy violations.",
		},
		{
			ID:          uuid.New(),
			Name:        "Closed",
			Prefix:      "CLS",
			Description: "Account closed by user request or administrative action.",
		},
	}
	for _, t := range types {
		t.CreatedAt = now
		t.UpdatedAt = now
		t.CreatedByID = userID
		t.UpdatedByID = userID
		t.OrganizationID = organizationID
		t.BranchID = branchID
	}

	if err := fc.Manager.CreateMany(types); err != nil {
		return nil, fmt.Errorf("failed to seed member types: %w", err)
	}
	return types, nil
}
