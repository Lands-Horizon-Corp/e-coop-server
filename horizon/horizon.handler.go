package horizon

import (
	"context"
	"time"

	"github.com/fatih/color"
)

// HorizonHandler defines the function signature for a handler
type HorizonHandler func(ctx context.Context, service *HorizonService) error

// HorizonRunnerParams defines the interface for runner parameters
type HorizonRunnerParams interface {
	Timeout() time.Duration
	OnStartMessage() string
	OnStopMessage() string
	Handler() HorizonHandler
	ForceLifetime() bool
}

// DefaultHorizonRunnerParams implements HorizonRunnerParams
type DefaultHorizonRunnerParams struct {
	TimeoutValue       time.Duration
	OnStartMessageText string
	OnStopMessageText  string
	HandlerFunc        HorizonHandler
	ForceLifetimeFunc  *bool
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

func WithHorizonService(params HorizonRunnerParams) error {
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
		if err := handler(ctx, service); err != nil {
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
}
