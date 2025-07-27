package event

import (
	"context"

	"github.com/labstack/echo/v4"
)

type PaymentEvent struct {
}

func (e *Event) Payment(ctx context.Context, echoCtx echo.Context, data PaymentEvent) {

}

func (e *Event) Withdraw(ctx context.Context, echoCtx echo.Context, data PaymentEvent) {

}

func (e *Event) Deposit(ctx context.Context, echoCtx echo.Context, data PaymentEvent) {

}
