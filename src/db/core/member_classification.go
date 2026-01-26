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

func MemberClassificationManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberClassification, types.MemberClassificationResponse, types.MemberClassificationRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.MemberClassification, types.MemberClassificationResponse, types.MemberClassificationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberClassification) *types.MemberClassificationResponse {
			if data == nil {
				return nil
			}
			return &types.MemberClassificationResponse{
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
				Icon:           data.Icon,
				Description:    data.Description,
			}
		},

		Created: func(data *types.MemberClassification) registry.Topics {
			return []string{
				"member_classification.create",
				fmt.Sprintf("member_classification.create.%s", data.ID),
				fmt.Sprintf("member_classification.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberClassification) registry.Topics {
			return []string{
				"member_classification.update",
				fmt.Sprintf("member_classification.update.%s", data.ID),
				fmt.Sprintf("member_classification.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberClassification) registry.Topics {
			return []string{
				"member_classification.delete",
				fmt.Sprintf("member_classification.delete.%s", data.ID),
				fmt.Sprintf("member_classification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func memberClassificationSeed(context context.Context,
	service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberClassifications := []*types.MemberClassification{
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Gold",
			Icon:           "sunrise",
			Description:    "Gold membership is reserved for top-tier members with excellent credit scores and consistent loyalty.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Silver",
			Icon:           "moon-star",
			Description:    "Silver membership is designed for members with good credit history and regular engagement.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bronze",
			Icon:           "cloud",
			Description:    "Bronze membership is for new or casual members who are starting their journey with us.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Platinum",
			Icon:           "gem",
			Description:    "Platinum membership offers ZEDE benefits to elite members with outstanding history and contributions.",
		},
	}
	for _, data := range memberClassifications {
		if err := MemberClassificationManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member classification %s", data.Name)
		}
	}
	return nil
}

func MemberClassificationCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.MemberClassification, error) {
	return MemberClassificationManager(service).Find(context, &types.MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
