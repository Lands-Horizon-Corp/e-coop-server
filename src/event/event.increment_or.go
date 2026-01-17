package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Automatically increments
func IncrementOfficialReceipt(
	context context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	referenceNumber string,
	source types.GeneralLedgerSource,
	userOrg *types.UserOrganization,
) error {
	branchSetting, err := core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{})
	if err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to find branch setting")
	}
	userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
	if err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get user organization by ID")
	}
	switch source {
	case types.GeneralLedgerSourcePayment:
		if userOrganization.PaymentORUseDateOR || branchSetting.WithdrawCommonOR != "" {
			break
		}
		if userOrganization.PaymentORUnique {
			transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
			if err != nil {
				return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get current transaction batch")
			}
			payments, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				TransactionBatchID: &transactionBatch.ID,
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				Source:             types.GeneralLedgerSourcePayment,
				ReferenceNumber:    referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find payments")
			}
			if len(payments) > 0 {
				return eris.New("IncrementOfficialReceipt: payment with the same reference number already exists")
			}
		}
		userOrganization.PaymentORCurrent++
		if userOrganization.PaymentORCurrent > userOrganization.PaymentOREnd {
			userOrganization.PaymentORIteration++
			userOrganization.PaymentORCurrent = userOrganization.PaymentORStart
		}
	case types.GeneralLedgerSourceWithdraw:
		if branchSetting.WithdrawUseDateOR || branchSetting.WithdrawCommonOR != "" {
			break
		}
		branchSetting.WithdrawORCurrent++
		if branchSetting.WithdrawORCurrent > branchSetting.WithdrawOREnd {
			branchSetting.WithdrawORIteration++
			branchSetting.WithdrawORCurrent = branchSetting.WithdrawORStart
		}
	case types.GeneralLedgerSourceDeposit:
		if branchSetting.DepositUseDateOR || branchSetting.DepositCommonOR != "" {
			break
		}
		branchSetting.DepositORCurrent++
		if branchSetting.DepositORCurrent > branchSetting.DepositOREnd {
			branchSetting.DepositORIteration++
			branchSetting.DepositORCurrent = branchSetting.DepositORStart
		}
	case types.GeneralLedgerSourceJournalVoucher:
		if branchSetting.JournalVoucherORUnique {
			journalVouchers, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceJournalVoucher,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find journal vouchers")
			}
			if len(journalVouchers) > 0 {
				return eris.New("IncrementOfficialReceipt: journal voucher with the same reference number already exists")
			}
		}
		branchSetting.JournalVoucherORCurrent++
	case types.GeneralLedgerSourceAdjustment:
		if branchSetting.AdjustmentVoucherORUnique {
			adjustments, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceAdjustment,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find adjustment vouchers")
			}
			if len(adjustments) > 0 {
				return eris.New("IncrementOfficialReceipt: adjustments with the same reference number already exists")
			}
		}
		branchSetting.AdjustmentVoucherORCurrent++
	case types.GeneralLedgerSourceCheckVoucher:
		if branchSetting.CheckVoucherGeneral {
			break
		}
		if branchSetting.CashCheckVoucherORUnique {
			checkVouchers, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceCheckVoucher,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find check vouchers")
			}
			if len(checkVouchers) > 0 {
				return eris.New("IncrementOfficialReceipt: check voucher with the same reference number already exists")
			}
		}
		branchSetting.CashCheckVoucherORCurrent++
	case types.GeneralLedgerSourceLoan:
		if branchSetting.CheckVoucherGeneral {
			break
		}
		if branchSetting.LoanVoucherORUnique {
			loans, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceLoan,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find loan vouchers")
			}
			if len(loans) > 0 {
				return eris.New("IncrementOfficialReceipt: loan with the same reference number already exists")
			}
		}
		branchSetting.LoanVoucherORCurrent++
	}

	if branchSetting.CheckVoucherGeneral {
		if branchSetting.CheckVoucherGeneralORUnique {
			loans, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceLoan,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: (general) failed to find loans")
			}
			if len(loans) > 0 {
				return eris.New("IncrementOfficialReceipt: (general) loan with the same reference number already exists")
			}
			checkVouchers, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          types.GeneralLedgerSourceCheckVoucher,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: (general) failed to find check vouchers")
			}
			if len(checkVouchers) > 0 {
				return eris.New("IncrementOfficialReceipt: (general) check voucher with the same reference number already exists")
			}
		}
		branchSetting.CheckVoucherGeneralORCurrent++
	}
	if err := core.BranchSettingManager(service).UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update branch setting with transaction")
	}
	if err := core.UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update user organization with transaction")
	}
	return nil
}
