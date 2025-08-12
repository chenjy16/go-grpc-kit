package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor 一元调用日志拦截器
func LoggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		// 调用处理器
		resp, err := handler(ctx, req)
		
		// 记录日志
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("code", code.String()),
		}
		
		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("gRPC unary call failed", fields...)
		} else {
			logger.Info("gRPC unary call completed", fields...)
		}
		
		return resp, err
	}
}

// LoggingStreamInterceptor 流式调用日志拦截器
func LoggingStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		
		// 调用处理器
		err := handler(srv, stream)
		
		// 记录日志
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("code", code.String()),
			zap.Bool("client_stream", info.IsClientStream),
			zap.Bool("server_stream", info.IsServerStream),
		}
		
		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("gRPC stream call failed", fields...)
		} else {
			logger.Info("gRPC stream call completed", fields...)
		}
		
		return err
	}
}