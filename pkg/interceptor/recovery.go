package interceptor

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor 一元调用恢复拦截器
func RecoveryUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 信息
				logger.Error("gRPC unary call panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				
				// 返回内部错误
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()
		
		return handler(ctx, req)
	}
}

// RecoveryStreamInterceptor 流式调用恢复拦截器
func RecoveryStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 信息
				logger.Error("gRPC stream call panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				
				// 返回内部错误
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()
		
		return handler(srv, stream)
	}
}