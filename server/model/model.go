package model

import (
	"github.com/go-playground/validator"
	"horizon.com/server/horizon"
)

type Model struct {
	validator *validator.Validate
	storage   *horizon.HorizonStorage
	qr        *horizon.HorizonQR
}

func NewModel(
	storage *horizon.HorizonStorage,
	qr *horizon.HorizonQR,
) (*Model, error) {
	return &Model{
		validator: validator.New(),
		storage:   storage,
		qr:        qr,
	}, nil
}
