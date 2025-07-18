package middleware

import (
	"context"

	"google.golang.org/grpc"
)

func GRPCTransaction(txManager TxManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var res interface{}
		err := txManager.ExecuteTx(ctx, func(txCtx context.Context) error {
			var innerErr error
			res, innerErr = handler(txCtx, req)
			return innerErr
		})

		return res, err
	}
}
