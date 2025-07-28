package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsRecorder interface {
	ObserveRequestDuration(method, status string, duration time.Duration)
	ObserveTotalRequests(method, status string)
}

func MetricsInterceptor(m MetricsRecorder) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		method := info.FullMethod
		statusCode := status.Code(err).String()

		m.ObserveRequestDuration(method, statusCode, duration)
		m.ObserveTotalRequests(method, statusCode)

		return resp, err
	}
}
