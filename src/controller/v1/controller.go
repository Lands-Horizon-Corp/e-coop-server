package controller_v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
	"github.com/google/uuid"
)

type Controller struct {
	// Services
	provider *src.Provider
	model    *model.Model
	event    *event.Event
	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken
	userToken             *cooperative_tokens.UserToken
	service               *service.TransactionService
}

func NewController(
	// Services
	provider *src.Provider,
	model *model.Model,
	event *event.Event,

	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken,
	userToken *cooperative_tokens.UserToken,
	service *service.TransactionService,

) (*Controller, error) {
	return &Controller{
		// Services
		provider: provider,
		model:    model,
		event:    event,

		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		service:               service,
	}, nil
}

func (c *Controller) Start() error {
	// Others
	c.Heartbeat()
	c.FormGeneratorController()

	// Basic Onboardding & Utilities
	c.BranchController()
	c.CategoryController()
	c.ContactController()
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
	c.UserMediaController()
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
	c.AdjustmentEntryTagController()
	return nil
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
