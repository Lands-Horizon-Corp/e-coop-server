package horizon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
)

// go test -v ./services/horizon_test/horizon.schedule_test.go
func TestHorizonSchedule_CreateAndListJobs(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	err := s.CreateJob(ctx, "job1", "@every 1s", func() {})
	assert.NoError(t, err)

	jobs, err := s.ListJobs(ctx)
	assert.NoError(t, err)
	assert.Contains(t, jobs, "job1")
}

func TestHorizonSchedule_ExecuteJob(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	var executed int32 = 0
	err := s.CreateJob(ctx, "job2", "@every 1s", func() {
		atomic.StoreInt32(&executed, 1)
	})
	assert.NoError(t, err)

	err = s.ExecuteJob(ctx, "job2")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Give some time to execute
	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
}

func TestHorizonSchedule_RemoveJob(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	err := s.CreateJob(ctx, "job3", "@every 1s", func() {})
	assert.NoError(t, err)

	err = s.RemoveJob(ctx, "job3")
	assert.NoError(t, err)

	jobs, err := s.ListJobs(ctx)
	assert.NoError(t, err)
	assert.NotContains(t, jobs, "job3")
}

func TestHorizonSchedule_ExecuteJob_NotFound(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	err := s.ExecuteJob(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job ID 'nonexistent' not found")
}

func TestHorizonSchedule_RemoveJob_NotFound(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	err := s.RemoveJob(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job ID 'nonexistent' not found")
}

func TestHorizonSchedule_StartAndStop(t *testing.T) {
	s := horizon.NewHorizonSchedule()
	ctx := context.Background()

	err := s.Run(ctx)
	assert.NoError(t, err)

	err = s.Stop(ctx)
	assert.NoError(t, err)
}
