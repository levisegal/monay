package server

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

type loggingInterceptor struct{}

func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	interceptor := &loggingInterceptor{}
	return interceptor.WrapUnary
}

func (i *loggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		resp, err := next(ctx, req)

		duration := time.Since(start)

		if err != nil {
			slog.Error("RPC failed",
				"procedure", req.Spec().Procedure,
				"duration", duration,
				"error", err,
			)
		} else {
			slog.Info("RPC completed",
				"procedure", req.Spec().Procedure,
				"duration", duration,
			)
		}

		return resp, err
	}
}
