package horizon_services

import (
	"context"
	"net/http"
	"reflect"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type Repository[TData any, TResponse any, TRequest any] interface {

	// Validate
	Validate(ctx echo.Context) (*TRequest, error)

	// Models
	ToModel(data *TData) *TResponse

	// Convert data to anything
	ToModels(data []*TData) []*TResponse

	// --- Retrieval ---
	// List retrieves all entities of type T, optionally with related entities specified in preloads.
	List(ctx context.Context, preloads ...string) ([]*TData, error)
	ListRaw(ctx context.Context, preloads ...string) ([]*TResponse, error)

	// GetByID retrieves a single entity by its UUID. Optionally preloads related entities.
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TData, error)
	GetByIDRaw(ctx context.Context, id uuid.UUID, preloads ...string) (*TResponse, error)

	// Find retrieves all entities that match the non-zero fields of the provided struct.
	// Optionally preloads related entities.
	Find(ctx context.Context, fields *TData, preloads ...string) ([]*TData, error)
	FindRaw(ctx context.Context, fields *TData, preloads ...string) ([]*TResponse, error)

	// FindOne retrieves a single entity that matches the non-zero fields of the provided struct.
	// Optionally preloads related entities.
	FindOne(ctx context.Context, fields *TData, preloads ...string) (*TData, error)
	FindOneRaw(ctx context.Context, fields *TData, preloads ...string) (*TResponse, error)

	// --- Aggregation ---

	// Count returns the number of records matching the given fields.
	Count(ctx context.Context, fields *TData) (int64, error)

	// CountWithTx performs Count using the provided GORM transaction.
	CountWithTx(ctx context.Context, tx *gorm.DB, fields *TData) (int64, error)

	// --- Creation ---

	// Create inserts a new record into the database.
	// Optionally preloads related entities after creation.
	Create(ctx context.Context, entity *TData, preloads ...string) error

	// CreateWithTx performs Create within the provided transaction.
	CreateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// CreateMany inserts multiple records in a batch.
	CreateMany(ctx context.Context, entities []*TData, preloads ...string) error

	// CreateManyWithTx performs CreateMany within the provided transaction.
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Update ---

	// Update modifies an existing record in the database.
	Update(ctx context.Context, entity *TData, preloads ...string) error

	// UpdateWithTx performs Update within the provided transaction.
	UpdateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// UpdateByID updates a record by its UUID.
	UpdateByID(ctx context.Context, id uuid.UUID, entity *TData, preloads ...string) error

	// UpdateByIDWithTx performs UpdateByID within the provided transaction.
	UpdateByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, entity *TData, preloads ...string) error

	// UpdateFields updates only the non-zero fields of the entity for the given UUID.
	UpdateFields(ctx context.Context, id uuid.UUID, fields *TData, preloads ...string) error

	// UpdateFieldsWithTx performs UpdateFields within the provided transaction.
	UpdateFieldsWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, fields *TData, preloads ...string) error

	// UpdateMany performs a batch update on multiple entities.
	UpdateMany(ctx context.Context, entities []*TData, preloads ...string) error

	// UpdateManyWithTx performs UpdateMany within the provided transaction.
	UpdateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Upsert ---

	// Upsert inserts a new record or updates it if it already exists.
	Upsert(ctx context.Context, entity *TData, preloads ...string) error

	// UpsertWithTx performs Upsert within the provided transaction.
	UpsertWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// UpsertMany performs batch upsert for multiple entities.
	UpsertMany(ctx context.Context, entities []*TData, preloads ...string) error

	// UpsertManyWithTx performs UpsertMany within the provided transaction.
	UpsertManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Deletion ---

	// Delete removes the specified entity from the database.
	Delete(ctx context.Context, entity *TData) error

	// DeleteWithTx performs Delete within the provided transaction.
	DeleteWithTx(ctx context.Context, tx *gorm.DB, entity *TData) error

	// DeleteByID deletes a record by its UUID.
	DeleteByID(ctx context.Context, id uuid.UUID) error

	// DeleteByIDWithTx performs DeleteByID within the provided transaction.
	DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error

	// DeleteMany deletes multiple entities in a batch.
	DeleteMany(ctx context.Context, entities []*TData) error

	// DeleteManyWithTx performs DeleteMany within the provided transaction.
	DeleteManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData) error
}

// RepositoryParams groups the constructor parameters for NewRepository
type RepositoryParams[TData any, TResponse any, TRequest any] struct {
	Service  *HorizonService
	Created  func(*TData) []string
	Updated  func(*TData) []string
	Deleted  func(*TData) []string
	Resource func(*TData) *TResponse
	Preloads []string
}

// CollectionManager is a generic implementation of Repository
type CollectionManager[TData any, TResponse any, TRequest any] struct {
	service  *HorizonService
	created  func(*TData) []string
	updated  func(*TData) []string
	deleted  func(*TData) []string
	resource func(*TData) *TResponse
	preloads []string
}

// NewRepository creates a new CollectionManager instance with the given parameters
func NewRepository[TData any, TResponse any, TRequest any](params RepositoryParams[TData, TResponse, TRequest]) Repository[TData, TResponse, TRequest] {
	return &CollectionManager[TData, TResponse, TRequest]{
		service:  params.Service,
		created:  params.Created,
		updated:  params.Updated,
		deleted:  params.Deleted,
		resource: params.Resource,
		preloads: params.Preloads,
	}
}

// ToModel implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

// ToModels implements Repository.
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

// Validate implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Validate(ctx echo.Context) (*TRequest, error) {
	var req TRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.service.Validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}

// Count implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Count(ctx context.Context, fields *TData) (int64, error) {
	var count int64
	if err := c.service.Database.Client().Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities")
	}
	return count, nil
}

// CountWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CountWithTx(ctx context.Context, tx *gorm.DB, fields *TData) (int64, error) {
	var count int64
	if err := tx.Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities in transaction")
	}
	return count, nil
}

// Create implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Create(ctx context.Context, entity *TData, preloads ...string) error {
	if err := c.service.Database.Client().Create(entity).Error; err != nil {
		return eris.Wrap(err, "failed to create entity")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload")
		}
		db := c.service.Database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads")
		}
	}
	c.CreatedBroadcast(ctx, entity)
	return nil
}

// CreateMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateMany(ctx context.Context, entities []*TData, preloads ...string) error {
	if err := c.service.Database.Client().Create(entities).Error; err != nil {
		return eris.Wrap(err, "failed to create entities")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		ids := make([]uuid.UUID, len(entities))
		for i, entity := range entities {
			id, err := getID(entity)
			if err != nil {
				return eris.Wrap(err, "failed to get ID for entity")
			}
			ids[i] = id
		}
		var reloaded []*TData
		db := c.service.Database.Client().Model(new(TData))
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
			return eris.Wrap(err, "failed to reload entities with preloads")
		}
		reloadedMap := make(map[uuid.UUID]*TData)
		for _, e := range reloaded {
			id, _ := getID(e)
			reloadedMap[id] = e
		}
		for _, entity := range entities {
			id, _ := getID(entity)
			if reloadedEntity, ok := reloadedMap[id]; ok {
				*entity = *reloadedEntity
			} else {
				return eris.Errorf("failed to find reloaded entity with ID %s", id)
			}
		}
	}
	return nil
}

// CreateManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	if err := tx.Create(entities).Error; err != nil {
		return eris.Wrap(err, "failed to create entities in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		ids := make([]uuid.UUID, len(entities))
		for i, entity := range entities {
			id, err := getID(entity)
			if err != nil {
				return eris.Wrap(err, "failed to get ID for entity in transaction")
			}
			ids[i] = id
		}
		var reloaded []*TData
		db := tx.Model(new(TData))
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
			return eris.Wrap(err, "failed to reload entities with preloads in transaction")
		}
		reloadedMap := make(map[uuid.UUID]*TData)
		for _, e := range reloaded {
			id, _ := getID(e)
			reloadedMap[id] = e
		}
		for _, entity := range entities {
			id, _ := getID(entity)
			if reloadedEntity, ok := reloadedMap[id]; ok {
				*entity = *reloadedEntity
			} else {
				return eris.Errorf("failed to find reloaded entity with ID %s in transaction", id)
			}
		}
	}
	return nil
}

// CreateWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	if err := tx.Create(entity).Error; err != nil {
		return eris.Wrap(err, "failed to create entity in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload in transaction")
		}
		db := tx.Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads in transaction")
		}
	}
	return nil
}

// Delete implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Delete(ctx context.Context, entity *TData) error {
	if err := c.service.Database.Client().Delete(entity).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity")
	}
	c.DeletedBroadcast(ctx, entity)
	return nil
}

// DeleteByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByID(ctx context.Context, id uuid.UUID) error {
	entity := new(TData)

	if err := c.service.Database.Client().First(entity, "id = ?", id).Error; err != nil {
		return eris.Wrapf(err, "failed to load entity with id %s before deletion", id)
	}

	if err := c.service.Database.Client().Delete(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to delete entity with id %s", id)
	}

	c.DeletedBroadcast(ctx, entity)
	return nil

}

// DeleteByIDWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	entity := new(TData)
	if err := tx.First(entity, "id = ?", id).Error; err != nil {
		return eris.Wrapf(err, "failed to load entity with id %s before deletion in transaction", id)
	}
	if err := tx.Delete(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to delete entity with id %s in transaction", id)
	}
	c.DeletedBroadcast(ctx, entity)
	return nil
}

// DeleteMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteMany(ctx context.Context, entities []*TData) error {
	for _, entity := range entities {
		if err := c.Delete(ctx, entity); err != nil {
			return eris.Wrapf(err, "failed to delete entity: %+v", entity)
		}
	}
	return nil
}

// DeleteManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData) error {
	for _, entity := range entities {
		if err := c.DeleteWithTx(ctx, tx, entity); err != nil {
			return err
		}
	}
	return nil
}

// DeleteWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteWithTx(ctx context.Context, tx *gorm.DB, entity *TData) error {
	if err := tx.Delete(entity).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity in transaction")
	}
	c.DeletedBroadcast(ctx, entity)
	return nil
}

// Find implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Find(ctx context.Context, fields *TData, preloads ...string) ([]*TData, error) {
	var entities []*TData
	db := c.service.Database.Client().Model(fields).Where(fields)
	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities")
	}
	return entities, nil
}

// FindOne implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindOne(ctx context.Context, fields *TData, preloads ...string) (*TData, error) {
	var entity TData
	db := c.service.Database.Client().Model(fields).Where(fields)
	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("created_at DESC").First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entity")
	}
	return &entity, nil
}

// FindOneRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindOneRaw(ctx context.Context, fields *TData, preloads ...string) (*TResponse, error) {
	entity, err := c.FindOne(ctx, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModel(entity), nil
}

// FindRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindRaw(ctx context.Context, fields *TData, preloads ...string) ([]*TResponse, error) {
	entity, err := c.Find(ctx, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModels(entity), nil
}

// GetByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TData, error) {
	var entity TData
	db := c.service.Database.Client().Model(new(TData))
	preloads = horizon.MergeString(c.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entity with id: %s", id)
	}
	return &entity, nil
}

// GetByIDRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) GetByIDRaw(ctx context.Context, id uuid.UUID, preloads ...string) (*TResponse, error) {
	entity, err := c.GetByID(ctx, id, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModel(entity), nil
}

// List implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) List(ctx context.Context, preloads ...string) ([]*TData, error) {
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

// ListRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) ListRaw(ctx context.Context, preloads ...string) ([]*TResponse, error) {
	entity, err := c.List(ctx, preloads...)
	if err != nil {
		return nil, err
	}
	return c.ToModels(entity), nil
}

// Update implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Update(ctx context.Context, entity *TData, preloads ...string) error {
	if err := c.service.Database.Client().Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload after update")
		}
		db := c.service.Database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads after update")
		}
	}
	c.UpdatedBroadcast(ctx, entity)
	return nil
}

// UpdateByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByID(ctx context.Context, id uuid.UUID, entity *TData, preloads ...string) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "failed to set entity ID")
	}
	if err := c.service.Database.Client().Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity by ID")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		db := c.service.Database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity after update by ID")
		}
	}
	c.UpdatedBroadcast(ctx, entity)
	return nil
}

// UpdateByIDWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, entity *TData, preloads ...string) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "failed to set entity ID in transaction")
	}
	if err := tx.Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity by ID in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		db := tx.Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity after update by ID in transaction")
		}
	}
	c.UpdatedBroadcast(ctx, entity)
	return nil
}

// UpdateFields implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFields(ctx context.Context, id uuid.UUID, fields *TData, preloads ...string) error {
	if err := c.service.Database.Client().Model(new(TData)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return eris.Wrap(err, "failed to update fields")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	db := c.service.Database.Client().Model(new(TData)).Where("id = ?", id)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.First(fields).Error; err != nil {
		return eris.Wrap(err, "failed to reload entity after updating fields")
	}
	c.UpdatedBroadcast(ctx, fields)
	return nil
}

// UpdateFieldsWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFieldsWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, fields *TData, preloads ...string) error {
	if err := tx.Model(new(TData)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return eris.Wrap(err, "failed to update fields in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	db := tx.Model(new(TData)).Where("id = ?", id)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.First(fields).Error; err != nil {
		return eris.Wrap(err, "failed to reload entity after updating fields in transaction")
	}
	c.UpdatedBroadcast(ctx, fields)
	return nil
}

// UpdateMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateMany(ctx context.Context, entities []*TData, preloads ...string) error {
	for _, entity := range entities {
		if err := c.Update(ctx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

// UpdateManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	for _, entity := range entities {
		if err := c.UpdateWithTx(ctx, tx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

// UpdateWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	if err := tx.Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload after update in transaction")
		}
		db := tx.Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads after update in transaction")
		}
	}
	c.UpdatedBroadcast(ctx, entity)
	return nil
}

// Upsert implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Upsert(ctx context.Context, entity *TData, preloads ...string) error {
	preloads = horizon.MergeString(c.preloads, preloads)
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "failed to get ID for upsert")
	}
	if id == uuid.Nil {
		return c.Create(ctx, entity, preloads...)
	}
	var existing TData
	if err := c.service.Database.Client().Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return c.Create(ctx, entity, preloads...)
		}
		return eris.Wrap(err, "failed to check existing entity for upsert")
	}
	return c.Update(ctx, entity, preloads...)
}

// UpsertMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertMany(ctx context.Context, entities []*TData, preloads ...string) error {
	for _, entity := range entities {
		if err := c.Upsert(ctx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

// UpsertManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	for _, entity := range entities {
		if err := c.UpsertWithTx(ctx, tx, entity, preloads...); err != nil {
			return err
		}
	}
	return nil
}

// UpsertWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "failed to get ID for upsert in transaction")
	}
	preloads = horizon.MergeString(c.preloads, preloads)
	if id == uuid.Nil {
		return c.CreateWithTx(ctx, tx, entity, preloads...)
	}
	var existing TData
	if err := tx.Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return c.CreateWithTx(ctx, tx, entity, preloads...)
		}
		return eris.Wrap(err, "failed to check existing entity for upsert in transaction")
	}
	return c.UpdateWithTx(ctx, tx, entity, preloads...)
}

func (c *CollectionManager[TData, TResponse, TRequest]) CreatedBroadcast(ctx context.Context, entity *TData) {
	go func() {
		topics := c.created(entity)
		payload := c.ToModel(entity)
		c.service.Broker.Dispatch(ctx, topics, payload)
	}()
}

func (c *CollectionManager[TData, TResponse, TRequest]) DeletedBroadcast(ctx context.Context, entity *TData) {
	go func() {
		topics := c.updated(entity)
		payload := c.ToModel(entity)
		c.service.Broker.Dispatch(ctx, topics, payload)
	}()
}

func (c *CollectionManager[TData, TResponse, TRequest]) UpdatedBroadcast(ctx context.Context, entity *TData) {
	go func() {
		topics := c.updated(entity)
		payload := c.ToModel(entity)
		c.service.Broker.Dispatch(ctx, topics, payload)
	}()
}

func getID[T any](entity *T) (uuid.UUID, error) {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil, eris.New("ID field not found in entity")
	}
	id, ok := idField.Interface().(uuid.UUID)
	if !ok {
		return uuid.Nil, eris.New("ID field is not a uuid.UUID")
	}
	return id, nil
}

func setID[T any](entity *T, id uuid.UUID) error {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return eris.New("ID field not found in entity")
	}
	if !idField.CanSet() {
		return eris.New("ID field cannot be set")
	}
	idField.Set(reflect.ValueOf(id))
	return nil
}
