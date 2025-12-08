package pagination

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"gorm.io/gorm"
)

func (f *Pagination[T]) DataGorm(
	db *gorm.DB,
	filterRoot Root,
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
	query := f.query(db, filterRoot)
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

func (f *Pagination[T]) DataGormNoPage(
	db *gorm.DB,
	filterRoot Root,
) ([]*T, error) {
	query := f.query(db, filterRoot)
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	return data, nil
}

func (f *Pagination[T]) Count(
	db *gorm.DB,
	filterRoot Root,
) (int64, error) {
	query := f.query(db, filterRoot)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return totalCount, nil
}

func (f *Pagination[T]) Tabular(
	db *gorm.DB,
	filterRoot Root,
	getter func(data *T) map[string]any,
) ([]byte, error) {
	data, err := f.DataGormNoPage(db, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)
	if len(data) == 0 {
		return buf.Bytes(), nil
	}
	firstRowFields := getter(data[0])
	fieldNames := make([]string, 0, len(firstRowFields))
	for k := range firstRowFields {
		fieldNames = append(fieldNames, k)
	}
	if err := csvWriter.Write(fieldNames); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	for _, item := range data {
		itemFields := getter(item)
		record := make([]string, len(fieldNames))
		for i, fieldName := range fieldNames {
			if value, exists := itemFields[fieldName]; exists {
				record[i] = fmt.Sprintf("%v", value)
			} else {
				record[i] = ""
			}
		}
		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %w", err)
	}
	return buf.Bytes(), nil
}
