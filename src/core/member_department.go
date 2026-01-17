package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func MemberDepartmentManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberDepartment, types.MemberDepartmentResponse, types.MemberDepartmentRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberDepartment, types.MemberDepartmentResponse, types.MemberDepartmentRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", },
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberDepartment) *types.MemberDepartmentResponse {
			if data == nil {
				return nil
			}
			return &types.MemberDepartmentResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				Name:           data.Name,
				Description:    data.Description,
				Icon:           data.Icon,
			}
		},
		Created: func(data *types.MemberDepartment) registry.Topics {
			return []string{
				"member_department.create",
				fmt.Sprintf("member_department.create.%s", data.ID),
				fmt.Sprintf("member_department.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberDepartment) registry.Topics {
			return []string{
				"member_department.update",
				fmt.Sprintf("member_department.update.%s", data.ID),
				fmt.Sprintf("member_department.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberDepartment) registry.Topics {
			return []string{
				"member_department.delete",
				fmt.Sprintf("member_department.delete.%s", data.ID),
				fmt.Sprintf("member_department.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberDepartmentSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberDepartments := []*types.MemberDepartment{
		{
			Name:        "General Services",
			Description: "Handles general inquiries, member registration, and basic services.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Loans and Credit",
			Description: "Manages loan applications, approvals, and credit services.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Savings and Deposits",
			Description: "Handles savings accounts, deposits, and withdrawal transactions.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Accounting and Finance",
			Description: "Manages financial records, budgeting, and accounting operations.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Member Relations",
			Description: "Focuses on member engagement, complaints, and relationship management.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Information Technology",
			Description: "Manages IT systems, software, hardware, and technical support.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Marketing and Communications",
			Description: "Handles promotional activities, communications, and public relations.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Human Resources",
			Description: "Manages employee relations, recruitment, and organizational development.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Audit and Compliance",
			Description:    "Ensures regulatory compliance and conducts internal audits.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Legal Affairs",
			Description: "Handles legal matters, contracts, and regulatory compliance.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Insurance Services",
			Description: "Manages insurance products and services for members.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Education and Training",
			Description:    "Provides financial literacy and cooperative education programs.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Security and Risk Management",
			Description:    "Handles security protocols and risk assessment procedures.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Operations Management",
			Description:    "Oversees daily operations and process improvement initiatives.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
	}

	for _, data := range memberDepartments {
		if err := MemberDepartmentManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member department %s", data.Name)
		}
	}
	return nil
}

func MemberDepartmentCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberDepartment, error) {
	return MemberDepartmentManager(service).Find(context, &types.MemberDepartment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
