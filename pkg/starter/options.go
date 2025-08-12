package starter

import (
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

// AppOption 配置选项
type AppOption func(*GrpcApplication)

// WithConfig 设置配置
func WithConfig(cfg *config.Config) AppOption {
	return func(app *GrpcApplication) {
		app.config = cfg
	}
}

// WithAppLogger 设置日志器
func WithAppLogger(logger *zap.Logger) AppOption {
	return func(app *GrpcApplication) {
		app.logger = logger
	}
}

// WithGrpcPort 设置 gRPC 端口
func WithGrpcPort(port int) AppOption {
	return func(app *GrpcApplication) {
		if app.config == nil {
			app.config = &config.Config{}
		}
		app.config.Server.GRPCPort = port
	}
}

// WithMetricsPort 设置指标端口
func WithMetricsPort(port int) AppOption {
	return func(app *GrpcApplication) {
		if app.config == nil {
			app.config = &config.Config{}
		}
		app.config.Metrics.Port = port
	}
}

// WithAppDiscovery 启用服务发现
func WithAppDiscovery(enabled bool) AppOption {
	return func(app *GrpcApplication) {
		if app.config == nil {
			app.config = &config.Config{}
		}
		// 通过设置类型来启用/禁用服务发现
		if enabled {
			app.config.Discovery.Type = "etcd"
		} else {
			app.config.Discovery.Type = ""
		}
	}
}

// WithEtcdEndpoints 设置 etcd 端点
func WithEtcdEndpoints(endpoints []string) AppOption {
	return func(app *GrpcApplication) {
		if app.config == nil {
			app.config = &config.Config{}
		}
		app.config.Discovery.Endpoints = endpoints
	}
}

// WithAppMetrics 启用指标
func WithAppMetrics(enabled bool) AppOption {
	return func(app *GrpcApplication) {
		if app.config == nil {
			app.config = &config.Config{}
		}
		app.config.Metrics.Enabled = enabled
	}
}

// DefaultOptions 默认配置选项
func DefaultOptions() []AppOption {
	return []AppOption{
		WithGrpcPort(9090),
		WithMetricsPort(8081),
		WithAppMetrics(true),
		WithAppDiscovery(false), // 默认关闭服务发现
	}
}