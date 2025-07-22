//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"notification/internal/application/event"
	infevent "notification/internal/infrastructure/event"
	"notification/internal/infrastructure/grpc/weather"
	"notification/internal/interface/rabbitmq"
	"notification/internal/test/containers"
	"notification/internal/test/stub"
	"testing"
	"time"
)

type ConsumerTestSuite struct {
	suite.Suite
	Rabbit *containers.RabbitMQContainer
}

type weatherUpdatedPayload struct {
	Email          string `json:"email"`
	City           string `json:"city"`
	Frequency      string `json:"frequency"`
	UnsubscribeURL string `json:"unsubscribe_url"`
}

func (suite *ConsumerTestSuite) SetupSuite() {
	ctx := context.Background()
	rabbit, err := containers.SetupRabbitMQContainer(ctx)
	suite.Require().NoError(err)
	suite.Rabbit = rabbit
}

func (suite *ConsumerTestSuite) TearDownSuite() {
	if suite.Rabbit != nil {
		ctx := context.Background()
		suite.Rabbit.Cleanup(ctx)
	}
}

func (suite *ConsumerTestSuite) SetupTest() {
	suite.Rabbit.Channel.QueueDelete(
		string(event.WeatherUpdatedEventName),
		false,
		false,
		false,
	)
}

func (suite *ConsumerTestSuite) TestWeatherUpdateFlow() {
	ctx := context.Background()

	consumer := rabbitmq.NewConsumer(suite.Rabbit.Channel)
	dispatcher := event.NewDispatcher(consumer)

	weatherClient := &stub.WeatherClient{
		DailyUpdateFunc: func(ctx context.Context, city string) (*weather.WeatherDaily, error) {
			suite.Equal("Kyiv", city)
			return &weather.WeatherDaily{
				Location: city,
			}, nil
		},
	}

	var sentEmail struct {
		templateName string
		to           string
		subject      string
		data         interface{}
		called       bool
	}

	sender := &stub.Sender{
		SendFunc: func(templateName, to, subject string, data any) error {
			sentEmail.templateName = templateName
			sentEmail.to = to
			sentEmail.subject = subject
			sentEmail.data = data
			sentEmail.called = true
			return nil
		},
	}

	handler := infevent.NewWeatherUpdateHandler(weatherClient, sender)

	err := suite.Rabbit.DeclareQueue(string(event.WeatherUpdatedEventName))
	suite.Require().NoError(err)

	err = dispatcher.Register(ctx, handler)
	suite.Require().NoError(err)

	testPayload := weatherUpdatedPayload{
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		UnsubscribeURL: "https://example.com/unsubscribe/123",
	}

	payloadBytes, err := json.Marshal(testPayload)
	suite.Require().NoError(err)

	err = suite.Rabbit.PublishMessage(ctx, string(event.WeatherUpdatedEventName), payloadBytes)
	suite.Require().NoError(err)

	suite.Eventually(func() bool {
		return sentEmail.called
	}, 10*time.Second, 100*time.Millisecond, "Email should be sent")

	suite.True(sentEmail.called, "Email sender should be called")
	suite.Equal("daily.html", sentEmail.templateName)
	suite.Equal("test@example.com", sentEmail.to)
	suite.Equal("Your weather daily forecast", sentEmail.subject)
}

func (suite *ConsumerTestSuite) TestWeatherUpdate_InvalidPayload_ShouldNotCallSender() {
	ctx := context.Background()

	consumer := rabbitmq.NewConsumer(suite.Rabbit.Channel)
	dispatcher := event.NewDispatcher(consumer)

	weatherClient := &stub.WeatherClient{}

	sender := &stub.Sender{
		SendFunc: func(templateName, to, subject string, data any) error {
			suite.Fail("Sender should not be called with invalid payload")
			return nil
		},
	}

	handler := infevent.NewWeatherUpdateHandler(weatherClient, sender)

	err := suite.Rabbit.DeclareQueue(string(event.WeatherUpdatedEventName))
	suite.Require().NoError(err)

	err = dispatcher.Register(ctx, handler)
	suite.Require().NoError(err)

	payload := []byte(`{"email": "test@example.com", "city": 123}`)

	err = suite.Rabbit.PublishMessage(ctx, string(event.WeatherUpdatedEventName), payload)
	suite.Require().NoError(err)
}

func (suite *ConsumerTestSuite) TestWeatherUpdate_WeatherClientError_ShouldNotSendEmail() {
	ctx := context.Background()

	consumer := rabbitmq.NewConsumer(suite.Rabbit.Channel)
	dispatcher := event.NewDispatcher(consumer)

	weatherClient := &stub.WeatherClient{
		DailyUpdateFunc: func(ctx context.Context, city string) (*weather.WeatherDaily, error) {
			return nil, assert.AnError
		},
	}

	sender := &stub.Sender{
		SendFunc: func(templateName, to, subject string, data any) error {
			suite.Fail("Sender should not be called when weatherClient fails")
			return nil
		},
	}

	handler := infevent.NewWeatherUpdateHandler(weatherClient, sender)

	err := suite.Rabbit.DeclareQueue(string(event.WeatherUpdatedEventName))
	suite.Require().NoError(err)

	err = dispatcher.Register(ctx, handler)
	suite.Require().NoError(err)

	testPayload := weatherUpdatedPayload{
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		UnsubscribeURL: "https://example.com/unsubscribe/123",
	}

	payloadBytes, err := json.Marshal(testPayload)
	suite.Require().NoError(err)

	err = suite.Rabbit.PublishMessage(ctx, string(event.WeatherUpdatedEventName), payloadBytes)
	suite.Require().NoError(err)

}

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}
