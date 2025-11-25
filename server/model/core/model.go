package core

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	IDSRequest struct {
		IDs uuid.UUIDs `json:"ids"`
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

	Core struct {
		provider  *server.Provider
		Migration []any

		BankManager                          registry.Registry[Bank, BankResponse, BankRequest]
		BranchManager                        registry.Registry[Branch, BranchResponse, BranchRequest]
		BranchSettingManager                 registry.Registry[BranchSetting, BranchSettingResponse, BranchSettingRequest]
		CategoryManager                      registry.Registry[Category, CategoryResponse, CategoryRequest]
		ContactUsManager                     registry.Registry[ContactUs, ContactUsResponse, ContactUsRequest]
		CurrencyManager                      registry.Registry[Currency, CurrencyResponse, CurrencyRequest]
		FeedbackManager                      registry.Registry[Feedback, FeedbackResponse, FeedbackRequest]
		FootstepManager                      registry.Registry[Footstep, FootstepResponse, FootstepRequest]
		GeneratedReportManager               registry.Registry[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]
		GeneratedReportsDownloadUsersManager registry.Registry[GeneratedReportsDownloadUsers, GeneratedReportsDownloadUsersResponse, GeneratedReportsDownloadUsersRequest]
		GeneratedSavingsInterestManager      registry.Registry[GeneratedSavingsInterest, GeneratedSavingsInterestResponse, GeneratedSavingsInterestRequest]
		GenerateSavingsInterestEntryManager  registry.Registry[GenerateSavingsInterestEntry, GenerateSavingsInterestEntryResponse, GenerateSavingsInterestEntryRequest]
		InvitationCodeManager                registry.Registry[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]
		MediaManager                         registry.Registry[Media, MediaResponse, MediaRequest]
		NotificationManager                  registry.Registry[Notification, NotificationResponse, any]
		OrganizationCategoryManager          registry.Registry[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]
		OrganizationDailyUsageManager        registry.Registry[OrganizationDailyUsage, OrganizationDailyUsageResponse, OrganizationDailyUsageRequest]
		OrganizationManager                  registry.Registry[Organization, OrganizationResponse, OrganizationRequest]
		OrganizationMediaManager             registry.Registry[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest]
		PermissionTemplateManager            registry.Registry[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]
		SubscriptionPlanManager              registry.Registry[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]
		UserOrganizationManager              registry.Registry[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]
		UserManager                          registry.Registry[User, UserResponse, UserRegisterRequest]
		UserRatingManager                    registry.Registry[UserRating, UserRatingResponse, UserRatingRequest]
		MemberProfileMediaManager            registry.Registry[MemberProfileMedia, MemberProfileMediaResponse, MemberProfileMediaRequest]
		MemberProfileArchiveManager          registry.Registry[MemberProfileArchive, MemberProfileArchiveResponse, MemberProfileArchiveRequest]
		// Members
		MemberAddressManager                registry.Registry[MemberAddress, MemberAddressResponse, MemberAddressRequest]
		MemberAssetManager                  registry.Registry[MemberAsset, MemberAssetResponse, MemberAssetRequest]
		MemberBankCardManager               registry.Registry[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest]
		MemberCenterHistoryManager          registry.Registry[MemberCenterHistory, MemberCenterHistoryResponse, MemberCenterHistoryRequest]
		MemberCenterManager                 registry.Registry[MemberCenter, MemberCenterResponse, MemberCenterRequest]
		MemberClassificationManager         registry.Registry[MemberClassification, MemberClassificationResponse, MemberClassificationRequest]
		MemberClassificationHistoryManager  registry.Registry[MemberClassificationHistory, MemberClassificationHistoryResponse, MemberClassificationHistoryRequest]
		MemberCloseRemarkManager            registry.Registry[MemberCloseRemark, MemberCloseRemarkResponse, MemberCloseRemarkRequest]
		MemberContactReferenceManager       registry.Registry[MemberContactReference, MemberContactReferenceResponse, MemberContactReferenceRequest]
		MemberDamayanExtensionEntryManager  registry.Registry[MemberDamayanExtensionEntry, MemberDamayanExtensionEntryResponse, MemberDamayanExtensionEntryRequest]
		MemberEducationalAttainmentManager  registry.Registry[MemberEducationalAttainment, MemberEducationalAttainmentResponse, MemberEducationalAttainmentRequest]
		MemberExpenseManager                registry.Registry[MemberExpense, MemberExpenseResponse, MemberExpenseRequest]
		MemberGenderHistoryManager          registry.Registry[MemberGenderHistory, MemberGenderHistoryResponse, MemberGenderHistoryRequest]
		MemberGenderManager                 registry.Registry[MemberGender, MemberGenderResponse, MemberGenderRequest]
		MemberGovernmentBenefitManager      registry.Registry[MemberGovernmentBenefit, MemberGovernmentBenefitResponse, MemberGovernmentBenefitRequest]
		MemberGroupHistoryManager           registry.Registry[MemberGroupHistory, MemberGroupHistoryResponse, MemberGroupHistoryRequest]
		MemberGroupManager                  registry.Registry[MemberGroup, MemberGroupResponse, MemberGroupRequest]
		MemberIncomeManager                 registry.Registry[MemberIncome, MemberIncomeResponse, MemberIncomeRequest]
		MemberJointAccountManager           registry.Registry[MemberJointAccount, MemberJointAccountResponse, MemberJointAccountRequest]
		MemberMutualFundHistoryManager      registry.Registry[MemberMutualFundHistory, MemberMutualFundHistoryResponse, MemberMutualFundHistoryRequest]
		MemberOccupationHistoryManager      registry.Registry[MemberOccupationHistory, MemberOccupationHistoryResponse, MemberOccupationHistoryRequest]
		MemberOccupationManager             registry.Registry[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]
		MemberOtherInformationEntryManager  registry.Registry[MemberOtherInformationEntry, MemberOtherInformationEntryResponse, MemberOtherInformationEntryRequest]
		MemberRelativeAccountManager        registry.Registry[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]
		MemberTypeHistoryManager            registry.Registry[MemberTypeHistory, MemberTypeHistoryResponse, MemberTypeHistoryRequest]
		MemberTypeManager                   registry.Registry[MemberType, MemberTypeResponse, MemberTypeRequest]
		MemberVerificationManager           registry.Registry[MemberVerification, MemberVerificationResponse, MemberVerificationRequest]
		MemberProfileManager                registry.Registry[MemberProfile, MemberProfileResponse, MemberProfileRequest]
		CollectorsMemberAccountEntryManager registry.Registry[CollectorsMemberAccountEntry, CollectorsMemberAccountEntryResponse, CollectorsMemberAccountEntryRequest]
		MemberDepartmentManager             registry.Registry[MemberDepartment, MemberDepartmentResponse, MemberDepartmentRequest]
		MemberDepartmentHistoryManager      registry.Registry[MemberDepartmentHistory, MemberDepartmentHistoryResponse, MemberDepartmentHistoryRequest]

		// Employee Feature
		TimesheetManager registry.Registry[Timesheet, TimesheetResponse, TimesheetRequest]
		CompanyManager   registry.Registry[Company, CompanyResponse, CompanyRequest]

		// GL/FS
		FinancialStatementDefinitionManager             registry.Registry[FinancialStatementDefinition, FinancialStatementDefinitionResponse, FinancialStatementDefinitionRequest]
		FinancialStatementGroupingManager               registry.Registry[FinancialStatementGrouping, FinancialStatementGroupingResponse, FinancialStatementGroupingRequest]
		GeneralLedgerAccountsGroupingManager            registry.Registry[GeneralLedgerAccountsGrouping, GeneralLedgerAccountsGroupingResponse, GeneralLedgerAccountsGroupingRequest]
		GeneralLedgerDefinitionManager                  registry.Registry[GeneralLedgerDefinition, GeneralLedgerDefinitionResponse, GeneralLedgerDefinitionRequest]
		GeneralAccountGroupingNetSurplusPositiveManager registry.Registry[GeneralAccountGroupingNetSurplusPositive, GeneralAccountGroupingNetSurplusPositiveResponse, GeneralAccountGroupingNetSurplusPositiveRequest]
		Generalaccountgroupingnetsurplusnegativemanager registry.Registry[GeneralAccountGroupingNetSurplusNegative, GeneralAccountGroupingNetSurplusNegativeResponse, GeneralAccountGroupingNetSurplusNegativeRequest]

		// MAINTENANCE TABLE FOR ACCOUNTING
		AccountClassificationManager registry.Registry[AccountClassification, AccountClassificationResponse, AccountClassificationRequest]
		AccountCategoryManager       registry.Registry[AccountCategory, AccountCategoryResponse, AccountCategoryRequest]
		PaymentTypeManager           registry.Registry[PaymentType, PaymentTypeResponse, PaymentTypeRequest]
		BillAndCoinsManager          registry.Registry[BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest]

		// ACCOUNT
		AccountManager           registry.Registry[Account, AccountResponse, AccountRequest]
		AccountTagManager        registry.Registry[AccountTag, AccountTagResponse, AccountTagRequest]
		AccountHistoryManager    registry.Registry[AccountHistory, AccountHistoryResponse, AccountHistoryRequest]
		UnbalancedAccountManager registry.Registry[UnbalancedAccount, UnbalancedAccountResponse, UnbalancedAccountRequest]

		// LEDGERS
		GeneralLedgerManager          registry.Registry[GeneralLedger, GeneralLedgerResponse, GeneralLedgerRequest]
		GeneralLedgerTagManager       registry.Registry[GeneralLedgerTag, GeneralLedgerTagResponse, GeneralLedgerTagRequest]
		MemberAccountingLedgerManager registry.Registry[MemberAccountingLedger, MemberAccountingLedgerResponse, MemberAccountingLedgerRequest]

		// TRANSACTION START
		TransactionBatchManager registry.Registry[TransactionBatch, TransactionBatchResponse, TransactionBatchRequest]
		CheckRemittanceManager  registry.Registry[CheckRemittance, CheckRemittanceResponse, CheckRemittanceRequest]
		OnlineRemittanceManager registry.Registry[OnlineRemittance, OnlineRemittanceResponse, OnlineRemittanceRequest]
		CashCountManager        registry.Registry[CashCount, CashCountResponse, CashCountRequest]
		BatchFundingManager     registry.Registry[BatchFunding, BatchFundingResponse, BatchFundingRequest]
		TransactionManager      registry.Registry[Transaction, TransactionResponse, TransactionRequest]
		TransactionTagManager   registry.Registry[TransactionTag, TransactionTagResponse, TransactionTagRequest]

		// Disbursements
		DisbursementTransactionManager registry.Registry[DisbursementTransaction, DisbursementTransactionResponse, DisbursementTransactionRequest]
		DisbursementManager            registry.Registry[Disbursement, DisbursementResponse, DisbursementRequest]

		// LOAN START
		ComputationSheetManager                      registry.Registry[ComputationSheet, ComputationSheetResponse, ComputationSheetRequest]
		FinesMaturityManager                         registry.Registry[FinesMaturity, FinesMaturityResponse, FinesMaturityRequest]
		InterestMaturityManager                      registry.Registry[InterestMaturity, InterestMaturityResponse, InterestMaturityRequest]
		IncludeNegativeAccountManager                registry.Registry[IncludeNegativeAccount, IncludeNegativeAccountResponse, IncludeNegativeAccountRequest]
		AutomaticLoanDeductionManager                registry.Registry[AutomaticLoanDeduction, AutomaticLoanDeductionResponse, AutomaticLoanDeductionRequest]
		BrowseExcludeIncludeAccountsManager          registry.Registry[BrowseExcludeIncludeAccounts, BrowseExcludeIncludeAccountsResponse, BrowseExcludeIncludeAccountsRequest]
		MemberClassificationInterestRateManager      registry.Registry[MemberClassificationInterestRate, MemberClassificationInterestRateResponse, MemberClassificationInterestRateRequest]
		LoanGuaranteedFundPerMonthManager            registry.Registry[LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest]
		LoanStatusManager                            registry.Registry[LoanStatus, LoanStatusResponse, LoanStatusRequest]
		LoanGuaranteedFundManager                    registry.Registry[LoanGuaranteedFund, LoanGuaranteedFundResponse, LoanGuaranteedFundRequest]
		LoanTransactionManager                       registry.Registry[LoanTransaction, LoanTransactionResponse, LoanTransactionRequest]
		LoanClearanceAnalysisManager                 registry.Registry[LoanClearanceAnalysis, LoanClearanceAnalysisResponse, LoanClearanceAnalysisRequest]
		LoanClearanceAnalysisInstitutionManager      registry.Registry[LoanClearanceAnalysisInstitution, LoanClearanceAnalysisInstitutionResponse, LoanClearanceAnalysisInstitutionRequest]
		LoanComakerMemberManager                     registry.Registry[LoanComakerMember, LoanComakerMemberResponse, LoanComakerMemberRequest]
		ComakerMemberProfileManager                  registry.Registry[ComakerMemberProfile, ComakerMemberProfileResponse, ComakerMemberProfileRequest]
		ComakerCollateralManager                     registry.Registry[ComakerCollateral, ComakerCollateralResponse, ComakerCollateralRequest]
		LoanTransactionEntryManager                  registry.Registry[LoanTransactionEntry, LoanTransactionEntryResponse, LoanTransactionEntryRequest]
		LoanTagManager                               registry.Registry[LoanTag, LoanTagResponse, LoanTagRequest]
		LoanTermsAndConditionSuggestedPaymentManager registry.Registry[LoanTermsAndConditionSuggestedPayment, LoanTermsAndConditionSuggestedPaymentResponse, LoanTermsAndConditionSuggestedPaymentRequest]
		LoanTermsAndConditionAmountReceiptManager    registry.Registry[LoanTermsAndConditionAmountReceipt, LoanTermsAndConditionAmountReceiptResponse, LoanTermsAndConditionAmountReceiptRequest]
		LoanPurposeManager                           registry.Registry[LoanPurpose, LoanPurposeResponse, LoanPurposeRequest]
		LoanAccountManager                           registry.Registry[LoanAccount, LoanAccountResponse, LoanAccountRequest]

		// Maintenance
		CollateralManager                                                   registry.Registry[Collateral, CollateralResponse, CollateralRequest]
		TagTemplateManager                                                  registry.Registry[TagTemplate, TagTemplateResponse, TagTemplateRequest]
		HolidayManager                                                      registry.Registry[Holiday, HolidayResponse, HolidayRequest]
		GroceryComputationSheetManager                                      registry.Registry[GroceryComputationSheet, GroceryComputationSheetResponse, GroceryComputationSheetRequest]
		GroceryComputationSheetMonthlyManager                               registry.Registry[GroceryComputationSheetMonthly, GroceryComputationSheetMonthlyResponse, GroceryComputationSheetMonthlyRequest]
		InterestRateSchemeManager                                           registry.Registry[InterestRateScheme, InterestRateSchemeResponse, InterestRateSchemeRequest]
		InterestRateByTermsHeaderManager                                    registry.Registry[InterestRateByTermsHeader, InterestRateByTermsHeaderResponse, InterestRateByTermsHeaderRequest]
		InterestRateByTermManager                                           registry.Registry[InterestRateByTerm, InterestRateByTermResponse, InterestRateByTermRequest]
		InterestRatePercentageManager                                       registry.Registry[InterestRatePercentage, InterestRatePercentageResponse, InterestRatePercentageRequest]
		MemberTypeReferenceManager                                          registry.Registry[MemberTypeReference, MemberTypeReferenceResponse, MemberTypeReferenceRequest]
		MemberTypeReferenceByAmountManager                                  registry.Registry[MemberTypeReferenceByAmount, MemberTypeReferenceByAmountResponse, MemberTypeReferenceByAmountRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDateManager        registry.Registry[MemberTypeReferenceInterestRateByUltimaMembershipDate, MemberTypeReferenceInterestRateByUltimaMembershipDateResponse, MemberTypeReferenceInterestRateByUltimaMembershipDateRequest]
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearManager registry.Registry[MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse, MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearRequest]
		MemberDeductionEntryManager                                         registry.Registry[MemberDeductionEntry, MemberDeductionEntryResponse, MemberDeductionEntryRequest]
		PostDatedCheckManager                                               registry.Registry[PostDatedCheck, PostDatedCheckResponse, PostDatedCheckRequest]

		// TIME DEPOSIT
		TimeDepositTypeManager                   registry.Registry[TimeDepositType, TimeDepositTypeResponse, TimeDepositTypeRequest]
		TimeDepositComputationManager            registry.Registry[TimeDepositComputation, TimeDepositComputationResponse, TimeDepositComputationRequest]
		TimeDepositComputationPreMatureManager   registry.Registry[TimeDepositComputationPreMature, TimeDepositComputationPreMatureResponse, TimeDepositComputationPreMatureRequest]
		ChargesRateSchemeManager                 registry.Registry[ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest]
		ChargesRateSchemeAccountManager          registry.Registry[ChargesRateSchemeAccount, ChargesRateSchemeAccountResponse, ChargesRateSchemeAccountRequest]
		ChargesRateByRangeOrMinimumAmountManager registry.Registry[ChargesRateByRangeOrMinimumAmount, ChargesRateByRangeOrMinimumAmountResponse, ChargesRateByRangeOrMinimumAmountRequest]
		ChargesRateByTermManager                 registry.Registry[ChargesRateByTerm, ChargesRateByTermResponse, ChargesRateByTermRequest]

		// ACCOUNTING ENTRY
		AdjustmentEntryManager           registry.Registry[AdjustmentEntry, AdjustmentEntryResponse, AdjustmentEntryRequest]
		AdjustmentTagManager             registry.Registry[AdjustmentTag, AdjustmentTagResponse, AdjustmentTagRequest]
		VoucherPayToManager              registry.Registry[VoucherPayTo, VoucherPayToResponse, VoucherPayToRequest]
		CashCheckVoucherManager          registry.Registry[CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest]
		CashCheckVoucherEntryManager     registry.Registry[CashCheckVoucherEntry, CashCheckVoucherEntryResponse, CashCheckVoucherEntryRequest]
		CashCheckVoucherTagManager       registry.Registry[CashCheckVoucherTag, CashCheckVoucherTagResponse, CashCheckVoucherTagRequest]
		CancelledCashCheckVoucherManager registry.Registry[CancelledCashCheckVoucher, CancelledCashCheckVoucherResponse, CancelledCashCheckVoucherRequest]
		JournalVoucherManager            registry.Registry[JournalVoucher, JournalVoucherResponse, JournalVoucherRequest]
		JournalVoucherEntryManager       registry.Registry[JournalVoucherEntry, JournalVoucherEntryResponse, JournalVoucherEntryRequest]
		JournalVoucherTagManager         registry.Registry[JournalVoucherTag, JournalVoucherTagResponse, JournalVoucherTagRequest]

		FundsManager                          registry.Registry[Funds, FundsResponse, FundsRequest]
		ChargesRateSchemeModeOfPaymentManager registry.Registry[ChargesRateSchemeModeOfPayment, ChargesRateSchemeModeOfPaymentResponse, ChargesRateSchemeModeOfPaymentRequest]
		BrowseReferenceManager                registry.Registry[BrowseReference, BrowseReferenceResponse, BrowseReferenceRequest]
		InterestRateByYearManager             registry.Registry[InterestRateByYear, InterestRateByYearResponse, InterestRateByYearRequest]
		InterestRateByDateManager             registry.Registry[InterestRateByDate, InterestRateByDateResponse, InterestRateByDateRequest]
		InterestRateByAmountManager           registry.Registry[InterestRateByAmount, InterestRateByAmountResponse, InterestRateByAmountRequest]
	}
)

func NewCore(provider *server.Provider) (*Core, error) {
	return &Core{
		provider: provider,
	}, nil
}

// Start initializes all model managers and performs auto-migration setup
func (m *Core) Start() error {

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
	m.browseReference()
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
	m.generatedReportsDownloadUsers()
	m.generatedSavingsInterest()
	m.generateSavingsInterestEntry()
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
	m.interestRateByYear()
	m.interestRateByDate()
	m.interestRateByAmount()
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
	m.loanPurpose()
	m.loanStatus()
	m.loanTag()
	m.loanTermsAndConditionAmountReceipt()
	m.loanTermsAndConditionSuggestedPayment()
	m.loanTransactionEntry()
	m.loanTransaction()
	m.loanAccount()
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
	m.memberProfileArchive()

	return nil
}
func (m *Core) GlobalSeeder(ctx context.Context) error {
	if err := m.currencySeed(ctx); err != nil {
		return err
	}
	if err := m.categorySeed(ctx); err != nil {
		return err
	}
	if err := m.subscriptionPlanSeed(ctx); err != nil {
		return err
	}
	return nil
}

// OrganizationSeeder seeds initial data for a new organization including default accounts, payment types, and templates
func (m *Core) OrganizationSeeder(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
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
	if err := m.accountSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := m.loanPurposeSeed(context, tx, userID, organizationID, branchID); err != nil {
		return nil
	}
	if err := m.accountClassificationSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
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
	if err := m.permissionTemplateSeed(context, tx, userID, organizationID, branchID); err != nil {
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
	if err := m.memberProfileSeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg.IsSeeded = true
	if err := m.UserOrganizationManager.UpdateByIDWithTx(context, tx, userOrg.ID, userOrg); err != nil {
		return err
	}
	if err := m.companySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

// OrganizationDestroyer cleans up and removes all data associated with an organization branch
func (m *Core) OrganizationDestroyer(ctx context.Context, tx *gorm.DB, organizationID uuid.UUID, branchID uuid.UUID) error {
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
		if err := m.HolidayManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := m.BillAndCoinsManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bill or coin %s", data.Name)
		}
	}
	for _, data := range banks {
		if err := m.BankManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bank %s", data.Name)
		}
	}
	for _, data := range invitationCodes {
		if err := m.InvitationCodeManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberTypeManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberOccupationManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberGroupManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberGenderManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberCenterManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.MemberClassificationManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.GeneralLedgerDefinitionManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.GeneralLedgerAccountsGroupingManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.FinancialStatementGroupingManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.PaymentTypeManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.DisbursementManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
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
		if err := m.CollateralManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy collateral %s", data.Name)
		}
	}

	// Delete Accounts
	accounts, err := m.AccountManager.Find(ctx, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get accounts")
	}
	for _, data := range accounts {
		if err := m.AccountManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account %s", data.Name)
		}
	}

	// Delete LoanPurpose
	loanPurposes, err := m.LoanPurposeManager.Find(ctx, &LoanPurpose{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan purposes")
	}
	for _, data := range loanPurposes {
		if err := m.LoanPurposeManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan purpose %s", data.Description)
		}
	}

	// Delete AccountCategory
	accountCategories, err := m.AccountCategoryManager.Find(ctx, &AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get account categories")
	}
	for _, data := range accountCategories {
		if err := m.AccountCategoryManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account category %s", data.Name)
		}
	}

	// Delete TagTemplate
	tagTemplates, err := m.TagTemplateManager.Find(ctx, &TagTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get tag templates")
	}
	for _, data := range tagTemplates {
		if err := m.TagTemplateManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy tag template %s", data.Name)
		}
	}

	// Delete LoanStatus
	loanStatuses, err := m.LoanStatusManager.Find(ctx, &LoanStatus{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan statuses")
	}
	for _, data := range loanStatuses {
		if err := m.LoanStatusManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan status %s", data.Name)
		}
	}

	// Delete MemberProfile
	memberProfiles, err := m.MemberProfileManager.Find(ctx, &MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member profiles")
	}
	for _, data := range memberProfiles {
		if err := m.MemberProfileManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member profile %s %s", data.FirstName, data.LastName)
		}
	}

	// Delete Company
	companies, err := m.CompanyManager.Find(ctx, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get companies")
	}
	for _, data := range companies {
		if err := m.CompanyManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy company %s", data.Name)
		}
	}

	// Delete MemberDepartment
	memberDepartments, err := m.MemberDepartmentManager.Find(ctx, &MemberDepartment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member departments")
	}
	for _, data := range memberDepartments {
		if err := m.MemberDepartmentManager.DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member department %s", data.Name)
		}
	}

	return nil
}
