package cable

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
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

func MigrateDatabase() {
	color.Blue("Running database migrations...")

	migrator, err := InitializeDatabaseMigrator()
	if err != nil {
		log.Fatalf("Failed to initialize database migrator: %v", err)
	}

	timeout := 30 * time.Minute
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := migrator.Migrate(ctx); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	color.Green("Database migration completed successfully.")
}
