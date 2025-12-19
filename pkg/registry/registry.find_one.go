package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) FindOne(
	ctx context.Context,
	fields *TData,
	preloads ...string,
) (*TData, error) {

	fmt.Println("---- Registry.FindOne START ----")

	if r == nil {
		panic("Registry.FindOne: r is nil")
	}

	if r.database == nil {
		panic("Registry.FindOne: r.database is nil")
	}

	if fields == nil {
		panic("Registry.FindOne: fields is nil")
	}

	fmt.Printf("Registry: %+v\n", r)
	fmt.Printf("Fields type: %T\n", fields)
	fmt.Printf("Preloads: %+v\n", preloads)

	db := r.Client(ctx)
	if db == nil {
		panic("Registry.FindOne: db is nil after Client()")
	}

	fmt.Println("Calling NormalFindOne...")

	entity, err := r.pagination.
		NormalFindOne(db, fields, r.preload(preloads...)...)

	if err != nil {
		fmt.Println("NormalFindOne error:", err)
		return nil, err
	}

	if entity == nil {
		fmt.Println("NormalFindOne returned nil entity (no error)")
		return nil, nil
	}

	fmt.Println("---- Registry.FindOne END ----")
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLock(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.NormalFindOneWithLock(r.Client(context), fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOne(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.ArrFindOne(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLock(
	context context.Context,
	tx *gorm.DB,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.ArrFindOneWithLock(tx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOne(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.StructuredFindOne(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLock(
	context context.Context,
	tx *gorm.DB,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(tx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.NormalFindOne(r.Client(context), fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.NormalFindOneWithLock(tx, fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.ArrFindOne(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.ArrFindOneWithLock(tx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.StructuredFindOne(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	filter query.StructuredFilter,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(tx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindOne(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) (*TData, error) {
	var db *gorm.DB
	if filter != nil {
		db = filter.Model(new(TData))
	} else {
		db = r.Client(context)
	}
	entity, err := r.pagination.RawFindOne(db, preloads...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindOneWithLock(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) (*TData, error) {
	var db *gorm.DB
	if filter != nil {
		db = filter.Model(new(TData))
	} else {
		db = r.Client(context)
	}
	entity, err := r.pagination.RawFindOneWithLock(db, preloads...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindOneRaw(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.RawFindOne(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindOneWithLockRaw(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.RawFindOneWithLock(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}
