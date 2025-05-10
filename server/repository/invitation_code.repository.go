package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
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
