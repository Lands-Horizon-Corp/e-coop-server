package model

import (
	"context"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lands-horizon/horizon-server/src"
	"gorm.io/gorm"
)

type (
	QRMemberProfile struct {
		Firstname       string `json:"first_name"`
		Lastname        string `json:"last_name"`
		Middlename      string `json:"middle_name"`
		ContactNumber   string `json:"contact_number"`
		MemberProfileID string `json:"member_profile_id"`
		BranchID        string `json:"branch_id"`
		OrganizationID  string `json:"organization_id"`
		Email           string `json:"email"`
	}
	QRInvitationCode struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"user_type"`
		Code           string `json:"code"`
		CurrentUse     int    `json:"current_use"`
		Description    string `json:"description"`
	}

	QRUser struct {
		UserID        string `json:"user_id"`
		Email         string `json:"email"`
		ContactNumber string `json:"contact_number"`
		Username      string `json:"user_name"`
		Name          string `json:"name"`
		Lastname      string `json:"last_name"`
		Firstname     string `json:"first_name"`
		Middlename    string `json:"middle_name"`
	}
	Model struct {
		provider *src.Provider

		// Managers
		Migration []any

		BankManager                   horizon_services.Repository[Bank, BankResponse, BankRequest]
		BranchManager                 horizon_services.Repository[Branch, BranchResponse, BranchRequest]
		CategoryManager               horizon_services.Repository[Category, CategoryResponse, CategoryRequest]
		ContactUsManager              horizon_services.Repository[ContactUs, ContactUsResponse, ContactUsRequest]
		FeedbackManager               horizon_services.Repository[Feedback, FeedbackResponse, FeedbackRequest]
		FootstepManager               horizon_services.Repository[Footstep, FootstepResponse, any]
		GeneratedReportManager        horizon_services.Repository[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]
		InvitationCodeManager         horizon_services.Repository[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]
		MediaManager                  horizon_services.Repository[Media, MediaResponse, MediaRequest]
		NotificationManager           horizon_services.Repository[Notification, NotificationResponse, any]
		OrganizationCategoryManager   horizon_services.Repository[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]
		OrganizationDailyUsageManager horizon_services.Repository[OrganizationDailyUsage, OrganizationDailyUsageResponse, OrganizationDailyUsageRequest]
		OrganizationManager           horizon_services.Repository[Organization, OrganizationResponse, OrganizationRequest]
		PermissionTemplateManager     horizon_services.Repository[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]
		SubscriptionPlanManager       horizon_services.Repository[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]
		UserOrganizationManager       horizon_services.Repository[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]
		UserManager                   horizon_services.Repository[User, UserResponse, UserRegisterRequest]
		UserRatingManager             horizon_services.Repository[UserRating, UserRatingResponse, UserRatingRequest]

		// Members
		MemberAddressManager                horizon_services.Repository[MemberAddress, MemberAddressReponse, MemberAddressRequest]
		MemberAssetManager                  horizon_services.Repository[MemberAsset, MemberAssetResponse, MemberAssetRequest]
		MemberBankCardManager               horizon_services.Repository[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest]
		MemberCenterHistoryManager          horizon_services.Repository[MemberCenterHistory, MemberCenterHistoryResponse, MemberCenterHistoryRequest]
		MemberCenterManager                 horizon_services.Repository[MemberCenter, MemberCenterResponse, MemberCenterRequest]
		MemberClassificationManager         horizon_services.Repository[MemberClassification, MemberClassificationResponse, MemberClassificationRequest]
		MemberClassificationHistoryManager  horizon_services.Repository[MemberClassificationHistory, MemberClassificationHistoryResponse, MemberClassificationHistoryRequest]
		MemberCloseRemarkManager            horizon_services.Repository[MemberCloseRemark, MemberCloseRemarkResponse, MemberCloseRemarkRequest]
		MemberContactReferenceManager       horizon_services.Repository[MemberContactReference, MemberContactReferenceResponse, MemberContactReferenceRequest]
		MemberDamayanExtensionEntryManager  horizon_services.Repository[MemberDamayanExtensionEntry, MemberDamayanExtensionEntryResponse, MemberDamayanExtensionEntryRequest]
		MemberEducationalAttainmentManager  horizon_services.Repository[MemberEducationalAttainment, MemberEducationalAttainmentResponse, MemberEducationalAttainmentRequest]
		MemberExpenseManager                horizon_services.Repository[MemberExpense, MemberExpenseResponse, MemberExpenseRequest]
		MemberGenderHistoryManager          horizon_services.Repository[MemberGenderHistory, MemberGenderHistoryResponse, MemberGenderHistoryRequest]
		MemberGenderManager                 horizon_services.Repository[MemberGender, MemberGenderResponse, MemberGenderRequest]
		MemberGovernmentBenefitManager      horizon_services.Repository[MemberGovernmentBenefit, MemberGovernmentBenefitResponse, MemberGovernmentBenefitRequest]
		MemberGroupHistoryManager           horizon_services.Repository[MemberGroupHistory, MemberGroupHistoryResponse, MemberGroupHistoryRequest]
		MemberGroupManager                  horizon_services.Repository[MemberGroup, MemberGroupResponse, MemberGroupRequest]
		MemberIncomeManager                 horizon_services.Repository[MemberIncome, MemberIncomeResponse, MemberIncomeRequest]
		MemberJointAccountManager           horizon_services.Repository[MemberJointAccount, MemberJointAccountResponse, MemberJointAccountRequest]
		MemberMutualFundHistoryManager      horizon_services.Repository[MemberMutualFundHistory, MemberMutualFundHistoryResponse, MemberMutualFundHistoryRequest]
		MemberOccupationHistoryManager      horizon_services.Repository[MemberOccupationHistory, MemberOccupationHistoryResponse, MemberOccupationHistoryRequest]
		MemberOccupationManager             horizon_services.Repository[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]
		MemberOtherInformationEntryManager  horizon_services.Repository[MemberOtherInformationEntry, MemberOtherInformationEntryResponse, MemberOtherInformationEntryRequest]
		MemberRelativeAccountManager        horizon_services.Repository[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]
		MemberTypeHistoryManager            horizon_services.Repository[MemberTypeHistory, MemberTypeHistoryResponse, MemberTypeHistoryRequest]
		MemberTypeManager                   horizon_services.Repository[MemberType, MemberTypeResponse, MemberTypeRequest]
		MemberVerificationManager           horizon_services.Repository[MemberVerification, MemberVerificationResponse, MemberVerificationRequest]
		MemberProfileManager                horizon_services.Repository[MemberProfile, MemberProfileResponse, MemberProfileRequest]
		CollectorsMemberAccountEntryManager horizon_services.Repository[CollectorsMemberAccountEntry, CollectorsMemberAccountEntryResponse, CollectorsMemberAccountEntryRequest]

		// Employee Feature
		TimesheetManager horizon_services.Repository[Timesheet, TimesheetResponse, TimesheetRequest]

		// GL/FS
		FinancialStatementDefinitionManager             horizon_services.Repository[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]
		FinancialStatementAccountsGroupingManager       horizon_services.Repository[FinancialStatementAccountsGrouping, FinancialStatementAccountsGroupingResponse, FinancialStatementAccountsGroupingRequest]
		GeneralLedgerAccountsGroupingManager            horizon_services.Repository[GeneralLedgerAccountsGrouping, GeneralLedgerAccountsGroupingResponse, GeneralLedgerAccountsGroupingRequest]
		GeneralLedgerDefinitionManager                  horizon_services.Repository[GeneralLedgerDefinition, GeneralLedgerDefinitionResponse, GeneralLedgerDefinitionRequest]
		GeneralAccountGroupingNetSurplusPositiveManager horizon_services.Repository[GeneralAccountGroupingNetSurplusPositive, GeneralAccountGroupingNetSurplusPositiveResponse, GeneralAccountGroupingNetSurplusPositiveRequest]
		GeneralAccountGroupingNetSurplusNegativeManager horizon_services.Repository[GeneralAccountGroupingNetSurplusNegative, GeneralAccountGroupingNetSurplusNegativeResponse, GeneralAccountGroupingNetSurplusNegativeRequest]

		// MAINTENANCE TABLE FOR ACCOUNTING
		AccountClassificationManager horizon_services.Repository[AccountClassification, AccountClassificationResponse, AccountClassificationRequest]
		AccountCategoryManager       horizon_services.Repository[AccountCategory, AccountCategoryResponse, AccountCategoryRequest]
		PaymentTypeManager           horizon_services.Repository[PaymentType, PaymentTypeResponse, PaymentTypeRequest]
		BillAndCoinsManager          horizon_services.Repository[BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest]

		// ACCOUNT
		AccountManager    horizon_services.Repository[Account, AccountResponse, AccountRequest]
		AccountTagManager horizon_services.Repository[AccountTag, AccountTagResponse, AccountTagRequest]

		// LEDGERS
		GeneralAccountingLedgerManager       horizon_services.Repository[GeneralAccountingLedger, GeneralAccountingLedgerResponse, GeneralAccountingLedgerRequest]
		GeneralAccountingLedgerTagManager    horizon_services.Repository[GeneralAccountingLedgerTag, GeneralAccountingLedgerTagResponse, GeneralAccountingLedgerTagRequest]
		GeneralLedgerTransactionManager      horizon_services.Repository[GeneralLedgerTransaction, GeneralLedgerTransactionResponse, GeneralLedgerTransactionRequest]
		GeneralLedgerTransactionEntryManager horizon_services.Repository[GeneralLedgerTransactionEntry, GeneralLedgerTransactionEntryResponse, GeneralLedgerTransactionEntryRequest]
		MemberAccountingLedgerManager        horizon_services.Repository[MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest]

		// TRANSACTION START
		TransactionBatchManager horizon_services.Repository[TransactionBatch, TransactionBatchResponse, TransactionBatchRequest]
		CheckRemittanceManager  horizon_services.Repository[CheckRemittance, CheckRemittanceResponse, CheckRemittanceRequest]
		OnlineRemittanceManager horizon_services.Repository[OnlineRemittance, OnlineRemittanceResponse, OnlineRemittanceRequest]
		CashCountManager        horizon_services.Repository[CashCount, CashCountResponse, CashCountRequest]
		BatchFundingManager     horizon_services.Repository[BatchFunding, BatchFundingResponse, BatchFundingRequest]
		TransactionManager      horizon_services.Repository[Transaction, TransactionResponse, TransactionRequest]
		TransactionTagManager   horizon_services.Repository[TransactionTag, TransactionTagResponse, TransactionTagRequest]
		// Entries
		CheckEntryManager       horizon_services.Repository[CheckEntry, CheckEntryResponse, CheckEntryRequest]
		OnlineEntryManager      horizon_services.Repository[OnlineEntry, OnlineEntryResponse, OnlineEntryRequest]
		WithdrawalEntryManager  horizon_services.Repository[WithdrawalEntry, WithdrawalEntryResponse, WithdrawalEntryRequest]
		DepositEntryManager     horizon_services.Repository[DepositEntry, DepositEntryResponse, DepositEntryRequest]
		TransactionEntryManager horizon_services.Repository[TransactionEntry, TransactionEntryResponse, TransactionEntryRequest]
		// Disbursements
		DisbursementTransactionManager horizon_services.Repository[DisbursementTransaction, DisbursementTransactionResponse, DisbursementTransactionRequest]
		DisbursementManager            horizon_services.Repository[Disbursement, DisbursementResponse, DisbursementRequest]

		// LOAN START
		ComputationSheetManager                      horizon_services.Repository[ComputationSheet, ComputationSheetResponse, ComputationSheetRequest]
		FinesMaturityManager                         horizon_services.Repository[FinesMaturity, FinesMaturityResponse, FinesMaturityRequest]
		InterestMaturityManager                      horizon_services.Repository[InterestMaturity, InterestMaturityResponse, InterestMaturityRequest]
		IncludeNegativeAccountManager                horizon_services.Repository[IncludeNegativeAccount, IncludeNegativeAccountResponse, IncludeNegativeAccountRequest]
		AutomaticLoanDeductionManager                horizon_services.Repository[AutomaticLoanDeduction, AutomaticLoanDeductionResponse, AutomaticLoanDeductionRequest]
		BrowseExcludeIncludeAccountsManager          horizon_services.Repository[BrowseExcludeIncludeAccounts, BrowseExcludeIncludeAccountsResponse, BrowseExcludeIncludeAccountsRequest]
		MemberClassificationInterestRateManager      horizon_services.Repository[MemberClassificationInterestRate, MemberClassificationInterestRateResponse, MemberClassificationInterestRateRequest]
		LoanGuaranteedFundPerMonthManager            horizon_services.Repository[LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest]
		LoanStatusManager                            horizon_services.Repository[LoanStatus, LoanStatusResponse, LoanStatusRequest]
		LoanGuaranteedFundManager                    horizon_services.Repository[LoanGuaranteedFund, LoanGuaranteedFundResponse, LoanGuaranteedFundRequest]
		LoanTransactionManager                       horizon_services.Repository[LoanTransaction, LoanTransactionResponse, LoanTransactionRequest]
		LoanClearanceAnalysisManager                 horizon_services.Repository[LoanClearanceAnalysis, LoanClearanceAnalysisResponse, LoanClearanceAnalysisRequest]
		LoanClearanceAnalysisInstitutionManager      horizon_services.Repository[LoanClearanceAnalysisInstitution, LoanClearanceAnalysisInstitutionResponse, LoanClearanceAnalysisInstitutionRequest]
		LoanComakerMemberManager                     horizon_services.Repository[LoanComakerMember, LoanComakerMemberResponse, LoanComakerMemberRequest]
		LoanTransactionEntryManager                  horizon_services.Repository[LoanTransactionEntry, LoanTransactionEntryResponse, LoanTransactionEntryRequest]
		LoanTagManager                               horizon_services.Repository[LoanTag, LoanTagResponse, LoanTagRequest]
		LoanTermsAndConditionSuggestedPaymentManager horizon_services.Repository[LoanTermsAndConditionSuggestedPayment, LoanTermsAndConditionSuggestedPaymentResponse, LoanTermsAndConditionSuggestedPaymentRequest]
		LoanTermsAndConditionAmountReceiptManager    horizon_services.Repository[LoanTermsAndConditionAmountReceipt, LoanTermsAndConditionAmountReceiptResponse, LoanTermsAndConditionAmountReceiptRequest]
		LoanPurposeManager                           horizon_services.Repository[LoanPurpose, LoanPurposeResponse, LoanPurposeRequest]
		LoanLedgerManager                            horizon_services.Repository[LoanLedger, LoanLedgerResponse, LoanLedgerRequest]

		// Maintenance
		CollateralManager                                                   horizon_services.Repository[Collateral, CollateralResponse, CollateralRequest]
		TagTemplateManager                                                  horizon_services.Repository[TagTemplate, TagTemplateResponse, TagTemplateRequest]
		HolidayManager                                                      horizon_services.Repository[Holiday, HolidayResponse, HolidayRequest]
		GroceryComputationSheetManager                                      horizon_services.Repository[GroceryComputationSheet, GroceryComputationSheetResponse, GroceryComputationSheetRequest]
		GroceryComputationSheetMonthlyManager                               horizon_services.Repository[GroceryComputationSheetMonthly, GroceryComputationSheetMonthlyResponse, GroceryComputationSheetMonthlyRequest]
		InterestRateSchemeManager                                           horizon_services.Repository[InterestRateScheme, InterestRateSchemeResponse, InterestRateSchemeRequest]
		InterestRateByTermsHeaderManager                                    horizon_services.Repository[InterestRateByTermsHeader, InterestRateByTermsHeaderResponse, InterestRateByTermsHeaderRequest]
		InterestRateByTermManager                                           horizon_services.Repository[InterestRateByTerm, InterestRateByTermResponse, InterestRateByTermRequest]
		InterestRatePercentageManager                                       horizon_services.Repository[InterestRatePercentage, InterestRatePercentageResponse, InterestRatePercentageRequest]
		MemberTypeReferenceManager                                          horizon_services.Repository[MemberTypeReference, MemberTypeReferenceResponse, MemberTypeReferenceRequest]
		MemberTypeReferenceByAmountManager                                  horizon_services.Repository[MemberTypeReferenceByAmount, MemberTypeReferenceByAmountResponse, MemberTypeReferenceByAmountRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDateManager        horizon_services.Repository[MemberTypeReferenceInterestRateByUltimaMembershipDate, MemberTypeReferenceInterestRateByUltimaMembershipDateResponse, MemberTypeReferenceInterestRateByUltimaMembershipDateRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearManager horizon_services.Repository[MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearRequest]
		MemberDeductionEntryManager                                         horizon_services.Repository[MemberDeductionEntry, MemberDeductionEntryResponse, MemberDeductionEntryRequest]
		PostDatedCheckManager                                               horizon_services.Repository[PostDatedCheck, PostDatedCheckResponse, PostDatedCheckRequest]

		// TIME DEPOSIT
		TimeDepositTypeManager                    horizon_services.Repository[TimeDepositType, TimeDepositTypeResponse, TimeDepositTypeRequest]
		TimeDepositComputationHeaderManager       horizon_services.Repository[TimeDepositComputationHeader, TimeDepositComputationHeaderResponse, TimeDepositComputationHeaderRequest]
		TimeDepositComputationManager             horizon_services.Repository[TimeDepositComputation, TimeDepositComputationResponse, TimeDepositComputationRequest]
		TimeDepositComputationPreMatureManager    horizon_services.Repository[TimeDepositComputationPreMature, TimeDepositComputationPreMatureResponse, TimeDepositComputationPreMatureRequest]
		ChargesRateSchemeManager                  horizon_services.Repository[ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest]
		ChargesRateSchemeAccountManager           horizon_services.Repository[ChargesRateSchemeAccount, ChargesRateSchemeAccountResponse, ChargesRateSchemeAccountRequest]
		ChargesRateByRangeOrMinimumAmountManager  horizon_services.Repository[ChargesRateByRangeOrMinimumAmount, ChargesRateByRangeOrMinimumAmountResponse, ChargesRateByRangeOrMinimumAmountRequest]
		ChargesRateByTermHeaderManager            horizon_services.Repository[ChargesRateByTermHeader, ChargesRateByTermHeaderResponse, ChargesRateByTermHeaderRequest]
		ChargesRateByTermManager                  horizon_services.Repository[ChargesRateByTerm, ChargesRateByTermResponse, ChargesRateByTermRequest]
		ChargesRateMemberTypeModeOfPaymentManager horizon_services.Repository[ChargesRateMemberTypeModeOfPayment, ChargesRateMemberTypeModeOfPaymentResponse, ChargesRateMemberTypeModeOfPaymentRequest]

		// ACCOUNTING ENTRY
		AdjustmentEntryManager                   horizon_services.Repository[AdjustmentEntry, AdjustmentEntryResponse, AdjustmentEntryRequest]
		AdjustmentEntryTagManager                horizon_services.Repository[AdjustmentEntryTag, AdjustmentEntryTagResponse, AdjustmentEntryTagRequest]
		VoucherPayToManager                      horizon_services.Repository[VoucherPayTo, VoucherPayToResponse, VoucherPayToRequest]
		CashCheckVoucherManager                  horizon_services.Repository[CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest]
		CashCheckVoucherTagManager               horizon_services.Repository[CashCheckVoucherTag, CashCheckVoucherTagResponse, CashCheckVoucherTagRequest]
		CashCheckVoucherEntryManager             horizon_services.Repository[CashCheckVoucherEntry, CashCheckVoucherEntryResponse, CashCheckVoucherEntryRequest]
		CashCheckVoucherDisbursementEntryManager horizon_services.Repository[CashCheckVoucherDisbursementEntry, CashCheckVoucherDisbursementEntryResponse, CashCheckVoucherDisbursementEntryRequest]
		JournalVoucherManager                    horizon_services.Repository[JournalVoucher, JournalVoucherResponse, JournalVoucherRequest]
		JournalVoucherTagManager                 horizon_services.Repository[JournalVoucherTag, JournalVoucherTagResponse, JournalVoucherTagRequest]
		JournalVoucherEntryManager               horizon_services.Repository[JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest]
	}
)

func NewModel(provider *src.Provider) (*Model, error) {
	return &Model{
		provider: provider,
	}, nil
}

// Setting up Validator, Broadcaster, Model, and Automigration
/*
x = x.replace(" ","").replace(".go","").replace("└──","").replace("├──","").replace(".", "")
for i in x.split("\n"):
    print(f'c.{i.replace("_", " ").title().replace(" ", "")}()')
*/
func (c *Model) Start() error {

	// Models
	c.AccountCategory()
	c.AccountClassification()
	c.Account()
	c.AccountTag()
	c.AdjustmentEntry()
	c.AdjustmentEntryTag()
	c.AutomaticLoanDeduction()
	c.Bank()
	c.BatchFunding()
	c.BillAndCoins()
	c.Branch()
	c.BrowseExcludeIncludeAccounts()
	c.CashCheckVoucherDisbursementEntry()
	c.CashCheckVoucherEntry()
	c.CashCheckVoucher()
	c.CashCheckVoucherTag()
	c.CashCount()
	c.Category()
	c.ChargesRateByRangeOrMinimumAmount()
	c.ChargesRateByTerm()
	c.ChargesRateByTermHeader()
	c.ChargesRateMemberTypeModeOfPayment()
	c.ChargesRateSchemeAccount()
	c.ChargesRateScheme()
	c.CheckEntry()
	c.CheckRemittance()
	c.Collateral()
	c.CollectorsMemberAccountEntry()
	c.ComputationSheet()
	c.ContactUs()
	c.DepositEntry()
	c.Disbursement()
	c.DisbursementTransaction()
	c.Feedback()
	c.FinancialStatementAccountsGrouping()
	c.FinancialStatementDefinition()
	c.FinesMaturity()
	c.Footstep()
	c.GeneralAccountGroupingNetSurplusNegative()
	c.GeneralAccountGroupingNetSurplusPositive()
	c.GeneralAccountingLedger()
	c.GeneralAccountingLedgerTag()
	c.GeneralLedgerAccountsGrouping()
	c.GeneralLedgerTransactionEntry()
	c.GeneralLedgerTransaction()
	c.GeneratedReport()
	c.GeneralLedgerDefinition()
	c.GroceryComputationSheet()
	c.GroceryComputationSheetMonthly()
	c.Holiday()
	c.IncludeNegativeAccount()
	c.InterestMaturity()
	c.InterestRateByTerm()
	c.InterestRateByTermsHeader()
	c.InterestRatePercentage()
	c.InterestRateScheme()
	c.InvitationCode()
	c.JournalVoucherEntry()
	c.JournalVoucher()
	c.JournalVoucherTag()
	c.LoanClearanceAnalysis()
	c.LoanClearanceAnalysisInstitution()
	c.LoanComakerMember()
	c.LoanGuaranteedFund()
	c.LoanGuaranteedFundPerMonth()
	c.LoanLedger()
	c.LoanPurpose()
	c.LoanStatus()
	c.LoanTag()
	c.LoanTermsAndConditionAmountReceipt()
	c.LoanTermsAndConditionSuggestedPayment()
	c.LoanTransactionEntry()
	c.LoanTransaction()
	c.Media()
	c.MemberAccountingLedger()
	c.MemberAddress()
	c.MemberAsset()
	c.MemberBankCard()
	c.MemberCenter()
	c.MemberCenterHistory()
	c.MemberClassification()
	c.MemberClassificationInterestRate()
	c.MemberCloseRemark()
	c.MemberContactReference()
	c.MemberDamayanExtensionEntry()
	c.MemberDeductionEntry()
	c.MemberEducationalAttainment()
	c.MemberExpense()
	c.MemberGender()
	c.MemberGenderHistory()
	c.MemberGovernmentBenefit()
	c.MemberGroup()
	c.MemberGroupHistory()
	c.MemberIncome()
	c.MemberJointAccount()
	c.MemberMutualFundHistory()
	c.MemberOccupation()
	c.MemberOccupationHistory()
	c.MemberOtherInformationEntry()
	c.MemberProfile()
	c.MemberRelativeAccount()
	c.MemberType()
	c.MemberTypeHistory()
	c.MemberTypeReferenceByAmount()
	c.MemberTypeReference()
	c.MemberTypeReferenceInterestRateByUltimaMembershipDate()
	c.MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear()
	c.MemberVerification()
	c.Notification()
	c.OnlineEntry()
	c.OnlineRemittance()
	c.OrganizationCategory()
	c.OrganizationDailyUsage()
	c.Organization()
	c.PaymentType()
	c.PermissionTemplate()
	c.PostDatedCheck()
	c.SubscriptionPlan()
	c.TagTemplate()
	c.TimeDepositComputation()
	c.TimeDepositComputationHeader()
	c.TimeDepositComputationPreMature()
	c.TimeDepositType()
	c.Timesheet()
	c.TransactionBatch()
	c.TransactionEntry()
	c.Transaction()
	c.TransactionTag()
	c.User()
	c.UserOrganization()
	c.UserRating()
	c.VoucherPayTo()
	c.WithdrawalEntry()

	if err := c.provider.Service.Database.Client().AutoMigrate(c.Migration...); err != nil {
		return err
	}
	return nil
}

func (c *Model) OrganizationSeeder(context context.Context, tx *gorm.DB, userId uuid.UUID, organizationId uuid.UUID, branchId uuid.UUID) error {
	return nil
}

func (c *Model) OrganizationDestroyer(context context.Context, tx *gorm.DB, userId uuid.UUID, organizationId uuid.UUID, branchId uuid.UUID) error {
	return nil
}
