package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
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
		Processing:     false,
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

			time.Sleep(1 * time.Second)

			// ============================================================
			// Process Loan Logic Here
			// ============================================================
			LoanProcessing, err := e.LoanProcessing(processContext, userOrg, &entry.ID)
			if err != nil {
				e.provider.Service.Logger.Error("failed to process loan transaction",
					zap.Error(err),
					zap.String("loanTransactionID", entry.ID.String()),
					zap.String("organizationID", userOrg.OrganizationID.String()),
					zap.String("branchID", (*userOrg.BranchID).String()),
					zap.Int("iteration", i+1),
					zap.Int("total", len(loanTransaction)))
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
				AccountName: func() string {
					if entry.Account != nil {
						return entry.Account.Name
					}
					return ""
				}(),
				MemberName: func() string {
					if entry.MemberProfile != nil {
						return entry.MemberProfile.FullName
					}
					return ""
				}(),
			}); err != nil {
				return
			}
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
