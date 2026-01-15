package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Automatically increments
func IncrementOfficialReceipt(
	context context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	referenceNumber string,
	source core.GeneralLedgerSource,
	userOrg *core.UserOrganization,
) error {
	branchSetting, err := core.BranchSettingManager(service).FindOne(context, &core.BranchSetting{})
	if err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to find branch setting")
	}
	userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
	if err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get user organization by ID")
	}
	switch source {
	case core.GeneralLedgerSourcePayment:
		if userOrganization.PaymentORUseDateOR || branchSetting.WithdrawCommonOR != "" {
			break
		}
		if userOrganization.PaymentORUnique {
			transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
			if err != nil {
				return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get current transaction batch")
			}
			payments, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				TransactionBatchID: &transactionBatch.ID,
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				Source:             core.GeneralLedgerSourcePayment,
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
	case core.GeneralLedgerSourceWithdraw:
		if branchSetting.WithdrawUseDateOR || branchSetting.WithdrawCommonOR != "" {
			break
		}
		branchSetting.WithdrawORCurrent++
		if branchSetting.WithdrawORCurrent > branchSetting.WithdrawOREnd {
			branchSetting.WithdrawORIteration++
			branchSetting.WithdrawORCurrent = branchSetting.WithdrawORStart
		}
	case core.GeneralLedgerSourceDeposit:
		if branchSetting.DepositUseDateOR || branchSetting.DepositCommonOR != "" {
			break
		}
		branchSetting.DepositORCurrent++
		if branchSetting.DepositORCurrent > branchSetting.DepositOREnd {
			branchSetting.DepositORIteration++
			branchSetting.DepositORCurrent = branchSetting.DepositORStart
		}
	case core.GeneralLedgerSourceJournalVoucher:
		if branchSetting.JournalVoucherORUnique {
			journalVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceJournalVoucher,
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
	case core.GeneralLedgerSourceAdjustment:
		if branchSetting.AdjustmentVoucherORUnique {
			adjustments, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceAdjustment,
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
	case core.GeneralLedgerSourceCheckVoucher:
		if branchSetting.CheckVoucherGeneral {
			break
		}
		if branchSetting.CashCheckVoucherORUnique {
			checkVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceCheckVoucher,
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
	case core.GeneralLedgerSourceLoan:
		if branchSetting.CheckVoucherGeneral {
			break
		}
		if branchSetting.LoanVoucherORUnique {
			loans, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceLoan,
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
			loans, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceLoan,
				ReferenceNumber: referenceNumber,
			})
			if err != nil {
				return eris.Wrap(err, "IncrementOfficialReceipt: (general) failed to find loans")
			}
			if len(loans) > 0 {
				return eris.New("IncrementOfficialReceipt: (general) loan with the same reference number already exists")
			}
			checkVouchers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				Source:          core.GeneralLedgerSourceCheckVoucher,
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
