package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/client"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// Application 应用程序
type Application struct {
	config          *config.Config
	logger          *zap.Logger
	grpcServer      *server.Server
	httpServer      *http.Server
	serviceManager  *discovery.ServiceManager
	clientFactory   *client.ClientFactory
	services        []server.ServiceRegistrar
	mu              sync.RWMutex
	shutdownTimeout time.Duration
}

// New 创建新的应用程序
func New(opts ...Option) *Application {
	app := &Application{
		services:        make([]server.ServiceRegistrar, 0),
		shutdownTimeout: 30 * time.Second,
	}
	
	// 应用选项
	for _, opt := range opts {
		opt(app)
	}
	
	// 如果没有配置，加载默认配置
	if app.config == nil {
		cfg, err := config.Load("")
		if err != nil {
			// 使用默认配置
			cfg = config.Get()
		}
		app.config = cfg
	}
	
	// 如果没有日志器，创建默认日志器
	if app.logger == nil {
		app.logger = app.createLogger()
	}
	
	return app
}

// Option 应用程序选项
type Option func(*Application)

// WithConfig 设置配置
func WithConfig(cfg *config.Config) Option {
	return func(app *Application) {
		app.config = cfg
	}
}

// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) Option {
	return func(app *Application) {
		app.logger = logger
	}
}

// WithShutdownTimeout 设置关闭超时
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.shutdownTimeout = timeout
	}
}

// RegisterService 注册服务
func (app *Application) RegisterService(service server.ServiceRegistrar) {
	app.mu.Lock()
	defer app.mu.Unlock()
	
	app.services = append(app.services, service)
}

// GetClient 获取gRPC客户端连接
// serviceName 可以是服务发现中的服务名，也可以是DNS地址（如 "example.com:9090"）
func (app *Application) GetClient(serviceName string) (*grpc.ClientConn, error) {
	if app.clientFactory == nil {
		return nil, fmt.Errorf("client factory not initialized")
	}
	return app.clientFactory.GetClient(serviceName)
}

// Run 运行应用程序
func (app *Application) Run() error {
	app.logger.Info("Starting application...")
	
	// 初始化组件
	if err := app.initialize(); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	
	// 启动服务
	if err := app.start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	
	// 等待信号
	app.waitForShutdown()
	
	// 优雅关闭
	return app.shutdown()
}

// initialize 初始化组件
func (app *Application) initialize() error {
	var registry discovery.Registry
	
	// 创建服务发现注册器（如果配置了的话）
	if app.config.Discovery.Type != "" {
		var err error
		registry, err = discovery.NewRegistry(&app.config.Discovery, app.logger)
		if err != nil {
			return fmt.Errorf("failed to create registry: %w", err)
		}
		
		// 创建服务管理器
		app.serviceManager = discovery.NewServiceManager(registry, app.logger)
	}
	
	// 创建客户端工厂（支持DNS解析器）
	app.clientFactory = client.NewClientFactory(app.config, registry, app.logger)
	
	// 创建 gRPC 服务器
	app.grpcServer = server.New(app.config, app.logger)
	
	// 注册业务服务
	for _, service := range app.services {
		app.grpcServer.RegisterService(service)
	}
	
	// 创建 HTTP 服务器（用于指标和健康检查）
	if app.config.Metrics.Enabled {
		app.httpServer = app.createHTTPServer()
	}
	
	return nil
}

// start 启动服务
func (app *Application) start() error {
	// 启动 gRPC 服务器
	if err := app.grpcServer.Start(); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	
	// 注册服务到服务发现（如果启用了服务发现）
	if app.serviceManager != nil {
		serviceInfo := &discovery.ServiceInfo{
			Name:    "grpc-service", // TODO: 从配置获取服务名
			Address: app.config.Server.Host,
			Port:    app.config.Server.GRPCPort,
			Metadata: map[string]string{
				"version": "1.0.0",
			},
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := app.serviceManager.RegisterService(ctx, serviceInfo); err != nil {
			app.logger.Warn("Failed to register service to discovery", zap.Error(err))
		}
	}
	
	// 启动 HTTP 服务器
	if app.httpServer != nil {
		go func() {
			addr := fmt.Sprintf(":%d", app.config.Metrics.Port)
			app.logger.Info("Starting HTTP server", zap.String("address", addr))
			if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				app.logger.Error("HTTP server failed", zap.Error(err))
			}
		}()
	}
	
	app.logger.Info("Application started successfully")
	return nil
}

// waitForShutdown 等待关闭信号
func (app *Application) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	app.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
}

// shutdown 优雅关闭
func (app *Application) shutdown() error {
	app.logger.Info("Shutting down application...")
	
	ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancel()
	
	var wg sync.WaitGroup
	
	// 关闭 HTTP 服务器
	if app.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.httpServer.Shutdown(ctx); err != nil {
				app.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
			}
		}()
	}
	
	// 注销服务
	if app.serviceManager != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.serviceManager.DeregisterAll(ctx); err != nil {
				app.logger.Error("Failed to deregister services", zap.Error(err))
			}
		}()
	}
	
	// 关闭客户端工厂
	if app.clientFactory != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.clientFactory.Close(); err != nil {
				app.logger.Error("Failed to close client factory", zap.Error(err))
			}
		}()
	}
	
	// 关闭 gRPC 服务器
	if app.grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.grpcServer.Stop(ctx); err != nil {
				app.logger.Error("Failed to stop gRPC server", zap.Error(err))
			}
		}()
	}
	
	// 等待所有组件关闭
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		app.logger.Info("Application shutdown completed")
	case <-ctx.Done():
		app.logger.Warn("Application shutdown timeout")
	}
	
	return nil
}

// createLogger 创建日志器
func (app *Application) createLogger() *zap.Logger {
	var level zapcore.Level
	switch app.config.Logging.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         app.config.Logging.Format,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	
	logger, _ := config.Build()
	return logger
}

// createHTTPServer 创建 HTTP 服务器
func (app *Application) createHTTPServer() *http.Server {
	mux := http.NewServeMux()
	
	// 指标端点
	mux.Handle(app.config.Metrics.Path, promhttp.Handler())
	
	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if app.grpcServer.IsHealthy() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable"))
		}
	})
	
	// 就绪检查端点
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})
	
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.Metrics.Port),
		Handler: mux,
	}
}