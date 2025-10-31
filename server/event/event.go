package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

type Event struct {
	modelcore               *modelcore.ModelCore
	user_organization_token *tokens.UserOrganizationToken
	user_token              *tokens.UserToken
	provider                *server.Provider
	usecase                 *usecase.TransactionService
}

func NewEvent(
	modelcore *modelcore.ModelCore,
	user_organization_token *tokens.UserOrganizationToken,
	user_token *tokens.UserToken,
	provider *server.Provider,
	usecase *usecase.TransactionService,
) (*Event, error) {
	return &Event{
		user_organization_token: user_organization_token,
		user_token:              user_token,
		modelcore:               modelcore,
		provider:                provider,
		usecase:                 usecase,
	}, nil
}
