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

	msg := st.Message()

	switch st.Code() {
	case codes.InvalidArgument:
		return pkgErrors.New(internalErrors.ErrInvalidInput, msg)
	case codes.NotFound:
		return pkgErrors.New(internalErrors.ErrNotFound, msg)
	case codes.AlreadyExists:
		return pkgErrors.New(internalErrors.ErrConflict, msg)
	default:
		return pkgErrors.New(internalErrors.ErrInvalidInput, msg)
	}
}
