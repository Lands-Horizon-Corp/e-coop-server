package event

import (
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
	"github.com/lands-horizon/horizon-server/src/service"
)

type Event struct {
	model                 *model.Model
	userOrganizationToken *cooperative_tokens.UserOrganizationToken
	userToken             *cooperative_tokens.UserToken
	provider              *src.Provider
	service               *service.TransactionService
}

func NewEvent(
	model *model.Model,
	userOrganizationToken *cooperative_tokens.UserOrganizationToken,
	userToken *cooperative_tokens.UserToken,
	provider *src.Provider,
	service *service.TransactionService,
) (*Event, error) {
	return &Event{
		userOrganizationToken: userOrganizationToken,
		userToken:             userToken,
		model:                 model,
		provider:              provider,
		service:               service,
	}, nil
}
