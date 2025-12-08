package pagination

import (
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (p Pagination[T]) Count(
	db *gorm.DB,
	filters []FilterSQL,
) (int64, error) {
	var count int64
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	if err := db.Model(new(T)).Count(&count).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to count entities with %d filters", len(filters))
	}
	return count, nil
}

func (p *Pagination[T]) Find(
	db *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order("updated_at DESC")
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with %d filters: %w", len(filters), err)
	}
	return entities, nil
}

func (r *Pagination[T]) FindLock(
	db *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = r.applyJoinsForFilters(db, filters)
	db = r.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = r.applySort(db, sorts)
	} else {
		db = db.Order("updated_at DESC")
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Find(&entities).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entities with %d filters and lock", len(filters))
	}
	return entities, nil
}

func (p *Pagination[T]) FindOne(
	db *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order("updated_at DESC")
	}
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with %d filters: %w", len(filters), err)
	}
	return &entity, nil
}

func (p *Pagination[T]) FindOneWithLock(
	db *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order("updated_at DESC")
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with lock and %d filters: %w", len(filters), err)
	}
	return &entity, nil
}

func (p *Pagination[T]) Exists(
	db *gorm.DB,
	filters []FilterSQL,
) (bool, error) {
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	var dummy int
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ExistsByID(
	db *gorm.DB,
	id any,
) (bool, error) {
	var dummy int
	subQuery := db.Model(new(T)).Select("1").Where("id = ?", id).Limit(1)
	err := db.Raw("SELECT EXISTS (?)", subQuery).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence by ID: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ExistsIncludingDeleted(
	db *gorm.DB,
	filters []FilterSQL,
) (bool, error) {
	db = db.Unscoped()
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	var dummy int
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence including deleted: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) GetMax(
	db *gorm.DB,
	field string,
	filters []FilterSQL,
) (any, error) {
	var result any
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	err := db.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMin(
	db *gorm.DB,
	field string,
	filters []FilterSQL,
) (any, error) {
	var result any
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	err := db.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMaxLock(
	tx *gorm.DB,
	field string,
	filters []FilterSQL,
) (any, error) {
	var result any
	tx = p.applyJoinsForFilters(tx, filters)
	tx = p.applySQLFilters(tx, filters)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMinLock(
	tx *gorm.DB,
	field string,
	filters []FilterSQL,
) (any, error) {
	var result any
	tx = p.applyJoinsForFilters(tx, filters)
	tx = p.applySQLFilters(tx, filters)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetByID(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v: %w", id, err)
	}

	return &entity, nil
}

func (p *Pagination[T]) GetByIDLock(
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v with lock: %w", id, err)
	}

	return &entity, nil
}
