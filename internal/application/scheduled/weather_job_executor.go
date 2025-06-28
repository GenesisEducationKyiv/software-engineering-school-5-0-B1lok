package scheduled

import (
	"context"
	"log"
	"sync"
	"time"

	"weather-api/internal/domain"
)

type GroupedSubscriptionReader interface {
	FindGroupedSubscriptions(
		ctx context.Context, frequency *domain.Frequency,
	) ([]*domain.GroupedSubscription, error)
}

type WeatherData interface {
	*domain.WeatherHourly | *domain.WeatherDaily
}

type Task[T WeatherData] struct {
	Subscription *domain.Subscription
	WeatherData  T
}

type WeatherJobExecutor[T WeatherData] struct {
	subscriptionRepo GroupedSubscriptionReader
	workerCount      int
	frequency        domain.Frequency
	getWeatherFunc   func(string) (T, error)
	notifyFunc       func(*domain.Subscription, T) error
}

func NewWeatherJobExecutor[T WeatherData](
	subscriptionRepo GroupedSubscriptionReader,
	frequency domain.Frequency,
	getWeatherFunc func(string) (T, error),
	notifyFunc func(*domain.Subscription, T) error,
) *WeatherJobExecutor[T] {
	return &WeatherJobExecutor[T]{
		subscriptionRepo: subscriptionRepo,
		workerCount:      10,
		frequency:        frequency,
		getWeatherFunc:   getWeatherFunc,
		notifyFunc:       notifyFunc,
	}
}

func (e *WeatherJobExecutor[T]) Execute(ctx context.Context) error {
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

func (e *WeatherJobExecutor[T]) startWorkers(
	groups []*domain.GroupedSubscription,
) (chan Task[T], chan error, *sync.WaitGroup) {
	totalSubscriptions := 0
	for _, group := range groups {
		totalSubscriptions += len(group.Subscriptions)
	}

	taskChan := make(chan Task[T], min(totalSubscriptions, 1000))
	errChan := make(chan error, e.workerCount)

	var wg sync.WaitGroup
	for i := 0; i < e.workerCount; i++ {
		wg.Add(1)
		go e.worker(context.Background(), taskChan, errChan, &wg)
	}

	return taskChan, errChan, &wg
}

func (e *WeatherJobExecutor[T]) dispatchTasks(
	ctx context.Context, groups []*domain.GroupedSubscription, taskChan chan<- Task[T],
) error {
	for _, group := range groups {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		weatherData, err := e.getWeatherFunc(group.City)
		if err != nil {
			log.Printf("Failed to fetch weather for city %s: %v", group.City, err)
			continue
		}

		for _, sub := range group.Subscriptions {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			taskChan <- Task[T]{
				Subscription: sub,
				WeatherData:  weatherData,
			}
		}
	}
	return nil
}

func (e *WeatherJobExecutor[T]) collectErrors(
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

func (e *WeatherJobExecutor[T]) worker(
	ctx context.Context,
	taskChan <-chan Task[T],
	errChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
			err := e.notifyFunc(task.Subscription, task.WeatherData)
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
