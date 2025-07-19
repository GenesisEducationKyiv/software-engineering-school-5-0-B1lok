package scheduled

import (
	"context"
)

type DailyWeatherUpdateJob struct {
	executor *WeatherJobExecutor
}

func NewDailyWeatherUpdateJob(
	subscriptionRepo GroupedSubscriptionReader,
	dispatcher EventDispatcher,
) *DailyWeatherUpdateJob {
	exec := NewWeatherJobExecutor(
		subscriptionRepo,
		"daily",
		dispatcher,
	)

	return &DailyWeatherUpdateJob{
		executor: exec,
	}
}

func (d *DailyWeatherUpdateJob) Name() string {
	return "DailyWeatherUpdateJob"
}

func (d *DailyWeatherUpdateJob) Schedule() string {
	return "0 0 8 * * *"
}

func (d *DailyWeatherUpdateJob) Run(ctx context.Context) error {
	return d.executor.Execute(ctx)
}
