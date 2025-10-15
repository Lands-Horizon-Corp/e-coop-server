package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
)

type Event struct {
	model_core              *model_core.ModelCore
	user_organization_token *cooperative_tokens.UserOrganizationToken
	user_token              *cooperative_tokens.UserToken
	provider                *src.Provider
	service                 *service.TransactionService
}

func NewEvent(
	model_core *model_core.ModelCore,
	user_organization_token *cooperative_tokens.UserOrganizationToken,
	user_token *cooperative_tokens.UserToken,
	provider *src.Provider,
	service *service.TransactionService,
) (*Event, error) {
	return &Event{
		user_organization_token: user_organization_token,
		user_token:              user_token,
		model_core:              model_core,
		provider:                provider,
		service:                 service,
	}, nil
}
