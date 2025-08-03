package controller_v1

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) GeneralLedgerEntriesController() {

	req := c.provider.Service.Request

	// BRANCH GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/branch/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/branch/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/branch/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current branch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})
	// ME
	// GET /api/v1/general-ledger/current/search
	// GET /api/v1/general-ledger/current/check-entry/search
	// GET /api/v1/general-ledger/current/online-entry/search
	// GET /api/v1/general-ledger/current/cash-entry/search
	// GET /api/v1/general-ledger/current/payment-entry/search
	// GET /api/v1/general-ledger/current/withdraw-entry/search
	// GET /api/v1/general-ledger/current/deposit-entry/search
	// GET /api/v1/general-ledger/current/journal-entry/search
	// GET /api/v1/general-ledger/current/adjustment-entry/search
	// GET /api/v1/general-ledger/current/journal-voucher/search
	// GET /api/v1/general-ledger/current/check-voucher/search
	// ME GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/current/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/current/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/current/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries of the current user with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// EMPLOYEE
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/check-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/online-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/cash-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/payment-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/withdraw-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/deposit-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/journal-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/adjustment-entry/search
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/journal-voucher
	// GET /api/v1/general-ledger/user-organization/:user_organization_id/check-voucher
	// EMPLOYEE GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified employee (by user organization ID) with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/journal-voucher
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/journal-voucher",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/user-organization/:user_organization_id/check-voucher
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/user-organization/:user_organization_id/check-voucher",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified employee with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// MEMBER
	// MEMBER GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/search/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/search/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member profile with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// MEMBER ACCOUNT
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search
	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search
	// MEMBER ACCOUNT GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/member-profile/:member_profile_id/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified member account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// TRANSACTION BATCH
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/search/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search
	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search
	// TRANSACTION BATCH GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/search/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/search/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction-batch/:transaction_batch_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction batch with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// TRANSACTION
	// GET /api/v1/general-ledger/transaction/:transaction_id/search/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/check-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/online-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/cash-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/payment-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search
	// GET /api/v1/general-ledger/transaction/:transaction_id/check-voucher/search
	// TRANSACTION GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/transaction/:transaction_id/search/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/search/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/transaction/:transaction_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/transaction/:transaction_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified transaction with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// ACCOUNTS
	// GET /api/v1/general-ledger/account/:account_id/search/search
	// GET /api/v1/general-ledger/account/:account_id/check-entry/search
	// GET /api/v1/general-ledger/account/:account_id/online-entry/search
	// GET /api/v1/general-ledger/account/:account_id/cash-entry/search
	// GET /api/v1/general-ledger/account/:account_id/payment-entry/search
	// GET /api/v1/general-ledger/account/:account_id/withdraw-entry/search
	// GET /api/v1/general-ledger/account/:account_id/deposit-entry/search
	// GET /api/v1/general-ledger/account/:account_id/journal-entry/search
	// GET /api/v1/general-ledger/account/:account_id/adjustment-entry/search
	// GET /api/v1/general-ledger/account/:account_id/journal-voucher/search
	// GET /api/v1/general-ledger/account/:account_id/check-voucher/search
	// ACCOUNTS GENERAL LEDGER ROUTES

	// GET /api/v1/general-ledger/account/:account_id/search/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/search/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/check-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/online-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/online-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all online entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/cash-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/cash-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all cash entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/payment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/payment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all payment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/withdraw-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/withdraw-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all withdraw entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/deposit-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/deposit-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all deposit entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/journal-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/adjustment-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/adjustment-entry/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all adjustment entry general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/journal-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/journal-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all journal voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

	// GET /api/v1/general-ledger/account/:account_id/check-voucher/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger/account/:account_id/check-voucher/search",
		Method:       "GET",
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Returns all check voucher general ledger entries for the specified account with pagination.",
	}, func(ctx echo.Context) error {
		return nil
	})

}
