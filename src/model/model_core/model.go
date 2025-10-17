package model_core

import (
	"context"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	IDSRequest struct {
		IDs []string `json:"ids"`
	}

	QRMemberProfile struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		MiddleName      string `json:"middle_name"`
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
	ModelCore struct {
		provider *src.Provider

		// Managers
		Migration []any

		BankManager                   horizon_services.Repository[Bank, BankResponse, BankRequest]
		BranchManager                 horizon_services.Repository[Branch, BranchResponse, BranchRequest]
		BranchSettingManager          horizon_services.Repository[BranchSetting, BranchSettingResponse, BranchSettingRequest]
		CategoryManager               horizon_services.Repository[Category, CategoryResponse, CategoryRequest]
		ContactUsManager              horizon_services.Repository[ContactUs, ContactUsResponse, ContactUsRequest]
		CurrencyManager               horizon_services.Repository[Currency, CurrencyResponse, CurrencyRequest]
		FeedbackManager               horizon_services.Repository[Feedback, FeedbackResponse, FeedbackRequest]
		FootstepManager               horizon_services.Repository[Footstep, FootstepResponse, FootstepRequest]
		GeneratedReportManager        horizon_services.Repository[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]
		InvitationCodeManager         horizon_services.Repository[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]
		MediaManager                  horizon_services.Repository[Media, MediaResponse, MediaRequest]
		NotificationManager           horizon_services.Repository[Notification, NotificationResponse, any]
		OrganizationCategoryManager   horizon_services.Repository[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]
		OrganizationDailyUsageManager horizon_services.Repository[OrganizationDailyUsage, OrganizationDailyUsageResponse, OrganizationDailyUsageRequest]
		OrganizationManager           horizon_services.Repository[Organization, OrganizationResponse, OrganizationRequest]
		OrganizationMediaManager      horizon_services.Repository[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest]
		PermissionTemplateManager     horizon_services.Repository[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]
		SubscriptionPlanManager       horizon_services.Repository[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]
		UserOrganizationManager       horizon_services.Repository[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]
		UserManager                   horizon_services.Repository[User, UserResponse, UserRegisterRequest]
		UserRatingManager             horizon_services.Repository[UserRating, UserRatingResponse, UserRatingRequest]
		UserMediaManager              horizon_services.Repository[UserMedia, UserMediaResponse, UserMediaRequest]

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
		MemberDepartmentManager             horizon_services.Repository[MemberDepartment, MemberDepartmentResponse, MemberDepartmentRequest]
		MemberDepartmentHistoryManager      horizon_services.Repository[MemberDepartmentHistory, MemberDepartmentHistoryResponse, MemberDepartmentHistoryRequest]

		// Employee Feature
		TimesheetManager horizon_services.Repository[Timesheet, TimesheetResponse, TimesheetRequest]
		CompanyManager   horizon_services.Repository[Company, CompanyResponse, CompanyRequest]

		// GL/FS
		FinancialStatementDefinitionManager             horizon_services.Repository[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]
		FinancialStatementGroupingManager               horizon_services.Repository[FinancialStatementGrouping, FinancialStatementGroupingResponse, FinancialStatementGroupingRequest]
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
		GeneralLedgerManager          horizon_services.Repository[GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest]
		GeneralLedgerTagManager       horizon_services.Repository[GeneralLedgerTag, GeneralLedgerTagResponse, GeneralLedgerTagRequest]
		MemberAccountingLedgerManager horizon_services.Repository[MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest]

		// TRANSACTION START
		TransactionBatchManager horizon_services.Repository[TransactionBatch, TransactionBatchResponse, TransactionBatchRequest]
		CheckRemittanceManager  horizon_services.Repository[CheckRemittance, CheckRemittanceResponse, CheckRemittanceRequest]
		OnlineRemittanceManager horizon_services.Repository[OnlineRemittance, OnlineRemittanceResponse, OnlineRemittanceRequest]
		CashCountManager        horizon_services.Repository[CashCount, CashCountResponse, CashCountRequest]
		BatchFundingManager     horizon_services.Repository[BatchFunding, BatchFundingResponse, BatchFundingRequest]
		TransactionManager      horizon_services.Repository[Transaction, TransactionResponse, TransactionRequest]
		TransactionTagManager   horizon_services.Repository[TransactionTag, TransactionTagResponse, TransactionTagRequest]

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
		ComakerMemberProfileManager                  horizon_services.Repository[ComakerMemberProfile, ComakerMemberProfileResponse, ComakerMemberProfileRequest]
		ComakerCollateralManager                     horizon_services.Repository[ComakerCollateral, ComakerCollateralResponse, ComakerCollateralRequest]
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
		TimeDepositTypeManager                   horizon_services.Repository[TimeDepositType, TimeDepositTypeResponse, TimeDepositTypeRequest]
		TimeDepositComputationManager            horizon_services.Repository[TimeDepositComputation, TimeDepositComputationResponse, TimeDepositComputationRequest]
		TimeDepositComputationPreMatureManager   horizon_services.Repository[TimeDepositComputationPreMature, TimeDepositComputationPreMatureResponse, TimeDepositComputationPreMatureRequest]
		ChargesRateSchemeManager                 horizon_services.Repository[ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest]
		ChargesRateSchemeAccountManager          horizon_services.Repository[ChargesRateSchemeAccount, ChargesRateSchemeAccountResponse, ChargesRateSchemeAccountRequest]
		ChargesRateByRangeOrMinimumAmountManager horizon_services.Repository[ChargesRateByRangeOrMinimumAmount, ChargesRateByRangeOrMinimumAmountResponse, ChargesRateByRangeOrMinimumAmountRequest]
		ChargesRateByTermHeaderManager           horizon_services.Repository[ChargesRateByTermHeader, ChargesRateByTermHeaderResponse, ChargesRateByTermHeaderRequest]
		ChargesRateByTermManager                 horizon_services.Repository[ChargesRateByTerm, ChargesRateByTermResponse, ChargesRateByTermRequest]

		// ACCOUNTING ENTRY
		AdjustmentEntryManager           horizon_services.Repository[AdjustmentEntry, AdjustmentEntryResponse, AdjustmentEntryRequest]
		AdjustmentTagManager             horizon_services.Repository[AdjustmentTag, AdjustmentTagResponse, AdjustmentTagRequest]
		VoucherPayToManager              horizon_services.Repository[VoucherPayTo, VoucherPayToResponse, VoucherPayToRequest]
		CashCheckVoucherManager          horizon_services.Repository[CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest]
		CashCheckVoucherEntryManager     horizon_services.Repository[CashCheckVoucherEntry, CashCheckVoucherEntryResponse, CashCheckVoucherEntryRequest]
		CashCheckVoucherTagManager       horizon_services.Repository[CashCheckVoucherTag, CashCheckVoucherTagResponse, CashCheckVoucherTagRequest]
		CancelledCashCheckVoucherManager horizon_services.Repository[CancelledCashCheckVoucher, CancelledCashCheckVoucherResponse, CancelledCashCheckVoucherRequest]
		JournalVoucherManager            horizon_services.Repository[JournalVoucher, JournalVoucherResponse, JournalVoucherRequest]
		JournalVoucherEntryManager       horizon_services.Repository[JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest]
		JournalVoucherTagManager         horizon_services.Repository[JournalVoucherTag, JournalVoucherTagResponse, JournalVoucherTagRequest]

		FundsManager                          horizon_services.Repository[Funds, FundsResponse, FundsRequest]
		ChargesRateSchemeModeOfPaymentManager horizon_services.Repository[ChargesRateSchemeModeOfPayment, ChargesRateSchemeModeOfPaymentResponse, ModeOfPayment]
	}
)

func NewModelCore(provider *src.Provider) (*ModelCore, error) {
	return &ModelCore{
		provider: provider,
	}, nil
}

// Setting up Validator, Broadcaster, Model, and Automigration
/*
x = x.replace(" ","").replace(".go","").replace("└──","").replace("├──","").replace(".", "")
for i in x.split("\n"):
    print(f'c.{i.replace("_", " ").title().replace(" ", "")}()')
*/
func (c *ModelCore) Start(context context.Context) error {

	// Models
	c.AccountCategory()
	c.AccountClassification()

	c.AdjustmentEntry()
	c.AdjustmentTag()
	c.AutomaticLoanDeduction()
	c.Bank()
	c.BatchFunding()
	c.BillAndCoins()
	c.Branch()
	c.BrowseExcludeIncludeAccounts()
	c.CashCheckVoucher()
	c.CashCheckVoucherEntry()
	c.CashCheckVoucherTag()
	c.CancelledCashCheckVoucher()
	c.CashCount()
	c.Category()
	c.Currency()
	c.ChargesRateByRangeOrMinimumAmount()
	c.ChargesRateByTerm()
	c.ChargesRateByTermHeader()
	c.ChargesRateSchemeAccount()
	c.ChargesRateScheme()
	c.CheckRemittance()
	c.Collateral()
	c.CollectorsMemberAccountEntry()
	c.ComputationSheet()
	c.ContactUs()
	c.Disbursement()
	c.DisbursementTransaction()
	c.Feedback()
	c.FinancialStatementGrouping()
	c.FinancialStatementDefinition()
	c.FinesMaturity()
	c.Footstep()
	c.GeneralAccountGroupingNetSurplusNegative()
	c.GeneralAccountGroupingNetSurplusPositive()
	c.GeneralLedger()
	c.GeneralLedgerTag()
	c.GeneralLedgerAccountsGrouping()
	c.GeneratedReport()
	c.GeneralLedgerDefinition()
	c.Account()
	c.AccountTag()
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
	c.JournalVoucher()
	c.JournalVoucherEntry()
	c.JournalVoucherTag()
	c.LoanClearanceAnalysis()
	c.LoanClearanceAnalysisInstitution()
	c.LoanComakerMember()
	c.ComakerMemberProfile()
	c.ComakerCollateral()
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
	c.MemberClassificationHistory()
	c.MemberClassificationInterestRate()
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
	c.MemberCloseRemark()
	c.MemberRelativeAccount()
	c.MemberType()
	c.MemberTypeHistory()
	c.MemberTypeReferenceByAmount()
	c.MemberTypeReference()
	c.MemberTypeReferenceInterestRateByUltimaMembershipDate()
	c.MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear()
	c.MemberVerification()
	c.Notification()
	c.OnlineRemittance()
	c.OrganizationCategory()
	c.OrganizationDailyUsage()
	c.Organization()
	c.OrganizationMedia()
	c.PaymentType()
	c.PermissionTemplate()
	c.PostDatedCheck()
	c.SubscriptionPlan()
	c.TagTemplate()
	c.TimeDepositComputation()
	c.TimeDepositComputationPreMature()
	c.TimeDepositType()
	c.Timesheet()
	c.TransactionBatch()
	c.Transaction()
	c.TransactionTag()
	c.User()
	c.UserOrganization()
	c.UserRating()
	c.UserMedia()
	c.VoucherPayTo()
	c.MemberDepartment()
	c.MemberDepartmentHistory()
	c.Funds()
	c.ChargesRateSchemeModeOfPayment()
	c.BranchSetting()
	c.Company()
	return nil
}

func (m *ModelCore) OrganizationSeeder(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	if err := m.InvitationCodeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.BankSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}

	if err := m.BillAndCoinsSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.HolidaySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberClassificationSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberGenderSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberGroupSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberCenterSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberOccupationSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberTypeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.MemberDepartmentSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.GeneralLedgerAccountsGroupingSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.FinancialStatementGroupingSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.PaymentTypeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.AccountSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.LoanPurposeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return nil
	}
	if err := m.AccountCategorySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.DisbursementSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.CollateralSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.TagTemplateSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.LoanStatusSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg, err := m.UserOrganizationManager.FindOne(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
	if err != nil {
		return err
	}
	if err := m.MemberProfileSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg.IsSeeded = true
	if err := m.UserOrganizationManager.UpdateByIDWithTx(context, tx, userOrg.ID, userOrg); err != nil {
		return err
	}
	if err := m.CompanySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

func (m *ModelCore) OrganizationDestroyer(ctx context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	invitationCodes, err := m.InvitationCodeManager.Find(ctx, &InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get invitation codes")
	}
	banks, err := m.BankManager.Find(ctx, &Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get banks")
	}
	billAndCoins, err := m.BillAndCoinsManager.Find(ctx, &BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get bill and coins")
	}
	// 4. Delete Holidays
	holidays, err := m.HolidayManager.Find(ctx, &Holiday{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get holidays")
	}
	for _, data := range holidays {
		if err := m.HolidayManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := m.BillAndCoinsManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bill or coin %s", data.Name)
		}
	}
	for _, data := range banks {
		if err := m.BankManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bank %s", data.Name)
		}
	}
	for _, data := range invitationCodes {
		if err := m.InvitationCodeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy invitation code %s", data.Code)
		}
	}

	// 1. Delete MemberType
	memberTypes, err := m.MemberTypeManager.Find(ctx, &MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member types")
	}
	for _, data := range memberTypes {
		if err := m.MemberTypeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member type %s", data.Name)
		}
	}

	// 2. Delete MemberOccupation
	memberOccupations, err := m.MemberOccupationManager.Find(ctx, &MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member occupations")
	}
	for _, data := range memberOccupations {
		if err := m.MemberOccupationManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member occupation %s", data.Name)
		}
	}

	// 3. Delete MemberGroup
	memberGroups, err := m.MemberGroupManager.Find(ctx, &MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member groups")
	}
	for _, data := range memberGroups {
		if err := m.MemberGroupManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member group %s", data.Name)
		}
	}

	// 4. Delete MemberGender
	memberGenders, err := m.MemberGenderManager.Find(ctx, &MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member genders")
	}
	for _, data := range memberGenders {
		if err := m.MemberGenderManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member gender %s", data.Name)
		}
	}

	// 5. Delete MemberCenter
	memberCenters, err := m.MemberCenterManager.Find(ctx, &MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member centers")
	}
	for _, data := range memberCenters {
		if err := m.MemberCenterManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member center %s", data.Name)
		}
	}

	// 6. Delete MemberClassification
	memberClassifications, err := m.MemberClassificationManager.Find(ctx, &MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member classifications")
	}
	for _, data := range memberClassifications {
		if err := m.MemberClassificationManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member classification %s", data.Name)
		}
	}

	generalLedgerDefinitions, err := m.GeneralLedgerDefinitionManager.Find(ctx, &GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger definitions")
	}
	for _, data := range generalLedgerDefinitions {
		if err := m.GeneralLedgerDefinitionManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger definition %s", data.Name)
		}
	}

	generalLedgerAccountsGroupings, err := m.GeneralLedgerAccountsGroupingManager.Find(ctx, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger accounts groupings")
	}
	for _, data := range generalLedgerAccountsGroupings {
		if err := m.GeneralLedgerAccountsGroupingManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger accounts grouping %s", data.Name)
		}
	}

	// Financial Statement Accounts Grouping destroyer
	FinancialStatementGroupings, err := m.FinancialStatementGroupingManager.Find(ctx, &FinancialStatementGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get financial statement accounts groupings")
	}
	for _, data := range FinancialStatementGroupings {
		if err := m.FinancialStatementGroupingManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy financial statement accounts grouping %s", data.Name)
		}
	}
	paymentTypes, err := m.PaymentTypeManager.Find(ctx, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get payment types")
	}
	for _, data := range paymentTypes {
		if err := m.PaymentTypeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy payment type %s", data.Name)
		}
	}
	disbursements, err := m.DisbursementManager.Find(ctx, &Disbursement{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get disbursements")
	}
	for _, data := range disbursements {
		if err := m.DisbursementManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy disbursement %s", data.Name)
		}
	}
	collaterals, err := m.CollateralManager.Find(ctx, &Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get collaterals")
	}
	for _, data := range collaterals {
		if err := m.CollateralManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy collateral %s", data.Name)
		}
	}
	return nil
}
