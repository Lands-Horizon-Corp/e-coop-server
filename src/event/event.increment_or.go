package event

import (
	"context"
	"fmt"

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
	fmt.Println("Entering IncrementOfficialReceipt function")
	branchSetting, err := core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{
		BranchID: *userOrg.BranchID,
	})
	fmt.Println("Fetched branchSetting")
	if err != nil {
		fmt.Println("Error fetching branchSetting")
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to find branch setting")
	}
	userOrganization, err := core.UserOrganizationManager(service).GetByID(context, userOrg.ID)
	fmt.Println("Fetched userOrganization")
	if err != nil {
		fmt.Println("Error fetching userOrganization")
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get user organization by ID")
	}
	fmt.Println("Starting switch on source")
	switch source {
	case types.GeneralLedgerSourcePayment:
		fmt.Println("Case: GeneralLedgerSourcePayment")
		if userOrganization.PaymentORUseDateOR || branchSetting.WithdrawCommonOR != "" {
			fmt.Println("Condition met for PaymentORUseDateOR or WithdrawCommonOR, breaking")
			break
		}
		if userOrganization.PaymentORUnique {
			fmt.Println("PaymentORUnique is true")
			transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
			fmt.Println("Fetched transactionBatch")
			if err != nil {
				fmt.Println("Error fetching transactionBatch")
				return eris.Wrapf(err, "IncrementOfficialReceipt: failed to get current transaction batch")
			}
			payments, err := core.GeneralLedgerManager(service).Find(context, &types.GeneralLedger{
				TransactionBatchID: &transactionBatch.ID,
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				Source:             types.GeneralLedgerSourcePayment,
				ReferenceNumber:    referenceNumber,
			})
			fmt.Println("Fetched payments")
			if err != nil {
				fmt.Println("Error fetching payments")
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find payments")
			}
			if len(payments) > 0 {
				fmt.Println("Payments with same reference number exist")
				return eris.New("IncrementOfficialReceipt: payment with the same reference number already exists")
			}
		}
		userOrganization.PaymentORCurrent++
		fmt.Println("Incremented PaymentORCurrent")
		if userOrganization.PaymentORCurrent > userOrganization.PaymentOREnd {
			fmt.Println("PaymentORCurrent exceeded PaymentOREnd")
			userOrganization.PaymentORIteration++
			userOrganization.PaymentORCurrent = userOrganization.PaymentORStart
			fmt.Println("Incremented PaymentORIteration and reset PaymentORCurrent")
		}
	case types.GeneralLedgerSourceWithdraw:
		fmt.Println("Case: GeneralLedgerSourceWithdraw")
		if branchSetting.WithdrawUseDateOR || branchSetting.WithdrawCommonOR != "" {
			fmt.Println("Condition met for WithdrawUseDateOR or WithdrawCommonOR, breaking")
			break
		}
		branchSetting.WithdrawORCurrent++
		fmt.Println("Incremented WithdrawORCurrent")
		if branchSetting.WithdrawORCurrent > branchSetting.WithdrawOREnd {
			fmt.Println("WithdrawORCurrent exceeded WithdrawOREnd")
			branchSetting.WithdrawORIteration++
			branchSetting.WithdrawORCurrent = branchSetting.WithdrawORStart
			fmt.Println("Incremented WithdrawORIteration and reset WithdrawORCurrent")
		}
	case types.GeneralLedgerSourceDeposit:
		fmt.Println("Case: GeneralLedgerSourceDeposit")
		if branchSetting.DepositUseDateOR || branchSetting.DepositCommonOR != "" {
			fmt.Println("Condition met for DepositUseDateOR or DepositCommonOR, breaking")
			break
		}
		branchSetting.DepositORCurrent++
		fmt.Println("Incremented DepositORCurrent")
		if branchSetting.DepositORCurrent > branchSetting.DepositOREnd {
			fmt.Println("DepositORCurrent exceeded DepositOREnd")
			branchSetting.DepositORIteration++
			branchSetting.DepositORCurrent = branchSetting.DepositORStart
			fmt.Println("Incremented DepositORIteration and reset DepositORCurrent")
		}
	case types.GeneralLedgerSourceJournalVoucher:
		fmt.Println("Case: GeneralLedgerSourceJournalVoucher")
		if branchSetting.JournalVoucherORUnique {
			fmt.Println("JournalVoucherORUnique is true")
			journalVouchers, err := core.JournalVoucherManager(service).Find(context, &types.JournalVoucher{
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				CashVoucherNumber: referenceNumber,
			})
			fmt.Println("Fetched journalVouchers")
			if err != nil {
				fmt.Println("Error fetching journalVouchers")
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find journal vouchers")
			}
			if len(journalVouchers) > 0 {
				fmt.Println("Journal vouchers with same reference number exist")
				return eris.New("IncrementOfficialReceipt: journal voucher with the same reference number already exists")
			}
		}
		branchSetting.JournalVoucherORCurrent++
		fmt.Println("Incremented JournalVoucherORCurrent")
	case types.GeneralLedgerSourceAdjustment:
		fmt.Println("Case: GeneralLedgerSourceAdjustment")
		if branchSetting.AdjustmentVoucherORUnique {
			fmt.Println("AdjustmentVoucherORUnique is true")
			adjustments, err := core.AdjustmentEntryManager(service).Find(context, &types.AdjustmentEntry{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				ReferenceNumber: referenceNumber,
			})
			fmt.Println("Fetched adjustments")
			if err != nil {
				fmt.Println("Error fetching adjustments")
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find adjustment vouchers")
			}
			if len(adjustments) > 0 {
				fmt.Println("Adjustments with same reference number exist")
				return eris.New("IncrementOfficialReceipt: adjustments with the same reference number already exists")
			}
		}
		branchSetting.AdjustmentVoucherORCurrent++
		fmt.Println("Incremented AdjustmentVoucherORCurrent")
	case types.GeneralLedgerSourceCheckVoucher:
		fmt.Println("Case: GeneralLedgerSourceCheckVoucher")
		if branchSetting.CheckVoucherGeneral {
			fmt.Println("CheckVoucherGeneral is true, breaking")
			break
		}
		if branchSetting.CashCheckVoucherORUnique {
			fmt.Println("CashCheckVoucherORUnique is true")
			checkVouchers, err := core.CashCheckVoucherManager(service).Find(context, &types.CashCheckVoucher{
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				CashVoucherNumber: referenceNumber,
			})
			fmt.Println("Fetched checkVouchers")
			if err != nil {
				fmt.Println("Error fetching checkVouchers")
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find check vouchers")
			}
			if len(checkVouchers) > 0 {
				fmt.Println("Check vouchers with same reference number exist")
				return eris.New("IncrementOfficialReceipt: check voucher with the same reference number already exists")
			}
		}
		branchSetting.CashCheckVoucherORCurrent++
		fmt.Println("Incremented CashCheckVoucherORCurrent")
	case types.GeneralLedgerSourceLoan:
		fmt.Println("Case: GeneralLedgerSourceLoan")
		if branchSetting.CheckVoucherGeneral {
			fmt.Println("CheckVoucherGeneral is true, breaking")
			break
		}
		if branchSetting.LoanVoucherORUnique {
			fmt.Println("LoanVoucherORUnique is true")
			loans, err := core.LoanTransactionManager(service).Find(context, &types.LoanTransaction{
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Voucher:        referenceNumber,
			})
			fmt.Println("Fetched loans")
			if err != nil {
				fmt.Println("Error fetching loans")
				return eris.Wrap(err, "IncrementOfficialReceipt: failed to find loan vouchers")
			}
			if len(loans) > 0 {
				fmt.Println("Loans with same reference number exist")
				return eris.New("IncrementOfficialReceipt: loan with the same reference number already exists")
			}
		}
		branchSetting.LoanVoucherORCurrent++
		fmt.Println("Incremented LoanVoucherORCurrent")
	}

	fmt.Println("Checking CheckVoucherGeneral")
	fmt.Printf("Source: %v\n", source)
	fmt.Printf("CheckVoucherGeneral: %t\n", branchSetting.CheckVoucherGeneral)
	if branchSetting.CheckVoucherGeneral && (source == types.GeneralLedgerSourceLoan || source == types.GeneralLedgerSourceCheckVoucher) {
		fmt.Println("CheckVoucherGeneral is true")
		if branchSetting.CheckVoucherGeneralORUnique {
			fmt.Println("CheckVoucherGeneralORUnique is true")
			loans, err := core.LoanTransactionManager(service).Find(context, &types.LoanTransaction{
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
				Voucher:        referenceNumber,
			})
			fmt.Println("Fetched loans (general)")
			if err != nil {
				fmt.Println("Error fetching loans (general)")
				return eris.Wrap(err, "IncrementOfficialReceipt: (general) failed to find loans")
			}
			if len(loans) > 0 {
				fmt.Println("Loans (general) with same reference number exist")
				return eris.New("IncrementOfficialReceipt: (general) loan with the same reference number already exists")
			}
			checkVouchers, err := core.CashCheckVoucherManager(service).Find(context, &types.CashCheckVoucher{
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				CashVoucherNumber: referenceNumber,
			})
			fmt.Println("Fetched checkVouchers (general)")
			if err != nil {
				fmt.Println("Error fetching checkVouchers (general)")
				return eris.Wrap(err, "IncrementOfficialReceipt: (general) failed to find check vouchers")
			}
			if len(checkVouchers) > 0 {
				fmt.Println("Check vouchers (general) with same reference number exist")
				return eris.New("IncrementOfficialReceipt: (general) check voucher with the same reference number already exists")
			}
		}
		branchSetting.CheckVoucherGeneralORCurrent++
		fmt.Println("Incremented CheckVoucherGeneralORCurrent")
	}
	fmt.Println("Updating branchSetting")
	if err := core.BranchSettingManager(service).UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
		fmt.Println("Error updating branchSetting")
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update branch setting with transaction")
	}
	fmt.Println("Updating userOrganization")
	if err := core.UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		fmt.Println("Error updating userOrganization")
		return eris.Wrapf(err, "IncrementOfficialReceipt: failed to update user organization with transaction")
	}
	fmt.Println("Exiting IncrementOfficialReceipt function successfully")
	return nil
}
