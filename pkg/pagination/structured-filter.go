package pagination

import (
	"fmt"

	"gorm.io/gorm"
)

func (f *Pagination[T]) StructuredPagination(
	db *gorm.DB,
	filterRoot StructuredFilter,
	pageIndex int,
	pageSize int,
) (*PaginationResult[T], error) {
	result := PaginationResult[T]{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	if result.PageIndex < 0 {
		result.PageIndex = 0
	}
	if result.PageSize <= 0 {
		result.PageSize = 30
	}
	query := f.structuredQuery(db, filterRoot)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	result.TotalSize = int(totalCount)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize
	offset := result.PageIndex * result.PageSize
	query = query.Offset(int(offset)).Limit(int(result.PageSize))
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	result.Data = data
	return &result, nil
}

func (f *Pagination[T]) StructuredCount(
	db *gorm.DB,
	filterRoot StructuredFilter,
) (int64, error) {
	query := f.structuredQuery(db, filterRoot)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return totalCount, nil
}

func (f *Pagination[T]) StructuredFind(
	db *gorm.DB,
	filterRoot StructuredFilter,
) ([]*T, error) {
	query := f.structuredQuery(db, filterRoot)
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	return data, nil
}

func (f *Pagination[T]) StructuredTabular(
	db *gorm.DB,
	filterRoot StructuredFilter,
	getter func(data *T) map[string]any,
) ([]byte, error) {
	data, err := f.StructuredFind(db, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}
