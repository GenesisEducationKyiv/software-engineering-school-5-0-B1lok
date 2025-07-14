package publisher

import "notification/internal/rabbitmq/consumer"

const (
	confirmationTemplateName = "confirm.html"
	dailyUpdateTemplateName  = "daily.html"
	hourlyUpdateTemplateName = "hourly.html"
)

func toConfirmationEmailMessage(
	message *consumer.ConfirmationEmailMessage,
) *ConfirmationEmailMessage {
	return &ConfirmationEmailMessage{
		BaseEmailMessage: BaseEmailMessage{
			To:           message.To,
			TemplateName: confirmationTemplateName,
		},
		TemplateData: ConfirmationEmailTemplate{
			City:      message.City,
			Frequency: message.Frequency,
			URL:       message.URL,
		},
	}
}

func toDailyUpdateMessage(
	message *consumer.WeatherUpdateMessage, template *DailyUpdateTemplate,
) *DailyUpdateMessage {
	return &DailyUpdateMessage{
		BaseEmailMessage: BaseEmailMessage{
			To:           message.To,
			TemplateName: dailyUpdateTemplateName,
		},
		TemplateData: WeatherTemplateData{
			Weather:        *template,
			Frequency:      message.Frequency,
			UnsubscribeURL: message.UnsubscribeURL,
		},
	}
}

func toHourlyUpdateMessage(
	message *consumer.WeatherUpdateMessage, template *HourlyUpdateTemplate,
) *HourlyUpdateMessage {
	return &HourlyUpdateMessage{
		BaseEmailMessage: BaseEmailMessage{
			To:           message.To,
			TemplateName: hourlyUpdateTemplateName,
		},
		TemplateData: WeatherTemplateData{
			Weather:        *template,
			Frequency:      message.Frequency,
			UnsubscribeURL: message.UnsubscribeURL,
		},
	}
}
