package scheduled

import (
	"context"
)

type HourlyWeatherUpdateJob struct {
	executor *WeatherJobExecutor
}

func NewHourlyWeatherUpdateJob(
	subscriptionRepo GroupedSubscriptionReader,
	dispatcher EventDispatcher,
) *HourlyWeatherUpdateJob {
	exec := NewWeatherJobExecutor(
		subscriptionRepo,
		"hourly",
		dispatcher,
	)

	return &HourlyWeatherUpdateJob{
		executor: exec,
	}
}

func (h *HourlyWeatherUpdateJob) Name() string {
	return "HourlyWeatherUpdateJob"
}

func (h *HourlyWeatherUpdateJob) Schedule() string {
	return "*/15 * * * * *"
}

func (h *HourlyWeatherUpdateJob) Run(ctx context.Context) error {
	return h.executor.Execute(ctx)
}
