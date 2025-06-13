package scheduled

import (
	"context"
	"log"
	"sync"
	"time"

	"weather-api/internal/application/email"
	"weather-api/internal/domain/models"
	"weather-api/internal/domain/repositories"
)

type WeatherData interface {
	*models.WeatherHourly | *models.WeatherDaily
}

type WeatherEmail interface {
	*email.WeatherHourlyEmail | *email.WeatherDailyEmail
}

type EmailTask interface {
	GetSubscription() *models.Subscription
}

type WeatherJobExecutor[T WeatherData, E EmailTask] struct {
	weatherRepo      repositories.WeatherRepository
	subscriptionRepo repositories.SubscriptionRepository
	sender           email.Sender
	host             string
	workerCount      int
	frequency        models.Frequency

	getWeatherFunc func(context.Context, string) (T, error)
	createTaskFunc func(*models.Subscription, T) E
	sendEmailFunc  func(E) error
}

func NewWeatherJobExecutor[T WeatherData, E EmailTask](
	weatherRepo repositories.WeatherRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	sender email.Sender,
	host string,
	frequency models.Frequency,
	getWeatherFunc func(context.Context, string) (T, error),
	createTaskFunc func(*models.Subscription, T) E,
	sendEmailFunc func(E) error,
) *WeatherJobExecutor[T, E] {
	return &WeatherJobExecutor[T, E]{
		weatherRepo:      weatherRepo,
		subscriptionRepo: subscriptionRepo,
		sender:           sender,
		host:             host,
		workerCount:      10,
		frequency:        frequency,
		getWeatherFunc:   getWeatherFunc,
		createTaskFunc:   createTaskFunc,
		sendEmailFunc:    sendEmailFunc,
	}
}

func (e *WeatherJobExecutor[T, E]) Execute(ctx context.Context) error {
	log.Printf("Weather job started for frequency: %s", e.frequency)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	groupedSubscriptions, err := e.subscriptionRepo.FindGroupedSubscriptions(ctx, &e.frequency)
	if err != nil {
		log.Printf("Failed to fetch subscriptions: %v", err)
		return err
	}

	taskChan, errChan, wg := e.startWorkers(groupedSubscriptions)

	if err := e.dispatchTasks(ctx, groupedSubscriptions, taskChan); err != nil {
		return err
	}

	close(taskChan)
	errors := e.collectErrors(ctx, errChan, wg)

	if len(errors) > 0 {
		log.Printf("Job completed with %d errors", len(errors))
		return errors[0]
	}

	log.Printf("Weather job completed successfully for frequency: %s", e.frequency)
	return nil
}

func (e *WeatherJobExecutor[T, E]) startWorkers(
	groups []*models.GroupedSubscription,
) (chan E, chan error, *sync.WaitGroup) {
	totalSubscriptions := 0
	for _, group := range groups {
		totalSubscriptions += len(group.Subscriptions)
	}

	taskChan := make(chan E, min(totalSubscriptions, 1000))
	errChan := make(chan error, e.workerCount)

	var wg sync.WaitGroup
	for i := 0; i < e.workerCount; i++ {
		wg.Add(1)
		go e.emailWorker(context.Background(), taskChan, errChan, &wg)
	}

	return taskChan, errChan, &wg
}

func (e *WeatherJobExecutor[T, E]) dispatchTasks(
	ctx context.Context, groups []*models.GroupedSubscription, taskChan chan<- E,
) error {
	for _, group := range groups {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		weatherData, err := e.getWeatherFunc(ctx, group.City)
		if err != nil {
			log.Printf("Failed to fetch weather for city %s: %v", group.City, err)
			continue
		}

		for _, sub := range group.Subscriptions {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			taskChan <- e.createTaskFunc(sub, weatherData)
		}
	}
	return nil
}

func (e *WeatherJobExecutor[T, E]) collectErrors(
	ctx context.Context, errChan chan error, wg *sync.WaitGroup,
) []error {
	errorCollector := make(chan []error, 1)

	go func() {
		var errs []error
		for err := range errChan {
			if err != nil {
				errs = append(errs, err)
			}
		}
		errorCollector <- errs
	}()

	wg.Wait()
	close(errChan)

	select {
	case errs := <-errorCollector:
		return errs
	case <-ctx.Done():
		return []error{ctx.Err()}
	}
}

func (e *WeatherJobExecutor[T, E]) emailWorker(
	ctx context.Context,
	taskChan <-chan E,
	errChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
			err := e.sendEmailFunc(task)
			if err != nil {
				select {
				case errChan <- err:
				default:
					log.Printf("Error channel full, additional error: %v", err)
				}
			}
		}
	}
}
