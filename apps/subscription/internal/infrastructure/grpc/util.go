package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
)

func MapGrpcErrorToDomain(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return pkgErrors.New(
			internalErrors.ErrInvalidInput, err.Error(),
		)
	}

	if st.Code() == codes.InvalidArgument {
		if validationErr := extractFieldViolationError(st); validationErr != nil {
			return validationErr
		}
	}

	errorMapping := map[codes.Code]error{
		codes.InvalidArgument: internalErrors.ErrInvalidInput,
		codes.NotFound:        internalErrors.ErrNotFound,
		codes.AlreadyExists:   internalErrors.ErrConflict,
	}

	mappedError, exists := errorMapping[st.Code()]
	if !exists {
		mappedError = internalErrors.ErrInvalidInput
	}

	return pkgErrors.New(mappedError, st.Message())
}

func extractFieldViolationError(st *status.Status) error {
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.BadRequest); ok {
			for _, violation := range info.FieldViolations {
				if violation.Field == "city" {
					return pkgErrors.New(internalErrors.ErrInvalidInput, violation.Description)
				}
			}
		}
	}
	return nil
}
