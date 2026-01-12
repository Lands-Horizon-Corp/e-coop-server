package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func accountHistory(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(handlers.Route{
		Method:       "GET",
		Route:        "/api/v1/account-history/account/:account_id",
		ResponseType: core.AccountHistoryResponse{},
		Note:         "Get account history by account ID",
	},
		func(ctx echo.Context) error {
			context := ctx.Request().Context()
			accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account_id: " + err.Error()})
			}
			userOrg, err := c.event.CurrentUserOrganization(context, ctx)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
			}
			accountHistory, err := c.core.GetAllAccountHistory(
				context,
				*accountID,
				userOrg.OrganizationID,
				*userOrg.BranchID,
			)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account history: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, accountHistory)
		})

	req.RegisterWebRoute(handlers.Route{
		Method:       "GET",
		Route:        "/api/v1/account-history/:account_history_id",
		ResponseType: core.AccountHistory{},
		Note:         "Get account history by account history ID",
	},
		func(ctx echo.Context) error {
			context := ctx.Request().Context()
			accountHistoryID, err := handlers.EngineUUIDParam(ctx, "account_history_id")
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account_history_id: " + err.Error()})
			}
			accountHistory, err := c.core.AccountHistoryManager().GetByID(context, *accountHistoryID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account history: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, accountHistory)
		})

	req.RegisterWebRoute(handlers.Route{
		Method:       "POST",
		Route:        "/api/v1/account-history/:account_history_id/restore",
		ResponseType: core.AccountHistory{},
		Note:         "Restore account history by account ID",
	},
		func(ctx echo.Context) error {
			context := ctx.Request().Context()
			accountHistoryID, err := handlers.EngineUUIDParam(ctx, "account_history_id")
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account_history_id: " + err.Error()})
			}
			userOrg, err := c.event.CurrentUserOrganization(context, ctx)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
			}
			accountHistory, err := c.core.AccountHistoryManager().GetByID(context, *accountHistoryID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account history: " + err.Error()})
			}
			account, err := c.core.AccountManager().GetByID(context, accountHistory.AccountID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
			}

			account.UpdatedByID = userOrg.UserID
			account.UpdatedAt = time.Now().UTC()
			account.BranchID = *userOrg.BranchID
			account.OrganizationID = userOrg.OrganizationID

			account.GeneralLedgerDefinitionID = accountHistory.GeneralLedgerDefinitionID
			account.FinancialStatementDefinitionID = accountHistory.FinancialStatementDefinitionID
			account.AccountClassificationID = accountHistory.AccountClassificationID
			account.AccountCategoryID = accountHistory.AccountCategoryID
			account.MemberTypeID = accountHistory.MemberTypeID
			account.Name = accountHistory.Name
			account.Description = accountHistory.Description
			account.MinAmount = accountHistory.MinAmount
			account.MaxAmount = accountHistory.MaxAmount
			account.Index = accountHistory.Index
			account.Type = accountHistory.Type
			account.IsInternal = accountHistory.IsInternal
			account.CashOnHand = accountHistory.CashOnHand
			account.PaidUpShareCapital = accountHistory.PaidUpShareCapital
			account.ComputationType = accountHistory.ComputationType
			account.FinesAmort = accountHistory.FinesAmort
			account.FinesMaturity = accountHistory.FinesMaturity
			account.InterestStandard = accountHistory.InterestStandard
			account.InterestSecured = accountHistory.InterestSecured
			account.ComputationSheetID = accountHistory.ComputationSheetID
			account.CohCibFinesGracePeriodEntryCashHand = accountHistory.CohCibFinesGracePeriodEntryCashHand
			account.CohCibFinesGracePeriodEntryCashInBank = accountHistory.CohCibFinesGracePeriodEntryCashInBank
			account.CohCibFinesGracePeriodEntryDailyAmortization = accountHistory.CohCibFinesGracePeriodEntryDailyAmortization
			account.CohCibFinesGracePeriodEntryDailyMaturity = accountHistory.CohCibFinesGracePeriodEntryDailyMaturity
			account.CohCibFinesGracePeriodEntryWeeklyAmortization = accountHistory.CohCibFinesGracePeriodEntryWeeklyAmortization
			account.CohCibFinesGracePeriodEntryWeeklyMaturity = accountHistory.CohCibFinesGracePeriodEntryWeeklyMaturity
			account.CohCibFinesGracePeriodEntryMonthlyAmortization = accountHistory.CohCibFinesGracePeriodEntryMonthlyAmortization
			account.CohCibFinesGracePeriodEntryMonthlyMaturity = accountHistory.CohCibFinesGracePeriodEntryMonthlyMaturity
			account.CohCibFinesGracePeriodEntrySemiMonthlyAmortization = accountHistory.CohCibFinesGracePeriodEntrySemiMonthlyAmortization
			account.CohCibFinesGracePeriodEntrySemiMonthlyMaturity = accountHistory.CohCibFinesGracePeriodEntrySemiMonthlyMaturity
			account.CohCibFinesGracePeriodEntryQuarterlyAmortization = accountHistory.CohCibFinesGracePeriodEntryQuarterlyAmortization
			account.CohCibFinesGracePeriodEntryQuarterlyMaturity = accountHistory.CohCibFinesGracePeriodEntryQuarterlyMaturity
			account.CohCibFinesGracePeriodEntrySemiAnnualAmortization = accountHistory.CohCibFinesGracePeriodEntrySemiAnnualAmortization
			account.CohCibFinesGracePeriodEntrySemiAnnualMaturity = accountHistory.CohCibFinesGracePeriodEntrySemiAnnualMaturity
			account.CohCibFinesGracePeriodEntryLumpsumAmortization = accountHistory.CohCibFinesGracePeriodEntryLumpsumAmortization
			account.CohCibFinesGracePeriodEntryLumpsumMaturity = accountHistory.CohCibFinesGracePeriodEntryLumpsumMaturity
			account.GeneralLedgerType = accountHistory.GeneralLedgerType
			account.LoanAccountID = accountHistory.LoanAccountID
			account.FinesGracePeriodAmortization = accountHistory.FinesGracePeriodAmortization
			account.AdditionalGracePeriod = accountHistory.AdditionalGracePeriod
			account.NoGracePeriodDaily = accountHistory.NoGracePeriodDaily
			account.FinesGracePeriodMaturity = accountHistory.FinesGracePeriodMaturity
			account.YearlySubscriptionFee = accountHistory.YearlySubscriptionFee
			account.CutOffDays = accountHistory.CutOffDays
			account.CutOffMonths = accountHistory.CutOffMonths
			account.LumpsumComputationType = accountHistory.LumpsumComputationType
			account.InterestFinesComputationDiminishing = accountHistory.InterestFinesComputationDiminishing
			account.InterestFinesComputationDiminishingStraightYearly = accountHistory.InterestFinesComputationDiminishingStraightYearly
			account.EarnedUnearnedInterest = accountHistory.EarnedUnearnedInterest
			account.LoanSavingType = accountHistory.LoanSavingType
			account.InterestDeduction = accountHistory.InterestDeduction
			account.OtherDeductionEntry = accountHistory.OtherDeductionEntry
			account.InterestSavingTypeDiminishingStraight = accountHistory.InterestSavingTypeDiminishingStraight
			account.OtherInformationOfAnAccount = accountHistory.OtherInformationOfAnAccount
			account.HeaderRow = accountHistory.HeaderRow
			account.CenterRow = accountHistory.CenterRow
			account.TotalRow = accountHistory.TotalRow
			account.GeneralLedgerGroupingExcludeAccount = accountHistory.GeneralLedgerGroupingExcludeAccount
			account.ShowInGeneralLedgerSourceWithdraw = accountHistory.ShowInGeneralLedgerSourceWithdraw
			account.ShowInGeneralLedgerSourceDeposit = accountHistory.ShowInGeneralLedgerSourceDeposit
			account.ShowInGeneralLedgerSourceJournal = accountHistory.ShowInGeneralLedgerSourceJournal
			account.ShowInGeneralLedgerSourcePayment = accountHistory.ShowInGeneralLedgerSourcePayment
			account.ShowInGeneralLedgerSourceAdjustment = accountHistory.ShowInGeneralLedgerSourceAdjustment
			account.ShowInGeneralLedgerSourceJournalVoucher = accountHistory.ShowInGeneralLedgerSourceJournalVoucher
			account.ShowInGeneralLedgerSourceCheckVoucher = accountHistory.ShowInGeneralLedgerSourceCheckVoucher
			account.CompassionFund = accountHistory.CompassionFund
			account.CompassionFundAmount = accountHistory.CompassionFundAmount
			account.Icon = accountHistory.Icon
			account.CashAndCashEquivalence = accountHistory.CashAndCashEquivalence
			account.InterestStandardComputation = accountHistory.InterestStandardComputation
			account.CurrencyID = accountHistory.CurrencyID

			if err := c.core.AccountManager().UpdateByID(context, account.ID, account); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account: " + err.Error()})
			}

			return ctx.JSON(http.StatusOK, c.core.AccountManager().ToModel(account))
		})

}
