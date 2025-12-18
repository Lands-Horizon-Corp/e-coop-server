package horizon

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*

dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable TimeZone=UTC"

database := NewGormDatabase(dsn, 10, 50, 30*time.Minute)

if err := database.Start(ctx); err != nil {
}
*/

type SQLDatabaseService interface {
	Run(ctx context.Context) error

	Stop(ctx context.Context) error

	Client() *gorm.DB

	Ping(ctx context.Context) error

	StartTransaction(ctx context.Context) (*gorm.DB, func(error) error)

	StartTransactionWithContext(ctx context.Context, fn func(*gorm.DB) error) error
}

type GormDatabase struct {
	dsn         string
	db          *gorm.DB
	maxIdleConn int
	maxOpenConn int
	maxLifetime time.Duration
}

func NewGormDatabase(dsn string, maxIdle, maxOpen int, maxLifetime time.Duration) SQLDatabaseService {
	return &GormDatabase{
		dsn:         dsn,
		maxIdleConn: maxIdle,
		maxOpenConn: maxOpen,
		maxLifetime: maxLifetime,
	}
}

func (g *GormDatabase) Client() *gorm.DB {
	return g.db
}

func (g *GormDatabase) Ping(ctx context.Context) error {
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

func (g *GormDatabase) Run(ctx context.Context) error {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
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

func (g *GormDatabase) Stop(_ context.Context) error {
	if g.db == nil {
		return nil
	}
	sqlDB, err := g.db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get generic database object")
	}
	return eris.Wrap(sqlDB.Close(), "failed to close database")
}

func (g *GormDatabase) StartTransaction(ctx context.Context) (*gorm.DB, func(error) error) {
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

func (g *GormDatabase) StartTransactionWithContext(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
