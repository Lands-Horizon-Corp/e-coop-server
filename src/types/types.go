package types

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
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

func Models() ([]any, []any) {
	first := []any{
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
		MemberProfile{},
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
		MemberAccountingLedger{},
		Area{},
	}
	second := []any{}
	return first, second
}

func AdminModels() ([]any, []any) {
	first := []any{
		License{},
		Admin{},
	}
	second := []any{}
	return first, second
}

func MigrateDB(db *gorm.DB, modelsFunc func() ([]any, []any), dbName string) error {
	parents, dependents := modelsFunc()
	if err := db.AutoMigrate(parents...); err != nil {
		return eris.Wrapf(err, "failed to migrate parent tables in %s", dbName)
	}
	if err := db.AutoMigrate(dependents...); err != nil {
		return eris.Wrapf(err, "failed to migrate dependent tables in %s", dbName)
	}
	return nil
}

func DropDB(db *gorm.DB, modelsFunc func() ([]any, []any), dbName string) error {
	parents, dependents := modelsFunc()
	if err := db.Migrator().DropTable(dependents...); err != nil {
		return eris.Wrapf(err, "failed to drop dependent tables in %s", dbName)
	}
	if err := db.Migrator().DropTable(parents...); err != nil {
		return eris.Wrapf(err, "failed to drop parent tables in %s", dbName)
	}
	return nil
}

func Migrate(service *horizon.HorizonService) error {
	if err := MigrateDB(service.Database.Client(), Models, "CoreDatabase"); err != nil {
		return err
	}
	if err := MigrateDB(service.AdminDatabase.Client(), AdminModels, "AdminDatabase"); err != nil {
		return err
	}
	return nil
}

func Drop(service *horizon.HorizonService) error {
	if err := DropDB(service.Database.Client(), Models, "CoreDatabase"); err != nil {
		return err
	}
	if err := DropDB(service.AdminDatabase.Client(), AdminModels, "AdminDatabase"); err != nil {
		return err
	}
	return nil
}
