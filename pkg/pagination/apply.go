package pagination

import "gorm.io/gorm"

func ApplyPresetConditions(db *gorm.DB, conditions any) *gorm.DB {
	if conditions == nil {
		return db
	}
	return db.Where(conditions)
}
