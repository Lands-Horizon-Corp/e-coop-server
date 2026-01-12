package core

import (
	"context"

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

func GlobalSeeder(ctx context.Context, service *horizon.HorizonService) error {
	if err := currencySeed(ctx, service); err != nil {
		return err
	}
	if err := categorySeed(ctx, service); err != nil {
		return err
	}
	if err := subscriptionPlanSeed(ctx, service); err != nil {
		return err
	}
	return nil
}

func OrganizationSeeder(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	if err := invitationCodeSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := bankSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := billAndCoinsSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := holidaySeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberClassificationSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberGenderSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberGroupSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberCenterSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberOccupationSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberTypeSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := memberDepartmentSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := generalLedgerAccountsGroupingSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := FinancialStatementAccountsGroupingSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := accountSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := loanPurposeSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return nil
	}
	if err := accountClassificationSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := accountCategorySeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := disbursementSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := collateralSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := tagTemplateSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := loanStatusSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := permissionTemplateSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg, err := UserOrganizationManager(service).FindOne(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
	if err != nil {
		return err
	}
	if err := memberProfileSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	userOrg.IsSeeded = true
	if err := UserOrganizationManager(service).UpdateByIDWithTx(context, tx, userOrg.ID, userOrg); err != nil {
		return err
	}
	if err := companySeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

func OrganizationDestroyer(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, organizationID uuid.UUID, branchID uuid.UUID) error {
	invitationCodes, err := InvitationCodeManager(service).Find(ctx, &InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get invitation codes")
	}
	banks, err := BankManager(service).Find(ctx, &Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get banks")
	}
	billAndCoins, err := BillAndCoinsManager(service).Find(ctx, &BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get bill and coins")
	}
	holidays, err := HolidayManager(service).Find(ctx, &Holiday{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get holidays")
	}
	for _, data := range holidays {
		if err := HolidayManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy holiday %s", data.Name)
		}
	}
	for _, data := range billAndCoins {
		if err := BillAndCoinsManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bill or coin %s", data.Name)
		}
	}
	for _, data := range banks {
		if err := BankManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy bank %s", data.Name)
		}
	}
	for _, data := range invitationCodes {
		if err := InvitationCodeManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy invitation code %s", data.Code)
		}
	}

	memberTypes, err := MemberTypeManager(service).Find(ctx, &MemberType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member types")
	}
	for _, data := range memberTypes {
		if err := MemberTypeManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member type %s", data.Name)
		}
	}

	memberOccupations, err := MemberOccupationManager(service).Find(ctx, &MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member occupations")
	}
	for _, data := range memberOccupations {
		if err := MemberOccupationManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member occupation %s", data.Name)
		}
	}

	memberGroups, err := MemberGroupManager(service).Find(ctx, &MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member groups")
	}
	for _, data := range memberGroups {
		if err := MemberGroupManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member group %s", data.Name)
		}
	}

	memberGenders, err := MemberGenderManager(service).Find(ctx, &MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member genders")
	}
	for _, data := range memberGenders {
		if err := MemberGenderManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member gender %s", data.Name)
		}
	}

	memberCenters, err := MemberCenterManager(service).Find(ctx, &MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member centers")
	}
	for _, data := range memberCenters {
		if err := MemberCenterManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member center %s", data.Name)
		}
	}

	memberClassifications, err := MemberClassificationManager(service).Find(ctx, &MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member classifications")
	}
	for _, data := range memberClassifications {
		if err := MemberClassificationManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member classification %s", data.Name)
		}
	}

	generalLedgerDefinitions, err := GeneralLedgerDefinitionManager(service).Find(ctx, &GeneralLedgerDefinition{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger definitions")
	}
	for _, data := range generalLedgerDefinitions {
		if err := GeneralLedgerDefinitionManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger definition %s", data.Name)
		}
	}

	generalLedgerAccountsGroupings, err := GeneralLedgerAccountsGroupingManager(service).Find(ctx, &GeneralLedgerAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get general ledger accounts groupings")
	}
	for _, data := range generalLedgerAccountsGroupings {
		if err := GeneralLedgerAccountsGroupingManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy general ledger accounts grouping %s", data.Name)
		}
	}

	FinancialStatementAccountsGroupings, err := FinancialStatementAccountsGroupingManager(service).Find(ctx, &FinancialStatementAccountsGrouping{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get financial statement accounts groupings")
	}
	for _, data := range FinancialStatementAccountsGroupings {
		if err := FinancialStatementAccountsGroupingManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy financial statement accounts grouping %s", data.Name)
		}
	}
	paymentTypes, err := PaymentTypeManager(service).Find(ctx, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get payment types")
	}
	for _, data := range paymentTypes {
		if err := PaymentTypeManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy payment type %s", data.Name)
		}
	}
	disbursements, err := DisbursementManager(service).Find(ctx, &Disbursement{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get disbursements")
	}
	for _, data := range disbursements {
		if err := DisbursementManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy disbursement %s", data.Name)
		}
	}
	collaterals, err := CollateralManager(service).Find(ctx, &Collateral{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get collaterals")
	}
	for _, data := range collaterals {
		if err := CollateralManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy collateral %s", data.Name)
		}
	}

	accounts, err := AccountManager(service).Find(ctx, &Account{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get accounts")
	}
	for _, data := range accounts {
		if err := AccountManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account %s", data.Name)
		}
	}

	loanPurposes, err := LoanPurposeManager(service).Find(ctx, &LoanPurpose{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan purposes")
	}
	for _, data := range loanPurposes {
		if err := LoanPurposeManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan purpose %s", data.Description)
		}
	}

	accountCategories, err := AccountCategoryManager(service).Find(ctx, &AccountCategory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get account categories")
	}
	for _, data := range accountCategories {
		if err := AccountCategoryManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy account category %s", data.Name)
		}
	}

	tagTemplates, err := TagTemplateManager(service).Find(ctx, &TagTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get tag templates")
	}
	for _, data := range tagTemplates {
		if err := TagTemplateManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy tag template %s", data.Name)
		}
	}

	loanStatuses, err := LoanStatusManager(service).Find(ctx, &LoanStatus{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get loan statuses")
	}
	for _, data := range loanStatuses {
		if err := LoanStatusManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy loan status %s", data.Name)
		}
	}

	memberProfiles, err := MemberProfileManager(service).Find(ctx, &MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member profiles")
	}
	for _, data := range memberProfiles {
		if err := MemberProfileManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member profile %s %s", data.FirstName, data.LastName)
		}
	}

	companies, err := CompanyManager(service).Find(ctx, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get companies")
	}
	for _, data := range companies {
		if err := CompanyManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy company %s", data.Name)
		}
	}

	memberDepartments, err := MemberDepartmentManager(service).Find(ctx, &MemberDepartment{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get member departments")
	}
	for _, data := range memberDepartments {
		if err := MemberDepartmentManager(service).DeleteWithTx(ctx, tx, data.ID); err != nil {
			return eris.Wrapf(err, "failed to destroy member department %s", data.Name)
		}
	}

	return nil
}
