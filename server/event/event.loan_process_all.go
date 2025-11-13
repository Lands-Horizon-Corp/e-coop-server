package event

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

type LoanProcessingEventResponse struct {
	Total       int       `json:"total"`
	Processed   int       `json:"processed"`
	StartTime   time.Time `json:"start_time"`
	CurrentTime time.Time `json:"current_time"`
	AccountName string    `json:"account_name"`
	MemberName  string    `json:"member_name"`
}

func (e *Event) ProcessAllLoans(processContext context.Context, ctx echo.Context) error {
	user, err := e.userOrganizationToken.CurrentUserOrganization(processContext, ctx)
	if err != nil {
		return eris.Wrap(err, "failed to get current user organization")
	}

	currentTime := time.Now().UTC()

	// Initial footstep and notification
	e.Footstep(ctx, FootstepEvent{
		Activity:    "loan-processing-started",
		Description: "Loan processing started",
		Module:      "Loan Processing",
	})
	e.OrganizationAdminsNotification(ctx, NotificationEvent{
		Title:       "Loan Processing",
		Description: "Loan processing started",
	})

	total := 1_000

	go func() {
		timeoutContext, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		// Process iterations
		for i := range total {
			// Add delay between iterations
			time.Sleep(3 * time.Second)

			err := e.provider.Service.Broker.Dispatch(timeoutContext, []string{
				fmt.Sprintf("loan.process.branch.%s", user.BranchID),
			}, LoanProcessingEventResponse{
				Total:       total,
				Processed:   i + 1,
				StartTime:   currentTime,
				CurrentTime: time.Now().UTC(),
				AccountName: "Test Account",
				MemberName:  "Test Member",
			})

			// Handle dispatch errors
			if err != nil {
				e.Footstep(ctx, FootstepEvent{
					Activity:    "loan-processing-dispatch-error",
					Description: "Failed to dispatch processing progress :" + err.Error(),
					Module:      "Loan Processing",
				})
			}

			// Check for timeout
			select {
			case <-timeoutContext.Done():
				e.Footstep(ctx, FootstepEvent{
					Activity:    "loan-processing-timeout",
					Description: fmt.Sprintf("Loan processing timed out at iteration %d", i+1),
					Module:      "Loan Processing",
				})
				return
			default:
			}
		}

		// Completion footstep and notification
		e.Footstep(ctx, FootstepEvent{
			Activity:    "loan-processing-completed",
			Description: "Loan processing completed successfully",
			Module:      "Loan Processing",
		})
		e.OrganizationAdminsNotification(ctx, NotificationEvent{
			Title:       "Loan Processing",
			Description: "Loan processing completed successfully",
		})
	}()

	return nil
}
