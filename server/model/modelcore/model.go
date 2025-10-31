package modelcore

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// IDSRequest represents a request containing multiple IDs
	IDSRequest struct {
		IDs []string `json:"ids"`
	}

	// QRMemberProfile represents QR code data for member profile information
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
	// QRInvitationCode represents QR code data for organization invitation codes
	QRInvitationCode struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"user_type"`
		Code           string `json:"code"`
		CurrentUse     int    `json:"current_use"`
		Description    string `json:"description"`
	}

	// QRUser represents QR code data for user information
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
	// modelcore represents the core model management structure containing all entity managers
	ModelCore struct {
		provider *server.Provider

		// Managers
		Migration []any

		BankManager                   services.Repository[Bank, BankResponse, BankRequest]
		BranchManager                 services.Repository[Branch, BranchResponse, BranchRequest]
		BranchSettingManager          services.Repository[BranchSetting, BranchSettingResponse, BranchSettingRequest]
		CategoryManager               services.Repository[Category, CategoryResponse, CategoryRequest]
		ContactUsManager              services.Repository[ContactUs, ContactUsResponse, ContactUsRequest]
		CurrencyManager               services.Repository[Currency, CurrencyResponse, CurrencyRequest]
		FeedbackManager               services.Repository[Feedback, FeedbackResponse, FeedbackRequest]
		FootstepManager               services.Repository[Footstep, FootstepResponse, FootstepRequest]
		GeneratedReportManager        services.Repository[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]
		InvitationCodeManager         services.Repository[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]
		MediaManager                  services.Repository[Media, MediaResponse, MediaRequest]
		NotificationManager           services.Repository[Notification, NotificationResponse, any]
		OrganizationCategoryManager   services.Repository[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]
		OrganizationDailyUsageManager services.Repository[OrganizationDailyUsage, OrganizationDailyUsageResponse, OrganizationDailyUsageRequest]
		OrganizationManager           services.Repository[Organization, OrganizationResponse, OrganizationRequest]
		OrganizationMediaManager      services.Repository[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest]
		PermissionTemplateManager     services.Repository[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]
		SubscriptionPlanManager       services.Repository[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]
		UserOrganizationManager       services.Repository[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]
		UserManager                   services.Repository[User, UserResponse, UserRegisterRequest]
		UserRatingManager             services.Repository[UserRating, UserRatingResponse, UserRatingRequest]
		MemberProfileMediaManager     services.Repository[MemberProfileMedia, MemberProfileMediaResponse, MemberProfileMediaRequest]

		// Members
		MemberAddressManager                services.Repository[MemberAddress, MemberAddressReponse, MemberAddressRequest]
		MemberAssetManager                  services.Repository[MemberAsset, MemberAssetResponse, MemberAssetRequest]
		MemberBankCardManager               services.Repository[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest]
		MemberCenterHistoryManager          services.Repository[MemberCenterHistory, MemberCenterHistoryResponse, MemberCenterHistoryRequest]
		MemberCenterManager                 services.Repository[MemberCenter, MemberCenterResponse, MemberCenterRequest]
		MemberClassificationManager         services.Repository[MemberClassification, MemberClassificationResponse, MemberClassificationRequest]
		MemberClassificationHistoryManager  services.Repository[MemberClassificationHistory, MemberClassificationHistoryResponse, MemberClassificationHistoryRequest]
		MemberCloseRemarkManager            services.Repository[MemberCloseRemark, MemberCloseRemarkResponse, MemberCloseRemarkRequest]
		MemberContactReferenceManager       services.Repository[MemberContactReference, MemberContactReferenceResponse, MemberContactReferenceRequest]
		MemberDamayanExtensionEntryManager  services.Repository[MemberDamayanExtensionEntry, MemberDamayanExtensionEntryResponse, MemberDamayanExtensionEntryRequest]
		MemberEducationalAttainmentManager  services.Repository[MemberEducationalAttainment, MemberEducationalAttainmentResponse, MemberEducationalAttainmentRequest]
		MemberExpenseManager                services.Repository[MemberExpense, MemberExpenseResponse, MemberExpenseRequest]
		MemberGenderHistoryManager          services.Repository[MemberGenderHistory, MemberGenderHistoryResponse, MemberGenderHistoryRequest]
		MemberGenderManager                 services.Repository[MemberGender, MemberGenderResponse, MemberGenderRequest]
		MemberGovernmentBenefitManager      services.Repository[MemberGovernmentBenefit, MemberGovernmentBenefitResponse, MemberGovernmentBenefitRequest]
		MemberGroupHistoryManager           services.Repository[MemberGroupHistory, MemberGroupHistoryResponse, MemberGroupHistoryRequest]
		MemberGroupManager                  services.Repository[MemberGroup, MemberGroupResponse, MemberGroupRequest]
		MemberIncomeManager                 services.Repository[MemberIncome, MemberIncomeResponse, MemberIncomeRequest]
		MemberJointAccountManager           services.Repository[MemberJointAccount, MemberJointAccountResponse, MemberJointAccountRequest]
		MemberMutualFundHistoryManager      services.Repository[MemberMutualFundHistory, MemberMutualFundHistoryResponse, MemberMutualFundHistoryRequest]
		MemberOccupationHistoryManager      services.Repository[MemberOccupationHistory, MemberOccupationHistoryResponse, MemberOccupationHistoryRequest]
		MemberOccupationManager             services.Repository[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]
		MemberOtherInformationEntryManager  services.Repository[MemberOtherInformationEntry, MemberOtherInformationEntryResponse, MemberOtherInformationEntryRequest]
		MemberRelativeAccountManager        services.Repository[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]
		MemberTypeHistoryManager            services.Repository[MemberTypeHistory, MemberTypeHistoryResponse, MemberTypeHistoryRequest]
		MemberTypeManager                   services.Repository[MemberType, MemberTypeResponse, MemberTypeRequest]
		MemberVerificationManager           services.Repository[MemberVerification, MemberVerificationResponse, MemberVerificationRequest]
		MemberProfileManager                services.Repository[MemberProfile, MemberProfileResponse, MemberProfileRequest]
		CollectorsMemberAccountEntryManager services.Repository[CollectorsMemberAccountEntry, CollectorsMemberAccountEntryResponse, CollectorsMemberAccountEntryRequest]
		MemberDepartmentManager             services.Repository[MemberDepartment, MemberDepartmentResponse, MemberDepartmentRequest]
		MemberDepartmentHistoryManager      services.Repository[MemberDepartmentHistory, MemberDepartmentHistoryResponse, MemberDepartmentHistoryRequest]

		// Employee Feature
		TimesheetManager services.Repository[Timesheet, TimesheetResponse, TimesheetRequest]
		CompanyManager   services.Repository[Company, CompanyResponse, CompanyRequest]

		// GL/FS
		FinancialStatementDefinitionManager             services.Repository[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]
		FinancialStatementGroupingManager               services.Repository[FinancialStatementGrouping, FinancialStatementGroupingResponse, FinancialStatementGroupingRequest]
		GeneralLedgerAccountsGroupingManager            services.Repository[GeneralLedgerAccountsGrouping, GeneralLedgerAccountsGroupingResponse, GeneralLedgerAccountsGroupingRequest]
		GeneralLedgerDefinitionManager                  services.Repository[GeneralLedgerDefinition, GeneralLedgerDefinitionResponse, GeneralLedgerDefinitionRequest]
		GeneralAccountGroupingNetSurplusPositiveManager services.Repository[GeneralAccountGroupingNetSurplusPositive, GeneralAccountGroupingNetSurplusPositiveResponse, GeneralAccountGroupingNetSurplusPositiveRequest]
		GeneralAccountGroupingNetSurplusNegativeManager services.Repository[GeneralAccountGroupingNetSurplusNegative, GeneralAccountGroupingNetSurplusNegativeResponse, GeneralAccountGroupingNetSurplusNegativeRequest]

		// MAINTENANCE TABLE FOR ACCOUNTING
		AccountClassificationManager services.Repository[AccountClassification, AccountClassificationResponse, AccountClassificationRequest]
		AccountCategoryManager       services.Repository[AccountCategory, AccountCategoryResponse, AccountCategoryRequest]
		PaymentTypeManager           services.Repository[PaymentType, PaymentTypeResponse, PaymentTypeRequest]
		BillAndCoinsManager          services.Repository[BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest]

		// ACCOUNT
		AccountManager           services.Repository[Account, AccountResponse, AccountRequest]
		AccountTagManager        services.Repository[AccountTag, AccountTagResponse, AccountTagRequest]
		AccountHistoryManager    services.Repository[AccountHistory, AccountHistoryResponse, AccountHistoryRequest]
		UnbalancedAccountManager services.Repository[UnbalancedAccount, UnbalancedAccountResponse, UnbalancedAccountRequest]

		// LEDGERS
		GeneralLedgerManager          services.Repository[GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest]
		GeneralLedgerTagManager       services.Repository[GeneralLedgerTag, GeneralLedgerTagResponse, GeneralLedgerTagRequest]
		MemberAccountingLedgerManager services.Repository[MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest]

		// TRANSACTION START
		TransactionBatchManager services.Repository[TransactionBatch, TransactionBatchResponse, TransactionBatchRequest]
		CheckRemittanceManager  services.Repository[CheckRemittance, CheckRemittanceResponse, CheckRemittanceRequest]
		OnlineRemittanceManager services.Repository[OnlineRemittance, OnlineRemittanceResponse, OnlineRemittanceRequest]
		CashCountManager        services.Repository[CashCount, CashCountResponse, CashCountRequest]
		BatchFundingManager     services.Repository[BatchFunding, BatchFundingResponse, BatchFundingRequest]
		TransactionManager      services.Repository[Transaction, TransactionResponse, TransactionRequest]
		TransactionTagManager   services.Repository[TransactionTag, TransactionTagResponse, TransactionTagRequest]

		// Disbursements
		DisbursementTransactionManager services.Repository[DisbursementTransaction, DisbursementTransactionResponse, DisbursementTransactionRequest]
		DisbursementManager            services.Repository[Disbursement, DisbursementResponse, DisbursementRequest]

		// LOAN START
		ComputationSheetManager                      services.Repository[ComputationSheet, ComputationSheetResponse, ComputationSheetRequest]
		FinesMaturityManager                         services.Repository[FinesMaturity, FinesMaturityResponse, FinesMaturityRequest]
		InterestMaturityManager                      services.Repository[InterestMaturity, InterestMaturityResponse, InterestMaturityRequest]
		IncludeNegativeAccountManager                services.Repository[IncludeNegativeAccount, IncludeNegativeAccountResponse, IncludeNegativeAccountRequest]
		AutomaticLoanDeductionManager                services.Repository[AutomaticLoanDeduction, AutomaticLoanDeductionResponse, AutomaticLoanDeductionRequest]
		BrowseExcludeIncludeAccountsManager          services.Repository[BrowseExcludeIncludeAccounts, BrowseExcludeIncludeAccountsResponse, BrowseExcludeIncludeAccountsRequest]
		MemberClassificationInterestRateManager      services.Repository[MemberClassificationInterestRate, MemberClassificationInterestRateResponse, MemberClassificationInterestRateRequest]
		LoanGuaranteedFundPerMonthManager            services.Repository[LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest]
		LoanStatusManager                            services.Repository[LoanStatus, LoanStatusResponse, LoanStatusRequest]
		LoanGuaranteedFundManager                    services.Repository[LoanGuaranteedFund, LoanGuaranteedFundResponse, LoanGuaranteedFundRequest]
		LoanTransactionManager                       services.Repository[LoanTransaction, LoanTransactionResponse, LoanTransactionRequest]
		LoanClearanceAnalysisManager                 services.Repository[LoanClearanceAnalysis, LoanClearanceAnalysisResponse, LoanClearanceAnalysisRequest]
		LoanClearanceAnalysisInstitutionManager      services.Repository[LoanClearanceAnalysisInstitution, LoanClearanceAnalysisInstitutionResponse, LoanClearanceAnalysisInstitutionRequest]
		LoanComakerMemberManager                     services.Repository[LoanComakerMember, LoanComakerMemberResponse, LoanComakerMemberRequest]
		ComakerMemberProfileManager                  services.Repository[ComakerMemberProfile, ComakerMemberProfileResponse, ComakerMemberProfileRequest]
		ComakerCollateralManager                     services.Repository[ComakerCollateral, ComakerCollateralResponse, ComakerCollateralRequest]
		LoanTransactionEntryManager                  services.Repository[LoanTransactionEntry, LoanTransactionEntryResponse, LoanTransactionEntryRequest]
		LoanTagManager                               services.Repository[LoanTag, LoanTagResponse, LoanTagRequest]
		LoanTermsAndConditionSuggestedPaymentManager services.Repository[LoanTermsAndConditionSuggestedPayment, LoanTermsAndConditionSuggestedPaymentResponse, LoanTermsAndConditionSuggestedPaymentRequest]
		LoanTermsAndConditionAmountReceiptManager    services.Repository[LoanTermsAndConditionAmountReceipt, LoanTermsAndConditionAmountReceiptResponse, LoanTermsAndConditionAmountReceiptRequest]
		LoanPurposeManager                           services.Repository[LoanPurpose, LoanPurposeResponse, LoanPurposeRequest]
		LoanLedgerManager                            services.Repository[LoanLedger, LoanLedgerResponse, LoanLedgerRequest]

		// Maintenance
		CollateralManager                                                   services.Repository[Collateral, CollateralResponse, CollateralRequest]
		TagTemplateManager                                                  services.Repository[TagTemplate, TagTemplateResponse, TagTemplateRequest]
		HolidayManager                                                      services.Repository[Holiday, HolidayResponse, HolidayRequest]
		GroceryComputationSheetManager                                      services.Repository[GroceryComputationSheet, GroceryComputationSheetResponse, GroceryComputationSheetRequest]
		GroceryComputationSheetMonthlyManager                               services.Repository[GroceryComputationSheetMonthly, GroceryComputationSheetMonthlyResponse, GroceryComputationSheetMonthlyRequest]
		InterestRateSchemeManager                                           services.Repository[InterestRateScheme, InterestRateSchemeResponse, InterestRateSchemeRequest]
		InterestRateByTermsHeaderManager                                    services.Repository[InterestRateByTermsHeader, InterestRateByTermsHeaderResponse, InterestRateByTermsHeaderRequest]
		InterestRateByTermManager                                           services.Repository[InterestRateByTerm, InterestRateByTermResponse, InterestRateByTermRequest]
		InterestRatePercentageManager                                       services.Repository[InterestRatePercentage, InterestRatePercentageResponse, InterestRatePercentageRequest]
		MemberTypeReferenceManager                                          services.Repository[MemberTypeReference, MemberTypeReferenceResponse, MemberTypeReferenceRequest]
		MemberTypeReferenceByAmountManager                                  services.Repository[MemberTypeReferenceByAmount, MemberTypeReferenceByAmountResponse, MemberTypeReferenceByAmountRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDateManager        services.Repository[MemberTypeReferenceInterestRateByUltimaMembershipDate, MemberTypeReferenceInterestRateByUltimaMembershipDateResponse, MemberTypeReferenceInterestRateByUltimaMembershipDateRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearManager services.Repository[MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearRequest]
		MemberDeductionEntryManager                                         services.Repository[MemberDeductionEntry, MemberDeductionEntryResponse, MemberDeductionEntryRequest]
		PostDatedCheckManager                                               services.Repository[PostDatedCheck, PostDatedCheckResponse, PostDatedCheckRequest]

		// TIME DEPOSIT
		TimeDepositTypeManager                   services.Repository[TimeDepositType, TimeDepositTypeResponse, TimeDepositTypeRequest]
		TimeDepositComputationManager            services.Repository[TimeDepositComputation, TimeDepositComputationResponse, TimeDepositComputationRequest]
		TimeDepositComputationPreMatureManager   services.Repository[TimeDepositComputationPreMature, TimeDepositComputationPreMatureResponse, TimeDepositComputationPreMatureRequest]
		ChargesRateSchemeManager                 services.Repository[ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest]
		ChargesRateSchemeAccountManager          services.Repository[ChargesRateSchemeAccount, ChargesRateSchemeAccountResponse, ChargesRateSchemeAccountRequest]
		ChargesRateByRangeOrMinimumAmountManager services.Repository[ChargesRateByRangeOrMinimumAmount, ChargesRateByRangeOrMinimumAmountResponse, ChargesRateByRangeOrMinimumAmountRequest]
		ChargesRateByTermManager                 services.Repository[ChargesRateByTerm, ChargesRateByTermResponse, ChargesRateByTermRequest]

		// ACCOUNTING ENTRY
		AdjustmentEntryManager           services.Repository[AdjustmentEntry, AdjustmentEntryResponse, AdjustmentEntryRequest]
		AdjustmentTagManager             services.Repository[AdjustmentTag, AdjustmentTagResponse, AdjustmentTagRequest]
		VoucherPayToManager              services.Repository[VoucherPayTo, VoucherPayToResponse, VoucherPayToRequest]
		CashCheckVoucherManager          services.Repository[CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest]
		CashCheckVoucherEntryManager     services.Repository[CashCheckVoucherEntry, CashCheckVoucherEntryResponse, CashCheckVoucherEntryRequest]
		CashCheckVoucherTagManager       services.Repository[CashCheckVoucherTag, CashCheckVoucherTagResponse, CashCheckVoucherTagRequest]
		CancelledCashCheckVoucherManager services.Repository[CancelledCashCheckVoucher, CancelledCashCheckVoucherResponse, CancelledCashCheckVoucherRequest]
		JournalVoucherManager            services.Repository[JournalVoucher, JournalVoucherResponse, JournalVoucherRequest]
		JournalVoucherEntryManager       services.Repository[JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest]
		JournalVoucherTagManager         services.Repository[JournalVoucherTag, JournalVoucherTagResponse, JournalVoucherTagRequest]

		FundsManager                          services.Repository[Funds, FundsResponse, FundsRequest]
		ChargesRateSchemeModeOfPaymentManager services.Repository[ChargesRateSchemeModeOfPayment, ChargesRateSchemeModeOfPaymentResponse, ChargesRateSchemeModeOfPaymentRequest]
	}
)

// Newmodelcore creates a new instance of modelcore with the provided service provider
func Newmodelcore(provider *server.Provider) (*ModelCore, error) {
	return &ModelCore{
		provider: provider,
	}, nil
}

// Start initializes all model managers and performs auto-migration setup
func (m *ModelCore) start(_ context.Context) error {

	// Models
	m.accountCategory()
	m.accountClassification()
	m.unbalancedAccount()
	m.adjustmentEntry()
	m.adjustmentTag()
	m.automaticLoanDeduction()
	m.bank()
	m.batchFunding()
	m.billAndCoins()
	m.branch()
	m.browseExcludeIncludeAccounts()
	m.cashCheckVoucher()
	m.cashCheckVoucherEntry()
	m.cashCheckVoucherTag()
	m.cancelledCashCheckVoucher()
	m.cashCount()
	m.category()
	m.currency()
	m.chargesRateByRangeOrMinimumAmount()
	m.chargesRateByTerm()
	m.chargesRateSchemeAccount()
	m.chargesRateScheme()
	m.checkRemittance()
	m.collateral()
	m.collectorsMemberAccountEntry()
	m.computationSheet()
	m.contactUs()
	m.disbursement()
	m.disbursementTransaction()
	m.feedback()
	m.financialStatementGrouping()
	m.financialStatementDefinition()
	m.finesMaturity()
	m.footstep()
	m.generalAccountGroupingNetSurplusNegative()
	m.generalAccountGroupingNetSurplusPositive()
	m.generalLedger()
	m.generalLedgerTag()
	m.generalLedgerAccountsGrouping()
	m.generatedReport()
	m.generalLedgerDefinition()
	m.account()
	m.accountTag()
	m.accountHistory()
	m.groceryComputationSheet()
	m.groceryComputationSheetMonthly()
	m.holiday()
	m.includeNegativeAccount()
	m.interestMaturity()
	m.interestRateByTerm()
	m.interestRateByTermsHeader()
	m.interestRatePercentage()
	m.interestRateScheme()
	m.invitationCode()
	m.journalVoucher()
	m.journalVoucherEntry()
	m.journalVoucherTag()
	m.loanClearanceAnalysis()
	m.loanClearanceAnalysisInstitution()
	m.loanComakerMember()
	m.comakerMemberProfile()
	m.comakerCollateral()
	m.loanGuaranteedFund()
	m.loanGuaranteedFundPerMonth()
	m.loanLedger()
	m.loanPurpose()
	m.loanStatus()
	m.loanTag()
	m.loanTermsAndConditionAmountReceipt()
	m.loanTermsAndConditionSuggestedPayment()
	m.loanTransactionEntry()
	m.loanTransaction()
	m.media()
	m.memberAccountingLedger()
	m.memberAddress()
	m.memberAsset()
	m.memberBankCard()
	m.memberCenter()
	m.memberCenterHistory()
	m.memberClassification()
	m.memberClassificationHistory()
	m.memberClassificationInterestRate()
	m.memberContactReference()
	m.memberDamayanExtensionEntry()
	m.memberDeductionEntry()
	m.memberEducationalAttainment()
	m.memberExpense()
	m.memberGender()
	m.memberGenderHistory()
	m.memberGovernmentBenefit()
	m.memberGroup()
	m.memberGroupHistory()
	m.memberIncome()
	m.memberJointAccount()
	m.memberMutualFundHistory()
	m.memberOccupation()
	m.memberOccupationHistory()
	m.memberOtherInformationEntry()
	m.memberProfile()
	m.memberCloseRemark()
	m.memberRelativeAccount()
	m.memberType()
	m.memberTypeHistory()
	m.memberTypeReferenceByAmount()
	m.memberTypeReference()
	m.memberTypeReferenceInterestRateByUltimaMembershipDate()
	m.memberTypeReferenceInterestRateByUltimaMembershipDatePerYear()
	m.memberVerification()
	m.notification()
	m.onlineRemittance()
	m.organizationCategory()
	m.organizationDailyUsage()
	m.organization()
	m.organizationMedia()
	m.paymentType()
	m.permissionTemplate()
	m.postDatedCheck()
	m.subscriptionPlan()
	m.tagTemplate()
	m.timeDepositComputation()
	m.timeDepositComputationPreMature()
	m.timeDepositType()
	m.timesheet()
	m.transactionBatch()
	m.transaction()
	m.transactionTag()
	m.user()
	m.userOrganization()
	m.userRating()
	m.memberProfileMedia()
	m.voucherPayTo()
	m.memberDepartment()
	m.memberDepartmentHistory()
	m.funds()
	m.chargesRateSchemeModeOfPayment()
	m.branchSetting()
	m.company()
	return nil
}

// OrganizationSeeder seeds initial data for a new organization including default accounts, payment types, and templates
func (m *ModelCore) organizationSeeder(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	if err := m.invitationCodeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.bankSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}

	if err := m.billAndCoinsSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.holidaySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberClassificationSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberGenderSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberGroupSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberCenterSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberOccupationSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberTypeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.memberDepartmentSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.generalLedgerAccountsGroupingSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.financialStatementGroupingSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.paymentTypeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.accountSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.loanPurposeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return nil
	}
	if err := m.accountCategorySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.disbursementSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.collateralSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.tagTemplateSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.loanStatusSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg, err := m.userOrganizationManager.FindOne(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
	if err != nil {
		return err
	}
	if err := m.memberProfileSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg.IsSeeded = true
	if err := m.userOrganizationManager.UpdateByIDWithTx(context, tx, userOrg.ID, userOrg); err != nil {
		return err
	}
	if err := m.companySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

// OrganizationDestroyer cleans up and removes all data associated with an organization branch
func (m *ModelCore) organizationDestroyer(ctx context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	invitationCodes, err := m.invitationCodeManager.Find(ctx, &InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get invitation codes")
	}
	banks, err := m.bankManager.Find(ctx, &Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get banks")
	}
	billAndCoins, err := m.billAndCoinsManager.Find(ctx, &BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get bill and coins")
	}
	// 4. Delete Holidays
	holidays, err := m.holidayManager.Find(ctx, &Holiday{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get holidays")
	}
	for _, data := range holidays {
		if err := m.holidayManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := m.billAndCoinsManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bill or coin %s", data.Name)
		}
	}
	for _, data := range banks {
		if err := m.bankManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bank %s", data.Name)
		}
	}
	for _, data := range invitationCodes {
		if err := m.invitationCodeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy invitation code %s", data.Code)
		}
	}

	// 1. Delete MemberType
	memberTypes, err := m.memberTypeManager.Find(ctx, &MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member types")
	}
	for _, data := range memberTypes {
		if err := m.memberTypeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member type %s", data.Name)
		}
	}

	// 2. Delete MemberOccupation
	memberOccupations, err := m.memberOccupationManager.Find(ctx, &MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member occupations")
	}
	for _, data := range memberOccupations {
		if err := m.memberOccupationManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member occupation %s", data.Name)
		}
	}

	// 3. Delete MemberGroup
	memberGroups, err := m.memberGroupManager.Find(ctx, &MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member groups")
	}
	for _, data := range memberGroups {
		if err := m.memberGroupManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member group %s", data.Name)
		}
	}

	// 4. Delete MemberGender
	memberGenders, err := m.memberGenderManager.Find(ctx, &MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member genders")
	}
	for _, data := range memberGenders {
		if err := m.memberGenderManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member gender %s", data.Name)
		}
	}

	// 5. Delete MemberCenter
	memberCenters, err := m.memberCenterManager.Find(ctx, &MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member centers")
	}
	for _, data := range memberCenters {
		if err := m.memberCenterManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member center %s", data.Name)
		}
	}

	// 6. Delete MemberClassification
	memberClassifications, err := m.memberClassificationManager.Find(ctx, &MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member classifications")
	}
	for _, data := range memberClassifications {
		if err := m.memberClassificationManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member classification %s", data.Name)
		}
	}

	generalLedgerDefinitions, err := m.generalLedgerDefinitionManager.Find(ctx, &GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger definitions")
	}
	for _, data := range generalLedgerDefinitions {
		if err := m.generalLedgerDefinitionManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger definition %s", data.Name)
		}
	}

	generalLedgerAccountsGroupings, err := m.generalLedgerAccountsGroupingManager.Find(ctx, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger accounts groupings")
	}
	for _, data := range generalLedgerAccountsGroupings {
		if err := m.generalLedgerAccountsGroupingManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger accounts grouping %s", data.Name)
		}
	}

	// Financial Statement Accounts Grouping destroyer
	FinancialStatementGroupings, err := m.financialStatementGroupingManager.Find(ctx, &FinancialStatementGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get financial statement accounts groupings")
	}
	for _, data := range FinancialStatementGroupings {
		if err := m.financialStatementGroupingManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy financial statement accounts grouping %s", data.Name)
		}
	}
	paymentTypes, err := m.paymentTypeManager.Find(ctx, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get payment types")
	}
	for _, data := range paymentTypes {
		if err := m.paymentTypeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy payment type %s", data.Name)
		}
	}
	disbursements, err := m.disbursementManager.Find(ctx, &Disbursement{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get disbursements")
	}
	for _, data := range disbursements {
		if err := m.disbursementManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy disbursement %s", data.Name)
		}
	}
	collaterals, err := m.collateralManager.Find(ctx, &Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get collaterals")
	}
	for _, data := range collaterals {
		if err := m.collateralManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy collateral %s", data.Name)
		}
	}

	// Delete Accounts
	accounts, err := m.accountManager.Find(ctx, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get accounts")
	}
	for _, data := range accounts {
		if err := m.accountManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account %s", data.Name)
		}
	}

	// Delete LoanPurpose
	loanPurposes, err := m.loanPurposeManager.Find(ctx, &LoanPurpose{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan purposes")
	}
	for _, data := range loanPurposes {
		if err := m.loanPurposeManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan purpose %s", data.Description)
		}
	}

	// Delete AccountCategory
	accountCategories, err := m.accountCategoryManager.Find(ctx, &AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get account categories")
	}
	for _, data := range accountCategories {
		if err := m.accountCategoryManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account category %s", data.Name)
		}
	}

	// Delete TagTemplate
	tagTemplates, err := m.tagTemplateManager.Find(ctx, &TagTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get tag templates")
	}
	for _, data := range tagTemplates {
		if err := m.tagTemplateManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy tag template %s", data.Name)
		}
	}

	// Delete LoanStatus
	loanStatuses, err := m.loanStatusManager.Find(ctx, &LoanStatus{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan statuses")
	}
	for _, data := range loanStatuses {
		if err := m.loanStatusManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan status %s", data.Name)
		}
	}

	// Delete MemberProfile
	memberProfiles, err := m.memberProfileManager.Find(ctx, &MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member profiles")
	}
	for _, data := range memberProfiles {
		if err := m.memberProfileManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member profile %s %s", data.FirstName, data.LastName)
		}
	}

	// Delete Company
	companies, err := m.companyManager.Find(ctx, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get companies")
	}
	for _, data := range companies {
		if err := m.companyManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy company %s", data.Name)
		}
	}

	// Delete MemberDepartment
	memberDepartments, err := m.memberDepartmentManager.Find(ctx, &MemberDepartment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member departments")
	}
	for _, data := range memberDepartments {
		if err := m.memberDepartmentManager.DeleteByIDWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member department %s", data.Name)
		}
	}

	return nil
}
