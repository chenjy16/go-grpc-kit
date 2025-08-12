package main

import (
	"log"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/app"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/application.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 显示当前的gRPC配置
	logger.Info("gRPC Server Configuration",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.Int("grpcPort", cfg.Server.GRPCPort),
		zap.Int("maxRecvMsgSize", cfg.GRPC.Server.MaxRecvMsgSize),
		zap.Int("maxSendMsgSize", cfg.GRPC.Server.MaxSendMsgSize),
		zap.Uint32("maxConcurrentStreams", cfg.GRPC.Server.MaxConcurrentStreams),
		zap.Duration("connectionTimeout", time.Duration(cfg.GRPC.Server.ConnectionTimeout)*time.Second),
		zap.Duration("keepaliveTime", time.Duration(cfg.GRPC.Server.KeepaliveTime)*time.Second),
		zap.Duration("keepaliveTimeout", time.Duration(cfg.GRPC.Server.KeepaliveTimeout)*time.Second),
		zap.Duration("keepaliveMinTime", time.Duration(cfg.GRPC.Server.KeepaliveMinTime)*time.Second),
		zap.Bool("enableReflection", cfg.GRPC.Server.EnableReflection),
		zap.Bool("enableCompression", cfg.GRPC.Server.EnableCompression),
		zap.String("compressionLevel", cfg.GRPC.Server.CompressionLevel),
		zap.Bool("enableLogging", cfg.GRPC.Server.EnableLogging),
		zap.Bool("enableMetrics", cfg.GRPC.Server.EnableMetrics),
		zap.Bool("enableRecovery", cfg.GRPC.Server.EnableRecovery),
		zap.Bool("enableTracing", cfg.GRPC.Server.EnableTracing),
	)

	logger.Info("gRPC Client Configuration",
		zap.Duration("timeout", time.Duration(cfg.GRPC.Client.Timeout)*time.Second),
		zap.Int("maxRecvMsgSize", cfg.GRPC.Client.MaxRecvMsgSize),
		zap.Int("maxSendMsgSize", cfg.GRPC.Client.MaxSendMsgSize),
		zap.String("loadBalancing", cfg.GRPC.Client.LoadBalancing),
		zap.Duration("keepaliveTime", time.Duration(cfg.GRPC.Client.KeepaliveTime)*time.Second),
		zap.Duration("keepaliveTimeout", time.Duration(cfg.GRPC.Client.KeepaliveTimeout)*time.Second),
		zap.Bool("permitWithoutStream", cfg.GRPC.Client.PermitWithoutStream),
		zap.Bool("enableCompression", cfg.GRPC.Client.EnableCompression),
		zap.String("compressionLevel", cfg.GRPC.Client.CompressionLevel),
		zap.Bool("enableLogging", cfg.GRPC.Client.EnableLogging),
		zap.Bool("enableMetrics", cfg.GRPC.Client.EnableMetrics),
		zap.Bool("enableTracing", cfg.GRPC.Client.EnableTracing),
	)

	logger.Info("Retry Policy Configuration",
		zap.Int("maxAttempts", cfg.GRPC.Client.RetryPolicy.MaxAttempts),
		zap.String("initialBackoff", cfg.GRPC.Client.RetryPolicy.InitialBackoff),
		zap.String("maxBackoff", cfg.GRPC.Client.RetryPolicy.MaxBackoff),
		zap.Float64("backoffMultiplier", cfg.GRPC.Client.RetryPolicy.BackoffMultiplier),
		zap.Strings("retryableStatusCodes", cfg.GRPC.Client.RetryPolicy.RetryableStatusCodes),
	)

	// 创建应用实例
	application := app.New(
		app.WithConfig(cfg),
		app.WithLogger(logger),
	)

	logger.Info("gRPC server starting with enhanced configuration...")
	logger.Info("Press Ctrl+C to stop the server")

	// 运行应用（这会阻塞直到收到停止信号）
	if err := application.Run(); err != nil {
		logger.Fatal("Failed to run application", zap.Error(err))
	}
}