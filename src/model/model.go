package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/rotisserie/eris"
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
		FinancialStatementAccountsGroupingManager       horizon_services.Repository[FinancialStatementsrouping, FinancialStatementAccountsGroupingResponse, FinancialStatementAccountsGroupingRequest]
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
func (c *Model) Start(context context.Context) error {

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
	return nil
}

func (m *Model) OrganizationSeeder(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now()
	expiration := now.AddDate(0, 1, 0) // 1 month from now

	invitationCodes := []*InvitationCode{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       "employee",
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         5,
			CurrentUse:     0,
			Description:    "Invitation code for employees (max 5 uses)",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       "member",
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         1000,
			CurrentUse:     0,
			Description:    "Invitation code for members (max 1000 uses)",
		},
	}

	for _, data := range invitationCodes {
		if err := m.InvitationCodeManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed invitation code for %s", data.UserType)
		}
	}
	banks := []*Bank{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "BDO Unibank, Inc.",
			Description:    "The largest bank in the Philippines by assets, BDO offers a wide range of financial services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bank of the Philippine Islands (BPI)",
			Description:    "One of the oldest banks in Southeast Asia, BPI provides banking and financial solutions.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Metropolitan Bank & Trust Company (Metrobank)",
			Description:    "A major universal bank in the Philippines, known for its extensive branch network.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Land Bank of the Philippines (Landbank)",
			Description:    "A government-owned bank focused on serving farmers and fishermen.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Philippine National Bank (PNB)",
			Description:    "One of the country’s largest banks, offering a full range of banking services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "China Banking Corporation (Chinabank)",
			Description:    "A leading private universal bank in the Philippines.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Security Bank Corporation",
			Description:    "A universal bank known for its innovative banking products.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Union Bank of the Philippines (UnionBank)",
			Description:    "A universal bank recognized for its digital banking leadership.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Development Bank of the Philippines (DBP)",
			Description:    "A government-owned development bank supporting infrastructure and social projects.",
		},
	}
	for _, data := range banks {
		if err := m.BankManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed bank %s", data.Name)
		}
	}
	billAndCoins := []*BillAndCoins{
		// Banknotes (New Generation Currency Series)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱1000 Bill", Value: 1000.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱500 Bill", Value: 500.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱200 Bill", Value: 200.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱100 Bill", Value: 100.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱50 Bill", Value: 50.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱20 Bill", Value: 20.00, CountryCode: "PHP"},

		// Coins (New Generation Currency Series)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱20 Coin", Value: 20.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱10 Coin", Value: 10.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱5 Coin", Value: 5.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱1 Coin", Value: 1.00, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Sentimo Coin", Value: 0.25, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Sentimo Coin", Value: 0.05, CountryCode: "PHP"},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Sentimo Coin", Value: 0.01, CountryCode: "PHP"},
	}
	year := now.Year()
	holidays := []*Holiday{
		// Regular Holidays
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 9, 0, 0, 0, 0, time.UTC), Name: "Araw ng Kagitingan", Description: "Day of Valor."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labor Day", Description: "Celebration of workers and laborers."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Commemorates Philippine independence from Spain."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 26, 0, 0, 0, 0, time.UTC), Name: "National Heroes Day", Description: "Honoring Philippine national heroes."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 30, 0, 0, 0, 0, time.UTC), Name: "Bonifacio Day", Description: "Commemorates the birth of Andres Bonifacio."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 30, 0, 0, 0, 0, time.UTC), Name: "Rizal Day", Description: "Commemorates the life of Dr. Jose Rizal."},

		// Special (Non-Working) Holidays
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 25, 0, 0, 0, 0, time.UTC), Name: "EDSA People Power Revolution", Description: "Commemorates the 1986 EDSA Revolution."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 21, 0, 0, 0, 0, time.UTC), Name: "Ninoy Aquino Day", Description: "Commemorates the assassination of Benigno Aquino Jr."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Honoring all the saints."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Feast of the Immaculate Conception", Description: "Catholic feast day."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC), Name: "New Year's Eve", Description: "Last day of the year."},

		// Religious Holidays (dates vary, set as placeholders)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian Holy Week."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian Holy Week."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 9, 0, 0, 0, 0, time.UTC), Name: "Black Saturday", Description: "Christian Holy Week."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian Holy Week."},

		// Islamic Holidays (dates vary each year, set as placeholders)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid'l Fitr", Description: "End of Ramadan (date varies)."},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid'l Adha", Description: "Feast of Sacrifice (date varies)."},
	}

	for _, data := range holidays {
		if err := m.HolidayManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := m.BillAndCoinsManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed bill or coin %s", data.Name)
		}
	}
	memberClassifications := []*MemberClassification{
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Gold",
			Icon:           "sunrise",
			Description:    "Gold membership is reserved for top-tier members with excellent credit scores and consistent loyalty.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Silver",
			Icon:           "moon-star",
			Description:    "Silver membership is designed for members with good credit history and regular engagement.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bronze",
			Icon:           "cloud",
			Description:    "Bronze membership is for new or casual members who are starting their journey with us.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Platinum",
			Icon:           "gem",
			Description:    "Platinum membership offers exclusive benefits to elite members with outstanding history and contributions.",
		},
	}
	for _, data := range memberClassifications {
		if err := m.MemberClassificationManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member classification %s", data.Name)
		}
	}

	memberCenter := []*MemberCenter{
		{
			Name:           "Main Wellness Center",
			Description:    "Provides health and wellness programs.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
		{

			Name:           "Training Hub",
			Description:    "Offers skill-building and training for members.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
		{

			Name:           "Community Support Center",
			Description:    "Focuses on community support services and events.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
	}
	for _, data := range memberCenter {
		if err := m.MemberCenterManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member center %s", data.Name)
		}
	}

	memberGenders := []*MemberGender{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Male",
			Description:    "Identifies as male.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Female",
			Description:    "Identifies as female.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Other",
			Description:    "Identifies outside the binary gender categories.",
		},
	}
	for _, data := range memberGenders {
		if err := m.MemberGenderManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member gender %s", data.Name)
		}
	}

	memberGroup := []*MemberGroup{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Single Moms",
			Description:    "Support group for single mothers in the community.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Athletes",
			Description:    "Members who actively participate in sports and fitness.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Tech",
			Description:    "Members involved in information technology or development.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Graphics Artists",
			Description:    "Creative members who specialize in digital and graphic design.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Accountants",
			Description:    "Finance-focused members responsible for budgeting and auditing.",
		},
	}
	for _, data := range memberGroup {
		if err := m.MemberGroupManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member group %s", data.Name)
		}
	}

	memberOccupations := []*MemberOccupation{
		{Name: "Farmer", Description: "Engaged in agriculture or crop cultivation.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Fisherfolk", Description: "Involved in fishing and aquaculture activities.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Agricultural Technician", Description: "Specializes in modern agricultural practices and tools.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Software Developer", Description: "Develops and maintains software systems.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "IT Specialist", Description: "Manages information technology infrastructure.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Accountant", Description: "Handles financial records and audits.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Teacher", Description: "Educates students in academic institutions.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Nurse", Description: "Provides healthcare and medical support.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Doctor", Description: "Licensed medical professional for diagnosing and treating patients.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Engineer", Description: "Designs and builds infrastructure or systems.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Construction Worker", Description: "Works on building and construction projects.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Driver", Description: "Professional vehicle operator (e.g., jeepney, tricycle, delivery).", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Vendor", Description: "Operates a small retail business or market stall.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Self-Employed", Description: "Independent worker managing their own business or services.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Housewife", Description: "Manages household responsibilities full-time.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Househusband", Description: "Male homemaker managing family and household duties.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Artist", Description: "Engaged in creative fields like painting, sculpture, or multimedia arts.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Graphic Designer", Description: "Creates visual content using design software.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Call Center Agent", Description: "Provides customer service through phone or chat support.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Unemployed", Description: "Currently without formal occupation.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Physicist", Description: "Studies the properties and interactions of matter and energy.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Pharmacist", Description: "Dispenses medications and advises on their safe use.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Chef", Description: "Creates recipes and prepares meals in restaurants or catering.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Mechanic", Description: "Repairs and maintains vehicles and machinery.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Electrician", Description: "Installs and repairs electrical systems and wiring.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Plumber", Description: "Installs and repairs piping systems for water and waste.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Architect", Description: "Designs buildings and ensures structural soundness.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Banker", Description: "Manages financial transactions and client relationships.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Lawyer", Description: "Provides legal advice and represents clients in court.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Journalist", Description: "Researches and reports news for print, online, or broadcast.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Social Worker", Description: "Supports individuals and families through counseling and services.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Caregiver", Description: "Provides in-home care and assistance to the elderly or disabled.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Security Guard", Description: "Protects property and enforces safety protocols.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Teacher’s Aide", Description: "Assists teachers in classroom management and lesson prep.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Student", Description: "Currently enrolled in an educational institution.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Retiree", Description: "Previously employed, now retired from active work.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Entrepreneur", Description: "Owns and operates one or more business ventures.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Musician", Description: "Performs, composes, or teaches music.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Writer", Description: "Crafts written content—books, articles, or scripts.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Pilot", Description: "Operates aircraft for commercial or private flights.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Scientist", Description: "Conducts research in natural or social sciences.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Lab Technician", Description: "Performs tests and experiments in scientific labs.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Receptionist", Description: "Manages front-desk operations and customer inquiries.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Janitor", Description: "Keeps buildings clean and well-maintained.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
	}
	for _, data := range memberOccupations {
		if err := m.MemberOccupationManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member ooccupation %s", data.Name)
		}
	}

	memberType := []*MemberType{
		{

			Name:           "New",
			Prefix:         "NEW",
			Description:    "Recently registered member, no activity yet.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Active",
			Prefix:         "ACT",
			Description:    "Regularly engaged member with no issues.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Loyal",
			Prefix:         "LOY",
			Description:    "Consistently active over a long period; high retention.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "VIP",
			Prefix:         "VIP",
			Description:    "Very high-value member with premium privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Reported",
			Prefix:         "RPT",
			Description:    "Flagged by community or system for review.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Suspended",
			Prefix:         "SUS",
			Description:    "Temporarily barred from activities pending resolution.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Banned",
			Prefix:         "BAN",
			Description:    "Permanently barred due to policy violations.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Closed",
			Prefix:         "CLS",
			Description:    "Account closed by user request or administrative action.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Alumni",
			Prefix:         "ALM",
			Description:    "Former member with notable contributions.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Pending",
			Prefix:         "PND",
			Description:    "Awaiting verification or approval.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Dormant",
			Prefix:         "DRM",
			Description:    "Inactive for a long period with no recent engagement.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Guest",
			Prefix:         "GST",
			Description:    "Limited access member without full privileges.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Moderator",
			Prefix:         "MOD",
			Description:    "Member with special privileges to manage content or users.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{

			Name:           "Admin",
			Prefix:         "ADM",
			Description:    "Administrator with full access and control.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
	}
	for _, data := range memberType {
		if err := m.MemberTypeManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member type %s", data.Name)
		}
	}
	return nil
}

func (m *Model) OrganizationDestroyer(ctx context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
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

	return nil
}
