package horizon

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rotisserie/eris"
)

type job struct {
	entryID  cron.EntryID
	schedule string
	task     func()
}

type ScheduleImpl struct {
	cron  *cron.Cron
	jobs  map[string]job
	mutex sync.Mutex
	cache *CacheImpl
}

func NewScheduleImpl(cache *CacheImpl) *ScheduleImpl {
	return &ScheduleImpl{
		cron:  cron.New(),
		jobs:  make(map[string]job),
		cache: cache,
	}
}

func (h *ScheduleImpl) Run() error {
	h.cron.Start()
	return nil
}

func (h *ScheduleImpl) Stop() error {
	h.cron.Stop()
	return nil
}

func (h *ScheduleImpl) CreateJob(jobID string, schedule string, task func()) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if _, exists := h.jobs[jobID]; exists {
		return nil
	}
	entryID, err := h.cron.AddFunc(schedule, task)
	if err != nil {
		return err
	}
	h.jobs[jobID] = job{entryID: entryID, task: task, schedule: schedule}
	return nil
}

func (h *ScheduleImpl) ListJobs() ([]string, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	jobs := make([]string, 0, len(h.jobs))
	for jobID := range h.jobs {
		jobs = append(jobs, jobID)
	}
	return jobs, nil

}

func (h *ScheduleImpl) RemoveJob(jobID string) error {
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

func (h *ScheduleImpl) ExecuteJob(jobID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	job, exists := h.jobs[jobID]
	if !exists {
		return eris.Errorf("failed to execute job: job ID '%s' not found", jobID)
	}
	job.task()
	return nil
}
