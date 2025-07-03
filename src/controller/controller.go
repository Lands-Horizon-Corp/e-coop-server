package controller

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

type Controller struct {
	// Services
	provider *src.Provider
	model    *model.Model
	event    *event.Event
	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken
	userToken             *cooperative_tokens.UserToken
}

func NewController(
	// Services
	provider *src.Provider,
	model *model.Model,
	event *event.Event,

	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken,
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken,
	userToken *cooperative_tokens.UserToken,

) (*Controller, error) {
	return &Controller{
		// Services
		provider: provider,
		model:    model,
		event:    event,

		// Tokens
		transactionBatchToken: transactionBatchToken,
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
	}, nil
}

func (c *Controller) Start() error {

	// Basic Onboardding & Utilities
	c.BranchController()
	c.CategoryController()
	c.ContactController()
	c.FeedbackController()
	c.FootstepController()
	c.GeneratedReports()
	c.InvitationCode()
	c.MediaController()
	c.NotificationController()
	c.OrganizationController()
	c.OrganizationDailyUsage()
	c.PermissionTemplateController()
	c.QRCodeController()
	c.SubscriptionPlanController()
	c.TimesheetController()
	c.UserController()
	c.UserOrganinzationController()
	c.UserRatingController()

	// Member Profile
	c.MemberGenderController()
	c.MemberCenterController()
	c.MemberTypeController()
	c.MemberClassificationController()
	c.MemberOccupationController()
	c.MemberGroupController()
	c.MemberProfileController()

	// member profile properties
	c.MemberEducationalAttainmentController()
	c.MemberAddressController()
	c.MemberContactReferenceController()
	c.MemberAssetController()
	c.MemberIncomeController()
	c.MemberExpenseController()
	c.MemberGovernmentBenefitController()
	c.MemberJointAccountController()
	c.MemberRelativeAccountController()

	// Account Maintenance
	c.BankController()
	c.HolidayController()
	c.BillAndCoinsController()

	// Transaction batch
	c.TransactionBatchController()
	c.CashCountController()
	c.BatchFundingController()
	c.CheckRemittanceController()
	c.TransactionBatchEntriesController()
	c.OnlineRemittanceController()

	// Accounts
	c.AccountController()
	c.GeneralLedgerController()
	c.FinancialStatementController()
	return nil
}

// Error responses
func (c *Controller) ErrorResponse(ctx echo.Context, statusCode int, message string) error {
	return ctx.JSON(statusCode, map[string]any{
		"success": false,
		"error":   message,
	})
}

func (c *Controller) BadRequest(ctx echo.Context, message string) error {
	return c.ErrorResponse(ctx, http.StatusBadRequest, message)
}

func (c *Controller) NotFound(ctx echo.Context, resource string) error {
	return c.ErrorResponse(ctx, http.StatusNotFound, fmt.Sprintf("%s not found", resource))
}

func (c *Controller) InternalServerError(ctx echo.Context, err error) error {
	return c.ErrorResponse(ctx, http.StatusInternalServerError, "Internal server error")
}
func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
