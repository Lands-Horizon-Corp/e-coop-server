package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
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

func (e *Event) ProcessAllLoans(processContext context.Context, userOrg *core.UserOrganization) error {
	if userOrg == nil {
		return eris.New("user organization is nil")
	}

	currentTime := time.Now().UTC()

	loanTransaction, err := e.core.LoanTransactionManager.FindIncludingDeleted(processContext, &core.LoanTransaction{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to get loan transactions for processing")
	}

	for _, entry := range loanTransaction {
		entry.Processing = true
		if err := e.core.LoanTransactionManager.UpdateByID(processContext, entry.ID, entry); err != nil {
			return eris.Wrap(err, "failed to mark loan transaction as processing")
		}
	}

	go func() {
		timeoutContext, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		// Process iterations
		for i, entry := range loanTransaction {
			if entry.Processing {
				continue
			}
			time.Sleep(1 * time.Second)

			// ============================================================
			// Process Loan Logic Here
			// ============================================================
			LoanProcessing, err := e.LoanProcessing(processContext, userOrg, &entry.ID)
			if err != nil {
				return
			}
			entry = LoanProcessing
			if err := e.provider.Service.Broker.Dispatch(timeoutContext, []string{
				fmt.Sprintf("loan.process.branch.%s", userOrg.BranchID),
			}, LoanProcessingEventResponse{
				Total:       len(loanTransaction),
				Processed:   i + 1,
				StartTime:   currentTime,
				CurrentTime: time.Now().UTC(),
				AccountName: entry.Account.Name,
				MemberName:  entry.MemberProfile.FullName,
			}); err != nil {
				return
			}

			// Check for timeout
			select {
			case <-timeoutContext.Done():

				return
			default:
			}
		}

		for _, entry := range loanTransaction {
			entry.Processing = false
			if err := e.core.LoanTransactionManager.UpdateByID(processContext, entry.ID, entry); err != nil {
				return
			}
		}

	}()
	return nil
}
