package registry

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Create(
	context context.Context,
	data *TData,
) error {
	db := r.Client(context)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	r.OnCreate(context, data)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateWithTx(
	ctx context.Context,
	tx *gorm.DB,
	data *TData,
) error {
	if tx == nil {
		return eris.New("tx is nil")
	}
	if tx.Error != nil {
		return fmt.Errorf("transaction already failed: %w", tx.Error)
	}
	if err := tx.WithContext(ctx).Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entity with transaction: %w", err)
	}
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateMany(
	context context.Context,
	data []*TData,
) error {
	db := r.Client(context)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entities: %w", err)
	}
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateManyWithTx(
	context context.Context,
	tx *gorm.DB,
	data []*TData,
) error {
	if err := tx.WithContext(context).Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entities with transaction: %w", err)
	}
	for _, entity := range data {
		r.OnCreate(context, entity)
	}
	return nil
}
