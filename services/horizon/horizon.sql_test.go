package horizon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)


func TestGormDatabaseLifecycle(t *testing.T) {
	env := NewEnvironmentService("../../.env")
	dsn := env.GetString("DATABASE_URL", "")

	if dsn == "" {
		t.Skip("TEST_DB_DSN environment variable not set")
	}

	db := NewGormDatabase(dsn, 5, 10, time.Minute)
	ctx := context.Background()

	err := db.Run(ctx)
	require.NoError(t, err, "should start database successfully")

	err = db.Ping(ctx)
	require.NoError(t, err, "should ping database successfully")

	client := db.Client()
	require.NotNil(t, client, "gorm client should not be nil")

	err = db.Stop(ctx)
	require.NoError(t, err, "should stop database successfully")
}
