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

// Example PostgreSQL DSN
// You can also build this from environment variables:
// host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable TimeZone=UTC
dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable TimeZone=UTC"

// Configure connection pool: 10 idle, 50 open, 30 minutes conn max lifetime
database := NewGormDatabase(dsn, 10, 50, 30*time.Minute)

// Start the database connection and pool
if err := database.Start(ctx); err != nil {
	panic(fmt.Errorf("DB start error: %w", err))
}
*/

// SQLDatabase defines the interface for PostgreSQL operations
type SQLDatabaseService interface {
	// Start initializes the connection pool with the database
	Run(ctx context.Context) error

	// Stop closes all database connections
	Stop(ctx context.Context) error

	// Client returns the active GORM database client
	Client() *gorm.DB

	// Ping checks if the database is reachable
	Ping(ctx context.Context) error
}

type GormDatabase struct {
	dsn         string
	db          *gorm.DB
	maxIdleConn int
	maxOpenConn int
	maxLifetime time.Duration
}

// NewGormDatabase constructs a new GormDatabase
func NewGormDatabase(dsn string, maxIdle, maxOpen int, maxLifetime time.Duration) SQLDatabaseService {
	return &GormDatabase{
		dsn:         dsn,
		maxIdleConn: maxIdle,
		maxOpenConn: maxOpen,
		maxLifetime: maxLifetime,
	}
}

// Client implements SQLDatabase.
func (g *GormDatabase) Client() *gorm.DB {
	return g.db
}

// Ping implements SQLDatabase.
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

// Start initializes the GORM connection pool
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

	// Ping to verify connectivity
	if err := sqlDB.PingContext(ctx); err != nil {
		return eris.Wrap(err, "database ping failed")
	}

	g.db = db
	return nil
}

// Stop implements SQLDatabase.
func (g *GormDatabase) Stop(ctx context.Context) error {
	if g.db == nil {
		return nil
	}
	sqlDB, err := g.db.DB()
	if err != nil {
		return eris.Wrap(err, "failed to get generic database object")
	}
	return eris.Wrap(sqlDB.Close(), "failed to close database")
}
