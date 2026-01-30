package core_admin

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
)

func GlobalSeeder(ctx context.Context, service *horizon.HorizonService) error {
	if err := licenseSeed(ctx, service); err != nil {
		return err
	}
	return nil
}
