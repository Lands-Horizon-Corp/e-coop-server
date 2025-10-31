package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
)

type Event struct {
	modelCore               *modelCore.ModelCore
	user_organization_token *cooperative_tokens.UserOrganizationToken
	user_token              *cooperative_tokens.UserToken
	provider                *src.Provider
	service                 *service.TransactionService
}

func NewEvent(
	modelCore *modelCore.ModelCore,
	user_organization_token *cooperative_tokens.UserOrganizationToken,
	user_token *cooperative_tokens.UserToken,
	provider *src.Provider,
	service *service.TransactionService,
) (*Event, error) {
	return &Event{
		user_organization_token: user_organization_token,
		user_token:              user_token,
		modelCore:               modelCore,
		provider:                provider,
		service:                 service,
	}, nil
}
