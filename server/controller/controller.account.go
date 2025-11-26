package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) accountController() {
	req := c.provider.Service.Request

	// GET: Search (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch. Only 'owner' and 'employee' roles are authorized. Returns paginated results.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/currency/:currency_id/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch and currency. Only 'owner' and 'employee' roles are authorized. Returns paginated results.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})
	// GET: /api/v1/account/deposit/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/deposit/search",
		Method:       "GET",
		Note:         "Retrieve all deposit accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/cash-and-cash-equivalence/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/currency/:currency_id/paid-up-shared-capital/search",
		Method:       "GET",
		Note:         "Retrieve all paid-up shared capital accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID: " + err.Error()})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			PaidUpShareCapital: true,
			CurrencyID:         currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash and cash equivalence accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/loan/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/loan/search",
		Method:       "GET",
		Note:         "Retrieve all loan accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeLoan,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		result := []*core.Account{}
		for _, acc := range accounts {
			cashAndCashEquivalence, _ := c.core.AccountManager.Find(context, &core.Account{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CashAndCashEquivalence: true,
				CurrencyID:             acc.CurrencyID,
			})
			if len(cashAndCashEquivalence) > 0 {
				result = append(result, acc)
			}
		}
		paginated, err := c.core.AccountManager.PaginationData(context, ctx, result)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paginated data: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

	// GET: /api/v1/account/cash-and-cash-equivalence/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/currency/:currency_id/loan/search",
		Method:       "GET",
		Note:         "Retrieve all loan accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID: " + err.Error()})
		}
		// If loan exist on settings
		// if branch settings has paid up shared capital account

		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     currencyID,
			Type:           core.AccountTypeLoan,
		})
		result := []*core.Account{}
		for _, acc := range accounts {
			cashAndCashEquivalence, _ := c.core.AccountManager.Find(context, &core.Account{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CashAndCashEquivalence: true,
				CurrencyID:             acc.CurrencyID,
			})
			if len(cashAndCashEquivalence) > 0 {
				result = append(result, acc)
			}
		}
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash and cash equivalence accounts: " + err.Error()})
		}
		paginated, err := c.core.AccountManager.PaginationData(context, ctx, result)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve paginated data: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paginated)
	})

	// GET: /api/v1/account/ar-ledger/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/ar-ledger/search",
		Method:       "GET",
		Note:         "Retrieve all A/R-Ledger accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeARLedger,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/ar-aging/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/ar-aging/search",
		Method:       "GET",
		Note:         "Retrieve all A/R-Aging accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeARAging,
		})

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/fines/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/fines/search",
		Method:       "GET",
		Note:         "Retrieve all fines accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeFines,
		})

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/interest/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/interest/search",
		Method:       "GET",
		Note:         "Retrieve all interest accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeInterest,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/svf-ledger/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/svf-ledger/search",
		Method:       "GET",
		Note:         "Retrieve all SVF-Ledger accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeSVFLedger,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/w-off/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/w-off/search",
		Method:       "GET",
		Note:         "Retrieve all W-Off accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeWOff,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/ap-ledger/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/ap-ledger/search",
		Method:       "GET",
		Note:         "Retrieve all A/P-Ledger accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeAPLedger,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/other/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/other/search",
		Method:       "GET",
		Note:         "Retrieve all other accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeOther,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: /api/v1/account/time-deposit/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/time-deposit/search",
		Method:       "GET",
		Note:         "Retrieve all time deposit accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}

		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Type:           core.AccountTypeTimeDeposit,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET: Search (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModels(accounts))
	})

	// POST: Create (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account",
		Method:       "POST",
		Note:         "Create a new account for the current branch.",
		ResponseType: core.AccountResponse{},
		RequestType:  core.AccountRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Account creation failed (/account), validation error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Account creation failed (/account), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for account (/account)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		account := &core.Account{
			CreatedAt:                             time.Now().UTC(),
			CreatedByID:                           userOrg.UserID,
			UpdatedAt:                             time.Now().UTC(),
			UpdatedByID:                           userOrg.UserID,
			BranchID:                              *userOrg.BranchID,
			OrganizationID:                        userOrg.OrganizationID,
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
			DefaultPaymentTypeID:                  req.DefaultPaymentTypeID,
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
			CohCibFinesGracePeriodEntrySemiAnnualAmortization:  req.CohCibFinesGracePeriodEntrySemiAnnualAmortization,
			CohCibFinesGracePeriodEntrySemiAnnualMaturity:      req.CohCibFinesGracePeriodEntrySemiAnnualMaturity,
			CohCibFinesGracePeriodEntryLumpsumAmortization:     req.CohCibFinesGracePeriodEntryLumpsumAmortization,
			CohCibFinesGracePeriodEntryLumpsumMaturity:         req.CohCibFinesGracePeriodEntryLumpsumMaturity,
			CohCibFinesGracePeriodEntryAnnualAmortization:      req.CohCibFinesGracePeriodEntryAnnualAmortization,
			CohCibFinesGracePeriodEntryAnnualMaturity:          req.CohCibFinesGracePeriodEntryAnnualMaturity,
			GeneralLedgerType:                   req.GeneralLedgerType,
			LoanAccountID:                       req.LoanAccountID,
			FinesGracePeriodAmortization:        req.FinesGracePeriodAmortization,
			AdditionalGracePeriod:               req.AdditionalGracePeriod,
			NoGracePeriodDaily:                  req.NoGracePeriodDaily,
			FinesGracePeriodMaturity:            req.FinesGracePeriodMaturity,
			YearlySubscriptionFee:               req.YearlySubscriptionFee,
			CutOffDays:                          req.CutOffDays,
			CutOffMonths:                        req.CutOffMonths,
			LumpsumComputationType:              req.LumpsumComputationType,
			InterestFinesComputationDiminishing: req.InterestFinesComputationDiminishing,
			InterestFinesComputationDiminishingStraightYearly: req.InterestFinesComputationDiminishingStraightYearly,
			EarnedUnearnedInterest:                            req.EarnedUnearnedInterest,
			LoanSavingType:                                    req.LoanSavingType,
			InterestDeduction:                                 req.InterestDeduction,
			OtherDeductionEntry:                               req.OtherDeductionEntry,
			InterestSavingTypeDiminishingStraight:             req.InterestSavingTypeDiminishingStraight,
			OtherInformationOfAnAccount:                       req.OtherInformationOfAnAccount,
			HeaderRow:                                         req.HeaderRow,
			CenterRow:                                         req.CenterRow,
			TotalRow:                                          req.TotalRow,
			GeneralLedgerGroupingExcludeAccount:               req.GeneralLedgerGroupingExcludeAccount,
			ShowInGeneralLedgerSourceWithdraw:                 req.ShowInGeneralLedgerSourceWithdraw,
			ShowInGeneralLedgerSourceDeposit:                  req.ShowInGeneralLedgerSourceDeposit,
			ShowInGeneralLedgerSourceJournal:                  req.ShowInGeneralLedgerSourceJournal,
			ShowInGeneralLedgerSourcePayment:                  req.ShowInGeneralLedgerSourcePayment,
			ShowInGeneralLedgerSourceAdjustment:               req.ShowInGeneralLedgerSourceAdjustment,
			ShowInGeneralLedgerSourceJournalVoucher:           req.ShowInGeneralLedgerSourceJournalVoucher,
			ShowInGeneralLedgerSourceCheckVoucher:             req.ShowInGeneralLedgerSourceCheckVoucher,
			CompassionFund:                                    req.CompassionFund,
			CompassionFundAmount:                              req.CompassionFundAmount,
			CashAndCashEquivalence:                            req.CashAndCashEquivalence,
			CurrencyID:                                        req.CurrencyID,

			Icon:                        req.Icon,
			InterestStandardComputation: req.InterestStandardComputation,
			InterestMaturity:            req.InterestMaturity,
			InterestAmortization:        req.InterestAmortization,
		}

		if err := c.core.AccountManager.Create(context, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Account creation failed (/account), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account: " + err.Error()})
		}

		// Create account history
		if err := c.core.CreateAccountHistory(context, nil, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-warning",
				Description: "Account created but history creation failed (/account): " + err.Error(),
				Module:      "Account",
			})
		}

		if len(req.AccountTags) > 0 {
			var tags []core.AccountTag
			for _, tagReq := range req.AccountTags {
				tags = append(tags, core.AccountTag{
					AccountID:      account.ID,
					Name:           tagReq.Name,
					Description:    tagReq.Description,
					Category:       tagReq.Category,
					Color:          tagReq.Color,
					Icon:           tagReq.Icon,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
				})
			}
			db := c.provider.Service.Database.Client()
			if err := db.Create(&tags).Error; err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Account tag creation failed (/account), db error: " + err.Error(),
					Module:      "Account",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create account tags: " + err.Error()})
			}
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created account (/account): " + account.Name,
			Module:      "Account",
		})
		return ctx.JSON(http.StatusCreated, account)
	})

	// GET: Get by ID (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id",
		Method:       "GET",
		Note:         "Retrieve a specific account by ID.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// PUT: Update (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id",
		Method:       "PUT",
		Note:         "Update an account by ID.",
		ResponseType: core.AccountResponse{},
		RequestType:  core.AccountRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account update failed (/account/:account_id), validation error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Account validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account update failed (/account/:account_id), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for account (/account/:account_id)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account update failed (/account/:account_id), invalid UUID: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account update failed (/account/:account_id), not found: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}

		account.UpdatedByID = userOrg.UserID
		account.UpdatedAt = userOrg.UserOrgTime()
		account.BranchID = *userOrg.BranchID
		account.OrganizationID = userOrg.OrganizationID

		account.GeneralLedgerDefinitionID = req.GeneralLedgerDefinitionID
		account.FinancialStatementDefinitionID = req.FinancialStatementDefinitionID
		account.AccountClassificationID = req.AccountClassificationID
		account.AccountCategoryID = req.AccountCategoryID
		account.MemberTypeID = req.MemberTypeID
		account.DefaultPaymentTypeID = req.DefaultPaymentTypeID

		account.Name = req.Name
		account.Description = req.Description
		account.MinAmount = req.MinAmount
		account.MaxAmount = req.MaxAmount
		account.Index = req.Index
		account.Type = req.Type
		account.IsInternal = req.IsInternal
		account.CashOnHand = req.CashOnHand
		account.PaidUpShareCapital = req.PaidUpShareCapital
		account.ComputationType = req.ComputationType
		account.FinesAmort = req.FinesAmort
		account.FinesMaturity = req.FinesMaturity
		account.InterestStandard = req.InterestStandard
		account.InterestSecured = req.InterestSecured
		account.ComputationSheetID = req.ComputationSheetID
		account.CohCibFinesGracePeriodEntryCashHand = req.CohCibFinesGracePeriodEntryCashHand
		account.CohCibFinesGracePeriodEntryCashInBank = req.CohCibFinesGracePeriodEntryCashInBank
		account.CohCibFinesGracePeriodEntryDailyAmortization = req.CohCibFinesGracePeriodEntryDailyAmortization
		account.CohCibFinesGracePeriodEntryDailyMaturity = req.CohCibFinesGracePeriodEntryDailyMaturity
		account.CohCibFinesGracePeriodEntryWeeklyAmortization = req.CohCibFinesGracePeriodEntryWeeklyAmortization
		account.CohCibFinesGracePeriodEntryWeeklyMaturity = req.CohCibFinesGracePeriodEntryWeeklyMaturity
		account.CohCibFinesGracePeriodEntryMonthlyAmortization = req.CohCibFinesGracePeriodEntryMonthlyAmortization
		account.CohCibFinesGracePeriodEntryMonthlyMaturity = req.CohCibFinesGracePeriodEntryMonthlyMaturity
		account.CohCibFinesGracePeriodEntrySemiMonthlyAmortization = req.CohCibFinesGracePeriodEntrySemiMonthlyAmortization
		account.CohCibFinesGracePeriodEntrySemiMonthlyMaturity = req.CohCibFinesGracePeriodEntrySemiMonthlyMaturity
		account.CohCibFinesGracePeriodEntryQuarterlyAmortization = req.CohCibFinesGracePeriodEntryQuarterlyAmortization
		account.CohCibFinesGracePeriodEntryQuarterlyMaturity = req.CohCibFinesGracePeriodEntryQuarterlyMaturity
		account.CohCibFinesGracePeriodEntrySemiAnnualAmortization = req.CohCibFinesGracePeriodEntrySemiAnnualAmortization
		account.CohCibFinesGracePeriodEntrySemiAnnualMaturity = req.CohCibFinesGracePeriodEntrySemiAnnualMaturity
		account.CohCibFinesGracePeriodEntryAnnualAmortization = req.CohCibFinesGracePeriodEntryAnnualAmortization
		account.CohCibFinesGracePeriodEntryAnnualMaturity = req.CohCibFinesGracePeriodEntryAnnualMaturity
		account.CohCibFinesGracePeriodEntryLumpsumAmortization = req.CohCibFinesGracePeriodEntryLumpsumAmortization
		account.CohCibFinesGracePeriodEntryLumpsumMaturity = req.CohCibFinesGracePeriodEntryLumpsumMaturity
		account.GeneralLedgerType = req.GeneralLedgerType
		account.LoanAccountID = req.LoanAccountID
		account.FinesGracePeriodAmortization = req.FinesGracePeriodAmortization
		account.AdditionalGracePeriod = req.AdditionalGracePeriod
		account.NoGracePeriodDaily = req.NoGracePeriodDaily
		account.FinesGracePeriodMaturity = req.FinesGracePeriodMaturity
		account.YearlySubscriptionFee = req.YearlySubscriptionFee
		account.CutOffDays = req.CutOffDays
		account.CutOffMonths = req.CutOffMonths
		account.LumpsumComputationType = req.LumpsumComputationType
		account.InterestFinesComputationDiminishing = req.InterestFinesComputationDiminishing
		account.InterestFinesComputationDiminishingStraightYearly = req.InterestFinesComputationDiminishingStraightYearly
		account.EarnedUnearnedInterest = req.EarnedUnearnedInterest
		account.LoanSavingType = req.LoanSavingType
		account.InterestDeduction = req.InterestDeduction
		account.OtherDeductionEntry = req.OtherDeductionEntry
		account.InterestSavingTypeDiminishingStraight = req.InterestSavingTypeDiminishingStraight
		account.OtherInformationOfAnAccount = req.OtherInformationOfAnAccount
		account.HeaderRow = req.HeaderRow
		account.CenterRow = req.CenterRow
		account.TotalRow = req.TotalRow
		account.GeneralLedgerGroupingExcludeAccount = req.GeneralLedgerGroupingExcludeAccount
		account.ShowInGeneralLedgerSourceWithdraw = req.ShowInGeneralLedgerSourceWithdraw
		account.ShowInGeneralLedgerSourceDeposit = req.ShowInGeneralLedgerSourceDeposit
		account.ShowInGeneralLedgerSourceJournal = req.ShowInGeneralLedgerSourceJournal
		account.ShowInGeneralLedgerSourcePayment = req.ShowInGeneralLedgerSourcePayment
		account.ShowInGeneralLedgerSourceAdjustment = req.ShowInGeneralLedgerSourceAdjustment
		account.ShowInGeneralLedgerSourceJournalVoucher = req.ShowInGeneralLedgerSourceJournalVoucher
		account.ShowInGeneralLedgerSourceCheckVoucher = req.ShowInGeneralLedgerSourceCheckVoucher
		account.CompassionFund = req.CompassionFund
		account.CompassionFundAmount = req.CompassionFundAmount
		account.Icon = req.Icon
		account.CashAndCashEquivalence = req.CashAndCashEquivalence
		account.InterestStandardComputation = req.InterestStandardComputation
		account.CurrencyID = req.CurrencyID
		account.InterestAmortization = req.InterestAmortization
		account.InterestMaturity = req.InterestMaturity

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before update failed (/account/:account_id): " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account update failed (/account/:account_id), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account: " + err.Error()})
		}

		if len(req.AccountTags) > 0 {
			for _, tagReq := range req.AccountTags {
				tag := &core.AccountTag{
					AccountID:      account.ID,
					Name:           tagReq.Name,
					Description:    tagReq.Description,
					Category:       tagReq.Category,
					Color:          tagReq.Color,
					Icon:           tagReq.Icon,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
				}
				if err := c.core.AccountTagManager.Create(context, tag); err != nil {
					c.event.Footstep(ctx, event.FootstepEvent{
						Activity:    "update-error",
						Description: "Account tag update failed (/account/:account_id), db error: " + err.Error(),
						Module:      "Account",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account tags: " + err.Error()})
				}
			}
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account (/account/:account_id): " + account.Name,
			Module:      "Account",
		})
		// Event notification
		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Description:      fmt.Sprintf("Account: the account has been updated - %s", account.Name),
			Title:            fmt.Sprintf("Updated account: %s", account.Name),
			NotificationType: core.NotificationSystem,
		})

		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// DELETE: Single (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/account/:account_id",
		Method: "DELETE",
		Note:   "Delete an account by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Account delete failed (/account/:account_id), invalid UUID: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}

		// Perform comprehensive delete validation
		if err := c.core.AccountDeleteCheck(context, *accountID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Account delete failed (/account/:account_id): " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Get account for logging purposes
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Account delete failed (/account/:account_id), not found: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}

		// Check user authorization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Account delete failed (/account/:account_id), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for account (/account/:account_id)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		// Perform deletion
		if err := c.core.AccountManager.Delete(context, account.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Account delete failed (/account/:account_id), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted account (/account/:account_id): " + account.Name,
			Module:      "Account",
		})

		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// DELETE: Bulk (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/account/bulk-delete",
		Method:      "DELETE",
		Note:        "Bulk delete multiple accounts by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete accounts (/account/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete accounts (/account/bulk-delete) | no IDs provided",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided."})
		}

		// Check user authorization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete accounts (/account/bulk-delete) | user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Unauthorized bulk delete attempt for accounts (/account/bulk-delete)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}

		// Validate each account before deletion
		var failedAccounts []string
		for _, accountID := range reqBody.IDs {
			if err := c.core.AccountDeleteCheck(context, accountID); err != nil {
				account, _ := c.core.AccountManager.GetByID(context, accountID)
				accountName := accountID.String()
				if account != nil {
					accountName = account.Name
				}
				failedAccounts = append(failedAccounts, fmt.Sprintf("%s: %s", accountName, err.Error()))
			}
		}

		if len(failedAccounts) > 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete accounts (/account/bulk-delete) | validation errors",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]any{
				"error":           "Some accounts cannot be deleted",
				"failed_accounts": failedAccounts,
			})
		}

		// Perform bulk deletion
		if err := c.core.AccountManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete accounts (/account/bulk-delete) | error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete accounts: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: fmt.Sprintf("Bulk deleted %d accounts (/account/bulk-delete)", len(reqBody.IDs)),
			Module:      "Account",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// PUT: Update index (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/index/:index",
		Method:       "PUT",
		Note:         "Update only the index field of an account using URL param.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account index update failed (/account/:account_id/index/:index), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized index update attempt for account (/account/:account_id/index/:index)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account index update failed (/account/:account_id/index/:index), invalid UUID: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}
		indexParam := ctx.Param("index")
		var newIndex int
		_, err = fmt.Sscanf(indexParam, "%d", &newIndex)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account index update failed (/account/:account_id/index/:index), invalid index: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index value: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account index update failed (/account/:account_id/index/:index), not found: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		account.Index = newIndex
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before index update failed (/account/:account_id/index/:index): " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account index update failed (/account/:account_id/index/:index), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account index: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated account index (/account/:account_id/index/:index): %s to %d", account.Name, newIndex),
			Module:      "Account",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// PUT: Remove GeneralLedgerDefinitionID (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/general-ledger-definition/remove",
		Method:       "PUT",
		Note:         "Remove the GeneralLedgerDefinitionID from an account.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove GL def failed (/account/:account_id/general-ledger-definition/remove), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized remove GL def attempt for account (/account/:account_id/general-ledger-definition/remove)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove GL def failed (/account/:account_id/general-ledger-definition/remove), invalid UUID: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove GL def failed (/account/:account_id/general-ledger-definition/remove), not found: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		account.GeneralLedgerDefinitionID = nil
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before GL def removal failed (/account/:account_id/general-ledger-definition/remove): " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove GL def failed (/account/:account_id/general-ledger-definition/remove), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to remove GeneralLedgerDefinitionID: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Removed GL def from account (/account/:account_id/general-ledger-definition/remove): %s", account.Name),
			Module:      "Account",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/financial-statement-definition/remove",
		Method:       "PUT",
		Note:         "Remove the GeneralLedgerDefinitionID from an account.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove FS def failed (/account/:account_id/financial-statement-definition/remove), user org error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized remove FS def attempt for account (/account/:account_id/financial-statement-definition/remove)",
				Module:      "Account",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove FS def failed (/account/:account_id/financial-statement-definition/remove), invalid UUID: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove FS def failed (/account/:account_id/financial-statement-definition/remove), not found: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		account.FinancialStatementDefinitionID = nil
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before FS def removal failed (/account/:account_id/financial-statement-definition/remove): " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account remove FS def failed (/account/:account_id/financial-statement-definition/remove), db error: " + err.Error(),
				Module:      "Account",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to remove FinancialStatementDefinitionID: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Removed FS def from account (/account/:account_id/financial-statement-definition/remove): %s", account.Name),
			Module:      "Account",
		})
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// Quick Search
	// GET: Search (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/withdraw/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                    userOrg.OrganizationID,
			BranchID:                          *userOrg.BranchID,
			ShowInGeneralLedgerSourceWithdraw: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/journal/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                   userOrg.OrganizationID,
			BranchID:                         *userOrg.BranchID,
			ShowInGeneralLedgerSourceJournal: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/payment/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                   userOrg.OrganizationID,
			BranchID:                         *userOrg.BranchID,
			ShowInGeneralLedgerSourcePayment: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/currency/:currency_id/payment/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch and currency. Only 'owner' and 'employee' roles are authorized. Returns paginated results.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owner and employee roles can view accounts."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                   userOrg.OrganizationID,
			BranchID:                         *userOrg.BranchID,
			CurrencyID:                       currencyID,
			ShowInGeneralLedgerSourcePayment: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Account retrieval failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/adjustment/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                      userOrg.OrganizationID,
			BranchID:                            *userOrg.BranchID,
			ShowInGeneralLedgerSourceAdjustment: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/journal-voucher/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                          userOrg.OrganizationID,
			BranchID:                                *userOrg.BranchID,
			ShowInGeneralLedgerSourceJournalVoucher: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/check-voucher/search",
		Method:       "GET",
		Note:         "Retrieve all accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:                        userOrg.OrganizationID,
			BranchID:                              *userOrg.BranchID,
			ShowInGeneralLedgerSourceCheckVoucher: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/cash-and-cash-equivalence/search",
		Method:       "GET",
		Note:         "Retrieve all cash and cash equivalence accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:         userOrg.OrganizationID,
			BranchID:               *userOrg.BranchID,
			CashAndCashEquivalence: true,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash and cash equivalence accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})
	//
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/currency/:currency_id/cash-and-cash-equivalence/search",
		Method:       "GET",
		Note:         "Retrieve all cash and cash equivalence accounts for the current branch.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized."})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID: " + err.Error()})
		}
		accounts, err := c.core.AccountManager.PaginationWithFields(context, ctx, &core.Account{
			OrganizationID:         userOrg.OrganizationID,
			BranchID:               *userOrg.BranchID,
			CashAndCashEquivalence: true,
			CurrencyID:             currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash and cash equivalence accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	// GET - api/v1/computation-sheet/:computation-sheet-id/accounts
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		Note:         "Returns all accounts connected to a computation sheet.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		computationSheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		accounts, err := c.core.AccountManager.FindRaw(context, &core.Account{
			ComputationSheetID: computationSheetID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No accounts found for the specified computation sheet"})
		}
		return ctx.JSON(http.StatusOK, accounts)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/computation-sheet/:computation_sheet_id/connect",
		Method:       "PUT",
		Note:         "Connect an account to a computation sheet.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		computationSheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		account.ComputationSheetID = computationSheetID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before computation sheet connection failed: " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account to computation sheet: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/computation-sheet/disconnect",
		Method:       "PUT",
		Note:         "Disconnect an account from a computation sheet.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		account.ComputationSheetID = nil
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before computation sheet disconnection failed: " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account to computation sheet: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// POST api/v1/account/:account_id/connect-to-loan/:account_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/connect-to-loan/:loan_id",
		Method:       "POST",
		Note:         "Connect an account to a loan.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		loanID, err := handlers.EngineUUIDParam(ctx, "loan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		loanAccount, err := c.core.AccountManager.GetByID(context, *loanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan not found"})
		}
		if loanAccount.Type != core.AccountTypeFines && loanAccount.Type != core.AccountTypeInterest && loanAccount.Type != core.AccountTypeSVFLedger {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "The specified loan account is not of a valid loan account type"})
		}
		loanAccount.LoanAccountID = &account.ID
		loanAccount.UpdatedAt = time.Now().UTC()
		loanAccount.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, loanAccount.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before loan connection failed: " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, loanAccount.ID, loanAccount); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account to loan: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(loanAccount))
	})

	// POST api/v1/account/:account_id/disconnect-account
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/disconnect-account",
		Method:       "POST",
		Note:         "Disconnect an account from a loan account.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		account.LoanAccountID = nil
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		// Create account history before update
		if err := c.core.CreateAccountHistoryBeforeUpdate(context, nil, account.ID, userOrg.UserID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-warning",
				Description: "Account history creation before loan account disconnection failed: " + err.Error(),
				Module:      "Account",
			})
		}

		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to disconnect account from loan account: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModel(account))
	})

	// GET /api/v1/account/loan-connectable-account-currency/:currency_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/loan-connectable-account-currency/:currency_id/search",
		Method:       "GET",
		Note:         "Retrieve all loan accounts for the current branch. Only Fines, Interest, SVF-Ledger",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID"})
		}

		accounts, err := c.core.FindAccountsByTypesAndBranch(
			context,
			userOrg.OrganizationID, *userOrg.BranchID, *currencyID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan accounts: " + err.Error()})
		}
		pagination, err := c.core.AccountManager.PaginationData(context, ctx, accounts)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to paginate loan accounts: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, pagination)
	})

	// GET api/v1/account/:account_id/loan-accounts
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/account/:account_id/loan-accounts",
		Method:       "GET",
		Note:         "Retrieve loan account connected to an account.",
		ResponseType: core.AccountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to fetch user organization: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		loanAccounts, err := c.core.FindLoanAccountsByID(context,
			userOrg.OrganizationID, *userOrg.BranchID, account.ID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Connected loan account not found"})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountManager.ToModels(loanAccounts))
	})
}
