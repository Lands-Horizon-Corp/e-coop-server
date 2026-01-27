package horizon

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// CommandConfig defines a Cobra command wrapper
type CommandConfig struct {
	Use     string
	Short   string
	RunFunc func(cmd *cobra.Command, args []string) error
}

// HorizonHandler now only takes context and service
type HorizonHandler func(ctx context.Context, service *HorizonService, cmd *cobra.Command, args []string) error

// HorizonRunnerParams defines the parameters for a Horizon service runner
type HorizonRunnerParams interface {
	Timeout() time.Duration
	OnStartMessage() string
	OnStopMessage() string
	Handler() HorizonHandler
	ForceLifetime() bool
	CommandUse() string
	CommandShort() string
}

// DefaultHorizonRunnerParams implements HorizonRunnerParams
type DefaultHorizonRunnerParams struct {
	TimeoutValue       time.Duration
	OnStartMessageText string
	OnStopMessageText  string
	HandlerFunc        HorizonHandler
	ForceLifetimeFunc  *bool
	CommandUseText     string
	CommandShortText   string
}

func (p DefaultHorizonRunnerParams) Timeout() time.Duration {
	if p.TimeoutValue > 0 {
		return p.TimeoutValue
	}
	return 30 * time.Minute
}

func (p DefaultHorizonRunnerParams) OnStartMessage() string {
	return p.OnStartMessageText
}

func (p DefaultHorizonRunnerParams) OnStopMessage() string {
	return p.OnStopMessageText
}

func (p DefaultHorizonRunnerParams) Handler() HorizonHandler {
	return p.HandlerFunc
}

func (p DefaultHorizonRunnerParams) ForceLifetime() bool {
	if p.ForceLifetimeFunc == nil {
		return false
	}
	return *p.ForceLifetimeFunc
}

func (p DefaultHorizonRunnerParams) CommandUse() string {
	return p.CommandUseText
}

func (p DefaultHorizonRunnerParams) CommandShort() string {
	return p.CommandShortText
}

func HorizonServiceRegister(params HorizonRunnerParams) CommandConfig {
	return CommandConfig{
		Use:   params.CommandUse(),
		Short: params.CommandShort(),
		RunFunc: func(cmd *cobra.Command, args []string) error {
			var (
				ctx  context.Context
				stop context.CancelFunc
			)
			if params.ForceLifetime() {
				ctx, stop = context.WithCancel(context.Background())
			} else {
				ctx, stop = context.WithTimeout(context.Background(), params.Timeout())
			}
			defer stop()
			service := NewHorizonService(params.ForceLifetime())
			if msg := params.OnStartMessage(); msg != "" {
				color.Blue(msg)
			}
			if err := service.Run(ctx); err != nil {
				return err
			}
			if handler := params.Handler(); handler != nil {
				if err := handler(ctx, service, cmd, args); err != nil {
					_ = service.Stop(context.Background())
					return err
				}
			}
			if params.ForceLifetime() {
				<-ctx.Done()
			}
			if err := service.Stop(context.Background()); err != nil {
				return err
			}
			if msg := params.OnStopMessage(); msg != "" {
				color.Yellow(msg)
			}
			return nil
		},
	}
}
func (c CommandConfig) ToCobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   c.Use,
		Short: c.Short,
		RunE:  c.RunFunc,
	}
}
