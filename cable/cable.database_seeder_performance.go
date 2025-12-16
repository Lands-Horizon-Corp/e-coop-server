package cable

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/wire"
)

type DatabasePerformanceSeeder struct {
	Provider *server.Provider
	Core     *core.Core
	Seeder   *seeder.Seeder
}

func NewDatabasePerformanceSeeder(provider *server.Provider, core *core.Core, seed *seeder.Seeder) *DatabasePerformanceSeeder {
	return &DatabasePerformanceSeeder{
		Provider: provider,
		Core:     core,
		Seeder:   seed,
	}
}

func (s *DatabasePerformanceSeeder) SeedPerformance(ctx context.Context, multiplier int32) error {
	if err := s.Provider.Service.RunDatabase(ctx); err != nil {
		return err
	}
	if err := s.Provider.Service.RunStorage(ctx); err != nil {
		return err
	}
	if err := s.Provider.Service.RunBroker(ctx); err != nil {
		return err
	}
	if err := s.Core.Start(); err != nil {
		return err
	}
	if err := s.Seeder.Run(ctx, multiplier); err != nil {
		return err
	}
	return nil
}

func InitializeDatabasePerformanceSeeder() (*DatabasePerformanceSeeder, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		seeder.NewSeeder,
		NewDatabasePerformanceSeeder,
	)
	return nil, nil
}
