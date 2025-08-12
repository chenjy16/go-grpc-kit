package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// TestService 测试服务
type TestService struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *TestService) RegisterService(server grpc.ServiceRegistrar) {
	grpc_health_v1.RegisterHealthServer(server, s)
}

func (s *TestService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			Port:     8080,
			GRPCPort: 9090,
		},
	}
	logger := zap.NewNop()

	server := New(cfg, logger)

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.config != cfg {
		t.Error("Expected config to be set")
	}

	if server.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestRegisterService(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	server := New(cfg, logger)

	service := &TestService{}
	server.RegisterService(service)

	if len(server.services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(server.services))
	}
}

func TestRegisterServiceAfterStart(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0, // 使用随机端口
		},
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 4 * 1024 * 1024,
				MaxSendMsgSize: 4 * 1024 * 1024,
			},
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// 尝试在启动后注册服务（应该被忽略）
	service := &TestService{}
	server.RegisterService(service)

	// 服务数量应该保持不变
	if len(server.services) != 0 {
		t.Errorf("Expected 0 services after start, got %d", len(server.services))
	}
}

func TestStartAndStop(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0, // 使用随机端口
		},
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 4 * 1024 * 1024,
				MaxSendMsgSize: 4 * 1024 * 1024,
			},
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	// 测试启动
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsHealthy() {
		t.Error("Expected server to be healthy after start")
	}

	// 测试重复启动
	err = server.Start()
	if err == nil {
		t.Error("Expected error when starting already started server")
	}

	// 测试停止
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsHealthy() {
		t.Error("Expected server to be unhealthy after stop")
	}

	// 测试重复停止
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error when stopping already stopped server: %v", err)
	}
}

func TestGetAddress(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0,
		},
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 4 * 1024 * 1024,
				MaxSendMsgSize: 4 * 1024 * 1024,
			},
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	// 启动前应该返回空字符串
	addr := server.GetAddress()
	if addr != "" {
		t.Errorf("Expected empty address before start, got %s", addr)
	}

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// 启动后应该返回有效地址
	addr = server.GetAddress()
	if addr == "" {
		t.Error("Expected non-empty address after start")
	}

	// 验证地址格式
	_, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		t.Errorf("Invalid address format: %s, error: %v", addr, err)
	}
}

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0,
		},
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 4 * 1024 * 1024,
				MaxSendMsgSize: 4 * 1024 * 1024,
			},
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// 创建客户端连接
	conn, err := grpc.NewClient(
		server.GetAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 测试健康检查
	client := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING status, got %v", resp.Status)
	}
}

func TestSetHealthStatus(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	server := New(cfg, logger)

	// 测试设置健康状态
	server.SetHealthStatus("test-service", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// 这里主要测试方法不会panic，具体的健康状态检查需要集成测试
}

func TestBuildServerOptions(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 1024,
				MaxSendMsgSize: 2048,
			},
		},
		TLS: config.TLSConfig{
			Enabled: false,
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	opts, err := server.buildServerOptions()
	if err != nil {
		t.Fatalf("Failed to build server options: %v", err)
	}

	if len(opts) == 0 {
		t.Error("Expected non-empty server options")
	}
}

func TestBuildServerOptionsWithTLS(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 1024,
				MaxSendMsgSize: 2048,
			},
		},
		TLS: config.TLSConfig{
			Enabled:  true,
			CertFile: "invalid-cert.pem",
			KeyFile:  "invalid-key.pem",
		},
	}
	logger := zap.NewNop()
	server := New(cfg, logger)

	_, err := server.buildServerOptions()
	if err == nil {
		t.Error("Expected error when building TLS options with invalid cert files")
	}
}

// BenchmarkServerStart 性能测试
func BenchmarkServerStart(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "localhost",
			GRPCPort: 0,
		},
		GRPC: config.GRPCConfig{
			Server: config.GRPCServerConfig{
				MaxRecvMsgSize: 4 * 1024 * 1024,
				MaxSendMsgSize: 4 * 1024 * 1024,
			},
		},
	}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := New(cfg, logger)
		err := server.Start()
		if err != nil {
			b.Fatalf("Failed to start server: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		server.Stop(ctx)
		cancel()
	}
}