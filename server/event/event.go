package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
)

type Event struct {
	modelcore               *modelcore.ModelCore
	user_organization_token *cooperative_tokens.UserOrganizationToken
	user_token              *cooperative_tokens.UserToken
	provider                *src.Provider
	service                 *service.TransactionService
}

func NewEvent(
	modelcore *modelcore.ModelCore,
	user_organization_token *cooperative_tokens.UserOrganizationToken,
	user_token *cooperative_tokens.UserToken,
	provider *src.Provider,
	service *service.TransactionService,
) (*Event, error) {
	return &Event{
		user_organization_token: user_organization_token,
		user_token:              user_token,
		modelcore:               modelcore,
		provider:                provider,
		service:                 service,
	}, nil
}
