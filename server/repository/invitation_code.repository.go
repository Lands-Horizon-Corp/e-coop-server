package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"horizon.com/server/server/model"
)

func (r *Repository) InvitationCodeList() ([]*model.InvitationCode, error) {
	var invitation_codes []*model.InvitationCode
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&invitation_codes).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list invitation_code")
	}
	return invitation_codes, nil
}

func (r *Repository) InvitationCodeCreate(data *model.InvitationCode) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create invitation_code")
	}
	r.publisher.InvitationCodeOnCreate(data)
	return nil
}

func (r *Repository) InvitationCodeUpdate(data *model.InvitationCode) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update invitation_code")
	}
	r.publisher.InvitationCodeOnUpdate(data)
	return nil
}

func (r *Repository) InvitationCodeDelete(data *model.InvitationCode) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete invitation_code")
	}
	r.publisher.InvitationCodeOnDelete(data)
	return nil
}

func (r *Repository) InvitationCodeGetByID(id uuid.UUID) (*model.InvitationCode, error) {
	var invitation_code model.InvitationCode
	if err := r.database.Client().First(&invitation_code, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find invitation_code with id: %s", id)
	}
	return &invitation_code, nil
}

func (r *Repository) InvitationCodeUpdateCreateTransaction(tx *gorm.DB, data *model.InvitationCode) error {
	var existing model.InvitationCode
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create invitation_code in UpdateCreate")
		}
		r.publisher.InvitationCodeOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update invitation_code in UpdateCreate")
		}
		r.publisher.InvitationCodeOnUpdate(data)
	}

	return nil
}

func (r *Repository) InvitationCodeUpdateCreate(data *model.InvitationCode) error {
	var existing model.InvitationCode
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create invitation_code in UpdateCreate")
		}
		r.publisher.InvitationCodeOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update invitation_code in UpdateCreate")
		}
		r.publisher.InvitationCodeOnUpdate(data)
	}
	return nil
}

func (r *Repository) InvitationCodeListByOrgBranch(
	orgID uuid.UUID,
	branchID uuid.UUID,
) ([]*model.InvitationCode, error) {
	var codes []*model.InvitationCode
	if err := r.database.Client().
		Where("organization_id = ? AND branch_id = ?", orgID, branchID).
		Order("created_at DESC").
		Find(&codes).Error; err != nil {
		return nil, eris.Wrapf(
			err,
			"failed to list invitation codes for org %s, branch %s",
			orgID, branchID,
		)
	}
	return codes, nil
}

func (r *Repository) InvitationCodeExists(
	orgID, branchID uuid.UUID,
	code string,
) (bool, error) {
	var count int64
	err := r.database.Client().
		Model(&model.InvitationCode{}).
		Where("organization_id = ? AND branch_id = ? AND code = ?", orgID, branchID, code).
		Count(&count).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check invitation code existence")
	}
	return count > 0, nil
}

func (r *Repository) InvitationCodeGetByCode(code string) (*model.InvitationCode, error) {
	var ic model.InvitationCode
	if err := r.database.Client().
		Where("code = ?", code).
		First(&ic).Error; err != nil {
		return nil, eris.Wrapf(err, "invitation code %s not found", code)
	}
	return &ic, nil
}

func (r *Repository) InvitationCodeRedeemTransaction(
	tx *gorm.DB,
	ic *model.InvitationCode,
) error {
	var existing model.InvitationCode
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&existing, "id = ?", ic.ID).Error; err != nil {
		return eris.Wrap(err, "failed to lock invitation code")
	}

	if existing.CurrentUse >= existing.MaxUse {
		return fmt.Errorf("invitation code has reached max uses")
	}
	if time.Now().UTC().After(existing.ExpirationDate) {
		return fmt.Errorf("invitation code has expired")
	}
	existing.CurrentUse++
	if err := tx.Save(&existing).Error; err != nil {
		return eris.Wrap(err, "failed to increment invitation code use")
	}
	r.publisher.InvitationCodeOnUpdate(&existing)
	*ic = existing
	return nil
}
