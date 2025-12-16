package cable

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/wire"
)

type DatabaseResetter struct {
	Provider *server.Provider
	Core     *core.Core
}

func NewDatabaseResetter(provider *server.Provider, core *core.Core) *DatabaseResetter {
	return &DatabaseResetter{
		Provider: provider,
		Core:     core,
	}
}

func (r *DatabaseResetter) Reset(ctx context.Context) error {
	if err := r.Provider.Service.RunDatabase(ctx); err != nil {
		return err
	}
	if err := r.Core.Start(); err != nil {
		return err
	}
	if err := r.Provider.Service.RunStorage(ctx); err != nil {
		return err
	}
	if err := r.Provider.Service.Storage.RemoveAllFiles(ctx); err != nil {
		return err
	}
	if err := r.Provider.Service.Database.Client().Migrator().DropTable(r.Core.Migration...); err != nil {
		return err
	}
	if err := r.Provider.Service.Database.Client().AutoMigrate(r.Core.Migration...); err != nil {
		return err
	}
	return nil
}

func InitializeDatabaseResetter() (*DatabaseResetter, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		NewDatabaseResetter,
	)
	return nil, nil
}
