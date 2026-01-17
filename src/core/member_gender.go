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

func MemberGenderManager(service *horizon.HorizonService) *registry.Registry[types.MemberGender, types.MemberGenderResponse, types.MemberGenderRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberGender, types.MemberGenderResponse, types.MemberGenderRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberGender) *types.MemberGenderResponse {
			if data == nil {
				return nil
			}
			return &types.MemberGenderResponse{
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
			}
		},

		Created: func(data *types.MemberGender) registry.Topics {
			return []string{
				"member_gender.create",
				fmt.Sprintf("member_gender.create.%s", data.ID),
				fmt.Sprintf("member_gender.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberGender) registry.Topics {
			return []string{
				"member_gender.update",
				fmt.Sprintf("member_gender.update.%s", data.ID),
				fmt.Sprintf("member_gender.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberGender) registry.Topics {
			return []string{
				"member_gender.delete",
				fmt.Sprintf("member_gender.delete.%s", data.ID),
				fmt.Sprintf("member_gender.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberGenderSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberGenders := []*types.MemberGender{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Male",
			Description:    "Identifies as male.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Female",
			Description:    "Identifies as female.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Other",
			Description:    "Identifies outside the binary gender categories.",
		},
	}
	branchSetting, err := BranchSettingManager(service).FindOne(context, &types.BranchSetting{
		BranchID: branchID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to get branch setting on member gender seed")
	}
	for _, data := range memberGenders {

		if err := MemberGenderManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member gender %s", data.Name)
		}
		if data.Name == "Female" {
			branchSetting.DefaultMemberGenderID = &data.ID
			if err := BranchSettingManager(service).UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
				return eris.Wrap(err, "failed to update branch settings with paid up share capital and cash on hand accounts")
			}
		}
	}
	return nil
}

func MemberGenderCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberGender, error) {
	return MemberGenderManager(service).Find(context, &types.MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
