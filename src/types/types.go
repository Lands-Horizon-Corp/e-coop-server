package types

import "github.com/google/uuid"

type (
	IDSRequest struct {
		IDs uuid.UUIDs `json:"ids"`
	}

	QRMemberProfile struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		MiddleName      string `json:"middle_name"`
		FullName        string `json:"full_name"`
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
)

func Models() []any {
	return []any{
		AccountCategory{},
		AccountClassification{},
		Account{},
		AccountHistory{},
		AccountTag{},
		AdjustmentEntry{},
		AdjustmentTag{},
		AutomaticLoanDeduction{},
		Bank{},
		BatchFunding{},
		BillAndCoins{},
		Branch{},
		BranchSetting{},
		BrowseExcludeIncludeAccounts{},
		BrowseReference{},
		CancelledCashCheckVoucher{},
		CashCheckVoucherEntry{},
		CashCheckVoucher{},
		CashCheckVoucherTag{},
		CashCount{},
		Category{},
		ChargesRateByRangeOrMinimumAmount{},
		ChargesRateByTerm{},
		ChargesRateSchemeAccount{},
		ChargesRateScheme{},
		ChargesRateSchemeModeOfPayment{},
		CheckRemittance{},
		Collateral{},
		CollectorsMemberAccountEntry{},
		ComakerCollateral{},
		ComakerMemberProfile{},
		Company{},
		ComputationSheet{},
		ContactUs{},
		Currency{},
		Disbursement{},
		DisbursementTransaction{},
		Feedback{},
		FinancialStatementAccountsGrouping{},
		FinancialStatementDefinition{},
		FinesMaturity{},
		Footstep{},
		Funds{},
		GeneralAccountGroupingNetSurplusNegative{},
		GeneralAccountGroupingNetSurplusPositive{},
		GeneralAccountingLedgerTag{},
		GeneralLedgerAccountsGrouping{},
		GeneralLedgerDefinition{},
		GeneralLedger{},
		GeneratedReport{},
		GeneratedReportsDownloadUsers{},
		GeneratedSavingsInterestEntry{},
		GeneratedSavingsInterest{},
		GroceryComputationSheet{},
		GroceryComputationSheetMonthly{},
		Holiday{},
		IncludeNegativeAccount{},
		InterestMaturity{},
		InterestRateByAmount{},
		InterestRateByDate{},
		InterestRateByTerm{},
		InterestRateByYear{},
		InterestRatePercentage{},
		InterestRateScheme{},
		InvitationCode{},
		JournalVoucherEntry{},
		JournalVoucher{},
		JournalVoucherTag{},
		LoanAccount{},
		LoanClearanceAnalysis{},
		LoanClearanceAnalysisInstitution{},
		LoanComakerMember{},
		LoanGuaranteedFund{},
		LoanGuaranteedFundPerMonth{},
		LoanPurpose{},
		LoanStatus{},
		LoanTag{},
		LoanTermsAndConditionAmountReceipt{},
		LoanTermsAndConditionSuggestedPayment{},
		LoanTransactionEntry{},
		LoanTransaction{},
		Media{},
		MemberAccountingLedger{},
		MemberAddress{},
		MemberAsset{},
		MemberBankCard{},
		MemberCenter{},
		MemberCenterHistory{},
		MemberClassification{},
		MemberClassificationHistory{},
		MemberClassificationInterestRate{},
		MemberCloseRemark{},
		MemberContactReference{},
		MemberDamayanExtensionEntry{},
		MemberDeductionEntry{},
		MemberDepartment{},
		MemberDepartmentHistory{},
		MemberEducationalAttainment{},
		MemberExpense{},
		MemberGender{},
		MemberGenderHistory{},
		MemberGovernmentBenefit{},
		MemberGroup{},
		MemberGroupHistory{},
		MemberIncome{},
		MemberJointAccount{},
		MemberMutualFundHistory{},
		MemberOccupation{},
		MemberOccupationHistory{},
		MemberOtherInformationEntry{},
		MemberProfileArchive{},
		MemberProfile{},
		MemberProfileMedia{},
		MemberRelativeAccount{},
		MemberType{},
		MemberTypeHistory{},
		MemberVerification{},
		MutualFundAdditionalMembers{},
		MutualFundEntry{},
		MutualFund{},
		MutualFundTable{},
		Notification{},
		OnlineRemittance{},
		OrganizationCategory{},
		OrganizationDailyUsage{},
		Organization{},
		OrganizationMedia{},
		PaymentType{},
		PermissionTemplate{},
		PostDatedCheck{},
		SubscriptionPlan{},
		TagTemplate{},
		TimeDepositComputation{},
		TimeDepositComputationPreMature{},
		TimeDepositType{},
		Timesheet{},
		TransactionBatch{},
		Transaction{},
		TransactionTag{},
		UnbalancedAccount{},
		User{},
		UserOrganization{},
		UserRating{},
		VoucherPayTo{},
		AccountTransaction{},
		AccountTransactionEntry{},
	}
}
