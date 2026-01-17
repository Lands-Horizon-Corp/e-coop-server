package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func LoanTransactionManager(service *horizon.HorizonService) *registry.Registry[types.LoanTransaction, types.LoanTransactionResponse, types.LoanTransactionRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanTransaction, types.LoanTransactionResponse, types.LoanTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "EmployeeUser", "EmployeeUser.Media",
			"TransactionBatch", "LoanPurpose", "LoanStatus",
			"ComakerDepositMemberAccountingLedger", "PreviousLoan", "ComakerDepositMemberAccountingLedger.Account",
			"Account",
			"Account.Currency",
			"MemberProfile", "MemberJointAccount", "SignatureMedia", "MemberProfile.Media",
			"MemberProfile.SignatureMedia", "MemberProfile.MemberType",
			"ReleasedBy", "PrintedBy", "ApprovedBy",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
			"LoanTags",
			"LoanTransactionEntries",
			"LoanTransactionEntries.Account",
			"LoanTransactionEntries.Account.Currency",
			"LoanTransactionEntries.AutomaticLoanDeduction",
			"LoanClearanceAnalysis",
			"LoanClearanceAnalysisInstitution",
			"LoanTermsAndConditionSuggestedPayment",
			"LoanTermsAndConditionAmountReceipt", "LoanTermsAndConditionAmountReceipt.Account",
			"ComakerMemberProfiles", "ComakerMemberProfiles.MemberProfile", "ComakerMemberProfiles.MemberProfile.Media",
			"ComakerCollaterals", "ComakerCollaterals.Collateral",
			"PreviousLoan.Account",
			"ReleasedBy", "PrintedBy", "ApprovedBy",
			"LoanAccounts", "LoanAccounts.Account", "LoanAccounts.Account.Currency",
			"ReleasedBy.Media", "PrintedBy.Media", "ApprovedBy.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanTransaction) *types.LoanTransactionResponse {
			if data == nil {
				return nil
			}
			return &types.LoanTransactionResponse{
				ID:                                     data.ID,
				CreatedAt:                              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                            data.CreatedByID,
				CreatedBy:                              UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                            data.UpdatedByID,
				UpdatedBy:                              UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:                         data.OrganizationID,
				Organization:                           OrganizationManager(service).ToModel(data.Organization),
				BranchID:                               data.BranchID,
				Branch:                                 BranchManager(service).ToModel(data.Branch),
				EmployeeUserID:                         data.EmployeeUserID,
				EmployeeUser:                           UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID:                     data.TransactionBatchID,
				TransactionBatch:                       TransactionBatchManager(service).ToModel(data.TransactionBatch),
				OfficialReceiptNumber:                  data.OfficialReceiptNumber,
				Voucher:                                data.Voucher,
				CheckDate:                              data.CheckDate,
				CheckNumber:                            data.CheckNumber,
				LoanPurposeID:                          data.LoanPurposeID,
				LoanPurpose:                            LoanPurposeManager(service).ToModel(data.LoanPurpose),
				LoanStatusID:                           data.LoanStatusID,
				LoanStatus:                             LoanStatusManager(service).ToModel(data.LoanStatus),
				ModeOfPayment:                          data.ModeOfPayment,
				ModeOfPaymentWeekly:                    data.ModeOfPaymentWeekly,
				ModeOfPaymentSemiMonthlyPay1:           data.ModeOfPaymentSemiMonthlyPay1,
				ModeOfPaymentSemiMonthlyPay2:           data.ModeOfPaymentSemiMonthlyPay2,
				ModeOfPaymentFixedDays:                 data.ModeOfPaymentFixedDays,
				ModeOfPaymentMonthlyExactDay:           data.ModeOfPaymentMonthlyExactDay,
				ComakerType:                            data.ComakerType,
				ComakerDepositMemberAccountingLedgerID: data.ComakerDepositMemberAccountingLedgerID,
				ComakerDepositMemberAccountingLedger:   MemberAccountingLedgerManager(service).ToModel(data.ComakerDepositMemberAccountingLedger),
				CollectorPlace:                         data.CollectorPlace,
				LoanType:                               data.LoanType,
				PreviousLoanID:                         data.PreviousLoanID,
				PreviousLoan:                           LoanTransactionManager(service).ToModel(data.PreviousLoan),
				Terms:                                  data.Terms,
				Amortization:                           data.Amortization,
				IsAddOn:                                data.IsAddOn,
				Applied1:                               data.Applied1,
				Applied2:                               data.Applied2,
				AccountID:                              data.AccountID,
				Account:                                AccountManager(service).ToModel(data.Account),
				MemberProfileID:                        data.MemberProfileID,
				MemberProfile:                          MemberProfileManager(service).ToModel(data.MemberProfile),
				MemberJointAccountID:                   data.MemberJointAccountID,
				MemberJointAccount:                     AccountManager(service).ToModel(data.MemberJointAccount),
				SignatureMediaID:                       data.SignatureMediaID,
				SignatureMedia:                         MediaManager(service).ToModel(data.SignatureMedia),
				MountToBeClosed:                        data.MountToBeClosed,
				DamayanFund:                            data.DamayanFund,
				ShareCapital:                           data.ShareCapital,
				LengthOfService:                        data.LengthOfService,
				ExcludeSunday:                          data.ExcludeSunday,
				ExcludeHoliday:                         data.ExcludeHoliday,
				ExcludeSaturday:                        data.ExcludeSaturday,
				RemarksOtherTerms:                      data.RemarksOtherTerms,
				RemarksPayrollDeduction:                data.RemarksPayrollDeduction,
				RecordOfLoanPaymentsOrLoanStatus:       data.RecordOfLoanPaymentsOrLoanStatus,
				CollateralOffered:                      data.CollateralOffered,
				AppraisedValue:                         data.AppraisedValue,
				AppraisedValueDescription:              data.AppraisedValueDescription,
				PrintedDate:                            data.PrintedDate,
				PrintNumber:                            data.PrintNumber,
				ApprovedDate:                           data.ApprovedDate,
				ReleasedDate:                           data.ReleasedDate,
				ReleasedByID:                           data.ReleasedByID,
				ReleasedBy:                             UserManager(service).ToModel(data.ReleasedBy),
				PrintedByID:                            data.PrintedByID,
				PrintedBy:                              UserManager(service).ToModel(data.PrintedBy),
				ApprovedByID:                           data.ApprovedByID,
				ApprovedBy:                             UserManager(service).ToModel(data.ApprovedBy),
				ApprovedBySignatureMediaID:             data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:               MediaManager(service).ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:                         data.ApprovedByName,
				ApprovedByPosition:                     data.ApprovedByPosition,
				PreparedBySignatureMediaID:             data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:               MediaManager(service).ToModel(data.PreparedBySignatureMedia),
				PreparedByName:                         data.PreparedByName,
				PreparedByPosition:                     data.PreparedByPosition,
				CertifiedBySignatureMediaID:            data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:              MediaManager(service).ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:                        data.CertifiedByName,
				CertifiedByPosition:                    data.CertifiedByPosition,
				VerifiedBySignatureMediaID:             data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:               MediaManager(service).ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:                         data.VerifiedByName,
				VerifiedByPosition:                     data.VerifiedByPosition,
				CheckBySignatureMediaID:                data.CheckBySignatureMediaID,
				CheckBySignatureMedia:                  MediaManager(service).ToModel(data.CheckBySignatureMedia),
				CheckByName:                            data.CheckByName,
				CheckByPosition:                        data.CheckByPosition,
				AcknowledgeBySignatureMediaID:          data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:            MediaManager(service).ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:                      data.AcknowledgeByName,
				AcknowledgeByPosition:                  data.AcknowledgeByPosition,
				NotedBySignatureMediaID:                data.NotedBySignatureMediaID,
				NotedBySignatureMedia:                  MediaManager(service).ToModel(data.NotedBySignatureMedia),
				NotedByName:                            data.NotedByName,
				NotedByPosition:                        data.NotedByPosition,
				PostedBySignatureMediaID:               data.PostedBySignatureMediaID,
				PostedBySignatureMedia:                 MediaManager(service).ToModel(data.PostedBySignatureMedia),
				PostedByName:                           data.PostedByName,
				PostedByPosition:                       data.PostedByPosition,
				PaidBySignatureMediaID:                 data.PaidBySignatureMediaID,
				PaidBySignatureMedia:                   MediaManager(service).ToModel(data.PaidBySignatureMedia),
				PaidByName:                             data.PaidByName,
				PaidByPosition:                         data.PaidByPosition,
				LoanTags:                               LoanTagManager(service).ToModels(data.LoanTags),
				LoanTransactionEntries:                 mapLoanTransactionEntries(service, data.LoanTransactionEntries),
				LoanClearanceAnalysis:                  LoanClearanceAnalysisManager(service).ToModels(data.LoanClearanceAnalysis),
				LoanClearanceAnalysisInstitution:       LoanClearanceAnalysisInstitutionManager(service).ToModels(data.LoanClearanceAnalysisInstitution),
				LoanTermsAndConditionSuggestedPayment:  LoanTermsAndConditionSuggestedPaymentManager(service).ToModels(data.LoanTermsAndConditionSuggestedPayment),
				LoanTermsAndConditionAmountReceipt:     LoanTermsAndConditionAmountReceiptManager(service).ToModels(data.LoanTermsAndConditionAmountReceipt),
				ComakerMemberProfiles:                  ComakerMemberProfileManager(service).ToModels(data.ComakerMemberProfiles),
				ComakerCollaterals:                     ComakerCollateralManager(service).ToModels(data.ComakerCollaterals),
				LoanAccounts:                           LoanAccountManager(service).ToModels(data.LoanAccounts),
				Count:                                  data.Count,
				Balance:                                data.Balance,
				LastPay:                                data.LastPay,
				Fines:                                  data.Fines,
				Interest:                               data.Interest,
				TotalDebit:                             data.TotalDebit,
				TotalCredit:                            data.TotalCredit,
				Processing:                             data.Processing,
			}
		},

		Created: func(data *types.LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.create",
				fmt.Sprintf("loan_transaction.create.%s", data.ID),
				fmt.Sprintf("loan_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.update",
				fmt.Sprintf("loan_transaction.update.%s", data.ID),
				fmt.Sprintf("loan_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanTransaction) registry.Topics {
			return []string{
				"loan_transaction.delete",
				fmt.Sprintf("loan_transaction.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanTransactionCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanTransaction, error) {
	return LoanTransactionManager(service).Find(context, &types.LoanTransaction{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func mapLoanTransactionEntries(service *horizon.HorizonService, entries []*types.LoanTransactionEntry) []*types.LoanTransactionEntryResponse {
	if entries == nil {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Index < entries[j].Index
	})
	var result []*types.LoanTransactionEntryResponse
	for _, entry := range entries {
		if entry != nil {
			result = append(result, LoanTransactionEntryManager(service).ToModel(entry))
		}
	}
	return result
}

func LoanTransactionWithDatesNotNull(ctx context.Context, service *horizon.HorizonService,
	memberID, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionsMemberAccount(ctx context.Context, service *horizon.HorizonService,
	memberID, accountID, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {

	account, err := AccountManager(service).GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Type != types.AccountTypeLoan {
		accountID = *account.LoanAccountID
	}
	filters := []query.ArrFilterSQL{
		{Field: "member_profile_id", Op: query.ModeEqual, Value: memberID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionDraft(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionPrinted(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionApproved(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionReleased(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}

func LoanTransactionReleasedCurrentDay(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*types.LoanTransaction, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "created_at", Op: query.ModeGTE, Value: startOfDay},
		{Field: "created_at", Op: query.ModeLT, Value: endOfDay},
	}

	return LoanTransactionManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
}
