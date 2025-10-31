package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

type Event struct {
	modelcore             *modelcore.ModelCore
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	provider              *server.Provider
	usecase               *usecase.TransactionService
}

func NewEvent(
	modelcore *modelcore.ModelCore,
	userOrganizationToken *tokens.UserOrganizationToken,
	userToken *tokens.UserToken,
	provider *server.Provider,
	usecase *usecase.TransactionService,
) (*Event, error) {
	return &Event{
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		modelcore:             modelcore,
		provider:              provider,
		usecase:               usecase,
	}, nil
}
