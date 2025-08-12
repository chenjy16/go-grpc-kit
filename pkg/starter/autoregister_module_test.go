package starter

import (
	"context"
	"testing"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestAutoRegisterModuleName(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: true,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	expectedName := "AutoRegister"
	
	if module.Name() != expectedName {
		t.Errorf("Expected module name '%s', got '%s'", expectedName, module.Name())
	}
}

func TestAutoRegisterModuleEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enabled",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disabled",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AutoRegister: config.AutoRegisterConfig{
					Enabled: tt.enabled,
				},
			}
			logger := zap.NewNop()

			module := NewAutoRegisterModule(cfg, logger)
			
			if module.Enabled() != tt.expected {
				t.Errorf("Expected enabled to be %v, got %v", tt.expected, module.Enabled())
			}
		})
	}
}

func TestAutoRegisterModuleInitialize(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled:     true,
			ScanDirs:    []string{"./services"},
			Patterns:    []string{"*.go"},
			Excludes:    []string{"*_test.go"},
			ServiceName: "test-service",
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	
	// 创建一个mock GrpcApplication
	app := &GrpcApplication{}
	
	err := module.Initialize(app)
	if err != nil {
		t.Errorf("Expected no error during initialization, got: %v", err)
	}

	// 验证autoRegister是否被初始化
	if module.autoRegister == nil {
		t.Error("Expected autoRegister to be initialized")
	}
}

func TestAutoRegisterModuleInitializeDisabled(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: false,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	app := &GrpcApplication{}
	
	err := module.Initialize(app)
	if err != nil {
		t.Errorf("Expected no error during initialization of disabled module, got: %v", err)
	}
}

func TestAutoRegisterModuleStart(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled:     true,
			ScanDirs:    []string{"./services"},
			Patterns:    []string{"*.go"},
			Excludes:    []string{"*_test.go"},
			ServiceName: "test-service",
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	app := &GrpcApplication{}
	
	// 先初始化
	err := module.Initialize(app)
	if err != nil {
		t.Fatalf("Failed to initialize module: %v", err)
	}

	// 创建context
	ctx := context.Background()
	
	err = module.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error during start, got: %v", err)
	}
}

func TestAutoRegisterModuleStartDisabled(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: false,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	app := &GrpcApplication{}
	
	err := module.Initialize(app)
	if err != nil {
		t.Fatalf("Failed to initialize disabled module: %v", err)
	}

	ctx := context.Background()
	
	err = module.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error during start of disabled module, got: %v", err)
	}
}

func TestAutoRegisterModuleStartWithoutInitialize(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: true,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	ctx := context.Background()
	
	// 不调用Initialize直接调用Start
	err := module.Start(ctx)
	if err != nil {
		// 这里实际上不会出错，因为Start方法没有检查初始化状态
		t.Logf("Start without initialization: %v", err)
	}
}

func TestAutoRegisterModuleStop(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: true,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	ctx := context.Background()
	
	err := module.Stop(ctx)
	if err != nil {
		t.Errorf("Expected no error during stop, got: %v", err)
	}
}

func TestAutoRegisterModuleScanAndRegister(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled:     true,
			ScanDirs:    []string{"./services"},
			Patterns:    []string{"*.go"},
			Excludes:    []string{"*_test.go"},
			ServiceName: "test-service",
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	app := &GrpcApplication{}
	
	// 初始化模块
	err := module.Initialize(app)
	if err != nil {
		t.Fatalf("Failed to initialize module: %v", err)
	}

	server := grpc.NewServer()
	
	err = module.ScanAndRegister(server)
	if err != nil {
		t.Errorf("Expected no error during scan and register, got: %v", err)
	}
}

func TestAutoRegisterModuleScanAndRegisterDisabled(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: false,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	app := &GrpcApplication{}
	
	err := module.Initialize(app)
	if err != nil {
		t.Fatalf("Failed to initialize disabled module: %v", err)
	}

	server := grpc.NewServer()
	
	err = module.ScanAndRegister(server)
	if err != nil {
		t.Errorf("Expected no error during scan and register of disabled module, got: %v", err)
	}
}

func TestAutoRegisterModuleScanAndRegisterWithoutInitialize(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled: true,
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	server := grpc.NewServer()
	
	// 不调用Initialize直接调用ScanAndRegister
	// 当前实现不检查初始化状态，所以不会出错
	err := module.ScanAndRegister(server)
	if err != nil {
		t.Logf("ScanAndRegister without initialization returned: %v", err)
	}
}

func TestAutoRegisterModuleIntegration(t *testing.T) {
	cfg := &config.Config{
		AutoRegister: config.AutoRegisterConfig{
			Enabled:     true,
			ScanDirs:    []string{"./services"},
			Patterns:    []string{"*.go"},
			Excludes:    []string{"*_test.go"},
			ServiceName: "test-service",
		},
	}
	logger := zap.NewNop()

	module := NewAutoRegisterModule(cfg, logger)
	server := grpc.NewServer()
	app := &GrpcApplication{}
	ctx := context.Background()
	
	// 完整的生命周期测试
	// 1. 检查模块是否启用
	if !module.Enabled() {
		t.Error("Expected module to be enabled")
	}

	// 2. 初始化
	err := module.Initialize(app)
	if err != nil {
		t.Fatalf("Failed to initialize module: %v", err)
	}

	// 3. 启动
	err = module.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start module: %v", err)
	}

	// 4. 扫描和注册
	err = module.ScanAndRegister(server)
	if err != nil {
		t.Fatalf("Failed to scan and register: %v", err)
	}

	// 5. 停止
	err = module.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop module: %v", err)
	}
}