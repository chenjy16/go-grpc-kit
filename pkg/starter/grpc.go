package starter

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GrpcApplication gRPC 应用启动器
type GrpcApplication struct {
	config   *config.Config
	logger   *zap.Logger
	services []ServiceRegistrar
	modules  []Module
}

// ServiceRegistrar 服务注册接口
type ServiceRegistrar interface {
	RegisterService(s grpc.ServiceRegistrar)
}

// Module 模块接口
type Module interface {
	Name() string
	Enabled() bool
	Initialize(app *GrpcApplication) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// New 创建新的 gRPC 应用
func New(opts ...AppOption) *GrpcApplication {
	// 加载默认配置
	cfg := config.Get()

	// 创建应用
	app := &GrpcApplication{
		config:   cfg,
		logger:   nil,
		services: make([]ServiceRegistrar, 0),
		modules:  make([]Module, 0),
	}

	// 应用选项
	for _, opt := range opts {
		opt(app)
	}

	// 创建日志器（如果没有设置）
	if app.logger == nil {
		app.logger = createDefaultLogger(cfg)
	}

	// 自动注册模块
	app.autoRegisterModules()

	return app
}



// RegisterService 注册服务
func (app *GrpcApplication) RegisterService(service ServiceRegistrar) *GrpcApplication {
	app.services = append(app.services, service)
	return app
}

// RegisterModule 注册模块
func (app *GrpcApplication) RegisterModule(module Module) *GrpcApplication {
	app.modules = append(app.modules, module)
	return app
}

// Run 运行应用
func (app *GrpcApplication) Run() error {
	app.logger.Info("Starting gRPC application",
		zap.String("service", "grpc-service"),
		zap.String("version", "1.0.0"))

	// 初始化模块
	if err := app.initializeModules(); err != nil {
		return fmt.Errorf("failed to initialize modules: %w", err)
	}

	// 启动模块
	ctx := context.Background()
	if err := app.startModules(ctx); err != nil {
		return fmt.Errorf("failed to start modules: %w", err)
	}

	app.logger.Info("Application started successfully")

	// 等待关闭信号
	app.waitForShutdown()

	// 优雅关闭
	return app.shutdown()
}

// autoRegisterModules 自动注册模块
func (app *GrpcApplication) autoRegisterModules() {
	// 注册 gRPC 服务器模块
	app.RegisterModule(NewGrpcServerModule(app.config, app.logger))

	// 根据配置注册其他模块
	if app.config.Metrics.Enabled {
		app.RegisterModule(NewMetricsModule(app.config, app.logger))
	}

	if app.config.Discovery.Type != "" {
		app.RegisterModule(NewDiscoveryModule(app.config, app.logger, "grpc-service"))
	}
}

// initializeModules 初始化模块
func (app *GrpcApplication) initializeModules() error {
	for _, module := range app.modules {
		if !module.Enabled() {
			app.logger.Info("Module disabled, skipping", zap.String("module", module.Name()))
			continue
		}

		app.logger.Info("Initializing module", zap.String("module", module.Name()))
		if err := module.Initialize(app); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", module.Name(), err)
		}
	}
	return nil
}

// startModules 启动模块
func (app *GrpcApplication) startModules(ctx context.Context) error {
	for _, module := range app.modules {
		if !module.Enabled() {
			continue
		}

		app.logger.Info("Starting module", zap.String("module", module.Name()))
		if err := module.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", module.Name(), err)
		}
	}
	return nil
}

// waitForShutdown 等待关闭信号
func (app *GrpcApplication) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	app.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
}

// shutdown 优雅关闭
func (app *GrpcApplication) shutdown() error {
	app.logger.Info("Shutting down application...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 反向停止模块
	for i := len(app.modules) - 1; i >= 0; i-- {
		module := app.modules[i]
		if !module.Enabled() {
			continue
		}

		app.logger.Info("Stopping module", zap.String("module", module.Name()))
		if err := module.Stop(ctx); err != nil {
			app.logger.Error("Failed to stop module",
				zap.String("module", module.Name()),
				zap.Error(err))
		}
	}

	app.logger.Info("Application shutdown completed")
	return nil
}

// createDefaultLogger 创建默认日志器
func createDefaultLogger(cfg *config.Config) *zap.Logger {
	var level zap.AtomicLevel
	switch cfg.Logging.Level {
	case "debug":
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config := zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         cfg.Logging.Format,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := config.Build()
	return logger
}