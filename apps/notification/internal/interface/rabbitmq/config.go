package rabbitmq

import "notification/internal/application/event"

func GetAppQueueConfigs() []QueueConfig {
	return []QueueConfig{
		defaultQueue(string(event.UserSubscribedEventName)),
		defaultQueue(string(event.WeatherUpdatedEventName)),
	}
}

func defaultQueue(name string) QueueConfig {
	return QueueConfig{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}
