package core

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
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
	}
)

func NewCore(provider *server.Provider) (*Core, error) {
	return &Core{
		provider: provider,
	}, nil
}

func (m *Core) Start() error {
	m.Migration = append(m.Migration, []any{
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
	}...)
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
	if err := m.FinancialStatementAccountsGroupingSeed(context, tx, userID, organizationID, branchID); err != nil {
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
	userOrg, err := m.UserOrganizationManager().FindOne(context, &UserOrganization{
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
	if err := m.UserOrganizationManager().UpdateByIDWithTx(context, tx, userOrg.ID, userOrg); err != nil {
		return err
	}
	if err := m.companySeed(context, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

func (m *Core) OrganizationDestroyer(ctx context.Context, tx *gorm.DB, organizationID uuid.UUID, branchID uuid.UUID) error {
	invitationCodes, err := m.InvitationCodeManager().Find(ctx, &InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get invitation codes")
	}
	banks, err := m.BankManager().Find(ctx, &Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get banks")
	}
	billAndCoins, err := m.BillAndCoinsManager().Find(ctx, &BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get bill and coins")
	}
	holidays, err := m.HolidayManager().Find(ctx, &Holiday{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get holidays")
	}
	for _, data := range holidays {
		if err := m.HolidayManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := m.BillAndCoinsManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bill or coin %s", data.Name)
		}
	}
	for _, data := range banks {
		if err := m.BankManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bank %s", data.Name)
		}
	}
	for _, data := range invitationCodes {
		if err := m.InvitationCodeManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy invitation code %s", data.Code)
		}
	}

	memberTypes, err := m.MemberTypeManager().Find(ctx, &MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member types")
	}
	for _, data := range memberTypes {
		if err := m.MemberTypeManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member type %s", data.Name)
		}
	}

	memberOccupations, err := m.MemberOccupationManager().Find(ctx, &MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member occupations")
	}
	for _, data := range memberOccupations {
		if err := m.MemberOccupationManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member occupation %s", data.Name)
		}
	}

	memberGroups, err := m.MemberGroupManager().Find(ctx, &MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member groups")
	}
	for _, data := range memberGroups {
		if err := m.MemberGroupManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member group %s", data.Name)
		}
	}

	memberGenders, err := m.MemberGenderManager().Find(ctx, &MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member genders")
	}
	for _, data := range memberGenders {
		if err := m.MemberGenderManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member gender %s", data.Name)
		}
	}

	memberCenters, err := m.MemberCenterManager().Find(ctx, &MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member centers")
	}
	for _, data := range memberCenters {
		if err := m.MemberCenterManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member center %s", data.Name)
		}
	}

	memberClassifications, err := m.MemberClassificationManager().Find(ctx, &MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member classifications")
	}
	for _, data := range memberClassifications {
		if err := m.MemberClassificationManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member classification %s", data.Name)
		}
	}

	generalLedgerDefinitions, err := m.GeneralLedgerDefinitionManager().Find(ctx, &GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger definitions")
	}
	for _, data := range generalLedgerDefinitions {
		if err := m.GeneralLedgerDefinitionManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger definition %s", data.Name)
		}
	}

	generalLedgerAccountsGroupings, err := m.GeneralLedgerAccountsGroupingManager().Find(ctx, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger accounts groupings")
	}
	for _, data := range generalLedgerAccountsGroupings {
		if err := m.GeneralLedgerAccountsGroupingManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger accounts grouping %s", data.Name)
		}
	}

	FinancialStatementAccountsGroupings, err := m.FinancialStatementAccountsGroupingManager().Find(ctx, &FinancialStatementAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get financial statement accounts groupings")
	}
	for _, data := range FinancialStatementAccountsGroupings {
		if err := m.FinancialStatementAccountsGroupingManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy financial statement accounts grouping %s", data.Name)
		}
	}
	paymentTypes, err := m.PaymentTypeManager().Find(ctx, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get payment types")
	}
	for _, data := range paymentTypes {
		if err := m.PaymentTypeManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy payment type %s", data.Name)
		}
	}
	disbursements, err := m.DisbursementManager().Find(ctx, &Disbursement{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get disbursements")
	}
	for _, data := range disbursements {
		if err := m.DisbursementManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy disbursement %s", data.Name)
		}
	}
	collaterals, err := m.CollateralManager().Find(ctx, &Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get collaterals")
	}
	for _, data := range collaterals {
		if err := m.CollateralManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy collateral %s", data.Name)
		}
	}

	accounts, err := m.AccountManager().Find(ctx, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get accounts")
	}
	for _, data := range accounts {
		if err := m.AccountManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account %s", data.Name)
		}
	}

	loanPurposes, err := m.LoanPurposeManager().Find(ctx, &LoanPurpose{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan purposes")
	}
	for _, data := range loanPurposes {
		if err := m.LoanPurposeManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan purpose %s", data.Description)
		}
	}

	accountCategories, err := m.AccountCategoryManager().Find(ctx, &AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get account categories")
	}
	for _, data := range accountCategories {
		if err := m.AccountCategoryManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account category %s", data.Name)
		}
	}

	tagTemplates, err := m.TagTemplateManager().Find(ctx, &TagTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get tag templates")
	}
	for _, data := range tagTemplates {
		if err := m.TagTemplateManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy tag template %s", data.Name)
		}
	}

	loanStatuses, err := m.LoanStatusManager().Find(ctx, &LoanStatus{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan statuses")
	}
	for _, data := range loanStatuses {
		if err := m.LoanStatusManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan status %s", data.Name)
		}
	}

	memberProfiles, err := m.MemberProfileManager().Find(ctx, &MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member profiles")
	}
	for _, data := range memberProfiles {
		if err := m.MemberProfileManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member profile %s %s", data.FirstName, data.LastName)
		}
	}

	companies, err := m.CompanyManager().Find(ctx, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get companies")
	}
	for _, data := range companies {
		if err := m.CompanyManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy company %s", data.Name)
		}
	}

	memberDepartments, err := m.MemberDepartmentManager().Find(ctx, &MemberDepartment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member departments")
	}
	for _, data := range memberDepartments {
		if err := m.MemberDepartmentManager().DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member department %s", data.Name)
		}
	}

	return nil
}
