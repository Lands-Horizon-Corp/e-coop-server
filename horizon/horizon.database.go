package horizon

import (
	"context"
	"log"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseImpl struct {
	dsn         string
	db          *gorm.DB
	maxIdleConn int
	maxOpenConn int
	maxLifetime time.Duration
}

func NewDatabaseImpl(dsn string, maxIdle, maxOpen int, maxLifetime time.Duration) *DatabaseImpl {
	log.Println("Database DSN: ", dsn)
	return &DatabaseImpl{
		dsn:         dsn,
		maxIdleConn: maxIdle,
		maxOpenConn: maxOpen,
		maxLifetime: maxLifetime,
	}
}

func (g *DatabaseImpl) Client() *gorm.DB {
	return g.db
}

func (g *DatabaseImpl) Ping(ctx context.Context) error {
	if g.db == nil {
		return eris.New("database not started")
	}
	sqlDB, err := g.db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get generic database object")
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return eris.Wrap(err, "ping failed")
	}
	return nil
}

func (g *DatabaseImpl) Run(ctx context.Context) error {
	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	db, err := gorm.Open(postgres.Open(g.dsn), config)
	if err != nil {
		return eris.Wrap(err, "failed to open database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get generic database object")
	}
	sqlDB.SetMaxIdleConns(g.maxIdleConn)
	sqlDB.SetMaxOpenConns(g.maxOpenConn)
	sqlDB.SetConnMaxLifetime(g.maxLifetime)
	if err := sqlDB.PingContext(ctx); err != nil {
		return eris.Wrap(err, "database ping failed")
	}
	g.db = db
	return nil
}

func (g *DatabaseImpl) Stop() error {
	if g.db == nil {
		return nil
	}
	sqlDB, err := g.db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get generic database object")
	}
	return eris.Wrap(sqlDB.Close(), "failed to close database")
}

func (g *DatabaseImpl) StartTransaction(ctx context.Context) (*gorm.DB, func(error) error) {
	tx := g.db.WithContext(ctx).Begin()
	end := func(err error) error {
		if err != nil {
			tx.Rollback()
			return err
		}
		if commitErr := tx.Commit().Error; commitErr != nil {
			return commitErr
		}
		return nil
	}
	return tx, end
}

func (g *DatabaseImpl) StartTransactionWithContext(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
