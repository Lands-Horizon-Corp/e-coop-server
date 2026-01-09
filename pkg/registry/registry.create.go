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
		fmt.Println("ERROR: transaction is nil")
		return eris.New("tx is nil")
	}

	// Check if transaction is already in an error state
	if tx.Error != nil {
		fmt.Printf("ERROR: transaction already has an error: %v\n", tx.Error)
		return fmt.Errorf("cannot create entity: transaction already failed: %w", tx.Error)
	}

	fmt.Printf("DEBUG: Creating entity with transaction: %+v\n", data)

	if err := tx.WithContext(ctx).Create(data).Error; err != nil {
		fmt.Printf("ERROR: failed to create entity: %v\n", err)
		return fmt.Errorf("failed to create entity with transaction: %w", err)
	}

	fmt.Printf("DEBUG: Successfully created entity: %+v\n", data)
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
