package controller

import (
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
)

type Controller struct {
	// Services
	provider *src.Provider

	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken
	userToken             *cooperative_tokens.UserToken

	// Models
	media    *model.MediaCollection
	feedback *model.FeedbackCollection
}

func NewController(
	// Services
	provider *src.Provider,

	// Tokens
	transactionBatchToken *cooperative_tokens.TransactionBatchToken,
	userOrganizationToken *cooperative_tokens.UserOrganizatonToken,
	userToken *cooperative_tokens.UserToken,

	// Models
	media *model.MediaCollection,
	feedback *model.FeedbackCollection,

) (*Controller, error) {
	return &Controller{
		// Services
		provider: provider,

		// Tokens
		transactionBatchToken: transactionBatchToken,
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,

		// Models
		media:    media,
		feedback: feedback,
	}, nil
}

func (c *Controller) Routes() {
	c.MediaController()
	c.FeedbackController()
}
