package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

func (e *Event) AccountTransactionProcess(
	context context.Context,
	userOrg core.UserOrganization,
	data core.AccountTransactionProcessGLRequest,
) error {
	return nil
}

func (e *Event) AccountTransactionLedgers(
	context context.Context,
	userOrg core.UserOrganization,
	year int,
	accountId *uuid.UUID,
) ([]*core.AccountTransactionLedgerResponse, error) {
	return nil, nil
}
