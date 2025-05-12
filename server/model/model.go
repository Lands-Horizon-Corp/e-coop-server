package model

import (
	"net/http"
	"reflect"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	Model struct {
		validator *validator.Validate
		storage   *horizon.HorizonStorage
		qr        *horizon.HorizonQR
	}

	CollectionManager[T any] interface {
		// --- Retrieval ---

		List(preloads ...string) ([]*T, error)
		GetByID(id uuid.UUID, preloads ...string) (*T, error)
		Find(fields *T, preloads ...string) ([]*T, error)
		FindOne(fields *T, preloads ...string) (*T, error)

		// --- Aggregation ---
		Count(fields *T) (int64, error)
		CountWithTx(tx *gorm.DB, fields *T) (int64, error)

		// --- Creation ---

		Create(entity *T, preloads ...string) error
		CreateWithTx(tx *gorm.DB, entity *T, preloads ...string) error
		CreateMany(entities []*T, preloads ...string) error
		CreateManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error

		// --- Update ---

		Update(entity *T, preloads ...string) error
		UpdateWithTx(tx *gorm.DB, entity *T, preloads ...string) error
		UpdateByID(id uuid.UUID, entity *T, preloads ...string) error
		UpdateByIDWithTx(tx *gorm.DB, id uuid.UUID, entity *T, preloads ...string) error
		UpdateFields(id uuid.UUID, fields *T, preloads ...string) error
		UpdateFieldsWithTx(tx *gorm.DB, id uuid.UUID, fields *T, preloads ...string) error
		UpdateMany(entities []*T, preloads ...string) error
		UpdateManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error

		// --- Upsert ---

		Upsert(entity *T, preloads ...string) error
		UpsertWithTx(tx *gorm.DB, entity *T, preloads ...string) error
		UpsertMany(entities []*T, preloads ...string) error
		UpsertManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error

		// --- Deletion ---

		Delete(entity *T) error
		DeleteWithTx(tx *gorm.DB, entity *T) error
		DeleteByID(id uuid.UUID) error
		DeleteByIDWithTx(tx *gorm.DB, id uuid.UUID) error
		DeleteMany(entities []*T) error
		DeleteManyWithTx(tx *gorm.DB, entities []*T) error
	}

	collectionManager[T any] struct {
		database  *horizon.HorizonDatabase
		broadcast *horizon.HorizonBroadcast
		created   func(*T) ([]string, any)
		updated   func(*T) ([]string, any)
		deleted   func(*T) ([]string, any)
		preloads  []string
	}
)

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
func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}
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
func NewcollectionManager[T any](
	database *horizon.HorizonDatabase,
	broadcast *horizon.HorizonBroadcast,
	created func(*T) ([]string, any),
	updated func(*T) ([]string, any),
	deleted func(*T) ([]string, any),
	preloads []string,
) CollectionManager[T] {
	return &collectionManager[T]{
		database:  database,
		broadcast: broadcast,
		created:   created,
		updated:   updated,
		deleted:   deleted,
		preloads:  preloads,
	}
}
func (r *collectionManager[T]) List(preloads ...string) ([]*T, error) {
	var entities []*T
	db := r.database.Client().Model(new(T))
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list entities")
	}
	return entities, nil
}

func (r *collectionManager[T]) GetByID(id uuid.UUID, preloads ...string) (*T, error) {
	var entity T
	db := r.database.Client().Model(new(T))
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entity with id: %s", id)
	}
	return &entity, nil
}

func (r *collectionManager[T]) Find(fields *T, preloads ...string) ([]*T, error) {
	var entities []*T
	db := r.database.Client().Model(fields).Where(fields)
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities")
	}
	return entities, nil
}

func (r *collectionManager[T]) FindOne(fields *T, preloads ...string) (*T, error) {
	var entity T
	db := r.database.Client().Model(fields).Where(fields)
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("created_at DESC").First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entity")
	}
	return &entity, nil
}

func (r *collectionManager[T]) Count(fields *T) (int64, error) {
	var count int64
	if err := r.database.Client().Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities")
	}
	return count, nil
}

func (r *collectionManager[T]) CountWithTx(tx *gorm.DB, fields *T) (int64, error) {
	var count int64
	if err := tx.Model(fields).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities in transaction")
	}
	return count, nil
}

func (r *collectionManager[T]) Create(entity *T, preloads ...string) error {
	if err := r.database.Client().Create(entity).Error; err != nil {
		return eris.Wrap(err, "failed to create entity")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload")
		}
		db := r.database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads")
		}
	}
	r.CreatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) CreateWithTx(tx *gorm.DB, entity *T, preloads ...string) error {
	if err := tx.Create(entity).Error; err != nil {
		return eris.Wrap(err, "failed to create entity in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
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
	r.CreatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) CreateMany(entities []*T, preloads ...string) error {
	if err := r.database.Client().Create(entities).Error; err != nil {
		return eris.Wrap(err, "failed to create entities")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		ids := make([]uuid.UUID, len(entities))
		for i, entity := range entities {
			id, err := getID(entity)
			if err != nil {
				return eris.Wrap(err, "failed to get ID for entity")
			}
			ids[i] = id
		}
		var reloaded []*T
		db := r.database.Client().Model(new(T))
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
			return eris.Wrap(err, "failed to reload entities with preloads")
		}
		reloadedMap := make(map[uuid.UUID]*T)
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
	r.CreatedBroadcastMany(entities)
	return nil
}

func (r *collectionManager[T]) CreateManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error {
	if err := tx.Create(entities).Error; err != nil {
		return eris.Wrap(err, "failed to create entities in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		ids := make([]uuid.UUID, len(entities))
		for i, entity := range entities {
			id, err := getID(entity)
			if err != nil {
				return eris.Wrap(err, "failed to get ID for entity in transaction")
			}
			ids[i] = id
		}
		var reloaded []*T
		db := tx.Model(new(T))
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.Where("id IN (?)", ids).Find(&reloaded).Error; err != nil {
			return eris.Wrap(err, "failed to reload entities with preloads in transaction")
		}
		reloadedMap := make(map[uuid.UUID]*T)
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
	r.CreatedBroadcastMany(entities)
	return nil
}

func (r *collectionManager[T]) Update(entity *T, preloads ...string) error {
	if err := r.database.Client().Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		id, err := getID(entity)
		if err != nil {
			return eris.Wrap(err, "failed to get entity ID for preload after update")
		}
		db := r.database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity with preloads after update")
		}
	}
	r.UpdatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) UpdateWithTx(tx *gorm.DB, entity *T, preloads ...string) error {
	if err := tx.Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
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
	r.UpdatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) UpdateByID(id uuid.UUID, entity *T, preloads ...string) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "failed to set entity ID")
	}
	if err := r.database.Client().Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity by ID")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		db := r.database.Client().Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity after update by ID")
		}
	}
	r.UpdatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) UpdateByIDWithTx(tx *gorm.DB, id uuid.UUID, entity *T, preloads ...string) error {
	if err := setID(entity, id); err != nil {
		return eris.Wrap(err, "failed to set entity ID in transaction")
	}
	if err := tx.Save(entity).Error; err != nil {
		return eris.Wrap(err, "failed to update entity by ID in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if len(preloads) > 0 {
		db := tx.Model(entity)
		for _, preload := range preloads {
			db = db.Preload(preload)
		}
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			return eris.Wrap(err, "failed to reload entity after update by ID in transaction")
		}
	}
	r.UpdatedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) UpdateFields(id uuid.UUID, fields *T, preloads ...string) error {
	if err := r.database.Client().Model(new(T)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return eris.Wrap(err, "failed to update fields")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	db := r.database.Client().Model(new(T)).Where("id = ?", id)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.First(fields).Error; err != nil {
		return eris.Wrap(err, "failed to reload entity after updating fields")
	}
	r.UpdatedBroadcast(fields)
	return nil
}

func (r *collectionManager[T]) UpdateFieldsWithTx(tx *gorm.DB, id uuid.UUID, fields *T, preloads ...string) error {
	if err := tx.Model(new(T)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return eris.Wrap(err, "failed to update fields in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	db := tx.Model(new(T)).Where("id = ?", id)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.First(fields).Error; err != nil {
		return eris.Wrap(err, "failed to reload entity after updating fields in transaction")
	}
	r.UpdatedBroadcast(fields)
	return nil
}

func (r *collectionManager[T]) UpdateMany(entities []*T, preloads ...string) error {
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, entity := range entities {
		if err := r.Update(entity, preloads...); err != nil {
			return eris.Wrap(err, "failed to update many entities")
		}
	}
	r.UpdatedBroadcast_many(entities)
	return nil
}

func (r *collectionManager[T]) UpdateManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error {
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, entity := range entities {
		if err := r.UpdateWithTx(tx, entity, preloads...); err != nil {
			return eris.Wrap(err, "failed to update many entities in transaction")
		}
	}
	r.UpdatedBroadcast_many(entities)
	return nil
}

func (r *collectionManager[T]) Upsert(entity *T, preloads ...string) error {
	preloads = horizon.MergeString(r.preloads, preloads)
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "failed to get ID for upsert")
	}
	if id == uuid.Nil {
		return r.Create(entity, preloads...)
	}
	var existing T
	if err := r.database.Client().Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return r.Create(entity, preloads...)
		}
		return eris.Wrap(err, "failed to check existing entity for upsert")
	}
	return r.Update(entity, preloads...)
}

func (r *collectionManager[T]) UpsertWithTx(tx *gorm.DB, entity *T, preloads ...string) error {
	id, err := getID(entity)
	if err != nil {
		return eris.Wrap(err, "failed to get ID for upsert in transaction")
	}
	preloads = horizon.MergeString(r.preloads, preloads)
	if id == uuid.Nil {
		return r.CreateWithTx(tx, entity, preloads...)
	}
	var existing T
	if err := tx.Where("id = ?", id).First(&existing).Error; err != nil {
		if eris.Is(err, gorm.ErrRecordNotFound) {
			return r.CreateWithTx(tx, entity, preloads...)
		}
		return eris.Wrap(err, "failed to check existing entity for upsert in transaction")
	}
	return r.UpdateWithTx(tx, entity, preloads...)
}

func (r *collectionManager[T]) UpsertMany(entities []*T, preloads ...string) error {
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, entity := range entities {
		if err := r.Upsert(entity, preloads...); err != nil {
			return eris.Wrap(err, "failed to upsert many entities")
		}
	}
	return nil
}

func (r *collectionManager[T]) UpsertManyWithTx(tx *gorm.DB, entities []*T, preloads ...string) error {
	preloads = horizon.MergeString(r.preloads, preloads)
	for _, entity := range entities {
		if err := r.UpsertWithTx(tx, entity, preloads...); err != nil {
			return eris.Wrap(err, "failed to upsert many entities in transaction")
		}
	}
	return nil
}

func (r *collectionManager[T]) Delete(entity *T) error {
	if err := r.database.Client().Delete(entity).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity")
	}
	r.DeletedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) DeleteWithTx(tx *gorm.DB, entity *T) error {
	if err := tx.Delete(entity).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity in transaction")
	}
	r.DeletedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) DeleteByID(id uuid.UUID) error {
	entity := new(T)

	if err := r.database.Client().First(entity, "id = ?", id).Error; err != nil {
		return eris.Wrapf(err, "failed to load entity with id %s before deletion", id)
	}

	if err := r.database.Client().Delete(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to delete entity with id %s", id)
	}

	r.DeletedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) DeleteByIDWithTx(tx *gorm.DB, id uuid.UUID) error {
	entity := new(T)
	if err := tx.First(entity, "id = ?", id).Error; err != nil {
		return eris.Wrapf(err, "failed to load entity with id %s before deletion in transaction", id)
	}
	if err := tx.Delete(entity).Error; err != nil {
		return eris.Wrapf(err, "failed to delete entity with id %s in transaction", id)
	}
	r.DeletedBroadcast(entity)
	return nil
}

func (r *collectionManager[T]) DeleteMany(entities []*T) error {
	if err := r.database.Client().Delete(entities).Error; err != nil {
		return eris.Wrap(err, "failed to delete entities")
	}
	r.DeletedBroadcastMany(entities)
	return nil
}

func (r *collectionManager[T]) DeleteManyWithTx(tx *gorm.DB, entities []*T) error {
	if err := tx.Delete(entities).Error; err != nil {
		return eris.Wrap(err, "failed to delete entities in transaction")
	}
	r.DeletedBroadcastMany(entities)
	return nil
}

func (r *collectionManager[T]) CreatedBroadcast(entity *T) {
	go func() {
		topics, payload := r.created(entity)
		r.broadcast.Dispatch(topics, payload)
	}()
}

func (r *collectionManager[T]) UpdatedBroadcast(entity *T) {
	go func() {
		topics, payload := r.updated(entity)
		r.broadcast.Dispatch(topics, payload)
	}()
}

func (r *collectionManager[T]) DeletedBroadcast(entity *T) {
	go func() {
		topics, payload := r.updated(entity)
		r.broadcast.Dispatch(topics, payload)
	}()
}

func (r *collectionManager[T]) CreatedBroadcastMany(entities []*T) {
	for _, entity := range entities {
		r.CreatedBroadcast(entity)
	}
}

func (r *collectionManager[T]) UpdatedBroadcast_many(entities []*T) {
	for _, entity := range entities {
		r.UpdatedBroadcast(entity)
	}
}

func (r *collectionManager[T]) DeletedBroadcastMany(entities []*T) {
	for _, entity := range entities {
		r.DeletedBroadcast(entity)
	}
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
