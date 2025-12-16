package cable

import (
	"context"

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

type MainServer struct {
	Controller            *v1.Controller
	Core                  *core.Core
	Provider              *server.Provider
	Event                 *event.Event
	Reports               *report.Reports
	Seeder                *seeder.Seeder
	UserToken             *tokens.UserToken
	UserOrganizationToken *tokens.UserOrganizationToken
	UsecaseService        *usecase.UsecaseService
}

func NewMainServer(
	ctrl *v1.Controller,
	core *core.Core,
	prov *server.Provider,
	ev *event.Event,
	reps *report.Reports,
	seed *seeder.Seeder,
	ut *tokens.UserToken,
	uot *tokens.UserOrganizationToken,
	ucs *usecase.UsecaseService,
) *MainServer {

	return &MainServer{
		Controller:            ctrl,
		Core:                  core,
		Provider:              prov,
		Event:                 ev,
		Reports:               reps,
		Seeder:                seed,
		UserToken:             ut,
		UserOrganizationToken: uot,
		UsecaseService:        ucs,
	}
}

func (s *MainServer) Start(ctx context.Context) error {
	if err := s.Controller.Start(); err != nil {
		return err
	}
	if err := s.Provider.Service.Run(ctx); err != nil {
		return err
	}
	if err := s.Core.Start(); err != nil {
		return err
	}
	return nil
}

func (s *MainServer) Stop(ctx context.Context) error {
	return s.Provider.Service.Stop(ctx)
}

func InitializeMainServer() (*MainServer, error) {
	wire.Build(
		v1.NewController,
		core.NewCore,
		server.NewProvider,
		event.NewEvent,
		report.NewReports,
		seeder.NewSeeder,
		tokens.NewUserToken,
		tokens.NewUserOrganizationToken,
		usecase.NewUsecaseService,
		NewMainServer,
	)
	return nil, nil
}
