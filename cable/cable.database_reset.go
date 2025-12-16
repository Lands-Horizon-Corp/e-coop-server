package cable

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/fatih/color"
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

func ResetDatabase() {
	color.Blue("Resetting database...")

	resetter, err := InitializeDatabaseResetter()
	if err != nil {
		log.Fatalf("Failed to initialize database resetter: %v", err)
	}

	timeout := 4 * time.Hour
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := resetter.Reset(ctx); err != nil {
		log.Fatalf("Database reset failed: %v", err)
	}

	color.Green("Database reset completed successfully.")
}
