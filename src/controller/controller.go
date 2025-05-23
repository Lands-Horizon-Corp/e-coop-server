package controller

import (
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
	c.MediaController()
	c.FeedbackController()
	return nil
}
