package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/tokens"
)

type Controller struct {
	// Services
	provider  *src.Provider
	modelcore *modelcore.ModelCore
	event     *event.Event
	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	service               *service.TransactionService
}

func NewController(
	// Services
	provider *src.Provider,
	modelcore *modelcore.ModelCore,
	event *event.Event,

	// Tokens
	userOrganizationToken *tokens.UserOrganizationToken,
	userToken *tokens.UserToken,
	service *service.TransactionService,

) (*Controller, error) {
	return &Controller{
		// Services
		provider:  provider,
		modelcore: modelcore,
		event:     event,

		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		service:               service,
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
	c.MemberAddressController()
	c.MemberContactReferenceController()
	c.MemberAssetController()
	c.MemberIncomeController()
	c.MemberExpenseController()
	c.MemberGovernmentBenefitController()
	c.MemberJointAccountController()
	c.MemberRelativeAccountController()

	// Account Maintenance
	c.BankController()
	c.CancelledCashCheckVoucherController()
	c.CashCheckVoucherController()
	c.CashCheckVoucherTagController()
	c.JournalVoucherTagController()
	c.AdjustmentTagController()
	c.HolidayController()
	c.BillAndCoinsController()

	// Transaction batch
	c.TransactionBatchController()
	c.CashCountController()
	c.BatchFundingController()
	c.CheckRemittanceController()
	c.TransactionController()
	c.OnlineRemittanceController()
	c.DisbursementController()
	c.DisbursementTransactionController()

	// Accounts
	c.AccountController()
	c.AccountHistory()
	c.MemberAccountingLedgerController()
	c.AccountCategoryController()
	c.AccountClassificationController()
	c.GeneralLedgerController()
	c.GeneralLedgerGroupingController()
	c.FinancialStatementController()
	c.AccountTagController()
	c.PaymentTypeController()
	c.PaymentController()

	// Loans
	c.LoanStatusController()
	c.LoanPurposeController()
	c.LoanTransactionController()
	c.LoanTransactionEntryController()
	c.CollateralController()
	c.ComputationSheetController()
	c.AutomaticLoanDeductionController()
	c.BrowseExcludeIncludeAccountsController()
	c.IncludeNegativeAccountController()
	c.MemberDepartmentController()
	c.LoanTagController()

	// Time deposit
	c.TimeDepositTypeController()
	c.TimeDepositComputationController()
	c.TimeDepositComputationPreMatureController()

	// Charges rate scheme
	c.ChargesRateSchemeController()
	c.ChargesRateByRangeOrMinimumAmountController()
	c.ChargesRateSchemeModeOfPaymentController()
	c.JournalVoucherController()
	c.AdjustmentEntryController()
	c.CompanyController()
	return nil
}
