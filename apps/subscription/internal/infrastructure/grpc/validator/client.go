package validator

import (
	"context"

	"subscription-service/internal/infrastructure/grpc"
)

type CityValidator struct {
	client CityValidationServiceClient
}

func NewCityValidator(client CityValidationServiceClient) *CityValidator {
	return &CityValidator{client: client}
}

func (v *CityValidator) Validate(ctx context.Context, city string) (*string, error) {
	resp, err := v.client.Validate(ctx, &CityRequest{City: city})
	if err != nil {
		return nil, grpc.MapGrpcErrorToDomain(err)
	}
	return &resp.City, nil
}
