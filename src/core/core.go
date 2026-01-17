package core

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

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
	if err := memberProfileSeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	if err := companySeed(context, service, tx, userID, organizationID, branchID); err != nil {
		return err
	}
	return nil
}

func OrganizationDestroyer(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, organizationID uuid.UUID, branchID uuid.UUID) error {
	invitationCodes, err := InvitationCodeManager(service).Find(ctx, &types.InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get invitation codes")
	}
	banks, err := BankManager(service).Find(ctx, &types.Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get banks")
	}
	billAndCoins, err := BillAndCoinsManager(service).Find(ctx, &types.BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
	if err != nil {
		return eris.Wrapf(err, "failed to get bill and coins")
	}
	holidays, err := HolidayManager(service).Find(ctx, &types.Holiday{
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

	memberTypes, err := MemberTypeManager(service).Find(ctx, &types.MemberType{
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

	memberOccupations, err := MemberOccupationManager(service).Find(ctx, &types.MemberOccupation{
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

	memberGroups, err := MemberGroupManager(service).Find(ctx, &types.MemberGroup{
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

	memberGenders, err := MemberGenderManager(service).Find(ctx, &types.MemberGender{
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

	memberCenters, err := MemberCenterManager(service).Find(ctx, &types.MemberCenter{
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

	memberClassifications, err := MemberClassificationManager(service).Find(ctx, &types.MemberClassification{
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

	generalLedgerDefinitions, err := GeneralLedgerDefinitionManager(service).Find(ctx, &types.GeneralLedgerDefinition{
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

	generalLedgerAccountsGroupings, err := GeneralLedgerAccountsGroupingManager(service).Find(ctx, &types.GeneralLedgerAccountsGrouping{
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

	FinancialStatementAccountsGroupings, err := FinancialStatementAccountsGroupingManager(service).Find(ctx, &types.FinancialStatementAccountsGrouping{
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
	paymentTypes, err := PaymentTypeManager(service).Find(ctx, &types.PaymentType{
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
	disbursements, err := DisbursementManager(service).Find(ctx, &types.Disbursement{
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
	collaterals, err := CollateralManager(service).Find(ctx, &types.Collateral{
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

	accounts, err := AccountManager(service).Find(ctx, &types.Account{
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

	loanPurposes, err := LoanPurposeManager(service).Find(ctx, &types.LoanPurpose{
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

	accountCategories, err := AccountCategoryManager(service).Find(ctx, &types.AccountCategory{
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

	tagTemplates, err := TagTemplateManager(service).Find(ctx, &types.TagTemplate{
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

	loanStatuses, err := LoanStatusManager(service).Find(ctx, &types.LoanStatus{
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

	memberProfiles, err := MemberProfileManager(service).Find(ctx, &types.MemberProfile{
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

	companies, err := CompanyManager(service).Find(ctx, &types.Company{
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

	memberDepartments, err := MemberDepartmentManager(service).Find(ctx, &types.MemberDepartment{
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
