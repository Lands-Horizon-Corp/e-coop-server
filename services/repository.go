package horizon_services

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

// Generic validation helper
func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}

// Model conversion helpers
func ToModel[T any, G any](data *T, mapFunc func(*T) *G) *G {
	if data == nil {
		return nil
	}
	return mapFunc(data)
}

func ToModels[T any, G any](data []*T, mapFunc func(*T) *G) []*G {
	if data == nil {
		return []*G{}
	}
	out := make([]*G, 0, len(data))
	for _, item := range data {
		if m := mapFunc(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}

// Filter operations
type FilterOp string

const (
	OpEq  FilterOp = "="
	OpGt  FilterOp = ">"
	OpGte FilterOp = ">="
	OpLt  FilterOp = "<"
	OpLte FilterOp = "<="
	OpNe  FilterOp = "<>"
	OpIn  FilterOp = "IN"
)

type Filter struct {
	Field string
	Op    FilterOp
	Value any
}

// Repository interface
type Repository[TData any, TResponse any, TRequest any] interface {
	Client() *gorm.DB
	Model() *TData
	Validate(ctx echo.Context) (*TRequest, error)
	ToModel(data *TData) *TResponse
	ToModels(data []*TData) []*TResponse

	Pagination(ctx context.Context, param echo.Context, data []*TData) horizon.PaginationResult[TResponse]
	Filtered(ctx context.Context, param echo.Context, data []*TData) []*TResponse
	// Filter operations
	FindWithFilters(ctx context.Context, filters []Filter, preloads ...string) ([]*TData, error)

	// Retrieval methods
	List(ctx context.Context, preloads ...string) ([]*TData, error)
	ListRaw(ctx context.Context, preloads ...string) ([]*TResponse, error)
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TData, error)
	GetByIDRaw(ctx context.Context, id uuid.UUID, preloads ...string) (*TResponse, error)
	Find(ctx context.Context, fields *TData, preloads ...string) ([]*TData, error)
	FindRaw(ctx context.Context, fields *TData, preloads ...string) ([]*TResponse, error)
	FindOne(ctx context.Context, fields *TData, preloads ...string) (*TData, error)
	FindOneRaw(ctx context.Context, fields *TData, preloads ...string) (*TResponse, error)
	FindWithConditions(ctx context.Context, conditions map[string]any, preloads ...string) ([]*TData, error)
	FindOneWithConditions(ctx context.Context, conditions map[string]any, preloads ...string) (*TData, error)

	// Aggregation
	Count(ctx context.Context, fields *TData) (int64, error)
	CountWithTx(ctx context.Context, tx *gorm.DB, fields *TData) (int64, error)

	// CRUD operations
	Create(ctx context.Context, entity *TData, preloads ...string) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error
	CreateMany(ctx context.Context, entities []*TData, preloads ...string) error
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error
	Update(ctx context.Context, entity *TData, preloads ...string) error
	UpdateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error
	UpdateByID(ctx context.Context, id uuid.UUID, entity *TData, preloads ...string) error
	UpdateByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, entity *TData, preloads ...string) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields *TData, preloads ...string) error
	UpdateFieldsWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, fields *TData, preloads ...string) error
	UpdateMany(ctx context.Context, entities []*TData, preloads ...string) error
	UpdateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error
	Upsert(ctx context.Context, entity *TData, preloads ...string) error
	UpsertWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error
	UpsertMany(ctx context.Context, entities []*TData, preloads ...string) error
	UpsertManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error
	Delete(ctx context.Context, entity *TData) error
	DeleteWithTx(ctx context.Context, tx *gorm.DB, entity *TData) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	DeleteMany(ctx context.Context, entities []*TData) error
	DeleteManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData) error
}

// Repository configuration
type RepositoryParams[TData any, TResponse any, TRequest any] struct {
	Service  *HorizonService
	Created  func(*TData) []string
	Updated  func(*TData) []string
	Deleted  func(*TData) []string
	Resource func(*TData) *TResponse
	Preloads []string
}

// CollectionManager implementation
type CollectionManager[TData any, TResponse any, TRequest any] struct {
	service  *HorizonService
	created  func(*TData) []string
	updated  func(*TData) []string
	deleted  func(*TData) []string
	resource func(*TData) *TResponse
	preloads []string
}

// NewRepository creates a new repository instance
func NewRepository[TData any, TResponse any, TRequest any](
	params RepositoryParams[TData, TResponse, TRequest],
) Repository[TData, TResponse, TRequest] {
	return &CollectionManager[TData, TResponse, TRequest]{
		service:  params.Service,
		created:  params.Created,
		updated:  params.Updated,
		deleted:  params.Deleted,
		resource: params.Resource,
		preloads: params.Preloads,
	}
}

// --- Core Methods ---
func (c *CollectionManager[TData, TResponse, TRequest]) Client() *gorm.DB {
	if c.service == nil || c.service.Database == nil {
		return nil
	}
	return c.service.Database.Client().Model(new(TData))
}

func (c *CollectionManager[TData, TResponse, TRequest]) Model() *TData {
	return new(TData)
}

func (c *CollectionManager[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

func (c *CollectionManager[TData, TResponse, TRequest]) ToModels(data []*TData) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := c.ToModel(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}

func (c *CollectionManager[TData, TResponse, TRequest]) Validate(ctx echo.Context) (*TRequest, error) {
	var req TRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := c.service.Validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}

// --- Pagination and Filtering ---
func (c *CollectionManager[TData, TResponse, TRequest]) Pagination(
	ctx context.Context,
	param echo.Context,
	data []*TData,
) horizon.PaginationResult[TResponse] {
	batchSize := 10_0000
	maxWorkers := runtime.NumCPU()
	filtered, err := horizon.Pagination(ctx, param, data, batchSize, maxWorkers)
	if err != nil {
		fmt.Println("Pagination error:", err)
		return horizon.PaginationResult[TResponse]{}
	}
	return horizon.PaginationResult[TResponse]{
		Data:      c.ToModels(filtered.Data),
		PageIndex: filtered.PageIndex,
		PageSize:  filtered.PageSize,
		TotalSize: filtered.TotalSize,
		Sort:      filtered.Sort,
		TotalPage: filtered.TotalPage,
	}
}

func (c *CollectionManager[TData, TResponse, TRequest]) Filtered(ctx context.Context, param echo.Context, data []*TData) []*TResponse {
	batchSize := 10_0000
	maxWorkers := runtime.NumCPU()
	filtered, err := horizon.FilterAndSortSlice(ctx, param, data, batchSize, maxWorkers)
	if err != nil {
		return c.ToModels(filtered)
	}
	return c.ToModels(filtered)
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindWithFilters(
	ctx context.Context,
	filters []Filter,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := c.service.Database.Client().Model(new(TData))

	for _, f := range filters {
		switch f.Op {
		case OpIn:
			db = db.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
		default:
			db = db.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entities with %d filters", len(filters))
	}
	return entities, nil
}

// --- Data Retrieval Methods ---
func (c *CollectionManager[TData, TResponse, TRequest]) List(
	ctx context.Context,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := c.service.Database.Client().Model(new(TData))

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list entities")
	}
	return entities, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) ListRaw(
	ctx context.Context,
	preloads ...string,
) ([]*TResponse, error) {
	entities, err := c.List(ctx, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModels(entities), nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) GetByID(
	ctx context.Context,
	id uuid.UUID,
	preloads ...string,
) (*TData, error) {
	var entity TData
	db := c.service.Database.Client().Model(new(TData))

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return nil, eris.Wrapf(err, "entity with ID %s not found", id)
		}
		return nil, eris.Wrapf(err, "failed to get entity by ID: %s", id)
	}
	return &entity, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) GetByIDRaw(
	ctx context.Context,
	id uuid.UUID,
	preloads ...string,
) (*TResponse, error) {
	entity, err := c.GetByID(ctx, id, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModel(entity), nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) Find(
	ctx context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := c.service.Database.Client().Model(fields).Where(fields)

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities by fields")
	}
	return entities, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindRaw(
	ctx context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	entities, err := c.Find(ctx, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModels(entities), nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindOne(
	ctx context.Context,
	fields *TData,
	preloads ...string,
) (*TData, error) {
	var entity TData
	db := c.service.Database.Client().Model(fields).Where(fields)

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").First(&entity).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return nil, eris.Wrap(err, "entity not found")
		}
		return nil, eris.Wrap(err, "failed to find entity by fields")
	}
	return &entity, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindOneRaw(
	ctx context.Context,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	entity, err := c.FindOne(ctx, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModel(entity), nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindWithConditions(
	ctx context.Context,
	conditions map[string]any,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := c.service.Database.Client().Model(new(TData))

	for field, value := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", field), value)
	}

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entities with %d conditions", len(conditions))
	}
	return entities, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) FindOneWithConditions(
	ctx context.Context,
	conditions map[string]any,
	preloads ...string,
) (*TData, error) {
	var entity TData
	db := c.service.Database.Client().Model(new(TData))

	for field, value := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", field), value)
	}

	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Order("updated_at DESC").First(&entity).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return nil, eris.Wrap(err, "entity not found with conditions")
		}
		return nil, eris.Wrapf(err, "failed to find entity with %d conditions", len(conditions))
	}
	return &entity, nil
}

// --- Aggregation Methods ---
func (c *CollectionManager[TData, TResponse, TRequest]) Count(
	ctx context.Context,
	fields *TData,
) (int64, error) {
	var count int64
	if err := c.service.Database.Client().Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities")
	}
	return count, nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) CountWithTx(
	ctx context.Context,
	tx *gorm.DB,
	fields *TData,
) (int64, error) {
	var count int64
	if err := tx.Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities in transaction")
	}
	return count, nil
}

// --- CRUD Operations ---
func (c *CollectionManager[TData, TResponse, TRequest]) Create(
	ctx context.Context,
	entity *TData,
	preloads ...string,
) error {
	if err := c.service.Database.Client().Create(entity).Error; err != nil {
		return eris.Wrap(err, "create operation failed")
	}

	if err := c.reloadWithPreloads(entity, preloads); err != nil {
		return eris.Wrap(err, "failed to reload after create")
	}

	c.CreatedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) CreateWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entity *TData,
	preloads ...string,
) error {
	if err := tx.Create(entity).Error; err != nil {
		return eris.Wrap(err, "create operation failed in transaction")
	}

	if err := c.reloadWithPreloadsTx(tx, entity, preloads); err != nil {
		return eris.Wrap(err, "failed to reload after create in transaction")
	}

	c.CreatedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) CreateMany(
	ctx context.Context,
	entities []*TData,
	preloads ...string,
) error {
	if err := c.service.Database.Client().Create(entities).Error; err != nil {
		return eris.Wrap(err, "batch create operation failed")
	}

	// Reload all entities with preloads
	return c.reloadManyWithPreloads(entities, preloads)
}

func (c *CollectionManager[TData, TResponse, TRequest]) CreateManyWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entities []*TData,
	preloads ...string,
) error {
	if err := tx.Create(entities).Error; err != nil {
		return eris.Wrap(err, "batch create operation failed in transaction")
	}

	// Reload all entities with preloads using transaction
	return c.reloadManyWithPreloadsTx(tx, entities, preloads)
}

func (c *CollectionManager[TData, TResponse, TRequest]) Update(
	ctx context.Context,
	entity *TData,
	preloads ...string,
) error {
	if err := c.service.Database.Client().Save(entity).Error; err != nil {
		return eris.Wrap(err, "update operation failed")
	}

	if err := c.reloadWithPreloads(entity, preloads); err != nil {
		return eris.Wrap(err, "failed to reload after update")
	}

	c.UpdatedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entity *TData,
	preloads ...string,
) error {
	if err := tx.Save(entity).Error; err != nil {
		return eris.Wrap(err, "update operation failed in transaction")
	}

	if err := c.reloadWithPreloadsTx(tx, entity, preloads); err != nil {
		return eris.Wrap(err, "failed to reload after update in transaction")
	}

	c.UpdatedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByID(
	ctx context.Context,
	id uuid.UUID,
	entity *TData,
	preloads ...string,
) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "invalid entity ID")
	}
	return c.Update(ctx, entity, preloads...)
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByIDWithTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	entity *TData,
	preloads ...string,
) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "invalid entity ID in transaction")
	}
	return c.UpdateWithTx(ctx, tx, entity, preloads...)
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFields(
	ctx context.Context,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {
	// Get field names using reflection
	t := reflect.TypeOf(new(TData)).Elem()
	fieldNames := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "ID" {
			continue
		}
		fieldNames = append(fieldNames, field.Name)
	}

	// Perform update with explicit field selection
	db := c.service.Database.Client().Model(new(TData)).Where("id = ?", id).Select(fieldNames).Updates(fields)
	if err := db.Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s", id)
	}

	// Reload with preloads
	preloads = horizon.MergeString(c.preloads, preloads)
	reloadDb := c.service.Database.Client().Model(new(TData)).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update", id)
	}

	c.UpdatedBroadcast(ctx, fields)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFieldsWithTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {
	// Get field names using reflection
	t := reflect.TypeOf(new(TData)).Elem()
	fieldNames := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "ID" {
			continue
		}
		fieldNames = append(fieldNames, field.Name)
	}

	// Perform update with explicit field selection
	db := tx.Model(new(TData)).Where("id = ?", id).Select(fieldNames).Updates(fields)
	if err := db.Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s in transaction", id)
	}

	// Reload with preloads
	preloads = horizon.MergeString(c.preloads, preloads)
	reloadDb := tx.Model(new(TData)).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update in transaction", id)
	}

	c.UpdatedBroadcast(ctx, fields)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateMany(
	ctx context.Context,
	entities []*TData,
	preloads ...string,
) error {
	for _, entity := range entities {
		if err := c.Update(ctx, entity, preloads...); err != nil {
			return eris.Wrapf(err, "failed to update entity in batch")
		}
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdateManyWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entities []*TData,
	preloads ...string,
) error {
	for _, entity := range entities {
		if err := c.UpdateWithTx(ctx, tx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) Delete(
	ctx context.Context,
	entity *TData,
) error {
	if err := c.service.Database.Client().Delete(entity).Error; err != nil {
		return eris.Wrap(err, "delete operation failed")
	}
	c.DeletedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeleteWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entity *TData,
) error {
	if err := tx.Delete(entity).Error; err != nil {
		return eris.Wrap(err, "delete operation failed in transaction")
	}
	c.DeletedBroadcast(ctx, entity)
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	entity, err := c.GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "failed to find entity %s for deletion", id)
	}
	return c.Delete(ctx, entity)
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByIDWithTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) error {
	entity, err := c.GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "failed to find entity %s for deletion in transaction", id)
	}
	return c.DeleteWithTx(ctx, tx, entity)
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeleteMany(
	ctx context.Context,
	entities []*TData,
) error {
	for _, entity := range entities {
		if err := c.Delete(ctx, entity); err != nil {
			return eris.Wrapf(err, "failed to delete entity in batch")
		}
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeleteManyWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entities []*TData,
) error {
	for _, entity := range entities {
		if err := c.DeleteWithTx(ctx, tx, entity); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) Upsert(
	ctx context.Context,
	entity *TData,
	preloads ...string,
) error {
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "invalid entity ID for upsert")
	}

	if id == uuid.Nil {
		return c.Create(ctx, entity, preloads...)
	}

	// Check if entity exists
	var existing TData
	if err := c.service.Database.Client().Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return c.Create(ctx, entity, preloads...)
		}
		return eris.Wrapf(err, "failed to check existence for entity %s", id)
	}

	return c.Update(ctx, entity, preloads...)
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpsertWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entity *TData,
	preloads ...string,
) error {
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "invalid entity ID for upsert in transaction")
	}

	if id == uuid.Nil {
		return c.CreateWithTx(ctx, tx, entity, preloads...)
	}

	// Check if entity exists
	var existing TData
	if err := tx.Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return c.CreateWithTx(ctx, tx, entity, preloads...)
		}
		return eris.Wrapf(err, "failed to check existence for entity %s in transaction", id)
	}

	return c.UpdateWithTx(ctx, tx, entity, preloads...)
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpsertMany(
	ctx context.Context,
	entities []*TData,
	preloads ...string,
) error {
	for _, entity := range entities {
		if err := c.Upsert(ctx, entity, preloads...); err != nil {
			return eris.Wrapf(err, "failed to upsert entity in batch")
		}
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpsertManyWithTx(
	ctx context.Context,
	tx *gorm.DB,
	entities []*TData,
	preloads ...string,
) error {
	for _, entity := range entities {
		if err := c.UpsertWithTx(ctx, tx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

// --- Helper Methods ---
func (c *CollectionManager[TData, TResponse, TRequest]) reloadWithPreloads(
	entity *TData,
	preloads []string,
) error {
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) == 0 {
		return nil
	}

	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "cannot reload without ID")
	}

	db := c.service.Database.Client().Model(new(TData))
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id = ?", id).First(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s", id)
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) reloadWithPreloadsTx(
	tx *gorm.DB,
	entity *TData,
	preloads []string,
) error {
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) == 0 {
		return nil
	}

	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "cannot reload without ID in transaction")
	}

	db := tx.Model(new(TData))
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id = ?", id).First(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s in transaction", id)
	}
	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) reloadManyWithPreloads(
	entities []*TData,
	preloads []string,
) error {
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(entities))
	for i, entity := range entities {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrapf(err, "cannot reload entity at index %d", i)
		}
		ids[i] = id
	}

	var reloaded []*TData
	db := c.service.Database.Client().Model(new(TData))
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
		return eris.Wrapf(err, "failed to reload %d entities", len(entities))
	}

	// Map reloaded entities by ID
	reloadedMap := make(map[uuid.UUID]*TData, len(reloaded))
	for _, e := range reloaded {
		id, _ := getID(e)
		reloadedMap[id] = e
	}

	// Replace original entities with reloaded ones
	for i, entity := range entities {
		id, _ := getID(entity)
		if reloaded, ok := reloadedMap[id]; ok {
			entities[i] = reloaded
		} else {
			return eris.Errorf("reloaded entity %s not found", id)
		}
	}

	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) reloadManyWithPreloadsTx(
	tx *gorm.DB,
	entities []*TData,
	preloads []string,
) error {
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(entities))
	for i, entity := range entities {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrapf(err, "cannot reload entity at index %d in transaction", i)
		}
		ids[i] = id
	}

	var reloaded []*TData
	db := tx.Model(new(TData))
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
		return eris.Wrapf(err, "failed to reload %d entities in transaction", len(entities))
	}

	// Map reloaded entities by ID
	reloadedMap := make(map[uuid.UUID]*TData, len(reloaded))
	for _, e := range reloaded {
		id, _ := getID(e)
		reloadedMap[id] = e
	}

	// Replace original entities with reloaded ones
	for i, entity := range entities {
		id, _ := getID(entity)
		if reloaded, ok := reloadedMap[id]; ok {
			entities[i] = reloaded
		} else {
			return eris.Errorf("reloaded entity %s not found in transaction", id)
		}
	}

	return nil
}

func (c *CollectionManager[TData, TResponse, TRequest]) CreatedBroadcast(
	ctx context.Context,
	entity *TData,
) {
	go func() {
		topics := c.created(entity)
		payload := c.ToModel(entity)
		if err := c.service.Broker.Dispatch(ctx, topics, payload); err != nil {
			if c.service.Logger != nil {
				c.service.Logger.Error("CreatedBroadcast dispatch error", zap.Error(err))
			}
		}
	}()
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdatedBroadcast(
	ctx context.Context,
	entity *TData,
) {
	go func() {
		topics := c.updated(entity)
		payload := c.ToModel(entity)
		if err := c.service.Broker.Dispatch(ctx, topics, payload); err != nil {
			if c.service.Logger != nil {
				c.service.Logger.Error("UpdatedBroadcast dispatch error", zap.Error(err))
			}
		}
	}()
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeletedBroadcast(
	ctx context.Context,
	entity *TData,
) {
	go func() {
		topics := c.deleted(entity)
		payload := c.ToModel(entity)
		if err := c.service.Broker.Dispatch(ctx, topics, payload); err != nil {
			if c.service.Logger != nil {
				c.service.Logger.Error("DeletedBroadcast dispatch error", zap.Error(err))
			}
		}
	}()
}

// --- ID Helpers ---
func getID[T any](entity *T) (uuid.UUID, error) {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil, eris.New("entity missing ID field")
	}

	id, ok := idField.Interface().(uuid.UUID)
	if !ok {
		return uuid.Nil, eris.New("ID field is not UUID type")
	}
	return id, nil
}

func setID[T any](entity *T, id uuid.UUID) error {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return eris.New("entity missing ID field")
	}
	if !idField.CanSet() {
		return eris.New("ID field cannot be set")
	}

	idField.Set(reflect.ValueOf(id))
	return nil
}
