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

func UserRatingManager(service *horizon.HorizonService) *registry.Registry[types.UserRating, types.UserRatingResponse, types.UserRatingRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.UserRating, types.UserRatingResponse, types.UserRatingRequest]{
		Preloads: []string{"Organization", "Branch", "RateeUser", "RaterUser"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.UserRating) *types.UserRatingResponse {
			if data == nil {
				return nil
			}
			return &types.UserRatingResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				CreatedByID: data.CreatedByID,
				CreatedBy:   UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID: data.UpdatedByID,
				UpdatedBy:   UserManager(service).ToModel(data.UpdatedBy),

				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				RateeUserID: data.RateeUserID,
				RateeUser:   UserManager(service).ToModel(&data.RateeUser),
				RaterUserID: data.RaterUserID,
				RaterUser:   UserManager(service).ToModel(&data.RaterUser),
				Rate:        data.Rate,
				Remark:      data.Remark,
			}
		},
		Created: func(data *types.UserRating) registry.Topics {
			return []string{
				"user_rating.create",
				fmt.Sprintf("user_rating.create.%s", data.ID),
				fmt.Sprintf("user_rating.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.UserRating) registry.Topics {
			return []string{
				"user_rating.update",
				fmt.Sprintf("user_rating.update.%s", data.ID),
				fmt.Sprintf("user_rating.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.UserRating) registry.Topics {
			return []string{
				"user_rating.delete",
				fmt.Sprintf("user_rating.delete.%s", data.ID),
				fmt.Sprintf("user_rating.delete.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetUserRatee(context context.Context, service *horizon.HorizonService, userID uuid.UUID) ([]*types.UserRating, error) {
	return UserRatingManager(service).Find(context, &types.UserRating{
		RateeUserID: userID,
	})
}

func GetUserRater(context context.Context, service *horizon.HorizonService, userID uuid.UUID) ([]*types.UserRating, error) {
	return UserRatingManager(service).Find(context, &types.UserRating{
		RaterUserID: userID,
	})
}

func UserRatingCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.UserRating, error) {
	return UserRatingManager(service).Find(context, &types.UserRating{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
