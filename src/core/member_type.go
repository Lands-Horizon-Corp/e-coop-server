package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	MemberType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Prefix                     string `gorm:"type:varchar(255)"`
		Name                       string `gorm:"type:varchar(255)"`
		Description                string `gorm:"type:text"`
		BrowseReferenceDescription string `gorm:"type:text"`

		BrowseReferences []*BrowseReference `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_references,omitempty"`
	}

	MemberTypeResponse struct {
		ID                         uuid.UUID                  `json:"id"`
		CreatedAt                  string                     `json:"created_at"`
		CreatedByID                uuid.UUID                  `json:"created_by_id"`
		CreatedBy                  *UserResponse              `json:"created_by,omitempty"`
		UpdatedAt                  string                     `json:"updated_at"`
		UpdatedByID                uuid.UUID                  `json:"updated_by_id"`
		UpdatedBy                  *UserResponse              `json:"updated_by,omitempty"`
		OrganizationID             uuid.UUID                  `json:"organization_id"`
		Organization               *OrganizationResponse      `json:"organization,omitempty"`
		BranchID                   uuid.UUID                  `json:"branch_id"`
		Branch                     *BranchResponse            `json:"branch,omitempty"`
		Prefix                     string                     `json:"prefix"`
		Name                       string                     `json:"name"`
		Description                string                     `json:"description"`
		BrowseReferenceDescription string                     `json:"browse_reference_description"`
		BrowseReferences           []*BrowseReferenceResponse `json:"browse_references,omitempty"`
	}

	MemberTypeRequest struct {
		Prefix                     string `json:"prefix,omitempty"`
		Name                       string `json:"name,omitempty"`
		Description                string `json:"description,omitempty"`
		BrowseReferenceDescription string `json:"browse_reference_description,omitempty"`
	}
)

func (m *Core) MemberTypeManager() *registry.Registry[MemberType, MemberTypeResponse, MemberTypeRequest] {
	return registry.NewRegistry(registry.RegistryParams[MemberType, MemberTypeResponse, MemberTypeRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Organization",
			"BrowseReferences",
			"BrowseReferences.Account",
			"BrowseReferences.MemberType",
		},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberType) *MemberTypeResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeResponse{
				ID:                         data.ID,
				CreatedAt:                  data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                data.CreatedByID,
				CreatedBy:                  m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:                  data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                data.UpdatedByID,
				UpdatedBy:                  m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:             data.OrganizationID,
				Organization:               m.OrganizationManager().ToModel(data.Organization),
				BranchID:                   data.BranchID,
				Branch:                     m.BranchManager().ToModel(data.Branch),
				Prefix:                     data.Prefix,
				Name:                       data.Name,
				Description:                data.Description,
				BrowseReferenceDescription: data.BrowseReferenceDescription,
				BrowseReferences:           m.BrowseReferenceManager().ToModels(data.BrowseReferences),
			}
		},

		Created: func(data *MemberType) registry.Topics {
			return []string{
				"member_type.create",
				fmt.Sprintf("member_type.create.%s", data.ID),
				fmt.Sprintf("member_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberType) registry.Topics {
			return []string{
				"member_type.update",
				fmt.Sprintf("member_type.update.%s", data.ID),
				fmt.Sprintf("member_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberType) registry.Topics {
			return []string{
				"member_type.delete",
				fmt.Sprintf("member_type.delete.%s", data.ID),
				fmt.Sprintf("member_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) memberTypeSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberType := []*MemberType{
		{

			Name:           "New",
			Prefix:         "NEW",
			Description:    "Recently registered member, no activity yet.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Active",
			Prefix:         "ACT",
			Description:    "Regularly engaged member with no issues.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Loyal",
			Prefix:         "LOY",
			Description:    "Consistently active over a long period; high retention.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "VIP",
			Prefix:         "VIP",
			Description:    "Very high-value member with premium privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Reported",
			Prefix:         "RPT",
			Description:    "Flagged by community or system for review.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Suspended",
			Prefix:         "SUS",
			Description:    "Temporarily barred from activities pending resolution.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Banned",
			Prefix:         "BAN",
			Description:    "Permanently barred due to policy violations.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Closed",
			Prefix:         "CLS",
			Description:    "Account closed by user request or administrative action.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Alumni",
			Prefix:         "ALM",
			Description:    "Former member with notable contributions.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Pending",
			Prefix:         "PND",
			Description:    "Awaiting verification or approval.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Dormant",
			Prefix:         "DRM",
			Description:    "Inactive for a long period with no recent engagement.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Guest",
			Prefix:         "GST",
			Description:    "Limited access member without full privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Moderator",
			Prefix:         "MOD",
			Description:    "Member with special privileges to manage content or users.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Admin",
			Prefix:         "ADM",
			Description:    "Administrator with full access and control.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
	}
	for _, data := range memberType {
		if err := m.MemberTypeManager().CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member type %s", data.Name)
		}
	}
	return nil
}

func (m *Core) MemberTypeCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberType, error) {
	return m.MemberTypeManager().Find(context, &MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
