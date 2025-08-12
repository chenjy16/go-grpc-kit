package starter

import (
	"context"
	"fmt"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/autoregister"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

// AutoRegisterModule 自动注册模块
type AutoRegisterModule struct {
	config       *config.Config
	logger       *zap.Logger
	autoRegister *autoregister.AutoRegister
}

// NewAutoRegisterModule 创建新的自动注册模块
func NewAutoRegisterModule(cfg *config.Config, logger *zap.Logger) *AutoRegisterModule {
	return &AutoRegisterModule{
		config:       cfg,
		logger:       logger,
		autoRegister: autoregister.NewAutoRegister(&cfg.AutoRegister, logger),
	}
}

// Name 返回模块名称
func (m *AutoRegisterModule) Name() string {
	return "AutoRegister"
}

// Enabled 返回模块是否启用
func (m *AutoRegisterModule) Enabled() bool {
	return m.config.AutoRegister.Enabled
}

// Initialize 初始化模块
func (m *AutoRegisterModule) Initialize(app *GrpcApplication) error {
	m.logger.Info("Initializing auto-register module")

	// 只在模块启用时验证配置
	if m.config.AutoRegister.Enabled {
		if err := m.autoRegister.ValidateConfig(); err != nil {
			return fmt.Errorf("auto-register config validation failed: %w", err)
		}
	}

	return nil
}

// Start 启动模块
func (m *AutoRegisterModule) Start(ctx context.Context) error {
	m.logger.Info("Starting auto-register module")

	// 这里我们需要获取 gRPC 服务器实例
	// 由于模块架构的限制，我们可能需要在 GrpcServerModule 启动后再执行自动注册
	// 或者在应用启动时提供一个回调机制

	m.logger.Info("Auto-register module started")
	return nil
}

// Stop 停止模块
func (m *AutoRegisterModule) Stop(ctx context.Context) error {
	m.logger.Info("Stopping auto-register module")
	return nil
}

// ScanAndRegister 扫描并注册服务（供外部调用）
func (m *AutoRegisterModule) ScanAndRegister(server interface{}) error {
	// 这里需要类型断言或接口适配
	// 暂时返回 nil，实际实现需要根据具体的服务器接口来调整
	m.logger.Info("Scanning and registering services...")
	
	// TODO: 实现实际的扫描和注册逻辑
	// return m.autoRegister.ScanAndRegister(server)
	
	return nil
}