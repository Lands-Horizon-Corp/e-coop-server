package horizon

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rotisserie/eris"
)

// Scheduler defines the interface for scheduling and managing cron jobs
type SchedulerService interface {
	// Start initializes the scheduler and in-memory job store
	Run(ctx context.Context) error

	// Stop gracefully shuts down the scheduler and clears all timers
	Stop(ctx context.Context) error

	// CreateJob registers a new cron-style job with the specified schedule
	CreateJob(ctx context.Context, jobID string, schedule string, task func()) error

	// ExecuteJob runs a job immediately
	ExecuteJob(ctx context.Context, jobID string) error

	// RemoveJob deletes a job by its ID from the scheduler
	RemoveJob(ctx context.Context, jobID string) error

	// ListJobs returns all registered job IDs
	ListJobs(ctx context.Context) ([]string, error)
}

type job struct {
	entryID  cron.EntryID
	schedule string
	task     func()
}

type HorizonSchedule struct {
	cron  *cron.Cron
	jobs  map[string]job
	mutex sync.Mutex
}

func NewHorizonSchedule() SchedulerService {
	return &HorizonSchedule{
		cron: cron.New(),
		jobs: make(map[string]job),
	}
}

// CreateJob implements Scheduler.
func (h *HorizonSchedule) CreateJob(ctx context.Context, jobID string, schedule string, task func()) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if _, exists := h.jobs[jobID]; exists {
		return nil // Job already exists
	}
	entryID, err := h.cron.AddFunc(schedule, task)
	if err != nil {
		return err
	}
	h.jobs[jobID] = job{entryID: entryID, task: task, schedule: schedule}
	return nil
}

// ListJobs implements Scheduler.
func (h *HorizonSchedule) ListJobs(ctx context.Context) ([]string, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	jobs := make([]string, 0, len(h.jobs))
	for jobID := range h.jobs {
		jobs = append(jobs, jobID)
	}
	return jobs, nil

}

// RemoveJob implements Scheduler.
func (h *HorizonSchedule) RemoveJob(ctx context.Context, jobID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	job, exists := h.jobs[jobID]
	if !exists {
		return eris.Errorf("failed to remove job: job ID '%s' not found", jobID)
	}
	h.cron.Remove(job.entryID)
	delete(h.jobs, jobID)
	return nil
}

// ExecuteJob implements Scheduler.
func (h *HorizonSchedule) ExecuteJob(ctx context.Context, jobID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	job, exists := h.jobs[jobID]
	if !exists {
		return eris.Errorf("failed to execute job: job ID '%s' not found", jobID)
	}
	job.task()
	return nil
}

// Start implements Scheduler.
func (h *HorizonSchedule) Run(ctx context.Context) error {
	h.cron.Start()
	return nil
}

// Stop implements Scheduler.
func (h *HorizonSchedule) Stop(ctx context.Context) error {
	h.cron.Stop()
	return nil
}
