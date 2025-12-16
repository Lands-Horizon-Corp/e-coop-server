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
	"github.com/fatih/color"
	"github.com/google/wire"
)

type DatabaseSeeder struct {
	Provider *server.Provider
	Core     *core.Core
	Seeder   *seeder.Seeder
}

func NewDatabaseSeeder(provider *server.Provider, core *core.Core, seed *seeder.Seeder) *DatabaseSeeder {
	return &DatabaseSeeder{
		Provider: provider,
		Core:     core,
		Seeder:   seed,
	}
}

func (s *DatabaseSeeder) Seed(ctx context.Context, multiplier int32) error {
	if err := s.Provider.Service.RunDatabase(ctx); err != nil {
		return err
	}
	if err := s.Provider.Service.RunStorage(ctx); err != nil {
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

func InitializeDatabaseSeeder() (*DatabaseSeeder, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		seeder.NewSeeder,
		NewDatabaseSeeder,
	)
	return nil, nil
}

func SeedDatabase() {
	color.Blue("Seeding database...")

	seederInstance, err := InitializeDatabaseSeeder()
	if err != nil {
		log.Fatalf("Failed to initialize database seeder: %v", err)
	}
	timeout := 3 * time.Hour
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := seederInstance.Seed(ctx, 5); err != nil {
		log.Fatalf("Database seeding failed: %v", err)
	}
	color.Green("Database seeding completed successfully.")
}
