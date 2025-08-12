package client

import (
	"context"
	"testing"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MockRegistry 模拟服务发现注册器
type MockRegistry struct {
	services map[string][]*discovery.ServiceInfo
}

func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		services: make(map[string][]*discovery.ServiceInfo),
	}
}

func (m *MockRegistry) Register(ctx context.Context, service *discovery.ServiceInfo) error {
	if m.services[service.Name] == nil {
		m.services[service.Name] = make([]*discovery.ServiceInfo, 0)
	}
	m.services[service.Name] = append(m.services[service.Name], service)
	return nil
}

func (m *MockRegistry) Deregister(ctx context.Context, service *discovery.ServiceInfo) error {
	services := m.services[service.Name]
	for i, s := range services {
		if s.Address == service.Address && s.Port == service.Port {
			m.services[service.Name] = append(services[:i], services[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockRegistry) Discover(ctx context.Context, serviceName string) ([]*discovery.ServiceInfo, error) {
	return m.services[serviceName], nil
}

func (m *MockRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*discovery.ServiceInfo, error) {
	ch := make(chan []*discovery.ServiceInfo, 1)
	ch <- m.services[serviceName]
	close(ch)
	return ch, nil
}

func (m *MockRegistry) Close() error {
	return nil
}

func TestNewClientFactory(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       30,
				MaxRetries:    3,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	factory := NewClientFactory(cfg, registry, logger)

	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	if factory.config != cfg {
		t.Error("Expected config to be set")
	}

	if factory.registry != registry {
		t.Error("Expected registry to be set")
	}

	if factory.logger != logger {
		t.Error("Expected logger to be set")
	}

	if factory.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
}

func TestGetClient(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       30,
				MaxRetries:    3,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	// 添加模拟服务
	service := &discovery.ServiceInfo{
		Name:    "test-service",
		Address: "localhost",
		Port:    9090,
	}
	registry.Register(context.Background(), service)

	factory := NewClientFactory(cfg, registry, logger)

	// 第一次获取客户端（会创建新连接）
	conn1, err := factory.GetClient("test-service")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if conn1 == nil {
		t.Fatal("Expected non-nil connection")
	}

	// 第二次获取客户端（应该返回缓存的连接）
	conn2, err := factory.GetClient("test-service")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if conn1 != conn2 {
		t.Error("Expected same connection instance from cache")
	}

	// 清理
	factory.Close()
}

func TestGetClientNonExistentService(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       1, // 短超时以快速失败
				MaxRetries:    1,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	factory := NewClientFactory(cfg, registry, logger)

	// 尝试获取不存在的服务
	_, err := factory.GetClient("non-existent-service")
	if err == nil {
		t.Error("Expected error when getting non-existent service")
	}

	factory.Close()
}

func TestBuildServiceConfig(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				MaxRetries:    5,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	factory := NewClientFactory(cfg, registry, logger)
	serviceConfig := factory.buildServiceConfig()

	if serviceConfig == "" {
		t.Error("Expected non-empty service config")
	}

	// 验证配置包含预期的值
	if !contains(serviceConfig, "round_robin") {
		t.Error("Expected service config to contain load balancing policy")
	}

	if !contains(serviceConfig, "5") {
		t.Error("Expected service config to contain max attempts")
	}
}

func TestBuildInterceptors(t *testing.T) {
	cfg := &config.Config{}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	factory := NewClientFactory(cfg, registry, logger)
	opts := factory.buildInterceptors()

	if len(opts) == 0 {
		t.Error("Expected non-empty interceptor options")
	}
}

func TestClose(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       30,
				MaxRetries:    3,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	// 添加模拟服务
	service := &discovery.ServiceInfo{
		Name:    "test-service",
		Address: "localhost",
		Port:    9090,
	}
	registry.Register(context.Background(), service)

	factory := NewClientFactory(cfg, registry, logger)

	// 创建一些客户端连接
	_, err := factory.GetClient("test-service")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	// 验证有连接存在
	if len(factory.clients) == 0 {
		t.Error("Expected at least one client connection")
	}

	// 关闭工厂
	err = factory.Close()
	if err != nil {
		t.Errorf("Unexpected error when closing factory: %v", err)
	}

	// 验证连接已清理
	if len(factory.clients) != 0 {
		t.Error("Expected all client connections to be closed")
	}
}

func TestConcurrentGetClient(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       30,
				MaxRetries:    3,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	// 添加模拟服务
	service := &discovery.ServiceInfo{
		Name:    "test-service",
		Address: "localhost",
		Port:    9090,
	}
	registry.Register(context.Background(), service)

	factory := NewClientFactory(cfg, registry, logger)
	defer factory.Close()

	// 并发获取客户端
	const numGoroutines = 10
	results := make(chan *grpc.ClientConn, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			conn, err := factory.GetClient("test-service")
			if err != nil {
				errors <- err
				return
			}
			results <- conn
		}()
	}

	// 收集结果
	var connections []*grpc.ClientConn
	for i := 0; i < numGoroutines; i++ {
		select {
		case conn := <-results:
			connections = append(connections, conn)
		case err := <-errors:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}

	// 验证所有连接都是同一个实例（缓存工作正常）
	if len(connections) != numGoroutines {
		t.Errorf("Expected %d connections, got %d", numGoroutines, len(connections))
	}

	firstConn := connections[0]
	for i, conn := range connections {
		if conn != firstConn {
			t.Errorf("Connection %d is different from first connection", i)
		}
	}
}

// BenchmarkGetClient 性能测试
func BenchmarkGetClient(b *testing.B) {
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Client: config.GRPCClientConfig{
				Timeout:       30,
				MaxRetries:    3,
				LoadBalancing: "round_robin",
			},
		},
	}
	registry := NewMockRegistry()
	logger := zap.NewNop()

	// 添加模拟服务
	service := &discovery.ServiceInfo{
		Name:    "test-service",
		Address: "localhost",
		Port:    9090,
	}
	registry.Register(context.Background(), service)

	factory := NewClientFactory(cfg, registry, logger)
	defer factory.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := factory.GetClient("test-service")
			if err != nil {
				b.Fatalf("Failed to get client: %v", err)
			}
		}
	})
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}