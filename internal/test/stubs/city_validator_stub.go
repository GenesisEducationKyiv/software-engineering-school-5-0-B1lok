package stubs

import "context"

type CityValidatorStub struct {
	ValidateFn func(city string) (*string, error)
}

func NewCityValidatorStub() *CityValidatorStub {
	return &CityValidatorStub{
		ValidateFn: nil,
	}
}

func (s *CityValidatorStub) Validate(ctx context.Context, city string) (*string, error) {
	if s.ValidateFn != nil {
		return s.ValidateFn(city)
	}
	return &city, nil
}
