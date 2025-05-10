package publisher

import (
	"fmt"

	"horizon.com/server/server/model"
)

func (b *Publisher) GeneratedReportOnCreate(data *model.GeneratedReport) {
	go func() {
		b.broadcast.Dispatch([]string{
			"generated_report.create",
			fmt.Sprintf("generated_report.create.%s", data.ID),
			fmt.Sprintf("generated_report.create.user.%s", data.UserID),
		}, b.model.GeneratedReportModel(data))
	}()
}

func (b *Publisher) GeneratedReportOnUpdate(data *model.GeneratedReport) {
	go func() {
		b.broadcast.Dispatch([]string{
			"generated_report.update",
			fmt.Sprintf("generated_report.update.%s", data.ID),
			fmt.Sprintf("generated_report.update.user.%s", data.UserID),
		}, b.model.GeneratedReportModel(data))
	}()

}

func (b *Publisher) GeneratedReportOnDelete(data *model.GeneratedReport) {
	go func() {
		b.broadcast.Dispatch([]string{
			"generated_report.delete",
			fmt.Sprintf("generated_report.delete.%s", data.ID),
			fmt.Sprintf("generated_report.delete.user.%s", data.UserID),
		}, b.model.GeneratedReportModel(data))
	}()
}
