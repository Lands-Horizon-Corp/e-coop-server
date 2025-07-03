package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) AccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/account/search",
		Method:   "GET",
		Response: "IAccount[]",
		Note:     "List all accounts for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		account, err := c.model.AccountManager.Find(context, &model.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountManager.ToModels(account))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account",
		Method:   "POST",
		Response: "IAccount",
		Note:     "Create a new account",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		account := &model.Account{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,

			GeneralLedgerDefinitionID:             req.GeneralLedgerDefinitionID,
			FinancialStatementDefinitionID:        req.FinancialStatementDefinitionID,
			AccountClassificationID:               req.AccountClassificationID,
			AccountCategoryID:                     req.AccountCategoryID,
			MemberTypeID:                          req.MemberTypeID,
			Name:                                  req.Name,
			Description:                           req.Description,
			MinAmount:                             req.MinAmount,
			MaxAmount:                             req.MaxAmount,
			Index:                                 req.Index,
			Type:                                  req.Type,
			IsInternal:                            req.IsInternal,
			CashOnHand:                            req.CashOnHand,
			PaidUpShareCapital:                    req.PaidUpShareCapital,
			ComputationType:                       req.ComputationType,
			FinesAmort:                            req.FinesAmort,
			FinesMaturity:                         req.FinesMaturity,
			InterestStandard:                      req.InterestStandard,
			InterestSecured:                       req.InterestSecured,
			ComputationSheetID:                    req.ComputationSheetID,
			CohCibFinesGracePeriodEntryCashHand:   req.CohCibFinesGracePeriodEntryCashHand,
			CohCibFinesGracePeriodEntryCashInBank: req.CohCibFinesGracePeriodEntryCashInBank,
			CohCibFinesGracePeriodEntryDailyAmortization:       req.CohCibFinesGracePeriodEntryDailyAmortization,
			CohCibFinesGracePeriodEntryDailyMaturity:           req.CohCibFinesGracePeriodEntryDailyMaturity,
			CohCibFinesGracePeriodEntryWeeklyAmortization:      req.CohCibFinesGracePeriodEntryWeeklyAmortization,
			CohCibFinesGracePeriodEntryWeeklyMaturity:          req.CohCibFinesGracePeriodEntryWeeklyMaturity,
			CohCibFinesGracePeriodEntryMonthlyAmortization:     req.CohCibFinesGracePeriodEntryMonthlyAmortization,
			CohCibFinesGracePeriodEntryMonthlyMaturity:         req.CohCibFinesGracePeriodEntryMonthlyMaturity,
			CohCibFinesGracePeriodEntrySemiMonthlyAmortization: req.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
			CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     req.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
			CohCibFinesGracePeriodEntryQuarterlyAmortization:   req.CohCibFinesGracePeriodEntryQuarterlyAmortization,
			CohCibFinesGracePeriodEntryQuarterlyMaturity:       req.CohCibFinesGracePeriodEntryQuarterlyMaturity,
			CohCibFinesGracePeriodEntrySemiAnualAmortization:   req.CohCibFinesGracePeriodEntrySemiAnualAmortization,
			CohCibFinesGracePeriodEntrySemiAnualMaturity:       req.CohCibFinesGracePeriodEntrySemiAnualMaturity,
			CohCibFinesGracePeriodEntryLumpsumAmortization:     req.CohCibFinesGracePeriodEntryLumpsumAmortization,
			CohCibFinesGracePeriodEntryLumpsumMaturity:         req.CohCibFinesGracePeriodEntryLumpsumMaturity,
			FinancialStatementType:                             string(req.FinancialStatementType),
			GeneralLedgerType:                                  req.GeneralLedgerType,
			AlternativeCode:                                    req.AlternativeCode,
			FinesGracePeriodAmortization:                       req.FinesGracePeriodAmortization,
			AdditionalGracePeriod:                              req.AdditionalGracePeriod,
			NumberGracePeriodDaily:                             req.NumberGracePeriodDaily,
			FinesGracePeriodMaturity:                           req.FinesGracePeriodMaturity,
			YearlySubscriptionFee:                              req.YearlySubscriptionFee,
			LoanCutOffDays:                                     req.LoanCutOffDays,
			LumpsumComputationType:                             string(req.LumpsumComputationType),
			InterestFinesComputationDiminishing:                string(req.InterestFinesComputationDiminishing),
			InterestFinesComputationDiminishingStraightYearly:  string(req.InterestFinesComputationDiminishingStraightYearly),
			EarnedUnearnedInterest:                             string(req.EarnedUnearnedInterest),
			LoanSavingType:                                     string(req.LoanSavingType),
			InterestDeduction:                                  string(req.InterestDeduction),
			OtherDeductionEntry:                                string(req.OtherDeductionEntry),
			InterestSavingTypeDiminishingStraight:              string(req.InterestSavingTypeDiminishingStraight),
			OtherInformationOfAnAccount:                        string(req.OtherInformationOfAnAccount),
			HeaderRow:                                          req.HeaderRow,
			CenterRow:                                          req.CenterRow,
			TotalRow:                                           req.TotalRow,
			GeneralLedgerGroupingExcludeAccount:                req.GeneralLedgerGroupingExcludeAccount,
		}

		if err := c.model.AccountManager.Create(context, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, account)
	})
}
