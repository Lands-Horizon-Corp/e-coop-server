package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func MemberEducationalAttainmentManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberEducationalAttainment, types.MemberEducationalAttainmentResponse, types.MemberEducationalAttainmentRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberEducationalAttainment, types.MemberEducationalAttainmentResponse, types.MemberEducationalAttainmentRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberEducationalAttainment) *types.MemberEducationalAttainmentResponse {
			if data == nil {
				return nil
			}
			return &types.MemberEducationalAttainmentResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          OrganizationManager(service).ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                BranchManager(service).ToModel(data.Branch),
				MemberProfileID:       data.MemberProfileID,
				MemberProfile:         MemberProfileManager(service).ToModel(data.MemberProfile),
				SchoolName:            data.SchoolName,
				SchoolYear:            data.SchoolYear,
				ProgramCourse:         data.ProgramCourse,
				EducationalAttainment: data.EducationalAttainment,
				Description:           data.Description,
			}
		},

		Created: func(data *types.MemberEducationalAttainment) registry.Topics {
			return []string{
				"member_educational_attainment.create",
				fmt.Sprintf("member_educational_attainment.create.%s", data.ID),
				fmt.Sprintf("member_educational_attainment.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_educational_attainment.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberEducationalAttainment) registry.Topics {
			return []string{
				"member_educational_attainment.update",
				fmt.Sprintf("member_educational_attainment.update.%s", data.ID),
				fmt.Sprintf("member_educational_attainment.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_educational_attainment.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberEducationalAttainment) registry.Topics {
			return []string{
				"member_educational_attainment.delete",
				fmt.Sprintf("member_educational_attainment.delete.%s", data.ID),
				fmt.Sprintf("member_educational_attainment.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_educational_attainment.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberEducationalAttainmentCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberEducationalAttainment, error) {
	return MemberEducationalAttainmentManager(service).Find(context, &types.MemberEducationalAttainment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
