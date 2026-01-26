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

func InvitationCodeManager(service *horizon.HorizonService) *registry.Registry[
	types.InvitationCode, types.InvitationCodeResponse, types.InvitationCodeRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.InvitationCode, types.InvitationCodeResponse, types.InvitationCodeRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Organization.Media",
			"Branch.Media",
			"Branch",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InvitationCode) *types.InvitationCodeResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}

			return &types.InvitationCodeResponse{
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

				UserType:       data.UserType,
				Code:           data.Code,
				ExpirationDate: data.ExpirationDate.Format(time.RFC3339),
				MaxUse:         data.MaxUse,
				CurrentUse:     data.CurrentUse,
				Description:    data.Description,

				PermissionName:        data.PermissionName,
				PermissionDescription: data.PermissionDescription,
				Permissions:           data.Permissions,
			}
		},
		Created: func(data *types.InvitationCode) registry.Topics {
			return []string{
				"invitation_code.create",
				fmt.Sprintf("invitation_code.create.%s", data.ID),
				fmt.Sprintf("invitation_code.create.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InvitationCode) registry.Topics {
			return []string{
				"invitation_code.update",
				fmt.Sprintf("invitation_code.update.%s", data.ID),
				fmt.Sprintf("invitation_code.update.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InvitationCode) registry.Topics {
			return []string{
				"invitation_code.delete",
				fmt.Sprintf("invitation_code.delete.%s", data.ID),
				fmt.Sprintf("invitation_code.delete.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func invitationCodeSeed(context context.Context, service *horizon.HorizonService,
	tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	expiration := now.AddDate(0, 1, 0)

	invitationCodes := []*types.InvitationCode{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       types.UserOrganizationTypeEmployee,
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         5,
			CurrentUse:     0,
			Description:    "Invitation code for employees (max 5 uses)",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       types.UserOrganizationTypeMember,
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         1000,
			CurrentUse:     0,
			Description:    "Invitation code for members (max 1000 uses)",
		},
	}
	for _, data := range invitationCodes {
		if err := InvitationCodeManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed invitation code for %s", data.UserType)
		}
	}
	return nil
}

func GetInvitationCodeByBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InvitationCode, error) {
	return InvitationCodeManager(service).Find(context, &types.InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func GetInvitationCodeByCode(context context.Context, service *horizon.HorizonService, code string) (*types.InvitationCode, error) {
	return InvitationCodeManager(service).FindOne(context, &types.InvitationCode{
		Code: code,
	})
}

func VerifyInvitationCodeByCode(context context.Context, service *horizon.HorizonService, code string) (*types.InvitationCode, error) {
	data, err := GetInvitationCodeByCode(context, service, code)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if now.After(data.ExpirationDate) {
		return nil, eris.Errorf("invitation code %q expired on %s", code, data.ExpirationDate.Format(time.RFC3339))
	}
	if data.CurrentUse >= data.MaxUse {
		return nil, eris.Errorf(
			"invitation code %q has already been used %d times (max %d)",
			code, data.CurrentUse, data.MaxUse,
		)
	}
	return data, nil
}

func RedeemInvitationCode(context context.Context, service *horizon.HorizonService, tx *gorm.DB, invitationCodeID uuid.UUID) error {
	data, err := InvitationCodeManager(service).GetByIDLock(context, tx, invitationCodeID)
	if err != nil {
		return eris.Wrap(err, "failed to lock invitation code for redemption")
	}
	data.CurrentUse++
	if err := InvitationCodeManager(service).UpdateByIDWithTx(context, tx, data.ID, data); err != nil {
		return eris.Wrapf(
			err,
			"failed to redeem invitation code %q (increment CurrentUse)",
			data.Code,
		)
	}

	return nil
}
