package repository

import (
	"strings"

	"gorm.io/gorm"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type UserRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.UserBroadcast
}

func NewUserRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.UserBroadcast,
) (*UserRepository, error) {
	return &UserRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *UserRepository) List() ([]*collection.User, error) {
	var users []*collection.User
	if err := r.database.Client().Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list user")
	}
	return users, nil
}
func (r *UserRepository) Create(data *collection.User) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create user")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *UserRepository) Update(data *collection.User) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update user")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *UserRepository) Delete(data *collection.User) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete user")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*collection.User, error) {
	var user collection.User
	if err := r.database.Client().First(&user, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with id: %s", id)
	}
	return &user, nil
}

func (r *UserRepository) UpdateCreateTransaction(tx *gorm.DB, data *collection.User) error {
	var existing collection.User
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}

	return nil
}

func (r *UserRepository) UpdateCreate(data *collection.User) error {
	var existing collection.User
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}
	return nil
}

func (r *UserRepository) UpdateFields(id uuid.UUID, fields *collection.User) error {
	return r.database.Client().Model(&collection.User{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *UserRepository) FindByEmail(email string) (*collection.User, error) {
	var user collection.User
	if err := r.database.Client().
		Where("email = ?", email).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with email %q", email)
	}
	return &user, nil
}

// FindByContactNumber looks up a user by contact number.
func (r *UserRepository) FindByContactNumber(contact string) (*collection.User, error) {
	var user collection.User
	if err := r.database.Client().
		Where("contact_number = ?", contact).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with contact number %q", contact)
	}
	return &user, nil
}

// FindByUserName looks up a user by username.
func (r *UserRepository) FindByUserName(username string) (*collection.User, error) {
	var user collection.User
	if err := r.database.Client().
		Where("user_name = ?", username).
		First(&user).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with username %q", username)
	}
	return &user, nil
}
func (r *UserRepository) FindByIdentifier(identifier string) (*collection.User, error) {
	if strings.Contains(identifier, "@") {
		if u, err := r.FindByEmail(identifier); err == nil {
			return u, nil
		}
	}
	numeric := strings.Trim(identifier, "+-0123456789")
	if numeric == "" {
		if u, err := r.FindByContactNumber(identifier); err == nil {
			return u, nil
		}
	}
	if u, err := r.FindByUserName(identifier); err == nil {
		return u, nil
	}
	return nil, eris.New("user not found by email, contact number, or username")
}
