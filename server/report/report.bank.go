package report

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

func (r *Reports) bankReport(ctx context.Context, data ReportData) (result []byte, err error) {
	result, err = data.generated.PDF("/api/v1/bank/search", func(params ...string) ([]byte, error) {
		return r.core.BankManager().StringTabular(ctx, data.generated.FilterSearch, &core.Bank{
			OrganizationID: data.generated.OrganizationID,
			BranchID:       data.generated.BranchID,
		})
	})

	return result, err
}
