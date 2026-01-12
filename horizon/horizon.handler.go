package horizon

import (
	"context"
	"time"

	"github.com/fatih/color"
)

type HorizonHandler func(ctx context.Context, service *HorizonService) error

type HorizonRunnerParams interface {
	Timeout() time.Duration
	OnStartMessage() string
	OnStopMessage() string
	Handler() HorizonHandler
	ForceLifetime() bool
}
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
	ctx, cancel := context.WithTimeout(
		context.Background(),
		params.Timeout(),
	)
	service := NewHorizonService(params.ForceLifetime())
	if msg := params.OnStartMessage(); msg != "" {
		color.Blue(msg)
	}
	defer func() {
		service.Stop(context.Background())
		cancel()
		if msg := params.OnStopMessage(); msg != "" {
			color.Yellow(msg+": %v", ctx.Err())
		}
	}()
	if err := service.Run(ctx); err != nil {
		return err
	}
	if params.Handler() == nil {
		return nil
	}
	return params.Handler()(ctx, service)
}
