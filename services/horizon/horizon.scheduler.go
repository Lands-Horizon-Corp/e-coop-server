package horizon

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rotisserie/eris"
)

type SchedulerService interface {
	Run(ctx context.Context) error

	Stop(ctx context.Context) error

	CreateJob(ctx context.Context, jobID string, schedule string, task func()) error

	ExecuteJob(ctx context.Context, jobID string) error

	RemoveJob(ctx context.Context, jobID string) error

	ListJobs(ctx context.Context) ([]string, error)
}

type job struct {
	entryID  cron.EntryID
	schedule string
	task     func()
}

type Schedule struct {
	cron  *cron.Cron
	jobs  map[string]job
	mutex sync.Mutex
}

func NewSchedule() SchedulerService {
	return &Schedule{
		cron: cron.New(),
		jobs: make(map[string]job),
	}
}

func (h *Schedule) CreateJob(_ context.Context, jobID string, schedule string, task func()) error {
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

func (h *Schedule) ListJobs(_ context.Context) ([]string, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	jobs := make([]string, 0, len(h.jobs))
	for jobID := range h.jobs {
		jobs = append(jobs, jobID)
	}
	return jobs, nil

}

func (h *Schedule) RemoveJob(_ context.Context, jobID string) error {
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

func (h *Schedule) ExecuteJob(_ context.Context, jobID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	job, exists := h.jobs[jobID]
	if !exists {
		return eris.Errorf("failed to execute job: job ID '%s' not found", jobID)
	}
	job.task()
	return nil
}

func (h *Schedule) Run(_ context.Context) error {
	h.cron.Start()
	return nil
}

func (h *Schedule) Stop(_ context.Context) error {
	h.cron.Stop()
	return nil
}
