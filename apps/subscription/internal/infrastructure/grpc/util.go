package grpc

import (
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
