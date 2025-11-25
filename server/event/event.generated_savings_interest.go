package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

func (e *Event) GenerateSavingsInterestPost(
	context context.Context,
	useOrg *core.UserOrganization,
	generatedSavingsInterestID uuid.UUID,
) error
