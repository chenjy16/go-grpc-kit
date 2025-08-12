package discovery

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewEtcdRegistry(t *testing.T) {
	endpoints := []string{"localhost:2379"}
	namespace := "/test"
	logger := zap.NewNop()

	registry, err := NewEtcdRegistry(endpoints, namespace, logger)
	if err != nil {
		t.Skipf("Skipping test due to etcd connection error: %v", err)
	}
	defer registry.Close()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	if registry.client == nil {
		t.Error("Expected etcd client to be initialized")
	}

	if registry.logger != logger {
		t.Error("Expected logger to be set")
	}

	if registry.namespace != namespace {
		t.Error("Expected namespace to be set")
	}
}

func TestServiceInfo(t *testing.T) {
	service := &ServiceInfo{
		Name:     "test-service",
		Address:  "localhost",
		Port:     9090,
		Metadata: map[string]string{"version": "1.0"},
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Failed to marshal service info: %v", err)
	}

	// 测试 JSON 反序列化
	var parsed ServiceInfo
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal service info: %v", err)
	}

	// 验证数据一致性
	if parsed.Name != service.Name {
		t.Errorf("Name mismatch: expected %s, got %s", service.Name, parsed.Name)
	}
	if parsed.Address != service.Address {
		t.Errorf("Address mismatch: expected %s, got %s", service.Address, parsed.Address)
	}
	if parsed.Port != service.Port {
		t.Errorf("Port mismatch: expected %d, got %d", service.Port, parsed.Port)
	}
	if parsed.Metadata["version"] != service.Metadata["version"] {
		t.Errorf("Metadata mismatch: expected %s, got %s", service.Metadata["version"], parsed.Metadata["version"])
	}
}

func TestBuildServiceKey(t *testing.T) {
	registry := &EtcdRegistry{
		namespace: "/test",
	}

	key := registry.buildServiceKey("test-service", "localhost", 9090)
	expected := "/test/services/test-service/localhost:9090"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestBuildServicePrefix(t *testing.T) {
	registry := &EtcdRegistry{
		namespace: "/test",
	}

	prefix := registry.buildServicePrefix("test-service")
	expected := "/test/services/test-service/"
	if prefix != expected {
		t.Errorf("Expected prefix %s, got %s", expected, prefix)
	}
}

// 注意：以下测试需要运行的 etcd 实例，在 CI/CD 环境中可能需要跳过
func TestEtcdRegistryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoints := []string{"localhost:2379"}
	namespace := "/test"
	logger := zap.NewNop()

	registry, err := NewEtcdRegistry(endpoints, namespace, logger)
	if err != nil {
		t.Skipf("Skipping test due to etcd connection error: %v", err)
	}
	defer registry.Close()

	ctx := context.Background()
	service := &ServiceInfo{
		Name:     "test-service",
		Address:  "localhost",
		Port:     9090,
		Metadata: map[string]string{"version": "1.0"},
	}

	// 测试注册
	err = registry.Register(ctx, service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 等待一下确保注册完成
	time.Sleep(100 * time.Millisecond)

	// 测试发现
	services, err := registry.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) == 0 {
		t.Error("Expected at least one service")
	}

	found := false
	for _, s := range services {
		if s.Address == service.Address && s.Port == service.Port {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find registered service")
	}

	// 测试注销
	err = registry.Deregister(ctx, service)
	if err != nil {
		t.Fatalf("Failed to deregister service: %v", err)
	}

	// 等待一下确保注销完成
	time.Sleep(100 * time.Millisecond)

	// 验证服务已注销
	services, err = registry.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service after deregister: %v", err)
	}

	for _, s := range services {
		if s.Address == service.Address && s.Port == service.Port {
			t.Error("Service should have been deregistered")
		}
	}
}

func TestEtcdRegistryWatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoints := []string{"localhost:2379"}
	namespace := "/test"
	logger := zap.NewNop()

	registry, err := NewEtcdRegistry(endpoints, namespace, logger)
	if err != nil {
		t.Skipf("Skipping test due to etcd connection error: %v", err)
	}
	defer registry.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 开始监听
	watchCh, err := registry.Watch(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to start watch: %v", err)
	}

	service := &ServiceInfo{
		Name:     "test-service",
		Address:  "localhost",
		Port:     9091,
		Metadata: map[string]string{"version": "1.0"},
	}

	// 在另一个 goroutine 中注册服务
	go func() {
		time.Sleep(100 * time.Millisecond)
		registry.Register(context.Background(), service)
	}()

	// 等待监听事件
	timeout := time.After(5 * time.Second)
	eventReceived := false

	for !eventReceived {
		select {
		case services := <-watchCh:
			for _, s := range services {
				if s.Address == service.Address && s.Port == service.Port {
					eventReceived = true
					break
				}
			}
		case <-timeout:
			t.Error("Timeout waiting for watch event")
			return
		}
	}

	// 清理
	registry.Deregister(context.Background(), service)
}

func TestEtcdRegistryError(t *testing.T) {
	// 测试无效的 etcd 配置
	endpoints := []string{"invalid:2379"}
	namespace := "/test"
	logger := zap.NewNop()

	_, err := NewEtcdRegistry(endpoints, namespace, logger)
	if err == nil {
		t.Error("Expected error when creating registry with invalid endpoints")
	}
}

// BenchmarkServiceInfoSerialization 性能测试
func BenchmarkServiceInfoSerialization(b *testing.B) {
	service := &ServiceInfo{
		Name:     "test-service",
		Address:  "localhost",
		Port:     9090,
		Metadata: map[string]string{
			"version":     "1.0",
			"environment": "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(service)
		if err != nil {
			b.Fatalf("Failed to marshal service info: %v", err)
		}

		var parsed ServiceInfo
		err = json.Unmarshal(data, &parsed)
		if err != nil {
			b.Fatalf("Failed to unmarshal service info: %v", err)
		}
	}
}