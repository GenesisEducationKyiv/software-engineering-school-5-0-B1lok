package consumer

type ConfirmationEmailMessage struct {
	To        string `json:"to"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	URL       string `json:"url"`
}

type WeatherUpdateMessage struct {
	To             string `json:"to"`
	City           string `json:"city"`
	Frequency      string `json:"frequency"`
	UnsubscribeURL string `json:"unsubscribe_url"`
}
