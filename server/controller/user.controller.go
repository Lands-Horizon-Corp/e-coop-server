package controller

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
)

type UserController struct {
	repository *repository.UserRepository
	collection *collection.UserCollection
	storage    *horizon.HorizonStorage
	broadcast  *horizon.HorizonBroadcast
}

func NewUserController(
	repository *repository.UserRepository,
	collection *collection.UserCollection,
	storage *horizon.HorizonStorage,
	broadcast *horizon.HorizonBroadcast,
) (*UserController, error) {
	return &UserController{
		repository: repository,
		collection: collection,
		storage:    storage,
		broadcast:  broadcast,
	}, nil
}

// api/v1/authentication/current
// api/v1/authentication/current/branch
// api/v1/authentication/current/org
// api/v1/authentication/current/user
// api/v1/authentication/login
// api/v1/authentication/logout
// api/v1/authentication/register
// api/v1/authentication/forgot-password
// api/v1/authentication/change-password
// api/v1/authentication/request-contact-number-verification
// api/v1/authentication/verify-contact-number-verification
// api/v1/authentication/request-email-verification
// api/v1/authentication/verify-email-verification

// api/v1/user/change-password
// api/v1/user/change-email
// api/v1/user/change-username
// api/v1/user/change-profile-picture
// api/v1/user/change-contact-number
// api/v1/user/security/verify-with-password
// api/v1/user/security/verify-with-contact-number
// api/v1/user/security/verify-with-email
