package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
)

type Controller struct {
	// Services
	provider  *src.Provider
	modelcore *modelcore.ModelCore
	event     *event.Event
	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken
	userToken             *cooperative_tokens.UserToken
	service               *service.TransactionService
}

func NewController(
	// Services
	provider *src.Provider,
	modelcore *modelcore.ModelCore,
	event *event.Event,

	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken,
	userToken *cooperative_tokens.UserToken,
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
	c.Heartbeat()
	c.FormGeneratorController()
	c.AuthenticationController()
	// Basic Onboardding & Utilities
	c.BranchController()
	c.CategoryController()
	c.ContactController()
	c.CurrencyController()
	c.FeedbackController()
	c.FootstepController()
	c.GeneratedReports()
	c.InvitationCode()
	c.MediaController()
	c.NotificationController()
	c.OrganizationController()
	c.OrganizationDailyUsage()
	c.PermissionTemplateController()
	c.QRCodeController()
	c.SubscriptionPlanController()
	c.TimesheetController()
	c.UserController()
	c.UserOrganinzationController()
	c.UserRatingController()
	c.MemberProfileMediaController()
	c.TagTemplateController()

	// Member Profile
	c.MemberGenderController()
	c.MemberCenterController()
	c.MemberTypeController()
	c.MemberClassificationController()
	c.MemberOccupationController()
	c.MemberGroupController()
	c.MemberProfileController()
	c.MemberTypeReferenceController()

	// member profile properties
	c.MemberEducationalAttainmentController()
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
