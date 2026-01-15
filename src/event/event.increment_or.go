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
	case core.GeneralLedgerSourceCheckVoucher:
		// TODO: implement logic
	case core.GeneralLedgerSourceJournalVoucher:
		// TODO: implement logic
	case core.GeneralLedgerSourceAdjustment:
		// TODO: implement logic
	case core.GeneralLedgerSourceLoan:
		// TODO: implement logic
	}

	if err := core.BranchSettingManager(service).UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update branch setting with transaction")
	}

	if err := core.UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update user organization with transaction")
	}

	return nil
}
