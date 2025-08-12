package starter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/interceptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GrpcServerModule gRPC 服务器模块
type GrpcServerModule struct {
	config     *config.Config
	logger     *zap.Logger
	grpcServer *grpc.Server
	listener   net.Listener
	healthSrv  *health.Server
	started    bool
	mu         sync.RWMutex
}

// NewGrpcServerModule 创建 gRPC 服务器模块
func NewGrpcServerModule(cfg *config.Config, logger *zap.Logger) *GrpcServerModule {
	return &GrpcServerModule{
		config:    cfg,
		logger:    logger,
		healthSrv: health.NewServer(),
	}
}

func (m *GrpcServerModule) Name() string {
	return "grpc-server"
}

func (m *GrpcServerModule) Enabled() bool {
	return true // gRPC 服务器总是启用
}

func (m *GrpcServerModule) Initialize(app *GrpcApplication) error {
	// 创建监听器
	addr := fmt.Sprintf("%s:%d", m.config.Server.Host, m.config.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.listener = listener

	// 构建服务器选项
	opts := m.buildServerOptions()

	// 创建 gRPC 服务器
	m.grpcServer = grpc.NewServer(opts...)

	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(m.grpcServer, m.healthSrv)

	// 注册反射服务
	reflection.Register(m.grpcServer)

	// 注册业务服务
	for _, service := range app.services {
		service.RegisterService(m.grpcServer)
	}

	m.logger.Info("gRPC server initialized",
		zap.String("address", addr),
		zap.Int("services", len(app.services)))

	return nil
}

func (m *GrpcServerModule) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	// 启动服务器
	go func() {
		if err := m.grpcServer.Serve(m.listener); err != nil {
			m.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// 设置健康状态
	m.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	m.started = true
	m.logger.Info("gRPC server started", zap.String("address", m.listener.Addr().String()))

	return nil
}

func (m *GrpcServerModule) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.logger.Info("Stopping gRPC server...")

	// 设置健康状态为不可用
	m.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// 优雅关闭
	done := make(chan struct{})
	go func() {
		m.grpcServer.GracefulStop()
		close(done)
	}()

	// 等待优雅关闭或超时
	select {
	case <-done:
		m.logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		m.logger.Warn("Force stopping gRPC server due to timeout")
		m.grpcServer.Stop()
	}

	m.started = false
	return nil
}

func (m *GrpcServerModule) buildServerOptions() []grpc.ServerOption {
	var opts []grpc.ServerOption

	// 设置消息大小限制
	opts = append(opts,
		grpc.MaxRecvMsgSize(m.config.GRPC.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(m.config.GRPC.Server.MaxSendMsgSize),
	)

	// 添加拦截器链
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptor.LoggingUnaryInterceptor(m.logger),
		interceptor.RecoveryUnaryInterceptor(m.logger),
		interceptor.MetricsUnaryInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptor.LoggingStreamInterceptor(m.logger),
		interceptor.RecoveryStreamInterceptor(m.logger),
		interceptor.MetricsStreamInterceptor(),
	}

	opts = append(opts,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	return opts
}

// GetAddress 获取服务器地址
func (m *GrpcServerModule) GetAddress() string {
	if m.listener == nil {
		return ""
	}
	return m.listener.Addr().String()
}

// MetricsModule 指标模块
type MetricsModule struct {
	config     *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	started    bool
	mu         sync.RWMutex
}

// NewMetricsModule 创建指标模块
func NewMetricsModule(cfg *config.Config, logger *zap.Logger) *MetricsModule {
	return &MetricsModule{
		config: cfg,
		logger: logger,
	}
}

func (m *MetricsModule) Name() string {
	return "metrics"
}

func (m *MetricsModule) Enabled() bool {
	return m.config.Metrics.Enabled
}

func (m *MetricsModule) Initialize(app *GrpcApplication) error {
	mux := http.NewServeMux()

	// 指标端点
	mux.Handle(m.config.Metrics.Path, promhttp.Handler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 就绪检查端点
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.config.Metrics.Port),
		Handler: mux,
	}

	m.logger.Info("Metrics module initialized", zap.Int("port", m.config.Metrics.Port))
	return nil
}

func (m *MetricsModule) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	go func() {
		m.logger.Info("Starting metrics server", zap.String("address", m.httpServer.Addr))
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	m.started = true
	return nil
}

func (m *MetricsModule) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.logger.Info("Stopping metrics server...")

	if err := m.httpServer.Shutdown(ctx); err != nil {
		m.logger.Error("Failed to shutdown metrics server", zap.Error(err))
		return err
	}

	m.started = false
	m.logger.Info("Metrics server stopped")
	return nil
}

// DiscoveryModule 服务发现模块
type DiscoveryModule struct {
	config         *config.Config
	logger         *zap.Logger
	serviceName    string
	serviceManager *discovery.ServiceManager
	registry       discovery.Registry
	started        bool
	mu             sync.RWMutex
}

// NewDiscoveryModule 创建服务发现模块
func NewDiscoveryModule(cfg *config.Config, logger *zap.Logger, serviceName string) *DiscoveryModule {
	return &DiscoveryModule{
		config:      cfg,
		logger:      logger,
		serviceName: serviceName,
	}
}

func (m *DiscoveryModule) Name() string {
	return "discovery"
}

func (m *DiscoveryModule) Enabled() bool {
	return m.config.Discovery.Type != ""
}

func (m *DiscoveryModule) Initialize(app *GrpcApplication) error {
	// 创建服务发现注册器
	registry, err := discovery.NewRegistry(&m.config.Discovery, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}
	m.registry = registry

	// 创建服务管理器
	m.serviceManager = discovery.NewServiceManager(registry, m.logger)

	m.logger.Info("Discovery module initialized",
		zap.String("type", m.config.Discovery.Type),
		zap.Strings("endpoints", m.config.Discovery.Endpoints))

	return nil
}

func (m *DiscoveryModule) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	// 注册服务到服务发现
	serviceInfo := &discovery.ServiceInfo{
		Name:    m.serviceName,
		Address: m.config.Server.Host,
		Port:    m.config.Server.GRPCPort,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	if err := m.serviceManager.RegisterService(ctx, serviceInfo); err != nil {
		m.logger.Warn("Failed to register service to discovery", zap.Error(err))
		// 不返回错误，允许应用继续运行
	} else {
		m.logger.Info("Service registered to discovery",
			zap.String("service", m.serviceName),
			zap.String("address", serviceInfo.Address),
			zap.Int("port", serviceInfo.Port))
	}

	m.started = true
	return nil
}

func (m *DiscoveryModule) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.logger.Info("Stopping discovery module...")

	// 注销所有服务
	if err := m.serviceManager.DeregisterAll(ctx); err != nil {
		m.logger.Error("Failed to deregister services", zap.Error(err))
	}

	// 关闭注册器
	if err := m.registry.Close(); err != nil {
		m.logger.Error("Failed to close registry", zap.Error(err))
	}

	m.started = false
	m.logger.Info("Discovery module stopped")
	return nil
}