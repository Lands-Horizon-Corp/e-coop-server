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
)

func TransactionManager(service *horizon.HorizonService) *registry.Registry[types.Transaction, types.TransactionResponse, types.TransactionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.Transaction, types.TransactionResponse, types.TransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch",
			"Organization", "SignatureMedia", "TransactionBatch", "EmployeeUser",
			"MemberProfile",
			"MemberProfile.Media",
			"MemberJointAccount.PictureMedia",
			"MemberJointAccount.SignatureMedia",
			"Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Transaction) *types.TransactionResponse {
			if data == nil {
				return nil
			}
			return &types.TransactionResponse{
				ID:                   data.ID,
				CreatedAt:            data.CreatedAt.Format(time.RFC3339),
				CreatedByID:          data.CreatedByID,
				CreatedBy:            UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:            data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:          data.UpdatedByID,
				UpdatedBy:            UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:       data.OrganizationID,
				Organization:         OrganizationManager(service).ToModel(data.Organization),
				BranchID:             data.BranchID,
				Branch:               BranchManager(service).ToModel(data.Branch),
				SignatureMediaID:     data.SignatureMediaID,
				SignatureMedia:       MediaManager(service).ToModel(data.SignatureMedia),
				TransactionBatchID:   data.TransactionBatchID,
				TransactionBatch:     TransactionBatchManager(service).ToModel(data.TransactionBatch),
				EmployeeUserID:       data.EmployeeUserID,
				EmployeeUser:         UserManager(service).ToModel(data.EmployeeUser),
				MemberProfileID:      data.MemberProfileID,
				MemberProfile:        MemberProfileManager(service).ToModel(data.MemberProfile),
				MemberJointAccountID: data.MemberJointAccountID,
				MemberJointAccount:   MemberJointAccountManager(service).ToModel(data.MemberJointAccount),
				LoanBalance:          data.LoanBalance,
				LoanDue:              data.LoanDue,
				TotalDue:             data.TotalDue,
				FinesDue:             data.FinesDue,
				TotalLoan:            data.TotalLoan,
				InterestDue:          data.InterestDue,
				ReferenceNumber:      data.ReferenceNumber,
				Amount:               data.Amount,
				Description:          data.Description,
				CurrencyID:           data.CurrencyID,
				Currency:             CurrencyManager(service).ToModel(data.Currency),
			}
		},

		Created: func(data *types.Transaction) registry.Topics {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.create.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.create.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.create",
				fmt.Sprintf("transaction.create.%s", data.ID),
				fmt.Sprintf("transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.create.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
		Updated: func(data *types.Transaction) registry.Topics {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.update.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.update.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.update",
				fmt.Sprintf("transaction.update.%s", data.ID),
				fmt.Sprintf("transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.update.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
		Deleted: func(data *types.Transaction) registry.Topics {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.update.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.update.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.delete",
				fmt.Sprintf("transaction.delete.%s", data.ID),
				fmt.Sprintf("transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.delete.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
	})
}

func TransactionCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Transaction, error) {
	return TransactionManager(service).Find(context, &types.Transaction{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func TransactionsByUserType(
	context context.Context,
	service *horizon.HorizonService,
	userID uuid.UUID,
	userType types.UserOrganizationType,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.Transaction, error) {
	var filter types.Transaction

	if userType == types.UserOrganizationTypeMember {
		memberProfile, err := MemberProfileManager(service).FindOne(context, &types.MemberProfile{
			UserID: &userID,
		})
		if err != nil {
			return nil, eris.Wrap(err, "failed to retrieve member profile")
		}
		filter.MemberProfileID = &memberProfile.ID
	} else {
		filter.EmployeeUserID = &userID
	}

	filter.OrganizationID = organizationID
	filter.BranchID = branchID

	return TransactionManager(service).Find(context, &filter)
}
