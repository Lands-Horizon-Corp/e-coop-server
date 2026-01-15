package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"gorm.io/gorm"
)

func IncrementOfficialReceipt(
	context context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	generalLedger *core.GeneralLedger) error {
	switch generalLedger.Source {
	case core.GeneralLedgerSourceWithdraw:
	case core.GeneralLedgerSourceDeposit:
	case core.GeneralLedgerSourcePayment:
	case core.GeneralLedgerSourceCheckVoucher:
	case core.GeneralLedgerSourceJournalVoucher:
	case core.GeneralLedgerSourceAdjustment:
	case core.GeneralLedgerSourceLoan:
	}
	return nil
}
