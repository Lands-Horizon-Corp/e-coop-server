package cmd

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/account"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/admin"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/charges"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/funds"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/journal"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/loan"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/member_profile"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/organization"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/reports"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/settings"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/time_deposit"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/transactions"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/user"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1/voucher"
	core_admin "github.com/Lands-Horizon-Corp/e-coop-server/src/db/admin"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func migrateDatabase() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       30 * time.Minute,
		OnStartMessageText: "🔄 Database migration in progress...",
		OnStopMessageText:  "✅ Database migration completed successfully",
		CommandUseText:     "db-migrate",
		CommandShortText:   "Apply database schema migrations",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			return types.Migrate(service)
		},
	})
}

func seedDatabase() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       2 * time.Hour,
		OnStartMessageText: "🌱 Seeding database with initial data...",
		OnStopMessageText:  "✅ Database seeding completed successfully",
		CommandUseText:     "db-seed",
		CommandShortText:   "Seed database with initial configuration",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {

			if err := seeder.Seed(ctx, service); err != nil {
				return err
			}
			return core_admin.GlobalSeeder(ctx, service)
		},
	})
}

func resetDatabase() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       30 * time.Minute,
		OnStartMessageText: "⚠️  Resetting database - this will drop all tables...",
		OnStopMessageText:  "✅ Database reset completed successfully",
		CommandUseText:     "db-reset",
		CommandShortText:   "Reset database (drop all tables and recreate)",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			if err := service.Storage.RemoveAllFiles(ctx); err != nil {
				return err
			}
			if err := types.Drop(service); err != nil {
				return err
			}
			return types.Migrate(service)
		},
	})
}

func refreshDatabase() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       2 * time.Hour,
		OnStartMessageText: "🔄 Refreshing database - resetting and reseeding...",
		OnStopMessageText:  "✅ Database refresh completed successfully",
		CommandUseText:     "db-refresh",
		CommandShortText:   "Full database refresh (reset + seed)",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			if err := service.Cache.Flush(ctx); err != nil {
				return err
			}
			if err := service.Storage.RemoveAllFiles(ctx); err != nil {
				return err
			}
			if err := types.Drop(service); err != nil {
				return err
			}
			if err := types.Migrate(service); err != nil {
				return err
			}
			return seeder.Seed(ctx, service)
		},
	})
}

func cleanCache() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       30 * time.Minute,
		OnStartMessageText: "🧹 Clearing application cache...",
		OnStopMessageText:  "✅ Cache cleared successfully",
		CommandUseText:     "cache-clean",
		CommandShortText:   "Clear all cached data",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			return service.Cache.Flush(ctx)
		},
	})
}

func enforceBlocklist() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       2 * time.Hour,
		OnStartMessageText: "🛡️  Enforcing HaGeZi blocklist updates...",
		OnStopMessageText:  "✅ Blocklist enforcement completed",
		CommandUseText:     "security-enforce-blocklist",
		CommandShortText:   "Update and enforce HaGeZi security blocklist",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			return service.Security.Firewall(ctx, func(ip, host string) {
				cacheKey := "blocked_ip:" + ip
				timestamp := float64(time.Now().Unix())

				if err := service.Cache.ZAdd(ctx, "blocked_ips_registry", timestamp, ip); err != nil {
					color.Red("❌ Failed to register IP %s: %v", ip, err)
					return
				}

				if err := service.Cache.Set(ctx, cacheKey, host, 60*24*time.Hour); err != nil {
					color.Red("❌ Failed to cache IP %s: %v", ip, err)
					return
				}

				color.Green("✅ Blocked IP %s (from %s)", ip, host)
			})
		},
	})
}

func clearBlockedIPs() horizon.CommandConfig {
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		TimeoutValue:       30 * time.Minute,
		OnStartMessageText: "🧹 Removing blocked IP entries...",
		OnStopMessageText:  "✅ Blocked IPs cleared successfully",
		CommandUseText:     "security-clear-blocked",
		CommandShortText:   "Clear all blocked IP entries",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			keys, err := service.Cache.Keys(ctx, "blocked_ip:*")
			if err != nil {
				color.Red("❌ Failed to retrieve blocked IP keys: %v", err)
				return err
			}

			count := 0
			for _, key := range keys {
				if err := service.Cache.Delete(ctx, key); err != nil {
					color.Yellow("⚠️  Failed to delete key %s: %v", key, err)
				} else {
					count++
				}
			}

			color.Green("✅ Removed %d blocked IP entries", count)
			return nil
		},
	})
}

func startServer() horizon.CommandConfig {
	forceLifeTime := true
	return horizon.HorizonServiceRegister(horizon.DefaultHorizonRunnerParams{
		ForceLifetimeFunc:  &forceLifeTime,
		TimeoutValue:       5 * time.Minute,
		OnStartMessageText: "🚀 Starting E-Cooperative Server...",
		OnStopMessageText:  "✅ Server is now running",
		CommandUseText:     "server",
		CommandShortText:   "Start the main application server",
		HandlerFunc: func(ctx context.Context, service *horizon.HorizonService, _ *cobra.Command, _ []string) error {
			admin.AdminController(service)
			admin.LicenseKeyController(service)
			admin.CommonController(service)

			settings.Heartbeat(service)
			settings.BranchController(service)
			settings.CategoryController(service)
			settings.ContactController(service)
			settings.CurrencyController(service)
			settings.FeedbackController(service)
			settings.FootstepController(service)
			settings.MediaController(service)
			settings.NotificationController(service)
			settings.PermissionTemplateController(service)
			settings.QRCodeController(service)
			settings.SubscriptionPlanController(service)
			settings.TimesheetController(service)
			settings.TagTemplateController(service)
			settings.HolidayController(service)
			settings.BankController(service)
			settings.CompanyController(service)
			settings.AreaController(service)

			user.AuthenticationController(service)
			user.UserController(service)
			user.UserOrganizationController(service)
			user.UserRatingController(service)
			user.InvitationCodeController(service)
			user.KYCController(service)
			user.FeedController(service)
			user.FeedCommentController(service)

			member_profile.MemberGenderController(service)
			member_profile.MemberCenterController(service)
			member_profile.MemberTypeController(service)
			member_profile.MemberClassificationController(service)
			member_profile.MemberOccupationController(service)
			member_profile.MemberGroupController(service)
			member_profile.MemberProfileController(service)
			member_profile.MemberEducationalAttainmentController(service)
			member_profile.MemberAddressController(service)
			member_profile.MemberContactReferenceController(service)
			member_profile.MemberAssetController(service)
			member_profile.MemberIncomeController(service)
			member_profile.MemberExpenseController(service)
			member_profile.MemberGovernmentBenefitController(service)
			member_profile.MemberJointAccountController(service)
			member_profile.MemberRelativeAccountController(service)
			member_profile.MemberDepartmentController(service)
			member_profile.MemberProfileMediaController(service)
			member_profile.MemberProfileArchiveController(service)
			member_profile.MemberProfileComakerController(service)
			member_profile.MemberAccountingLedgerController(service)

			organization.OrganizationController(service)
			organization.OrganizationDailyUsageController(service)
			organization.OrganizationMediaController(service)

			voucher.CancelledCashCheckVoucherController(service)
			voucher.CashCheckVoucherController(service)
			voucher.CashCheckVoucherTagController(service)
			voucher.BillAndCoinsController(service)
			voucher.CashCountController(service)

			journal.JournalVoucherController(service)
			journal.JournalVoucherTagController(service)

			transactions.TransactionBatchController(service)
			transactions.CheckRemittanceController(service)
			transactions.TransactionController(service)
			transactions.OnlineRemittanceController(service)
			transactions.DisbursementController(service)
			transactions.DisbursementTransactionController(service)
			transactions.AdjustmentEntryController(service)
			transactions.AdjustmentTagController(service)
			transactions.PaymentTypeController(service)
			transactions.PaymentController(service)

			account.AccountController(service)
			account.AccountHistoryController(service)
			account.AccountCategoryController(service)
			account.AccountClassificationController(service)
			account.AccountTagController(service)
			account.AccountTransactionController(service)
			account.BrowseExcludeIncludeAccountsController(service)
			account.BrowseReferenceController(service)
			account.IncludeNegativeAccountController(service)

			reports.GeneratedReportsController(service)
			reports.GeneralLedgerController(service)
			reports.GeneralLedgerGroupingController(service)
			reports.FinancialStatementController(service)
			reports.GeneratedSavingsInterestController(service)
			reports.GeneratedSavingsInterestEntryController(service)

			loan.LoanStatusController(service)
			loan.LoanPurposeController(service)
			loan.LoanTransactionController(service)
			loan.LoanTransactionEntryController(service)
			loan.CollateralController(service)
			loan.ComputationSheetController(service)
			loan.AutomaticLoanDeductionController(service)
			loan.LoanTagController(service)

			time_deposit.TimeDepositTypeController(service)
			time_deposit.TimeDepositComputationController(service)
			time_deposit.TimeDepositComputationPreMatureController(service)

			charges.ChargesRateSchemeController(service)
			charges.ChargesRateByRangeOrMinimumAmountController(service)
			charges.ChargesRateSchemeModeOfPaymentController(service)
			charges.ChargesRateByTermController(service)

			funds.FundsController(service)
			funds.BatchFundingController(service)
			funds.MutualFundsController(service)
			funds.MutualFundEntryController(service)

			return service.RunLifeTime(ctx)
		},
	})
}

func Register() []horizon.CommandConfig {
	return []horizon.CommandConfig{
		migrateDatabase(),
		seedDatabase(),
		resetDatabase(),
		refreshDatabase(),
		cleanCache(),
		enforceBlocklist(),
		clearBlockedIPs(),
		startServer(),
	}
}
