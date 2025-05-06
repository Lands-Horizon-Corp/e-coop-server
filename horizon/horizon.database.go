package horizon

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type HorizonDatabase struct {
	config *HorizonConfig
	client *gorm.DB
}

func NewHorizonDatabase(config *HorizonConfig) (*HorizonDatabase, error) {
	return &HorizonDatabase{
		config: config,
		client: nil,
	}, nil
}

func (hd *HorizonDatabase) run() error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		hd.config.PostgresHost,
		hd.config.PostgresPort,
		hd.config.PostgresUser,
		hd.config.PostgresPassword,
		hd.config.PostgresDB,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	hd.client = db
	return nil
}

func (hd *HorizonDatabase) stop() error {
	if hd.client == nil {
		return nil
	}
	sqlDB, err := hd.client.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	hd.client = nil
	return nil
}

func (hd *HorizonDatabase) Client() *gorm.DB {
	return hd.client
}

func (hd *HorizonDatabase) Ping() error {
	sqlDB, err := hd.client.DB()
	if err != nil {
		return fmt.Errorf("failed to get raw DB from GORM: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
