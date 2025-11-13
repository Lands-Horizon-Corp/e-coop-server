package event

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) ProcessAllLoans(processContext context.Context, ctx echo.Context) error {
	user, err := e.userOrganizationToken.CurrentUserOrganization(processContext, ctx)
	if err != nil {
		return eris.Wrap(err, "failed to get current user organization")
	}

	currentTime := time.Now().UTC()

	go func() {

		context, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		// Loop 100 times with delay
		for i := range 10_00 {

			// Add delay between iterations (e.g., 1 second)
			time.Sleep(1 * time.Second)

			e.provider.Service.Broker.Dispatch(context, []string{
				fmt.Sprintf("bank.update.branch.%s", user.BranchID),
			}, map[string]any{
				"total":     1_000,
				"processed": i,
				"start":     currentTime,
				"current":   time.Now().UTC(),
			})
			select {
			case <-context.Done():

				return
			default:
				// Continue processing
			}
		}

	}()

	return nil
}
