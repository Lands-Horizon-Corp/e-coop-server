package v1

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/account"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/charges"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/funds"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/journal"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/loan"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/member_profile"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/organization"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/reports"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/settings"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/time_deposit"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/transactions"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/user"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/voucher"
)

func Controllers(service *horizon.HorizonService) error {
	// Settings Module
	settings.Heartbeat(service)
	settings.BranchController(service)
	settings.CategoryController(service)
	settings.ContactController(service)
	settings.CurrencyController(service)
	settings.FeedbackController(service)
	settings.FootstepController(service)
	settings.MediaController(service)
	settings.NotificationController(service)
	settings.PermissionTemplateController(service)
	settings.QRCodeController(service)
	settings.SubscriptionPlanController(service)
	settings.TimesheetController(service)
	settings.TagTemplateController(service)
	settings.HolidayController(service)
	settings.BankController(service)
	settings.CompanyController(service)

	// User Module
	user.AuthenticationController(service)
	user.UserController(service)
	user.UserOrganizationController(service)
	user.UserRatingController(service)
	user.InvitationCodeController(service)
	user.KYCController(service)

	// Member Profile Module
	member_profile.MemberGenderController(service)
	member_profile.MemberCenterController(service)
	member_profile.MemberTypeController(service)
	member_profile.MemberClassificationController(service)
	member_profile.MemberOccupationController(service)
	member_profile.MemberGroupController(service)
	member_profile.MemberProfileController(service)
	member_profile.MemberEducationalAttainmentController(service)
	member_profile.MemberAddressController(service)
	member_profile.MemberContactReferenceController(service)
	member_profile.MemberAssetController(service)
	member_profile.MemberIncomeController(service)
	member_profile.MemberExpenseController(service)
	member_profile.MemberGovernmentBenefitController(service)
	member_profile.MemberJointAccountController(service)
	member_profile.MemberRelativeAccountController(service)
	member_profile.MemberDepartmentController(service)
	member_profile.MemberProfileMediaController(service)
	member_profile.MemberProfileArchiveController(service)
	member_profile.MemberProfileComakerController(service)
	member_profile.MemberAccountingLedgerController(service)

	// Organization Module
	organization.OrganizationController(service)
	organization.OrganizationDailyUsageController(service)
	organization.OrganizationMediaController(service)

	// Voucher Module
	voucher.CancelledCashCheckVoucherController(service)
	voucher.CashCheckVoucherController(service)
	voucher.CashCheckVoucherTagController(service)
	voucher.BillAndCoinsController(service)
	voucher.CashCountController(service)

	// Journal Module
	journal.JournalVoucherController(service)
	journal.JournalVoucherTagController(service)

	// Transactions Module
	transactions.TransactionBatchController(service)
	transactions.CheckRemittanceController(service)
	transactions.TransactionController(service)
	transactions.OnlineRemittanceController(service)
	transactions.DisbursementController(service)
	transactions.DisbursementTransactionController(service)
	transactions.AdjustmentEntryController(service)
	transactions.AdjustmentTagController(service)
	transactions.PaymentTypeController(service)
	transactions.PaymentController(service)

	// Account Module
	account.AccountController(service)
	account.AccountHistoryController(service)
	account.AccountCategoryController(service)
	account.AccountClassificationController(service)
	account.AccountTagController(service)
	account.AccountTransactionController(service)
	account.BrowseExcludeIncludeAccountsController(service)
	account.BrowseReferenceController(service)
	account.IncludeNegativeAccountController(service)

	// Reports Module
	reports.GeneratedReportsController(service)
	reports.GeneralLedgerController(service)
	reports.GeneralLedgerGroupingController(service)
	reports.FinancialStatementController(service)
	reports.GeneratedSavingsInterestController(service)
	reports.GeneratedSavingsInterestEntryController(service)

	// Loan Module
	loan.LoanStatusController(service)
	loan.LoanPurposeController(service)
	loan.LoanTransactionController(service)
	loan.LoanTransactionEntryController(service)
	loan.CollateralController(service)
	loan.ComputationSheetController(service)
	loan.AutomaticLoanDeductionController(service)
	loan.LoanTagController(service)

	// Time Deposit Module
	time_deposit.TimeDepositTypeController(service)
	time_deposit.TimeDepositComputationController(service)
	time_deposit.TimeDepositComputationPreMatureController(service)

	// Charges Module
	charges.ChargesRateSchemeController(service)
	charges.ChargesRateByRangeOrMinimumAmountController(service)
	charges.ChargesRateSchemeModeOfPaymentController(service)
	charges.ChargesRateByTermController(service)

	// Funds Module
	funds.FundsController(service)
	funds.BatchFundingController(service)
	funds.MutualFundsController(service)
	funds.MutualFundEntryController(service)

	// Common Controller
	commonController(service)

	return nil
}
