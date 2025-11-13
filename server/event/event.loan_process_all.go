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
		// Create a new context with timeout to avoid shadowing the outer context
		timeoutContext, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		// Loop 1000 times (fixed the range syntax)
		for i := range 1000 {
			// Add delay between iterations (e.g., 1 second)
			time.Sleep(1 * time.Second)

			// Dispatch progress update
			err := e.provider.Service.Broker.Dispatch(timeoutContext, []string{
				fmt.Sprintf("loan.process.branch.%s", user.BranchID),
			}, map[string]any{
				"total":        1000,
				"processed":    i + 1, // Add 1 to show actual progress (1-1000 instead of 0-999)
				"start":        currentTime,
				"current":      time.Now().UTC(),
				"account_name": "sample name",
				"member_name":  "member name sample",
			})

			// Log dispatch errors if any
			if err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "loan-processing-dispatch-error",
					Description: "Failed to dispatch loan processing progress: " + err.Error(),
					Module:      "Loan Processing",
				})
			}

			// Check for context cancellation
			select {
			case <-timeoutContext.Done():
				e.Footstep(ctx, FootstepEvent{
					Activity:    "loan-processing-timeout",
					Description: fmt.Sprintf("Loan processing timed out at iteration %d", i+1),
					Module:      "Loan Processing",
				})
				return
			default:
				// Continue processing
			}
		}

		// Final completion message
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-processing-completed",
			Description: "Successfully completed processing all 1000 loan iterations",
			Module:      "Loan Processing",
		})
	}()

	return nil
}
