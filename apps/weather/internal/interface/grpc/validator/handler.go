package validator

import (
	"context"

	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

type CityValidator interface {
	Validate(ctx context.Context, city string) (*string, error)
}

type Handler struct {
	UnimplementedCityValidationServiceServer
	validator CityValidator
}

func NewHandler(validator CityValidator) *Handler {
	return &Handler{validator: validator}
}

func (h *Handler) Validate(ctx context.Context, request *CityRequest) (*CityResponse, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}

	validatedCity, err := h.validator.Validate(ctx, request.GetCity())
	if err != nil {
		return nil, err
	}
	if validatedCity == nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")
	}

	return &CityResponse{City: *validatedCity}, nil
}
