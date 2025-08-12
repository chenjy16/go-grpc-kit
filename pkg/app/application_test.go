package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MockServiceRegistrar 模拟服务注册器
type MockServiceRegistrar struct {
	registered bool
}

func (m *MockServiceRegistrar) RegisterService(server grpc.ServiceRegistrar) {
	m.registered = true
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

	if app.shutdownTimeout != 30*time.Second {
		t.Errorf("Expected default shutdown timeout 30s, got %v", app.shutdownTimeout)
	}
}

func TestNewWithOptions(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 9091,
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Port:    8081,
			Path:    "/metrics",
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
	}
	logger := zap.NewNop()
	timeout := 60 * time.Second

	app := New(
		WithConfig(cfg),
		WithLogger(logger),
		WithShutdownTimeout(timeout),
	)

	if app.config != cfg {
		t.Error("Expected config to be set")
	}

	if app.logger != logger {
		t.Error("Expected logger to be set")
	}

	if app.shutdownTimeout != timeout {
		t.Errorf("Expected shutdown timeout %v, got %v", timeout, app.shutdownTimeout)
	}
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 9091,
		},
	}

	option := WithConfig(cfg)
	app := &Application{}
	option(app)

	if app.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestWithLogger(t *testing.T) {
	logger := zap.NewNop()
	option := WithLogger(logger)
	app := &Application{}
	option(app)

	if app.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	timeout := 60 * time.Second
	option := WithShutdownTimeout(timeout)
	app := &Application{}
	option(app)

	if app.shutdownTimeout != timeout {
		t.Errorf("Expected shutdown timeout %v, got %v", timeout, app.shutdownTimeout)
	}
}

func TestRegisterService(t *testing.T) {
	app := New()
	service := &MockServiceRegistrar{}

	app.RegisterService(service)

	if len(app.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(app.services))
	}

	if app.services[0] != service {
		t.Error("Expected service to be registered")
	}
}

func TestRegisterMultipleServices(t *testing.T) {
	app := New()
	service1 := &MockServiceRegistrar{}
	service2 := &MockServiceRegistrar{}

	app.RegisterService(service1)
	app.RegisterService(service2)

	if len(app.services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(app.services))
	}

	if app.services[0] != service1 {
		t.Error("Expected first service to be registered")
	}

	if app.services[1] != service2 {
		t.Error("Expected second service to be registered")
	}
}

func TestConcurrentRegisterService(t *testing.T) {
	app := New()
	
	// 并发注册服务
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			service := &MockServiceRegistrar{}
			app.RegisterService(service)
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if len(app.services) != 10 {
		t.Errorf("Expected 10 services, got %d", len(app.services))
	}
}

func TestInitialize(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0, // 使用随机端口
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Port:    0, // 使用随机端口
			Path:    "/metrics",
		},
		Discovery: config.DiscoveryConfig{
			Type: "", // 禁用服务发现
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	app := New(WithConfig(cfg))
	service := &MockServiceRegistrar{}
	app.RegisterService(service)

	err := app.initialize()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	if app.grpcServer == nil {
		t.Error("Expected gRPC server to be created")
	}

	if app.serviceManager != nil {
		t.Error("Expected service manager to be nil when discovery type is empty")
	}

	if app.httpServer == nil {
		t.Error("Expected HTTP server to be created when metrics enabled")
	}
}

func TestInitializeWithoutMetrics(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0,
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
		Discovery: config.DiscoveryConfig{
			Type: "",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	app := New(WithConfig(cfg))

	err := app.initialize()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	if app.httpServer != nil {
		t.Error("Expected HTTP server to be nil when metrics disabled")
	}
}

func TestCreateLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		format   string
		expected bool
	}{
		{"debug level", "debug", "json", true},
		{"info level", "info", "console", true},
		{"warn level", "warn", "json", true},
		{"error level", "error", "console", true},
		{"invalid level", "invalid", "json", true}, // 应该使用默认的 info 级别
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Logging: config.LoggingConfig{
					Level:  tt.level,
					Format: tt.format,
				},
			}

			app := &Application{config: cfg}
			logger := app.createLogger()

			if logger == nil {
				t.Error("Expected logger to be created")
			}
		})
	}
}

func TestCreateHTTPServer(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Port: 8081,
			Path: "/metrics",
		},
	}

	app := &Application{
		config: cfg,
		grpcServer: &server.Server{}, // 模拟 gRPC 服务器
	}

	httpServer := app.createHTTPServer()

	if httpServer == nil {
		t.Fatal("Expected HTTP server to be created")
	}

	if httpServer.Addr != ":8081" {
		t.Errorf("Expected server address ':8081', got '%s'", httpServer.Addr)
	}

	if httpServer.Handler == nil {
		t.Error("Expected HTTP handler to be set")
	}
}

func TestHTTPServerEndpoints(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Port: 8081,
			Path: "/metrics",
		},
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0,
		},
		Discovery: config.DiscoveryConfig{
			Type: "",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	app := New(WithConfig(cfg))
	
	// 初始化应用以创建 gRPC 服务器
	err := app.initialize()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	httpServer := app.createHTTPServer()

	// 测试健康检查端点 - 服务器未启动时应该返回不健康
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := &MockResponseWriter{}
	httpServer.Handler.ServeHTTP(rr, req)

	if rr.statusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr.statusCode)
	}

	if string(rr.body) != "Service Unavailable" {
		t.Errorf("Expected body 'Service Unavailable', got '%s'", string(rr.body))
	}

	// 测试就绪检查端点
	req, _ = http.NewRequest("GET", "/ready", nil)
	rr = &MockResponseWriter{}
	httpServer.Handler.ServeHTTP(rr, req)

	if rr.statusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.statusCode)
	}

	if string(rr.body) != "Ready" {
		t.Errorf("Expected body 'Ready', got '%s'", string(rr.body))
	}
}

func TestHTTPServerHealthCheckAfterStart(t *testing.T) {
	cfg := &config.Config{
		Metrics: config.MetricsConfig{
			Port: 0, // 使用随机端口
			Path: "/metrics",
		},
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0, // 使用随机端口
		},
		Discovery: config.DiscoveryConfig{
			Type: "",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	app := New(WithConfig(cfg))
	
	// 初始化应用
	err := app.initialize()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// 启动 gRPC 服务器
	err = app.grpcServer.Start()
	if err != nil {
		t.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.grpcServer.Stop(ctx)
	}()

	httpServer := app.createHTTPServer()

	// 测试健康检查端点 - 服务器启动后应该返回健康
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := &MockResponseWriter{}
	httpServer.Handler.ServeHTTP(rr, req)

	if rr.statusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.statusCode)
	}

	if string(rr.body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(rr.body))
	}
}

// GrpcServerInterface 定义 gRPC 服务器接口
type GrpcServerInterface interface {
	IsHealthy() bool
	RegisterService(service server.ServiceRegistrar)
	Start() error
	Stop(ctx context.Context) error
	GetAddress() string
}

// MockGrpcServer 模拟 gRPC 服务器
type MockGrpcServer struct {
	healthy bool
}

func (m *MockGrpcServer) IsHealthy() bool {
	return m.healthy
}

func (m *MockGrpcServer) RegisterService(service server.ServiceRegistrar) {}
func (m *MockGrpcServer) Start() error                                    { return nil }
func (m *MockGrpcServer) Stop(ctx context.Context) error                  { return nil }
func (m *MockGrpcServer) GetAddress() string                              { return "localhost:9090" }

// MockResponseWriter 模拟 HTTP ResponseWriter
type MockResponseWriter struct {
	statusCode int
	body       []byte
	headers    http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *MockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
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
			Host:     "localhost",
			GRPCPort: 9091,
		},
	}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := New(
			WithConfig(cfg),
			WithLogger(logger),
			WithShutdownTimeout(60*time.Second),
		)

		if app == nil {
			b.Fatal("Expected app to be created")
		}
	}
}

func BenchmarkRegisterService(b *testing.B) {
	app := New()
	service := &MockServiceRegistrar{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.RegisterService(service)
	}
}

func BenchmarkCreateLogger(b *testing.B) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	app := &Application{config: cfg}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := app.createLogger()
		if logger == nil {
			b.Fatal("Expected logger to be created")
		}
	}
}