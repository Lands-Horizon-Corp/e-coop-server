package repository

import (
	"strings"

	"gorm.io/gorm"
	"horizon.com/server/server/model"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Repository) UserList() ([]*model.User, error) {
	var users []*model.User
	if err := r.database.Client().Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list user")
	}
	return users, nil
}
func (r *Repository) UserCreate(data *model.User) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create user")
	}
	r.publisher.UserOnCreate(data)
	return nil
}

func (r *Repository) UserUpdate(data *model.User) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update user")
	}
	r.publisher.UserOnUpdate(data)
	return nil
}

func (r *Repository) UserDelete(data *model.User) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete user")
	}
	r.publisher.UserOnDelete(data)
	return nil
}

func (r *Repository) UserGetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.database.Client().First(&user, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with id: %s", id)
	}
	return &user, nil
}

func (r *Repository) UserUpdateCreateTransaction(tx *gorm.DB, data *model.User) error {
	var existing model.User
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user in UpdateCreate")
		}
		r.publisher.UserOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user in UpdateCreate")
		}
		r.publisher.UserOnUpdate(data)
	}

	return nil
}

func (r *Repository) UserUpdateCreate(data *model.User) error {
	var existing model.User
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user in UpdateCreate")
		}
		r.publisher.UserOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user in UpdateCreate")
		}
		r.publisher.UserOnUpdate(data)
	}
	return nil
}

func (r *Repository) UserUpdateFields(id uuid.UUID, fields *model.User) error {
	return r.database.Client().Model(&model.User{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *Repository) UserFindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.database.Client().
		Where("email = ?", email).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with email %q", email)
	}
	return &user, nil
}

// FindByContactNumber looks up a user by contact number.
func (r *Repository) UserFindByContactNumber(contact string) (*model.User, error) {
	var user model.User
	if err := r.database.Client().
		Where("contact_number = ?", contact).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with contact number %q", contact)
	}
	return &user, nil
}

// FindByUserName looks up a user by username.
func (r *Repository) UserFindByUserName(username string) (*model.User, error) {
	var user model.User
	if err := r.database.Client().
		Where("user_name = ?", username).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with username %q", username)
	}
	return &user, nil
}
func (r *Repository) FindByIdentifier(identifier string) (*model.User, error) {
	if strings.Contains(identifier, "@") {
		if u, err := r.UserFindByEmail(identifier); err == nil {
			return u, nil
		}
	}
	numeric := strings.Trim(identifier, "+-0123456789")
	if numeric == "" {
		if u, err := r.UserFindByContactNumber(identifier); err == nil {
			return u, nil
		}
	}
	if u, err := r.UserFindByUserName(identifier); err == nil {
		return u, nil
	}
	return nil, eris.New("user not found by email, contact number, or username")
}
