package cmd

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	v1 "github.com/Lands-Horizon-Corp/e-coop-server/server/controller/v1"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/wire"
)

func InitializeSeeder() (
	*server.Provider,
	*core.Core,
	*seeder.Seeder,
) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		seeder.NewSeeder,
	)
	return nil, nil, nil
}

func InitializeMigration() (
	*server.Provider,
	*core.Core,
) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
	)
	return nil, nil
}

func InitializeCache() *server.Provider {
	wire.Build(
		server.NewProvider,
	)
	return nil
}

func InitializeFirewall() (
	*server.Provider,
	*core.Core,
) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
	)
	return nil, nil
}

func InitializeServer() (
	*server.Provider,
	*core.Core,
	*v1.Controller,
) {
	wire.Build(
		server.NewProvider,
		core.NewCore,

		v1.NewController,
		event.NewEvent,
		report.NewReports,
		seeder.NewSeeder,

		tokens.NewUserToken,
		tokens.NewUserOrganizationToken,
		usecase.NewUsecaseService,
	)
	return nil, nil, nil
}
