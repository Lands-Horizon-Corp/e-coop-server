package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
)

type Controller struct {
	// Services
	provider *src.Provider
	model    *model.Model

	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken
	userToken             *cooperative_tokens.UserToken
}

func NewController(
	// Services
	provider *src.Provider,
	model *model.Model,

	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken,
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken,
	userToken *cooperative_tokens.UserToken,

) (*Controller, error) {
	return &Controller{
		// Services
		provider: provider,
		model:    model,

		// Tokens
		transactionBatchToken: transactionBatchToken,
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
	}, nil
}

func (c *Controller) Start() error {

	c.CategoryController()
	c.ContactController()
	c.FeedbackController()
	c.MediaController()
	c.QRCodeController()
	c.UserController()

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
