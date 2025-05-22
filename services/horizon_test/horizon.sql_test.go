package horizon_test

import (
	"context"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon_test/horizon.sql_test.go

func TestGormDatabaseLifecycle(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")
	dsn := env.GetString("DATABASE_URL", "")

	if dsn == "" {
		t.Skip("TEST_DB_DSN environment variable not set")
	}

	db := horizon.NewGormDatabase(dsn, 5, 10, time.Minute)
	ctx := context.Background()

	// Start the database
	err := db.Run(ctx)
	require.NoError(t, err, "should start database successfully")

	// Ping should work
	err = db.Ping(ctx)
	require.NoError(t, err, "should ping database successfully")

	// Client should not be nil
	client := db.Client()
	require.NotNil(t, client, "gorm client should not be nil")

	// Stop the database
	err = db.Stop(ctx)
	require.NoError(t, err, "should stop database successfully")
}
