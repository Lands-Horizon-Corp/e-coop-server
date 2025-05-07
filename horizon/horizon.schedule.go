package horizon

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type job struct {
	entryID  cron.EntryID
	schedule string
	task     func() error
}

type HorizonSchedule struct {
	log   *HorizonLog
	cron  *cron.Cron
	jobs  map[string]job
	mutex sync.Mutex
}

func NewHorizonSchedule(log *HorizonLog) (*HorizonSchedule, error) {
	return &HorizonSchedule{
		log:  log,
		cron: cron.New(),
		jobs: make(map[string]job),
	}, nil
}

func (hs *HorizonSchedule) Run() error {
	hs.cron.Start()
	return nil
}

func (hs *HorizonSchedule) Stop() error {
	hs.cron.Stop()
	hs.log.Log(LogEntry{
		Category: CategorySchedule,
		Level:    LevelInfo,
		Message:  "Scheduler stopped",
		Fields: []zap.Field{
			zap.String("action", "stop"),
		},
	})
	return nil
}

func (hs *HorizonSchedule) Create(jobID, schedule string, task func() error) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	// Check if the job already exists
	if _, exists := hs.jobs[jobID]; exists {
		err := eris.New(fmt.Sprintf("job with ID %s already exists", jobID))
		hs.log.Log(LogEntry{
			Category: CategorySchedule,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("job id", jobID),
				zap.String("schedule", schedule),
				zap.Error(err),
			},
		})
		return err
	}

	entryID, err := hs.cron.AddFunc(schedule, func() {
		hs.log.Log(LogEntry{
			Category: CategorySchedule,
			Level:    LevelInfo,
			Message:  fmt.Sprintf("Job %s started", jobID),
			Fields: []zap.Field{
				zap.String("job id", jobID),
				zap.String("action", "start"),
				zap.String("schedule", schedule),
			},
		})

		start := time.Now()
		defer func() {
			hs.log.Log(LogEntry{
				Category: CategorySchedule,
				Level:    LevelInfo,
				Message:  fmt.Sprintf("Job %s finished in %v", jobID, time.Since(start)),
				Fields: []zap.Field{
					zap.String("job id", jobID),
					zap.String("action", "end"),
					zap.String("schedule", schedule),
				},
			})

		}()

		if err := task(); err != nil {
			hs.log.Log(LogEntry{
				Category: CategorySchedule,
				Level:    LevelError,
				Message:  fmt.Sprintf("Job %s failed", jobID),
				Fields: []zap.Field{
					zap.String("job id", jobID),
					zap.String("action", "fail"),
					zap.String("schedule", schedule),
					zap.Error(err),
				},
			})
		}
	})

	if err != nil {
		err = eris.Wrapf(err, "failed to create job %s", jobID)
		hs.log.Log(LogEntry{
			Category: CategorySchedule,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("job id", jobID),
				zap.String("schedule", schedule),
				zap.Error(err),
			},
		})
		return err
	}

	hs.jobs[jobID] = job{entryID: entryID, task: task, schedule: schedule}
	hs.log.Log(LogEntry{
		Category: CategorySchedule,
		Level:    LevelInfo,
		Message:  fmt.Sprintf("Job %s scheduled with schedule %s", jobID, schedule),
		Fields: []zap.Field{
			zap.String("job id", jobID),
			zap.String("schedule", schedule),
		},
	})

	return nil
}

func (hs *HorizonSchedule) Execute(jobID string) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	job, exists := hs.jobs[jobID]
	if !exists {
		err := eris.New(fmt.Sprintf("job %s not found", jobID))
		hs.log.Log(LogEntry{
			Category: CategorySchedule,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("job id", jobID),
				zap.String("schedule", job.schedule),
				zap.Error(err),
			},
		})
		return err
	}

	go job.task()
	hs.log.Log(LogEntry{
		Category: CategorySchedule,
		Level:    LevelInfo,
		Message:  fmt.Sprintf("Job %s executed manually", jobID),
		Fields: []zap.Field{
			zap.String("job id", jobID),
			zap.String("schedule", job.schedule),
		},
	})
	return nil
}

func (hs *HorizonSchedule) Remove(jobID string) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	job, exists := hs.jobs[jobID]
	if !exists {
		err := eris.New(fmt.Sprintf("job %s not found", jobID))
		hs.log.Log(LogEntry{
			Category: CategorySchedule,
			Level:    LevelError,
			Message:  err.Error(),
			Fields: []zap.Field{
				zap.String("job id", jobID),
				zap.String("schedule", job.schedule),
				zap.Error(err),
			},
		})
		return err
	}

	hs.cron.Remove(job.entryID)
	delete(hs.jobs, jobID)
	hs.log.Log(LogEntry{
		Category: CategorySchedule,
		Level:    LevelInfo,
		Message:  fmt.Sprintf("Job %s removed", jobID),
		Fields: []zap.Field{
			zap.String("job id", jobID),
			zap.String("schedule", job.schedule),
			zap.String("action", "remove"),
		},
	})
	return nil
}

func (hs *HorizonSchedule) ListJobs() []string {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	jobs := make([]string, 0, len(hs.jobs))
	schedules := make([]string, 0, len(hs.jobs))
	for jobID, value := range hs.jobs {
		jobs = append(jobs, jobID)
		schedules = append(schedules, fmt.Sprintf("%s - %s ", jobID, value.schedule))
	}

	hs.log.Log(LogEntry{
		Category: CategorySchedule,
		Level:    LevelInfo,
		Message:  fmt.Sprintf("Listed jobs: %v", jobs),
		Fields: []zap.Field{
			zap.Strings("job ids", jobs),
			zap.Strings("schedule", schedules),
		},
	})
	return jobs
}
