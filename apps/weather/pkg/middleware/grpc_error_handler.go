package middleware

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

func GrpcErrorInterceptor() grpc.UnaryServerInterceptor {
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
			grpcCode := toGrpcCode(apiErr.Base)
			st := status.New(toGrpcCode(apiErr.Base), apiErr.Message)
			if grpcCode == codes.InvalidArgument {
				badReq := &errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{
							Field:       "city",
							Description: apiErr.Message,
						},
					},
				}
				stWithDetails, err := st.WithDetails(badReq)
				if err != nil {
					return nil, st.Err()
				}
				return nil, stWithDetails.Err()
			}

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
	case errors.Is(err, internalErrors.ErrServiceUnavailable):
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}
