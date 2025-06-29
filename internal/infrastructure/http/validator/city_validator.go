package validator

type Client interface {
	Validate(city string) (*string, error)
}

type CityValidator struct {
	client Client
}

func NewCityValidator(provider Client) *CityValidator {
	return &CityValidator{client: provider}
}

func (c CityValidator) Validate(city string) (*string, error) {
	return c.client.Validate(city)
}
