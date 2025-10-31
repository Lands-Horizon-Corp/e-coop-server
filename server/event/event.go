package event

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
)

// Event holds references required by event handlers.
//
// It wires model managers, tokens and services used by event handlers.
type Event struct {
	modelcore             *modelcore.ModelCore
	userOrganizationToken *tokens.UserOrganizationToken
	userToken             *tokens.UserToken
	provider              *server.Provider
	usecase               *usecase.TransactionService
}

// NewEvent constructs a new Event instance wiring domain services used
// by the package's event handlers.
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
