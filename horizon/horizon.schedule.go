// horizon/scheduler.go
package horizon

type HorizonSchedule struct {
	log *HorizonLog
}

func NewHorizonSchedule(log *HorizonLog) (*HorizonSchedule, error) {

	return &HorizonSchedule{log: log}, nil
}

func (hs *HorizonSchedule) Start() {

}
func (hs *HorizonSchedule) Stop()    {}
func (hs *HorizonSchedule) Create()  {}
func (hs *HorizonSchedule) Execute() {}
func (hs *HorizonSchedule) Remove()  {}

// // NewHorizonSchedule creates a new scheduler instance.
// func NewHorizonSchedule(log *HorizonLog) (*HorizonSchedule, error) {
// 	scheduler, err := gocron.NewScheduler()
// 	if err != nil {
// 		return nil, eris.Wrap(err, "failed to create scheduler")
// 	}
// 	return &HorizonSchedule{scheduler: &scheduler, log: log}, nil
// }

// // Start begins running scheduled jobs asynchronously.
// func (hs *HorizonSchedule) Start() {
// 	hs.scheduler.Start()
// }

// // Stop shuts down the scheduler gracefully.
// func (hs *HorizonSchedule) Stop() error {
// 	return hs.scheduler.Shutdown()
// }

// // Generate registers and schedules a new task under the given key tag.
// // The task func returns an error, which will be logged on failure.
// func (hs *HorizonSchedule) Generate(key string, interval time.Duration, task func() error) {
// 	if hs.existsTag(key) {
// 		hs.log.Log(LogEntry{
// 			Category: CategorySchedule,
// 			Level:    LevelWarn,
// 			Message:  fmt.Sprintf("Job already exists: %s", key),
// 			Fields:   []zap.Field{zap.String("key", key)},
// 		})
// 		return
// 	}

// 	wrapped := func() {
// 		start := time.Now()
// 		hs.log.Log(LogEntry{
// 			Category: CategorySchedule,
// 			Level:    LevelDebug,
// 			Message:  fmt.Sprintf("Starting job: %s", key),
// 			Fields: []zap.Field{
// 				zap.String("key", key),
// 				zap.Time("start_time", start),
// 			},
// 		})

// 		if err := task(); err != nil {
// 			hs.log.Log(LogEntry{
// 				Category: CategorySchedule,
// 				Level:    LevelError,
// 				Message:  fmt.Sprintf("Job %s failed", key),
// 				Fields:   []zap.Field{zap.String("key", key), zap.Error(err)},
// 			})
// 		} else {
// 			end := time.Now()
// 			hs.log.Log(LogEntry{
// 				Category: CategorySchedule,
// 				Level:    LevelDebug,
// 				Message:  fmt.Sprintf("Completed job: %s", key),
// 				Fields: []zap.Field{
// 					zap.String("key", key),
// 					zap.Time("end_time", end),
// 					zap.Duration("execution_duration", end.Sub(start)),
// 				},
// 			})
// 		}
// 	}

// 	// schedule & tag
// 	_, err := hs.scheduler.
// 		Every(interval).
// 		Tag(key).
// 		Do(wrapped)
// 	if err != nil {
// 		hs.log.Log(LogEntry{
// 			Category: CategorySchedule,
// 			Level:    LevelError,
// 			Message:  fmt.Sprintf("Failed to schedule job: %s", key),
// 			Fields:   []zap.Field{zap.String("key", key), zap.Error(err)},
// 		})
// 		return
// 	}

// 	hs.log.Log(LogEntry{
// 		Category: CategorySchedule,
// 		Level:    LevelInfo,
// 		Message:  fmt.Sprintf("Scheduled job: %s every %s", key, interval),
// 		Fields:   []zap.Field{zap.String("key", key), zap.Duration("interval", interval)},
// 	})
// }

// // Run triggers the job tagged with key immediately.
// func (hs *HorizonSchedule) Run(key string) {
// 	jobs := hs.findJobsByTag(key)
// 	if len(jobs) == 0 {
// 		hs.log.Log(LogEntry{
// 			Category: CategorySchedule, Level: LevelWarn,
// 			Message: fmt.Sprintf("No job to run: %s", key),
// 			Fields:  []zap.Field{zap.String("key", key)},
// 		})
// 		return
// 	}

// 	for _, job := range jobs {
// 		go job.RunNow()
// 	}

// 	hs.log.Log(LogEntry{
// 		Category: CategorySchedule, Level: LevelInfo,
// 		Message: fmt.Sprintf("Manually triggered: %s", key),
// 		Fields:  []zap.Field{zap.String("key", key)},
// 	})
// }

// // Remove unschedules and removes all jobs tagged with key.
// func (hs *HorizonSchedule) Remove(key string) {
// 	if !hs.existsTag(key) {
// 		hs.log.Log(LogEntry{
// 			Category: CategorySchedule,
// 			Level:    LevelWarn,
// 			Message:  fmt.Sprintf("No job to remove: %s", key),
// 			Fields:   []zap.Field{zap.String("key", key)},
// 		})
// 		return
// 	}

// 	hs.scheduler.RemoveByTags(key)
// 	hs.log.Log(LogEntry{
// 		Category: CategorySchedule,
// 		Level:    LevelInfo,
// 		Message:  fmt.Sprintf("Removed job: %s", key),
// 		Fields:   []zap.Field{zap.String("key", key)},
// 	})
// }

// // ListJobs returns all tags in use (i.e. the job keys).
// func (hs *HorizonSchedule) ListJobs() []string {
// 	var keys []string
// 	for _, job := range hs.scheduler.Jobs() {
// 		keys = append(keys, job.Tags()...)
// 	}
// 	return keys
// }

// func (hs *HorizonSchedule) existsTag(key string) bool {
// 	for _, job := range hs.scheduler.Jobs() {
// 		for _, tag := range job.Tags() {
// 			if tag == key {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func (hs *HorizonSchedule) findJobsByTag(key string) []gocron.Job {
// 	var matches []gocron.Job
// 	for _, job := range hs.scheduler.Jobs() {
// 		for _, tag := range job.Tags() {
// 			if tag == key {
// 				matches = append(matches, job)
// 				break
// 			}
// 		}
// 	}
// 	return matches
// }
