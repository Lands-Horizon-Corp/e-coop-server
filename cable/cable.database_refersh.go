package cable

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/fatih/color"
	"github.com/google/wire"
)

type DatabaseRefresher struct {
	Provider *server.Provider
	Core     *core.Core
	Seeder   *seeder.Seeder
	Reports  *report.Reports
}

func NewDatabaseRefresher(provider *server.Provider, core *core.Core, seeder *seeder.Seeder, reports *report.Reports) *DatabaseRefresher {
	return &DatabaseRefresher{
		Provider: provider,
		Core:     core,
		Seeder:   seeder,
		Reports:  reports,
	}
}

func (d *DatabaseRefresher) Refresh(ctx context.Context) error {
	if err := d.Provider.Service.RunDatabase(ctx); err != nil {
		return err
	}
	if err := d.Provider.Service.RunStorage(ctx); err != nil {
		return err
	}
	if err := d.Provider.Service.Storage.RemoveAllFiles(ctx); err != nil {
		return err
	}
	if err := d.Provider.Service.RunBroker(ctx); err != nil {
		return err
	}
	if err := d.Core.Start(); err != nil {
		return err
	}
	if err := d.Provider.Service.Database.Client().Migrator().DropTable(d.Core.Migration...); err != nil {
		return err
	}
	if err := d.Provider.Service.Database.Client().AutoMigrate(d.Core.Migration...); err != nil {
		return err
	}
	if err := d.Seeder.Run(ctx, 5); err != nil {
		return err
	}
	return nil
}

func InitializeDatabaseRefresher() (*DatabaseRefresher, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		seeder.NewSeeder,
		report.NewReports,
		NewDatabaseRefresher,
	)
	return nil, nil
}

func RefreshDatabase() {
	color.Blue("Refreshing database...")

	refresher, err := InitializeDatabaseRefresher()
	if err != nil {
		log.Fatalf("Failed to initialize database refresher: %v", err)
	}

	timeout := 3 * time.Hour
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := refresher.Refresh(ctx); err != nil {
		log.Fatalf("Database refresh failed: %v", err)
	}

	color.Green("Database reset completed successfully.")
}
