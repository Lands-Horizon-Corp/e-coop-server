package controller_v1

import (
	"github.com/google/uuid"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

type Controller struct {
	// Services
	provider *src.Provider
	model    *model.Model
	event    *event.Event
	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken
	userToken             *cooperative_tokens.UserToken
}

func NewController(
	// Services
	provider *src.Provider,
	model *model.Model,
	event *event.Event,

	// Tokens
	userOrganizationToken *cooperative_tokens.UserOrganizationToken,
	userToken *cooperative_tokens.UserToken,

) (*Controller, error) {
	return &Controller{
		// Services
		provider: provider,
		model:    model,
		event:    event,

		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
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
	c.CollateralController()
	c.ComputationSheetController()
	c.AutomaticLoanDeductionController()
	c.BrowseExcludeIncludeAccountsController()
	c.IncludeNegativeAccountController()
	c.MemberDepartmentController()
	c.TimeDepositTypeController()
	c.TimeDepositComputation()

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
