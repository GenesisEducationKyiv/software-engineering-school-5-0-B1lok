package rabbitmq

func GetAppQueueConfigs() []QueueConfig {
	return []QueueConfig{
		defaultQueue("user_subscribed"),
		defaultQueue("weather_updated"),
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
