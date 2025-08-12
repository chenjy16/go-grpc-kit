package autoregister

import (
	"fmt"
	"os"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// AutoRegister 自动注册器
type AutoRegister struct {
	config    *config.AutoRegisterConfig
	logger    *zap.Logger
	scanner   *Scanner
	generator *Generator
}

// NewAutoRegister 创建新的自动注册器
func NewAutoRegister(cfg *config.AutoRegisterConfig, logger *zap.Logger) *AutoRegister {
	return &AutoRegister{
		config:    cfg,
		logger:    logger,
		scanner:   NewScanner(cfg, logger),
		generator: NewGenerator(logger),
	}
}

// ScanAndRegister 扫描并注册服务
func (ar *AutoRegister) ScanAndRegister(server grpc.ServiceRegistrar) error {
	if !ar.config.Enabled {
		ar.logger.Info("Auto-register is disabled")
		return nil
	}

	ar.logger.Info("Starting auto-registration scan",
		zap.Strings("scan_dirs", ar.config.ScanDirs))

	// 扫描服务
	services, err := ar.scanner.ScanServices()
	if err != nil {
		return fmt.Errorf("failed to scan services: %w", err)
	}

	if len(services) == 0 {
		ar.logger.Info("No services found for auto-registration")
		return nil
	}

	ar.logger.Info("Found services for auto-registration",
		zap.Int("count", len(services)))

	// 这里可以选择两种方式：
	// 1. 直接注册（需要反射或插件机制）
	// 2. 生成注册代码（推荐）

	// 方式2: 生成注册代码
	outputPath := "./auto_register_generated.go"
	if err := ar.generator.GenerateRegistrationCode(services, outputPath); err != nil {
		return fmt.Errorf("failed to generate registration code: %w", err)
	}

	ar.logger.Info("Auto-registration completed",
		zap.String("output", outputPath))

	return nil
}

// ScanAndGenerate 扫描服务并生成注册代码到指定文件
func (ar *AutoRegister) ScanAndGenerate(outputPath string) error {
	if !ar.config.Enabled {
		ar.logger.Info("Auto-register is disabled")
		return nil
	}

	ar.logger.Info("Starting auto-registration scan",
		zap.Strings("scan_dirs", ar.config.ScanDirs))

	// 扫描服务
	services, err := ar.scanner.ScanServices()
	if err != nil {
		return fmt.Errorf("failed to scan services: %w", err)
	}

	ar.logger.Info("Found services for auto-registration",
		zap.Int("count", len(services)))

	// 生成注册代码
	if err := ar.generator.GenerateRegistrationCode(services, outputPath); err != nil {
		return fmt.Errorf("failed to generate registration code: %w", err)
	}

	ar.logger.Info("Auto-registration code generated",
		zap.String("output", outputPath))

	return nil
}

// ValidateConfig 验证配置
func (ar *AutoRegister) ValidateConfig() error {
	if len(ar.config.ScanDirs) == 0 {
		return fmt.Errorf("scan_dirs cannot be empty when auto-register is enabled")
	}

	// 检查扫描目录是否存在
	for _, dir := range ar.config.ScanDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			ar.logger.Warn("Scan directory does not exist", zap.String("dir", dir))
		}
	}

	return nil
}