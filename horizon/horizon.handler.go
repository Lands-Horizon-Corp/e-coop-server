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
	var lifecycleCtx context.Context
	var stop context.CancelFunc
	if params.ForceLifetime() {
		lifecycleCtx, stop = context.WithCancel(context.Background())
	} else {
		lifecycleCtx, stop = context.WithTimeout(context.Background(), params.Timeout())
	}
	defer stop()
	service := NewHorizonService(params.ForceLifetime())
	if msg := params.OnStartMessage(); msg != "" {
		color.Blue(msg)
	}
	if err := service.Run(lifecycleCtx); err != nil {
		return err
	}
	if params.Handler() != nil {
		handlerCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := params.Handler()(handlerCtx, service); err != nil {
			return service.Stop(context.Background())
		}
	}
	if params.ForceLifetime() {
		<-lifecycleCtx.Done()
	}
	if err := service.Stop(context.Background()); err != nil {
		return nil
	}
	if msg := params.OnStopMessage(); msg != "" {
		color.Yellow(msg)
	}
	return nil
}
