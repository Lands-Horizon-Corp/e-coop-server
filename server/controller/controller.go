package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

type Controller struct {
	// Services
	provider  *server.Provider
	modelcore *modelcore.ModelCore
	event     *event.Event
	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	usecase               *usecase.TransactionService
}

func NewController(
	// Services
	provider *server.Provider,
	modelcore *modelcore.ModelCore,
	event *event.Event,

	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken,
	userToken *tokens.UserToken,
	usecase *usecase.TransactionService,

) (*Controller, error) {
	return &Controller{
		// Services
		provider:  provider,
		modelcore: modelcore,
		event:     event,

		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		usecase:               usecase,
	}, nil
}

func (c *Controller) Start() error {
	// Others
	c.heartbeat()
	c.formGeneratorController()
	c.authenticationController()
	// Basic Onboardding & Utilities
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

	// Member Profile
	c.memberGenderController()
	c.memberCenterController()
	c.memberTypeController()
	c.memberClassificationController()
	c.memberOccupationController()
	c.memberGroupController()
	c.memberProfileController()
	c.memberTypeReferenceController()

	// member profile properties
	c.memberEducationalAttainmentController()
	c.memberAddressController()
	c.memberContactReferenceController()
	c.memberAssetController()
	c.memberIncomeController()
	c.memberExpenseController()
	c.memberGovernmentBenefitController()
	c.memberJointAccountController()
	c.memberRelativeAccountController()

	// Account Maintenance
	c.bankController()
	c.cancelledCashCheckVoucherController()
	c.cashCheckVoucherController()
	c.cashCheckVoucherTagController()
	c.journalVoucherTagController()
	c.adjustmentTagController()
	c.holidayController()
	c.billAndCoinsController()

	// Transaction batch
	c.transactionBatchController()
	c.cashCountController()
	c.batchFundingController()
	c.checkRemittanceController()
	c.transactionController()
	c.onlineRemittanceController()
	c.disbursementController()
	c.disbursementTransactionController()

	// Accounts
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

	// Loans
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

	// Time deposit
	c.timeDepositTypeController()
	c.timeDepositComputationController()
	c.timeDepositComputationPreMatureController()

	// Charges rate scheme
	c.chargesRateSchemeController()
	c.chargesRateByRangeOrMinimumAmountController()
	c.chargesRateSchemeModeOfPaymentController()
	c.journalVoucherController()
	c.adjustmentEntryController()
	c.companyController()

	c.organizationMediaController()
	c.fundsController()
	c.chargesRateByTermController()
	return nil
}
