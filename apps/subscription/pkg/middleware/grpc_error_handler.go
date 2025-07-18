package middleware

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
)

func GRPCErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		if apiErr, ok := pkgErrors.IsApiError(err); ok {
			st := status.New(toGrpcCode(apiErr.Base), apiErr.Message)
			return nil, st.Err()
		}

		st := status.New(codes.Internal, "Internal server error")
		return nil, st.Err()
	}
}

func toGrpcCode(err error) codes.Code {
	switch {
	case errors.Is(err, internalErrors.ErrNotFound):
		return codes.NotFound
	case errors.Is(err, internalErrors.ErrConflict):
		return codes.AlreadyExists
	case errors.Is(err, internalErrors.ErrInvalidInput):
		return codes.InvalidArgument
	case errors.Is(err, internalErrors.ErrInternal):
		return codes.Internal
	default:
		return codes.Unknown
	}
}
