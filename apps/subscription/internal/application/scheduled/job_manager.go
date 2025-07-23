package scheduled

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/robfig/cron/v3"
)

type JobManager struct {
	cron    *cron.Cron
	jobs    []Job
	context context.Context
}

func NewJobManager(ctx context.Context) *JobManager {
	return &JobManager{
		cron:    cron.New(cron.WithSeconds()),
		jobs:    []Job{},
		context: ctx,
	}
}

func (jm *JobManager) RegisterJob(job Job) {
	jm.jobs = append(jm.jobs, job)
}

func (jm *JobManager) StartScheduler() {
	for _, job := range jm.jobs {
		schedule := job.Schedule()
		if _, err := jm.cron.AddFunc(schedule, func() {
			if err := job.Run(jm.context); err != nil {
				log.Error().
					Str("job", job.Name()).
					Err(err).
					Msg("Error in job")
			} else {
				log.Info().
					Str("job", job.Name()).
					Msg("Job executed successfully")
			}
		}); err != nil {
			log.Error().
				Str("job", job.Name()).
				Err(err).
				Msg("Failed to schedule job")
		}
	}
	jm.cron.Start()
}
