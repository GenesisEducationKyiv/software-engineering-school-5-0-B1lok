package validator

import "context"

type Client interface {
	Validate(ctx context.Context, city string) (*string, error)
}

type CityValidator struct {
	client Client
}

func NewCityValidator(provider Client) *CityValidator {
	return &CityValidator{client: provider}
}

func (c CityValidator) Validate(ctx context.Context, city string) (*string, error) {
	return c.client.Validate(ctx, city)
}
