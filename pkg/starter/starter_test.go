package starter

import (
	"context"
	"testing"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MockService 模拟 gRPC 服务
type MockService struct{}

func (s *MockService) RegisterService(server grpc.ServiceRegistrar) {
	// 模拟服务注册
}

func TestNew(t *testing.T) {
	app := New()

	if app == nil {
		t.Fatal("Expected app to be created")
	}

	if app.config == nil {
		t.Error("Expected config to be initialized")
	}

	if app.logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if app.services == nil {
		t.Error("Expected services slice to be initialized")
	}

	if app.modules == nil {
		t.Error("Expected modules slice to be initialized")
	}
}

func TestNewWithOptions(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort: 9091,
		},
	}
	logger := zap.NewNop()

	app := New(
		WithConfig(cfg),
		WithAppLogger(logger),
		WithGrpcPort(9092),
		WithMetricsPort(8082),
		WithAppMetrics(true),
		WithAppDiscovery(false),
	)

	if app.config.Server.GRPCPort != 9092 {
		t.Errorf("Expected gRPC port 9092, got %d", app.config.Server.GRPCPort)
	}

	if app.config.Metrics.Port != 8082 {
		t.Errorf("Expected metrics port 8082, got %d", app.config.Metrics.Port)
	}

	if !app.config.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}

	if app.config.Discovery.Type != "" {
		t.Error("Expected discovery to be disabled")
	}

	if app.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort: 9091,
		},
	}

	option := WithConfig(cfg)
	app := &GrpcApplication{}
	option(app)

	if app.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestWithAppLogger(t *testing.T) {
	logger := zap.NewNop()
	option := WithAppLogger(logger)
	app := &GrpcApplication{}
	option(app)

	if app.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestWithGrpcPort(t *testing.T) {
	option := WithGrpcPort(9091)
	app := &GrpcApplication{}
	option(app)

	if app.config.Server.GRPCPort != 9091 {
		t.Errorf("Expected gRPC port 9091, got %d", app.config.Server.GRPCPort)
	}
}

func TestWithMetricsPort(t *testing.T) {
	option := WithMetricsPort(8082)
	app := &GrpcApplication{}
	option(app)

	if app.config.Metrics.Port != 8082 {
		t.Errorf("Expected metrics port 8082, got %d", app.config.Metrics.Port)
	}
}

func TestWithAppMetrics(t *testing.T) {
	option := WithAppMetrics(true)
	app := &GrpcApplication{}
	option(app)

	if !app.config.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}

	option = WithAppMetrics(false)
	app = &GrpcApplication{}
	option(app)

	if app.config.Metrics.Enabled {
		t.Error("Expected metrics to be disabled")
	}
}

func TestWithAppDiscovery(t *testing.T) {
	option := WithAppDiscovery(true)
	app := &GrpcApplication{}
	option(app)

	if app.config.Discovery.Type != "etcd" {
		t.Error("Expected discovery to be enabled with etcd type")
	}

	option = WithAppDiscovery(false)
	app = &GrpcApplication{}
	option(app)

	if app.config.Discovery.Type != "" {
		t.Error("Expected discovery to be disabled")
	}
}

func TestWithEtcdEndpoints(t *testing.T) {
	endpoints := []string{"localhost:2379", "localhost:2380"}
	option := WithEtcdEndpoints(endpoints)
	app := &GrpcApplication{}
	option(app)

	if len(app.config.Discovery.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(app.config.Discovery.Endpoints))
	}

	if app.config.Discovery.Endpoints[0] != "localhost:2379" {
		t.Errorf("Expected first endpoint 'localhost:2379', got '%s'", app.config.Discovery.Endpoints[0])
	}
}

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	if len(options) == 0 {
		t.Error("Expected default options to be provided")
	}

	// 应用默认选项
	app := &GrpcApplication{}
	for _, opt := range options {
		opt(app)
	}

	if app.config.Server.GRPCPort != 9090 {
		t.Errorf("Expected default gRPC port 9090, got %d", app.config.Server.GRPCPort)
	}

	if app.config.Metrics.Port != 8081 {
		t.Errorf("Expected default metrics port 8081, got %d", app.config.Metrics.Port)
	}

	if !app.config.Metrics.Enabled {
		t.Error("Expected metrics to be enabled by default")
	}

	if app.config.Discovery.Type != "" {
		t.Error("Expected discovery to be disabled by default")
	}
}

func TestRegisterService(t *testing.T) {
	app := New()
	service := &MockService{}

	app.RegisterService(service)

	if len(app.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(app.services))
	}
}

func TestRegisterMultipleServices(t *testing.T) {
	app := New()
	service1 := &MockService{}
	service2 := &MockService{}

	app.RegisterService(service1).RegisterService(service2)

	if len(app.services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(app.services))
	}
}

func TestRegisterModule(t *testing.T) {
	app := New()
	module := &MockModule{}

	app.RegisterModule(module)

	// 应该有默认模块 + 我们注册的模块
	found := false
	for _, m := range app.modules {
		if m == module {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected module to be registered")
	}
}

func TestAutoRegisterModules(t *testing.T) {
	app := New(
		WithAppMetrics(true),
		WithAppDiscovery(false),
	)

	// 验证模块是否被注册
	foundGrpc := false
	foundMetrics := false

	for _, module := range app.modules {
		switch module.(type) {
		case *GrpcServerModule:
			foundGrpc = true
		case *MetricsModule:
			foundMetrics = true
		}
	}

	if !foundGrpc {
		t.Error("Expected gRPC module to be registered")
	}

	if !foundMetrics {
		t.Error("Expected metrics module to be registered")
	}
}

func TestAutoRegisterModulesWithDiscovery(t *testing.T) {
	app := New(
		WithAppMetrics(false),
		WithAppDiscovery(true),
	)

	// 验证模块是否被注册
	foundGrpc := false
	foundDiscovery := false

	for _, module := range app.modules {
		switch module.(type) {
		case *GrpcServerModule:
			foundGrpc = true
		case *DiscoveryModule:
			foundDiscovery = true
		}
	}

	if !foundGrpc {
		t.Error("Expected gRPC module to be registered")
	}

	if !foundDiscovery {
		t.Error("Expected discovery module to be registered")
	}
}

func TestInitializeModules(t *testing.T) {
	app := New(
		WithGrpcPort(0),
		WithAppMetrics(false),
		WithAppDiscovery(false),
	)

	err := app.initializeModules()
	if err != nil {
		t.Fatalf("Failed to initialize modules: %v", err)
	}

	// 验证模块是否被初始化
	for _, module := range app.modules {
		if module == nil {
			t.Error("Module should not be nil after initialization")
		}
	}
}

func TestStartAndStopModules(t *testing.T) {
	app := New(
		WithGrpcPort(0),
		WithAppMetrics(false),
		WithAppDiscovery(false),
	)

	err := app.initializeModules()
	if err != nil {
		t.Fatalf("Failed to initialize modules: %v", err)
	}

	ctx := context.Background()

	// 启动模块
	err = app.startModules(ctx)
	if err != nil {
		t.Fatalf("Failed to start modules: %v", err)
	}

	// 停止模块
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 反向停止模块
	for i := len(app.modules) - 1; i >= 0; i-- {
		module := app.modules[i]
		if !module.Enabled() {
			continue
		}

		err := module.Stop(ctx)
		if err != nil {
			t.Errorf("Failed to stop module %s: %v", module.Name(), err)
		}
	}
}

// MockModule 模拟模块
type MockModule struct {
	name    string
	enabled bool
}

func (m *MockModule) Name() string {
	if m.name == "" {
		return "mock"
	}
	return m.name
}

func (m *MockModule) Enabled() bool {
	return m.enabled
}

func (m *MockModule) Initialize(app *GrpcApplication) error {
	return nil
}

func (m *MockModule) Start(ctx context.Context) error {
	return nil
}

func (m *MockModule) Stop(ctx context.Context) error {
	return nil
}

// BenchmarkNew 性能测试
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := New()
		if app == nil {
			b.Fatal("Expected app to be created")
		}
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort: 9091,
		},
	}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := New(
			WithConfig(cfg),
			WithAppLogger(logger),
			WithGrpcPort(9092),
			WithMetricsPort(8082),
			WithAppMetrics(true),
			WithAppDiscovery(false),
		)

		if app == nil {
			b.Fatal("Expected app to be created")
		}
	}
}