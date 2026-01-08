package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

type Controller struct {
	provider *server.Provider
	core     *core.Core
	event    *event.Event
}

func NewController(
	provider *server.Provider,
	core *core.Core,
	event *event.Event,

) (*Controller, error) {
	return &Controller{
		provider: provider,
		core:     core,
		event:    event,
	}, nil
}

func (c *Controller) Start() error {
	c.heartbeat()
	c.formGeneratorController()
	c.authenticationController()
	c.branchController()
	c.categoryController()
	c.contactController()
	c.currencyController()
	c.feedbackController()
	c.footstepController()
	c.generatedReports()
	c.invitationCode()
	c.mediaController()
	c.notificationController()
	c.organizationController()
	c.organizationDailyUsage()
	c.permissionTemplateController()
	c.qRCodeController()
	c.subscriptionPlanController()
	c.timesheetController()
	c.userController()
	c.userOrganinzationController()
	c.userRatingController()
	c.memberProfileMediaController()
	c.tagTemplateController()

	c.memberGenderController()
	c.memberCenterController()
	c.memberTypeController()
	c.memberClassificationController()
	c.memberOccupationController()
	c.memberGroupController()
	c.memberProfileController()

	c.memberEducationalAttainmentController()
	c.memberAddressController()
	c.memberContactReferenceController()
	c.memberAssetController()
	c.memberIncomeController()
	c.memberExpenseController()
	c.memberGovernmentBenefitController()
	c.memberJointAccountController()
	c.memberRelativeAccountController()

	c.bankController()
	c.cancelledCashCheckVoucherController()
	c.cashCheckVoucherController()
	c.cashCheckVoucherTagController()
	c.journalVoucherTagController()
	c.adjustmentTagController()
	c.holidayController()
	c.billAndCoinsController()

	c.transactionBatchController()
	c.cashCountController()
	c.batchFundingController()
	c.checkRemittanceController()
	c.transactionController()
	c.onlineRemittanceController()
	c.disbursementController()
	c.disbursementTransactionController()

	c.accountController()
	c.accountHistory()
	c.memberAccountingLedgerController()
	c.accountCategoryController()
	c.accountClassificationController()
	c.generalLedgerController()
	c.generalLedgerGroupingController()
	c.financialStatementController()
	c.accountTagController()
	c.paymentTypeController()
	c.paymentController()

	c.loanStatusController()
	c.loanPurposeController()
	c.loanTransactionController()
	c.loanTransactionEntryController()
	c.collateralController()
	c.computationSheetController()
	c.automaticLoanDeductionController()
	c.browseExcludeIncludeAccountsController()
	c.includeNegativeAccountController()
	c.memberDepartmentController()
	c.loanTagController()

	c.timeDepositTypeController()
	c.timeDepositComputationController()
	c.timeDepositComputationPreMatureController()

	c.chargesRateSchemeController()
	c.chargesRateByRangeOrMinimumAmountController()
	c.chargesRateSchemeModeOfPaymentController()
	c.journalVoucherController()
	c.adjustmentEntryController()
	c.companyController()

	c.organizationMediaController()
	c.fundsController()
	c.chargesRateByTermController()
	c.memberProfileArchiveController()
	c.commonController()
	c.memberProfileComaker()
	c.browseReferenceController()
	c.generateSavingsInterest()
	c.generatedSavingsInterestEntryController()
	c.mutualFundsController()
	c.mutualFundEntryController()
	c.kycController()
	c.accountTransactionController()
	return nil
}
