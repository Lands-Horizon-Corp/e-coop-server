package cable

import (
	"context"

	"github.com/google/wire"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

type DatabaseMigrator struct {
	Provider *server.Provider
	Core     *core.Core
}

func NewDatabaseMigrator(provider *server.Provider, core *core.Core) *DatabaseMigrator {
	return &DatabaseMigrator{
		Provider: provider,
		Core:     core,
	}
}

func (m *DatabaseMigrator) Migrate(ctx context.Context) error {
	if err := m.Provider.Service.RunDatabase(ctx); err != nil {
		return err
	}
	if err := m.Core.Start(); err != nil {
		return err
	}
	if err := m.Provider.Service.Database.Client().AutoMigrate(m.Core.Migration...); err != nil {
		return err
	}
	return nil
}
func InitializeDatabaseMigrator() (*DatabaseMigrator, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		NewDatabaseMigrator,
	)
	return nil, nil
}
